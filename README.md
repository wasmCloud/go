# wasmCloud Go

Go SDK, examples, and project templates for building WebAssembly components
that run on [wasmCloud](https://wasmcloud.com) v2.

- [`component`](./component) (`go.wasmcloud.dev/component`) — the component
  SDK: write HTTP handlers with standard `net/http` types and compile them
  to WASI components with big Go via
  [componentize-go](https://github.com/bytecodealliance/componentize-go).
  Two worlds are supported: a **sync WASI P2** world (default, stock Go)
  and an **async WASI P3** world (`wasi:http@0.3.0`, streaming bodies and
  native goroutine concurrency).
- [`examples/components`](./examples/components) — runnable example
  components, each a standalone Go module.
- [`templates`](./templates) — starter templates for `wash new`.
- [`x`](./x) — experimental libraries.

## Quickstart

```shell
wash new https://github.com/wasmCloud/go.git --name my-component \
  --subfolder templates/http-hello-world
cd my-component
wash dev
```

## Examples

| Example | Shows |
|---|---|
| [http-server](./examples/components/http-server) | HTTP routing and handlers via `net/http` |
| [http-client](./examples/components/http-client) | Outbound HTTP with `http.RoundTripper` |
| [http-password-checker](./examples/components/http-password-checker) | A small JSON API |
| [http-keyvalue-crud](./examples/components/http-keyvalue-crud) | HTTP CRUD over host-served `wasi:keyvalue` |
| [http-otel](./examples/components/http-otel) | OpenTelemetry tracing via the host's `wasi:otel` plugin |
| [http-p3-streaming](./examples/components/http-p3-streaming) | Async WASI P3: streaming echo + concurrent outbound fan-out |

## Migrating from wasmCloud v1

wasmCloud v2 removed wadm manifests, capability providers, and the
`wasmcloud:bus` interface; the v1-era provider SDK
(`go.wasmcloud.dev/provider`) and TinyGo component SDK line (`v0.0.x`) are
no longer developed on `main`, though existing module versions
keep resolving. See [MIGRATING.md](./MIGRATING.md).
