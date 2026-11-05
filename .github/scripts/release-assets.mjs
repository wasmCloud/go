#!/usr/bin/env node
// Builds the assets attached to a component/vX.Y.Z release: the
// go.wasmcloud.dev/component module zip and a SHA256SUMS for it.
//
// The zip is built from the tag with GOPROXY=direct, the same way the proxy
// builds it, and the go command checks it against the checksum database. It
// must match proxy.golang.org's byte for byte, so the attested file is the
// one `go get` downloads.
//
// Invoked from .github/workflows/release-tag.yaml:
//
//   node release-assets.mjs <vX.Y.Z> <out-dir>

import { execFileSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import { mkdirSync, mkdtempSync, readFileSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';

const MODULE = 'go.wasmcloud.dev/component';

function fail(msg) {
  console.error(`::error::${msg}`);
  process.exit(1);
}

const [version, outDir] = process.argv.slice(2);
if (!/^v\d+\.\d+\.\d+$/.test(version ?? '') || !outDir) {
  fail('usage: release-assets.mjs <vX.Y.Z> <out-dir>');
}

const work = mkdtempSync(join(process.env.RUNNER_TEMP ?? tmpdir(), 'release-assets-'));
const env = { ...process.env, GOMODCACHE: join(work, 'modcache'), GOPROXY: 'direct', GOFLAGS: '' };

function go(args) {
  try {
    return execFileSync('go', args, { cwd: work, env, encoding: 'utf8', stdio: ['ignore', 'pipe', 'inherit'] });
  } catch (err) {
    // `go mod download -json` reports a failed download as JSON on stdout.
    let detail = err.message;
    try {
      detail = JSON.parse(err.stdout).Error ?? detail;
    } catch {}
    fail(`go ${args.join(' ')} failed: ${detail}`);
  }
}

go(['mod', 'init', 'release.invalid/attest']);
const download = JSON.parse(go(['mod', 'download', '-json', `${MODULE}@${version}`]));
const zip = readFileSync(download.Zip);

const url = `https://proxy.golang.org/${MODULE}/@v/${version}.zip`;
const res = await fetch(url);
if (!res.ok) {
  fail(`${url}: HTTP ${res.status}`);
}
if (!zip.equals(Buffer.from(await res.arrayBuffer()))) {
  fail(`the module zip built from the tag differs from ${url}`);
}

const name = `component-${version}.zip`;
const sha256 = createHash('sha256').update(zip).digest('hex');
mkdirSync(outDir, { recursive: true });
writeFileSync(join(outDir, name), zip);
// The format `sha256sum --check` reads.
writeFileSync(join(outDir, 'SHA256SUMS'), `${sha256}  ${name}\n`);
console.log(`${name} ${download.Sum} sha256:${sha256}, matches ${url}`);
