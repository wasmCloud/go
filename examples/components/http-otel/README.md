# http-otel

An HTTP echo component instrumented with the OpenTelemetry Go SDK,
exporting traces through the wasmCloud host's **`wasi_otel` host plugin**
(`wasi:otel@0.2.0-rc.2`) instead of shipping OTLP over HTTP itself.

How it works:

- The component registers a `sdktrace.SpanProcessor` that forwards span
  lifecycle events to the host: `wasi:otel/tracing.on-start` when a span
  begins and `on-end` (with the full span data) when it completes.
- `wasi:otel/tracing.outer-span-context` returns the host's current span
  context, which the handler uses as the remote parent — so component
  spans nest inside the host's own request trace.
- The host plugin batches and exports the spans via its own OTLP pipeline;
  the component needs no collector endpoint, no batching, and no OTLP
  wire-format code.

Like the [http-keyvalue-crud](../http-keyvalue-crud) example, the extra
interfaces come from an app-local world ([wit/world.wit](./wit/world.wit))
that [componentize-go.toml](./componentize-go.toml) merges with the SDK's
`wasip2` world at build time. The `wasi_otel_*` bindings are generated
with:

```shell
go tool componentize-go --ignore-toml-files \
  -w wasmcloud:examples/http-otel@0.1.0 -d wit bindings -o .
```

> [!NOTE]
> The `wasi_otel` host plugin is not yet in a released wash — running this
> example requires a `wash` built from
> [wasmCloud main](https://github.com/wasmCloud/wasmCloud) (the plugin
> lives in `crates/wash-runtime/src/plugin/wasi_otel`). Released wash
> versions fail to link the component's `wasi:otel` imports.

The plugin is opt-in for `wash dev`: this example's
[.wash/config.yaml](./.wash/config.yaml) enables it with:

```yaml
dev:
  wasi_otel: true
```

The plugin exports spans over OTLP gRPC (default `http://localhost:4317`).

## Develop

```shell
wash dev
curl -X POST localhost:8000/ -d 'echo this back with a trace'
```

Spans emitted by the component are exported by the host's OTel pipeline —
configure the host's OTLP endpoint to see them in your collector of
choice.

## Build & deploy

```shell
wash build
wash oci push ghcr.io/<your-org>/http-otel:0.1.0 build/http_otel.wasm
kubectl apply -f deployment.yaml
```

[deployment.yaml](./deployment.yaml) declares `wasi:otel` under
`hostInterfaces` alongside `wasi:http`.
