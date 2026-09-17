# HTTP Keyvalue CRUD

A WebAssembly component exposing CRUD operations over HTTP, backed by the
host-served `wasi:keyvalue/store` interface.

This example also shows how a component adds interfaces beyond the SDK's
default world: [wit/world.wit](./wit/world.wit) declares a small world
importing `wasi:keyvalue/store`, [componentize-go.toml](./componentize-go.toml)
merges it with the SDK's `wasip2` world at build time, and the
`wasi_keyvalue_store/` bindings are generated with:

```shell
go tool componentize-go --ignore-toml-files \
  -w wasmcloud:examples/http-keyvalue-crud@0.1.0 -d wit bindings -o .
```

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
curl -X POST localhost:8000/crud/mykey -d '{"foo": "bar"}'   # -> {"message":"Set mykey","value":{"foo":"bar"}}
curl localhost:8000/crud/mykey                               # -> {"message":"Got mykey","value":{"foo":"bar"}}
curl -X DELETE localhost:8000/crud/mykey                     # -> {"message":"Deleted mykey"}
curl -i localhost:8000/crud/mykey                            # -> 404, mykey does not exist
```

The body of a `POST` must be a JSON document (anything else is a 400). It is
stored as sent and returned embedded in the response as JSON, not as a
quoted string.

## Build

```shell
wash build
```

## Deploy to wasmCloud on Kubernetes

Push the component to an OCI registry, point `image` in
[deploy/deployment.yaml](./deploy/deployment.yaml) at it, and apply the manifest:

```shell
wash oci push ghcr.io/<your-org>/http-keyvalue-crud:0.1.0 build/http_keyvalue_crud.wasm
kubectl apply -f deploy/deployment.yaml
```

Two things in the manifest have to match your cluster:

- **`config.host`** on the `wasi:http` entry. The host routes each request to
  a workload by its `Host` header, so set this to the hostname your ingress
  forwards to the `http-keyvalue-crud` Service. Keep it unique per workload: two
  workloads naming the same host become replicas of one route and split its
  traffic.
- **`wasi:keyvalue`** stays listed. The host serves `wasi:keyvalue/store` from
  its own backend (NATS JetStream KV on a stock install), and binds it only
  because the manifest names it.

See the wasmCloud [workload deployment
quickstart](https://wasmcloud.com/docs/quickstart/deploy-a-webassembly-workload/)
for cluster setup.
