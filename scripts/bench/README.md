# Benchmarks

End-to-end HTTP invoke benchmarks for the Go component SDK, modeled on
[wasmCloud/wasmCloud's bench workflow](https://github.com/wasmCloud/wasmCloud/actions/workflows/bench.yml).

Each bench builds an example component with `wash build`, serves it with
`wash dev`, and drives a closed-loop (concurrency 1) HTTP load against it,
measuring per-invoke latency. This tracks regressions across the SDK,
`componentize-go`, and wash versions.

| bench | example | scenarios |
|---|---|---|
| `http_invoke` | `http-server` | `GET /`, `POST /post` |
| `http_invoke_p3` | `http-p3-streaming` | `POST /echo` |

## Running locally

```sh
./scripts/bench/run-bench.sh http_invoke
```

Requires `go` and `wash` on PATH. Output lands in `bench-output/` (or
`$BENCH_OUTPUT_DIR`): `results.jsonl`, `summary.md`, `metadata.json`, and
the wash/dev logs.

## CI

`.github/workflows/bench.yml` runs a bench on demand (workflow_dispatch
with a `bench` choice and optional `ref`) and automatically when a
`component/v*` release is published. Results are rendered to the job
summary and uploaded as a 90-day artifact.

CI runs on GitHub-hosted runners: absolute numbers are noisy (shared
vCPUs), so read trends across runs, not single-run deltas.

## Result format

`results.jsonl` rows use the same schema as wasmCloud/wasmCloud's
`bench-tools jsonl` output — `bench`/`group`/`param`, flattened run
metadata (`sha`, `ref`, `timestamp`, …), `metric: "mean_ns"`, `value`,
and the `median_ns`/`std_dev_ns`/`ci_*_ns` siblings — so these rows can
join the same `history.json` aggregate if the S3 push is wired up later.
The bench → example mapping lives in `httpbench/main.go`; add new benches
there and to the workflow's `bench` choice list.
