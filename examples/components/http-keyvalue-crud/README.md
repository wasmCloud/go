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

## Develop

```shell
wash dev
curl -X POST localhost:8000/crud/mykey -d '{"foo": "bar"}'
curl localhost:8000/crud/mykey
curl -X DELETE localhost:8000/crud/mykey
```

## Build & deploy

```shell
wash build
wash oci push ghcr.io/<your-org>/http-keyvalue-crud:0.1.0 build/http_keyvalue_crud.wasm
kubectl apply -f deployment.yaml
```
