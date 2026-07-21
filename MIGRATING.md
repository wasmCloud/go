# Migrating from wasmCloud v1 to v2

This repo was rebuilt for wasmCloud v2 (2.5.2+). This guide maps the v1
concepts to their v2 replacements.

## Concept map

| wasmCloud v1 | wasmCloud v2 |
|---|---|
| wadm / OAM `Application` manifest (`wadm.yaml`) | Kubernetes `WorkloadDeployment` (`runtime.wasmcloud.dev/v1alpha1`) + `Service`, reconciled by the runtime operator |
| Capability providers (NATS + wRPC, `.par.gz`) | Host plugins (in-process) and service workloads; the WASI interface set (`wasi:http`, `wasi:keyvalue`, `wasi:blobstore`, `wasi:config`) is served by the host, gated by the deny-by-default `hostInterfaces` allowlist |
| `wasmcloud:bus/lattice` (link names, call targets) | Removed. Networking is explicit; in-process calls by default |
| `wasmcloud:secrets@0.1.0-draft` | Not yet available in 2.5.2; a `wasmcloud:secrets@1.0.0` host plugin lands in a future release. Use `wasi:config/store` for configuration |
| `wash up`, `wash app deploy`, `wash call`, `wash get inventory` | `wash dev` (local loop), `wash build`, `wash oci push`, `kubectl apply` |
| `wasmcloud.toml` | `.wash/config.yaml` |

## Go SDK changes (`go.wasmcloud.dev/component`)

**v0.0.x (TinyGo, WASI P2 only)** remains available at its tags for
existing projects; the TinyGo line ends at `v0.0.8`. **v0.2.0+** is a
rebuild on big Go (1.25+) and
[componentize-go](https://github.com/bytecodealliance/componentize-go):

- Build with `go tool componentize-go build` (or `wash build` with the
  build command in `.wash/config.yaml`) instead of TinyGo + wit-bindgen-go.
- No per-project WIT or generated bindings: the SDK ships its worlds and a
  `componentize-go.toml`; apps need only `go.mod` + code.
- `net/wasihttp` keeps the same API (`Handle`, `HandleFunc`, `Transport`,
  `DefaultClient`) on the sync WASI P2 world — most handler code ports
  unchanged.
- `net/wasihttp3` (new) targets the async WASI P3 world
  (`wasi:http@0.3.0`): streaming bodies, concurrent outbound requests with
  plain goroutines. Opt in with
  `go tool componentize-go -w wasmcloud:component-go/wasip3@0.2.0 build`.
  Requires componentize-go's auto-installed patched Go until
  [golang/go#76775](https://github.com/golang/go/pull/76775) merges.
- `wasmcloud.SetLinkName` / `CallTargetInterface` (bus) and
  `wasmcloud/secret.go` are removed (see concept map).
- `wasilog` and `wasmcloud.GetConfigOrDefault` work as before.

## Removed from this repo

- `provider/` + provider examples/templates — no v2 equivalent; the last
  v1 releases stay importable (`go.wasmcloud.dev/provider@v0.x`).
- `x/wasmbus` + example — v1 lattice RPC.
- `examples/component/invoke` — demonstrated `wasmcloud:bus`.
- `examples/component/sqldb-postgres-query` — to return rebuilt on the v2
  async PostgreSQL backend.
