# HTTP Server

A WebAssembly component that serves HTTP requests using Go's standard
`net/http` types via `go.bytecodealliance.org/pkg/wasihttp`, and logs through
`wasi:logging` with `go.bytecodealliance.org/pkg/wasilog`.

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
curl -X POST -d hello localhost:8000/post
curl localhost:8000/error    # -> 500
```

## Build

```shell
wash build
```

## Deploy to wasmCloud on Kubernetes

Push the component to an OCI registry, point `image` in
[deployment.yaml](./deployment.yaml) at it, and apply the manifest:

```shell
wash oci push ghcr.io/<your-org>/http-server:0.1.0 build/http_server.wasm
kubectl apply -f deployment.yaml
```

Two things in the manifest have to match your cluster:

- **`config.host`** on the `wasi:http` entry. The host routes each request to
  a workload by its `Host` header, so set this to the hostname your ingress
  forwards to the `http-server` Service. Keep it unique per workload: two
  workloads naming the same host become replicas of one route and split its
  traffic.
- **`wasi:logging`** stays listed. A real host binds only the interfaces the
  manifest names, and this component imports `wasi:logging`; without the
  entry it fails to start with `a matching implementation was not found in
  the linker`. `wash dev` derives it from the imports, so dev never shows it.

See [deployment.yaml](./deployment.yaml) for the `WorkloadDeployment`
definition and the wasmCloud [workload deployment
quickstart](https://wasmcloud.com/docs/quickstart/deploy-a-webassembly-workload/)
for cluster setup.
