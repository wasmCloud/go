#!/usr/bin/env bash
# Run a single HTTP-invoke bench: build the example component that backs
# it, serve it with `wash dev`, then drive load with scripts/bench/httpbench.
#
# Mirrors wasmCloud/wasmCloud's scripts/bench/run-bench.sh in shape:
# bench-name argument, a run log with host provenance, and structured
# output (results.jsonl / summary.md / metadata.json) left under
# $BENCH_OUTPUT_DIR for the workflow to summarize and archive.
#
# The bench-name → example-folder mapping lives in httpbench/main.go
# (-print-example); this script deliberately doesn't duplicate it.

set -euo pipefail

bench="${1:?bench name required (e.g. http_invoke)}"

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
out_dir="${BENCH_OUTPUT_DIR:-${repo_root}/bench-output}"
mkdir -p "$out_dir"

httpbench="${repo_root}/scripts/bench/httpbench"
example="$(cd "$httpbench" && go run . -bench "$bench" -print-example)"
example_dir="${repo_root}/examples/components/${example}"

log="${out_dir}/run-${bench}-${GITHUB_RUN_ID:-local}.log"
{
  echo "=== run-bench: ${bench} (example: ${example}) ==="
  echo "host:   $(hostname)"
  echo "kernel: $(uname -srm)"
  echo "cpus:   $(getconf _NPROCESSORS_ONLN)"
  echo "git:    $(git -C "$repo_root" rev-parse HEAD)"
  echo "ref:    ${WASMCLOUD_BENCH_REF:-${GITHUB_REF_NAME:-?}}"
  echo "go:     $(go version)"
  echo "wash:   $(wash --version)"
  echo "ts:     $(date -u +%FT%TZ)"
  echo
} | tee "$log"

cd "$example_dir"
wash build 2>&1 | tee -a "$log"

# wash dev serves HTTP before the workload is registered; httpbench
# handles readiness by retrying the first scenario until it answers 2xx.
wash dev --non-interactive >"${out_dir}/wash-dev-${bench}.log" 2>&1 &
dev_pid=$!
trap 'kill "$dev_pid" 2>/dev/null || true' EXIT

(cd "$httpbench" && go run . -bench "$bench" -out "$out_dir") 2>&1 | tee -a "$log"
