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
- [`component/nats`](./component/nats) — NATS-native messaging on
  `wasmcloud:nats@0.1.0`: core pub/sub and request-reply, JetStream, and
  JetStream KV, for what the portable
  [`component/messaging`](./component/messaging) package cannot
  express — acknowledged delivery and redelivery, replay from an arbitrary
  stream position, and compare-and-swap on a KV revision. Every function on
  the interface is async, so a component importing it targets WASI P3.
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

| Example | Shows | WASI P3 |
|---|---|---|
| [http-server](./examples/components/http-server) | HTTP routing and handlers via `net/http` | |
| [http-client](./examples/components/http-client) | Outbound HTTP with `http.RoundTripper` | |
| [http-password-checker](./examples/components/http-password-checker) | A small JSON API | |
| [http-keyvalue-crud](./examples/components/http-keyvalue-crud) | HTTP CRUD over host-served `wasi:keyvalue` | |
| [http-otel](./examples/components/http-otel) | OpenTelemetry tracing via the host's `wasi:otel` plugin | |
| [http-p3-streaming](./examples/components/http-p3-streaming) | Async WASI P3: streaming echo + concurrent outbound fan-out | ✅ |
| [nats-request-reply](./examples/components/nats-request-reply) | Core NATS request-reply: a responder plus an HTTP gateway that calls it | ✅ |
| [nats-jetstream-orders](./examples/components/nats-jetstream-orders) | JetStream delivery into KV, with ack/nak and compare-and-swap | ✅ |
| [nats-kv-watch](./examples/components/nats-kv-watch) | Watching a JetStream KV bucket and rebuilding derived state | ✅ |
| [nats-stream-replay](./examples/components/nats-stream-replay) | JetStream replay by sequence or range, and a pull consumer you settle yourself | ✅ |

The unmarked examples build as sync WASI P2 components with stock Go. The
marked ones target the async P3 world: `http-p3-streaming` asks for it
directly, and the `nats-*` examples inherit it because every function on
`wasmcloud:nats@0.1.0` is async. componentize-go selects a patched Go
toolchain for those automatically — see
[http-p3-streaming](./examples/components/http-p3-streaming#toolchain-note).

## Migrating from wasmCloud v1

wasmCloud v2 removed wadm manifests, capability providers, and the
`wasmcloud:bus` interface; the v1-era provider SDK
(`go.wasmcloud.dev/provider`) and TinyGo component SDK line (`v0.0.x`) are
no longer developed on `main`, though existing module versions
keep resolving. See [MIGRATING.md](./MIGRATING.md).
