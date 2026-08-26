# HTTP Server

A WebAssembly component that serves HTTP requests using Go's standard
`net/http` types via `go.wasmcloud.dev/component/net/wasihttp`.

## Prerequisites

- Go 1.25+
- [`wash`](https://wasmcloud.com/docs/installation) 2.x

## Develop

Start a local dev loop (builds the component and serves it, rebuilding on
change):

```shell
wash dev
```

Then try the endpoints:

```shell
curl localhost:8000
curl localhost:8000/headers
curl -X POST -d key=value localhost:8000/form
```

## Build

```shell
wash build
```

## Deploy to wasmCloud on Kubernetes

Push the component to an OCI registry and apply the deployment manifest:

```shell
wash oci push ghcr.io/<your-org>/http-server:0.1.0 build/http_server.wasm
kubectl apply -f deployment.yaml
```

See [deployment.yaml](./deployment.yaml) for the `WorkloadDeployment`
definition and the wasmCloud [workload deployment
quickstart](https://wasmcloud.com/docs/quickstart/deploy-a-webassembly-workload/)
for cluster setup.
