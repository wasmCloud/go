# http-hello-world

A Go WebAssembly component that handles HTTP requests with standard
`net/http` types, built with the
[wasmCloud Go SDK](https://github.com/wasmCloud/go).

## Prerequisites

- Go 1.27+
- [`wash`](https://wasmcloud.com/docs/installation) 2.x

`wash new` runs `go mod tidy` for you. If you copied this directory by hand
instead, run it yourself first — the template ships without a `go.sum`:

```shell
go mod tidy
```

## Develop

Start the dev loop — it builds the component, serves it locally, and
rebuilds on change:

```shell
wash dev
curl localhost:8000
```

## Build

```shell
wash build
```

The component is a standard WASI P2 component built with stock Go via
[componentize-go](https://github.com/bytecodealliance/componentize-go); the
target world is declared by the SDK's `componentize-go.toml`. To target the
async WASI P3 world instead, change the build command to
`go tool componentize-go -w wasmcloud:component-go/wasip3@0.2.0 build`. The
code stays the same: componentize-go sets the `componentizego_async` build tag
for async worlds, which switches `go.bytecodealliance.org/pkg/wasihttp` to its
P3 implementation.

## Deploy to wasmCloud on Kubernetes

Push the component to an OCI registry, point `image` in
[deploy/deployment.yaml](./deploy/deployment.yaml) at it, and apply the manifest:

```shell
wash oci push ghcr.io/<your-org>/http-hello-world:0.1.0 build/http_hello_world.wasm
kubectl apply -f deploy/deployment.yaml
```

Set `config.host` on the manifest's `wasi:http` entry to the hostname your
ingress forwards to the `http-hello-world` Service. The host routes each
request to a workload by its `Host` header, so keep it unique per workload:
two workloads naming the same host become replicas of one route and split its
traffic.

See the [wasmCloud workload deployment quickstart](https://wasmcloud.com/docs/quickstart/deploy-a-webassembly-workload/).
