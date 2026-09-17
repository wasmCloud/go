# HTTP Password Checker

A WebAssembly component that scores password strength over HTTP, using
[go-password-validator](https://github.com/wagslane/go-password-validator).

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
curl -X POST localhost:8000/api/v1/check -d '{"value":"hunter2"}'                         # -> 400, "valid":false
curl -X POST localhost:8000/api/v1/check -d '{"value":"C0rrect-Horse-Battery-Staple!9"}'  # -> 200, "valid":true
```

## Build

```shell
wash build
```

## Deploy to wasmCloud on Kubernetes

Push the component to an OCI registry, point `image` in
[deploy/deployment.yaml](./deploy/deployment.yaml) at it, and apply the manifest:

```shell
wash oci push ghcr.io/<your-org>/http-password-checker:0.1.0 build/http_password_checker.wasm
kubectl apply -f deploy/deployment.yaml
```

Set `config.host` on the manifest's `wasi:http` entry to the hostname your
ingress forwards to the `http-password-checker` Service. The host routes each
request to a workload by its `Host` header, so keep it unique per workload:
two workloads naming the same host become replicas of one route and split its
traffic.

See the wasmCloud [workload deployment
quickstart](https://wasmcloud.com/docs/quickstart/deploy-a-webassembly-workload/)
for cluster setup.
