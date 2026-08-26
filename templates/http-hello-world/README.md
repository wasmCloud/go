# http-hello-world

A Go WebAssembly component that handles HTTP requests with standard
`net/http` types, built with the
[wasmCloud Go SDK](https://github.com/wasmCloud/go).

## Prerequisites

- Go 1.25+
- [`wash`](https://wasmcloud.com/docs/installation) 2.x

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
`go tool componentize-go -w wasmcloud:component-go/wasip3@0.2.0 build` and
use `go.wasmcloud.dev/component/net/wasihttp3`.

## Deploy to wasmCloud on Kubernetes

```shell
wash oci push ghcr.io/<your-org>/http-hello-world:0.1.0 build/http_hello_world.wasm
kubectl apply -f deployment.yaml
```

See the [wasmCloud workload deployment quickstart](https://wasmcloud.com/docs/quickstart/deploy-a-webassembly-workload/).
