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

The `wasi_otel` host plugin ships in released `wash` (verified with 2.9.0),
but it is opt-in everywhere: a host that has not enabled it fails to link the
component's `wasi:otel` imports.

For `wash dev`: this example's
[.wash/config.yaml](./.wash/config.yaml) enables it with:

```yaml
dev:
  wasi_otel: true
```

The plugin exports spans over OTLP gRPC (default `http://localhost:4317`).

## Prerequisites

- Go 1.27+
- [`wash`](https://wasmcloud.com/docs/installation) 2.x

On a fresh clone, download the Go modules first:

```shell
go mod download
```

componentize-go reads the SDK's WIT straight out of the module cache and does
not fetch it itself, so until the modules are downloaded `wash dev` and
`wash build` fail with `failed to read path for WIT [wit]`.

## Develop

```shell
wash dev
curl -X POST localhost:8000/ -d 'echo this back with a trace'
```

Spans emitted by the component are exported by the host's OTel pipeline —
configure the host's OTLP endpoint to see them in your collector of
choice.

## Build

```shell
wash build
```

## Deploy to wasmCloud on Kubernetes

The host group has to run with the plugin enabled and an OTLP endpoint to
export to. In the `runtime-operator` chart:

```yaml
runtime:
  env:
    # Setting the endpoint is what turns export on. The host speaks OTLP
    # over gRPC only.
    - name: OTEL_EXPORTER_OTLP_ENDPOINT
      value: "http://<your-collector>:4317"
  extraArgs:
    - "--wasi-otel"
```

Then push the component, point `image` in [deployment.yaml](./deployment.yaml)
at it, and apply the manifest:

```shell
wash oci push ghcr.io/<your-org>/http-otel:0.1.0 build/http_otel.wasm
kubectl apply -f deployment.yaml
```

Two things in the manifest have to match your cluster:

- **`config.host`** on the `wasi:http` entry. The host routes each request to
  a workload by its `Host` header, so set this to the hostname your ingress
  forwards to the `http-otel` Service. Keep it unique per workload: two
  workloads naming the same host become replicas of one route and split its
  traffic.
- **`wasi:otel`** stays listed under `hostInterfaces` alongside `wasi:http`;
  a real host binds only the interfaces the manifest names.

See the wasmCloud [workload deployment
quickstart](https://wasmcloud.com/docs/quickstart/deploy-a-webassembly-workload/)
for cluster setup.
