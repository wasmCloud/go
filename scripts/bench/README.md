# Benchmarks

End-to-end HTTP invoke benchmarks for the Go component SDK, modeled on
[wasmCloud/wasmCloud's bench workflow](https://github.com/wasmCloud/wasmCloud/actions/workflows/bench.yml)
and publishing to the same
[arewefastyet](https://wasmcloud.github.io/arewefastyet/) site.

Each bench builds an example component with `wash build`, serves it with
`wash dev`, and drives a closed-loop (concurrency 1) HTTP load against it,
measuring per-invoke latency. This tracks regressions across the SDK,
`componentize-go`, and wash versions.

Bench names carry a `_go` suffix because both repos feed one shared
`history.json`: the site's timelines key on the `bench` field, and the
Rust repo already publishes `http_invoke`.

| bench | example | scenarios |
|---|---|---|
| `http_invoke_go` | `http-server` | `GET /`, `POST /post` |
| `http_invoke_p3_go` | `http-p3-streaming` | `POST /echo` |

## Running locally

```sh
./scripts/bench/run-bench.sh http_invoke_go
```

Requires `go` and `wash` on PATH. Output lands in `bench-output/` (or
`$BENCH_OUTPUT_DIR`): `results.jsonl`, `summary.md`, `metadata.json`, and
the wash/dev logs.

## CI

`.github/workflows/bench.yml` runs a bench on demand (workflow_dispatch
with a `bench` choice and optional `ref`, including `#N` for a PR) and
automatically when a `component/v*` release is published. It runs on the
dedicated Hetzner bench host (`[self-hosted, bench, hetzner]` — the same
box as the Rust repo's benches, so numbers are comparable), after
`.github/scripts/bench-preflight.mjs` verifies the host baseline.

Results are rendered to the job summary, uploaded as a 90-day artifact,
and — except for PR refs — pushed by
`.github/scripts/bench-push-results.mjs` to the shared S3 bucket:
per-run artifacts under `runs/<date>/<sha>/<run-id>/<bench>/`, plus a
read-merge-write of the public `history.json` aggregate (deduped on
`(sha, bench, group, param, run_attempt, metric)`) and a CloudFront
invalidation. The site renders whatever bench names appear in
`history.json` — no site change is needed for a new bench.

Repo prerequisites for the push (mirroring the Rust repo):
- access to the org's bench runner group
- `vars.WASMCLOUD_BENCH_HOSTNAME`
- secrets `WASMCLOUD_BENCH_AWS_ROLE_ARN`, `WASMCLOUD_BENCH_S3_REGION`,
  `WASMCLOUD_BENCH_S3_BUCKET`, `WASMCLOUD_BENCH_CF_DISTRIBUTION_ID`
- the AWS role's OIDC trust policy must include `repo:wasmCloud/go:*`

## Result format

`results.jsonl` rows use the same schema as wasmCloud/wasmCloud's
`bench-tools jsonl` output — `bench`/`group`/`param`, flattened run
metadata (`sha`, `ref`, `timestamp`, …), `metric: "mean_ns"`, `value`,
and the `median_ns`/`std_dev_ns`/`ci_*_ns` siblings — so these rows join
the same `history.json` aggregate as the Rust benches.
The bench → example mapping lives in `httpbench/main.go`; add new benches
there and to the workflow's `bench` choice list.
