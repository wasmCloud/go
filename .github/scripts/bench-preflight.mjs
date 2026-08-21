#!/usr/bin/env node
// Pre-flight checks for the wasmCloud bench host, ported from
// wasmCloud/wasmCloud's .github/scripts/bench-preflight.mjs. This repo's
// benches run on the SAME dedicated Hetzner host, so the host-baseline
// invariants are identical; only the toolchain check differs (go + wash
// instead of cargo). A drifted host produces meaningless numbers; better
// to fail fast than to publish them.
//
// Invoked from .github/workflows/bench.yml AFTER the setup-go/setup-wash
// steps (the toolchain check relies on their PATH additions). Reads:
//
//   WASMCLOUD_BENCH_HOSTNAME       expected hostname (workflow: vars.WASMCLOUD_BENCH_HOSTNAME)
//   WASMCLOUD_BENCH_ISOLATED_CPU   override for the isolated-CPU index (default: "5")
//   BENCH_OUTPUT_DIR               output dir (default: ./bench-output)

import { execFileSync, spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, readFileSync, statfsSync } from 'node:fs';
import { hostname, loadavg, platform } from 'node:os';
import { join } from 'node:path';

const EXPECTED_NPROC = 6;
const EXPECTED_GOVERNOR = 'performance';
const MIN_FREE_BYTES = 5 * 1024 ** 3; // 5 GiB
const MAX_LOAD1 = 1.0;
const LOAD_SETTLE_SECS = 240;
const LOAD_POLL_SECS = 5;

const isolatedCpu = process.env.WASMCLOUD_BENCH_ISOLATED_CPU ?? '5';
const outDir = process.env.BENCH_OUTPUT_DIR ?? join(process.cwd(), 'bench-output');

function fail(msg) {
  console.error(`::error::pre-flight: ${msg}`);
  process.exit(1);
}

function ok(msg) {
  console.log(`pre-flight: ${msg}`);
}

function readTrim(path) {
  return readFileSync(path, 'utf8').trim();
}

function runStdout(cmd, args = []) {
  return execFileSync(cmd, args, { encoding: 'utf8' }).trim();
}

function sleepSync(ms) {
  Atomics.wait(new Int32Array(new SharedArrayBuffer(4)), 0, 0, ms);
}

function readLoad1() {
  return loadavg()[0];
}

// Parse a Linux CPU mask string (`0-5`, `0,2-5`, empty) into a count.
function countCpuMask(mask) {
  if (!mask) return 0;
  let total = 0;
  for (const part of mask.split(',')) {
    const [lo, hi] = part.split('-').map(Number);
    total += hi === undefined ? 1 : hi - lo + 1;
  }
  return total;
}

// 0. Linux only — every check below reads /proc or /sys, and os.loadavg()
//    returns [0, 0, 0] on Windows.
if (platform() !== 'linux') {
  fail(`unsupported platform '${platform()}'; the bench host must be Linux`);
}

// 1. WASMCLOUD_BENCH_HOSTNAME must be exported, and we must be on that host.
const expectedHostname = process.env.WASMCLOUD_BENCH_HOSTNAME;
if (!expectedHostname) {
  fail('WASMCLOUD_BENCH_HOSTNAME not set (workflow: vars.WASMCLOUD_BENCH_HOSTNAME)');
}
const actualHostname = hostname();
if (actualHostname !== expectedHostname) {
  fail(`wrong host: ${actualHostname} (expected ${expectedHostname})`);
}
ok(`host: ${expectedHostname}`);

// 2. Online CPUs == 6 (nosmt collapsed 12 SMT → 6 physical). Sysfs
//    `online` is the only reading that tracks what's actually online;
//    see the Rust repo's preflight for why nproc variants are wrong.
const onlineMask = readTrim('/sys/devices/system/cpu/online');
const onlineCount = countCpuMask(onlineMask);
if (onlineCount !== EXPECTED_NPROC) {
  fail(
    `expected ${EXPECTED_NPROC} online CPUs (nosmt); got ${onlineCount} (online mask: '${onlineMask}')`,
  );
}
ok(`online CPUs: ${onlineCount}  (mask '${onlineMask}'; nosmt active)`);

// 3. cpufreq governor == "performance" on every CPU.
const governors = runStdout('sh', [
  '-c',
  'cat /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor',
]).split('\n');
for (const g of governors) {
  if (g !== EXPECTED_GOVERNOR) {
    fail(`scaling_governor = ${g} (expected ${EXPECTED_GOVERNOR})`);
  }
}
ok(`governor: ${EXPECTED_GOVERNOR} on every CPU`);

// 4. Isolated CPU matches the host baseline. This bench doesn't pin to
//    it, but a missing isolcpus= means the host was re-staged without the
//    baseline and every other invariant is suspect too.
if (!existsSync('/sys/devices/system/cpu/isolated')) {
  fail('/sys/devices/system/cpu/isolated not readable (kernel too old?)');
}
const isolated = readTrim('/sys/devices/system/cpu/isolated');
if (isolated !== isolatedCpu) {
  fail(`isolated CPU mismatch: kernel reports '${isolated}', expected '${isolatedCpu}'`);
}
ok(`isolcpus: CPU ${isolatedCpu} reserved`);

// 5. mdraid must not be resyncing — resync I/O would skew bench numbers.
let mdstat = '';
try {
  mdstat = readFileSync('/proc/mdstat', 'utf8');
} catch {
  // No mdraid configured; nothing to check.
}
if (mdstat.includes('resync')) {
  fail('mdraid resync in progress; refusing to bench');
}
ok('mdraid: clean (no resync)');

// 6. 1-min loadavg under MAX_LOAD1, waiting out the previous bench's
//    decaying load tail (benches run back-to-back on this one host,
//    including the Rust repo's).
let load1 = readLoad1();
if (load1 > MAX_LOAD1) {
  ok(`loadavg(1m)=${load1} > ${MAX_LOAD1}; waiting up to ${LOAD_SETTLE_SECS}s to settle`);
  const deadline = Date.now() + LOAD_SETTLE_SECS * 1000;
  while (load1 > MAX_LOAD1 && Date.now() < deadline) {
    sleepSync(LOAD_POLL_SECS * 1000);
    load1 = readLoad1();
  }
  if (load1 > MAX_LOAD1) {
    fail(`1-min loadavg=${load1} after ${LOAD_SETTLE_SECS}s settle wait (something else is busy)`);
  }
}
ok(`loadavg(1m): ${load1}`);

// 7. Output dir exists, is writable, and its mount has 5 GiB free.
mkdirSync(outDir, { recursive: true });
const access = spawnSync('test', ['-w', outDir]);
if (access.status !== 0) {
  fail(`${outDir} not writable`);
}
const fs = statfsSync(outDir);
const freeBytes = Number(fs.bavail) * Number(fs.bsize);
if (freeBytes < MIN_FREE_BYTES) {
  fail(`less than 5 GiB free at ${outDir}`);
}
ok(`output dir: ${outDir} (${Math.floor(freeBytes / 1024 ** 3)} GiB free)`);

// 8. Toolchain: go and wash must be on PATH (installed by the setup-go /
//    setup-wash steps that precede this one).
for (const tool of ['go', 'wash', 'zstd']) {
  const which = spawnSync('sh', ['-c', `command -v ${tool}`], { encoding: 'utf8' });
  if (which.status !== 0) {
    fail(`${tool} not on PATH`);
  }
  ok(`${tool}: ${which.stdout.trim()}`);
}
