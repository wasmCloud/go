#!/usr/bin/env node
// Upload one bench run's artifacts to S3, then update the public
// history.json aggregate and invalidate CloudFront.
//
// Ported from wasmCloud/wasmCloud's .github/scripts/bench-push-results.mjs,
// adapted to this repo's bench layout: httpbench has already written
// results.jsonl / summary.md / metadata.json / logs under
// $BENCH_OUTPUT_DIR (no cargo, no bench-tools invocation needed here).
// Both repos push into the SAME bucket and history.json; the Go rows are
// distinguished by their `bench` names (http_invoke_go, ...), and
// https://wasmcloud.github.io/arewefastyet/ picks them up from the
// merged aggregate automatically.
//
// Per-run layout (private; only the bench role can read):
//   s3://${WASMCLOUD_BENCH_S3_BUCKET}/runs/<date>/<short-sha>/<run-id>/<bench>/
//     ├─ results.jsonl       one JSON row per (group, param, metric)
//     ├─ summary.md          human-readable run summary
//     ├─ metadata.json       run-level facts (written by httpbench)
//     ├─ run.log             run-bench.sh stdout/stderr
//     └─ wash-dev.log.zst    wash dev output (zstd; it can be large)
//
// Aggregate (publicly readable through CloudFront):
//   s3://${WASMCLOUD_BENCH_S3_BUCKET}/history.json — same read-merge-write
//   and dedup key as the Rust repo: (sha, bench, group, param,
//   run_attempt, metric). Safe without locking only because each repo's
//   bench workflow queues on a single self-hosted runner; the two repos
//   share that runner, and GitHub serializes jobs on it.
//
// Reads (required):
//   WASMCLOUD_BENCH_NAME                 bench whose output we're uploading
//   WASMCLOUD_BENCH_S3_BUCKET            target bucket
//   WASMCLOUD_BENCH_CF_DISTRIBUTION_ID   CloudFront distribution to invalidate
//
// Reads (optional):
//   BENCH_OUTPUT_DIR        default ./bench-output
//   GITHUB_RUN_ID           default "local"

import { execFileSync, spawnSync } from 'node:child_process';
import { existsSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { basename, join } from 'node:path';

const bench = required('WASMCLOUD_BENCH_NAME');
const bucket = required('WASMCLOUD_BENCH_S3_BUCKET');
const distId = required('WASMCLOUD_BENCH_CF_DISTRIBUTION_ID');

const outDir = process.env.BENCH_OUTPUT_DIR ?? join(process.cwd(), 'bench-output');
const runId = process.env.GITHUB_RUN_ID ?? 'local';
const sha = run('git', ['rev-parse', 'HEAD']);
const shortSha = run('git', ['rev-parse', '--short=12', 'HEAD']);
const date = new Date().toISOString().slice(0, 10); // YYYY-MM-DD
const prefix = `runs/${date}/${shortSha}/${runId}/${bench}`;

const resultsPath = join(outDir, 'results.jsonl');
if (!existsSync(resultsPath)) {
  console.error(`::error::no results.jsonl at ${resultsPath}; refusing to push an empty run`);
  process.exit(1);
}
const jsonl = readFileSync(resultsPath, 'utf8');

const work = mkdtempSync(join(tmpdir(), 'push-bench-'));
process.on('exit', () => {
  try {
    rmSync(work, { recursive: true, force: true });
  } catch {
    // best-effort
  }
});

// 1. Compress the wash-dev log (16MB+ in practice) with zstd -19; the
//    other artifacts are small and upload as-is.
const washLog = join(outDir, `wash-dev-${bench}.log`);
if (existsSync(washLog)) {
  run('zstd', ['-19', '-T0', '-q', '-f', washLog, '-o', join(work, 'wash-dev.log.zst')]);
}
const runLog = join(outDir, `run-${bench}-${runId}.log`);
if (existsSync(runLog)) {
  run('cp', [runLog, join(work, 'run.log')]);
}
for (const f of ['results.jsonl', 'summary.md', 'metadata.json']) {
  const src = join(outDir, f);
  if (existsSync(src)) run('cp', [src, join(work, f)]);
}

// 2. Upload per-run artifacts.
console.log(`uploading per-run artifacts to s3://${bucket}/${prefix}/`);
for (const file of ['results.jsonl', 'summary.md', 'metadata.json', 'run.log', 'wash-dev.log.zst']) {
  const path = join(work, file);
  if (!existsSync(path)) continue;
  run('aws', ['s3', 'cp', '--no-progress', path, `s3://${bucket}/${prefix}/${basename(path)}`]);
}

// 3. Read-modify-write the public history.json aggregate (same dedup key
//    as the Rust repo's push script so the two repos merge cleanly).
console.log(`updating s3://${bucket}/history.json`);
let existing = [];
const head = spawnSync('aws', ['s3api', 'head-object', '--bucket', bucket, '--key', 'history.json'], {
  stdio: 'ignore',
});
if (head.status === 0) {
  const histPath = join(work, 'history-existing.json');
  run('aws', ['s3', 'cp', '--no-progress', `s3://${bucket}/history.json`, histPath]);
  existing = JSON.parse(readFileSync(histPath, 'utf8'));
}

const newRows = jsonl
  .split('\n')
  .filter((line) => line.length > 0)
  .map((line) => JSON.parse(line));

const dedupKey = (r) =>
  JSON.stringify([r.sha, r.bench, r.group, r.param, r.run_attempt, r.metric ?? null]);
const merged = new Map();
for (const row of existing) merged.set(dedupKey(row), row);
for (const row of newRows) merged.set(dedupKey(row), row); // new rows win on collision
const final = [...merged.values()].sort((a, b) => a.timestamp.localeCompare(b.timestamp));

const histOut = join(work, 'history.json');
writeFileSync(histOut, JSON.stringify(final));

run('aws', [
  's3',
  'cp',
  '--no-progress',
  '--content-type',
  'application/json',
  '--cache-control',
  'public, max-age=60',
  histOut,
  `s3://${bucket}/history.json`,
]);

// 4. Invalidate CloudFront so the next request hits a fresh edge cache.
console.log('invalidating CloudFront /history.json');
const invalidationId = run('aws', [
  'cloudfront',
  'create-invalidation',
  '--distribution-id',
  distId,
  '--paths',
  '/history.json',
  '--query',
  'Invalidation.Id',
  '--output',
  'text',
]);
console.log(`invalidation: ${invalidationId}`);

console.log(
  `::notice title=bench results::s3://${bucket}/${prefix}/  (history now ${final.length} rows)`,
);

// ─── helpers ────────────────────────────────────────────────────────────

function required(name) {
  const v = process.env[name];
  if (!v) {
    console.error(`${name} not set`);
    process.exit(1);
  }
  return v;
}

function run(cmd, args = []) {
  return execFileSync(cmd, args, { encoding: 'utf8' }).trim();
}
