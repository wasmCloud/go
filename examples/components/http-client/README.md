# HTTP Client

A WebAssembly component that serves HTTP requests and makes outbound HTTP
calls using `wasihttp.DefaultClient` (an `http.Client` over
`wasihttp.Transport`, an `http.RoundTripper` over
`wasi:http/outgoing-handler`), from `go.bytecodealliance.org/pkg/wasihttp`.

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
curl localhost:8000    # -> a random dog picture URL from dog.ceo
```

Outbound hosts are declared in [.wash/config.yaml](./.wash/config.yaml)
under `workload.allowedHosts`.

## Build

```shell
wash build
```

## Deploy to wasmCloud on Kubernetes

Push the component to an OCI registry, point `image` in
[deployment.yaml](./deployment.yaml) at it, and apply the manifest:

```shell
wash oci push ghcr.io/<your-org>/http-client:0.1.0 build/http_client.wasm
kubectl apply -f deployment.yaml
```

Two things in the manifest have to match your cluster:

- **`localResources.allowedHosts`** on the component. Outbound HTTP is
  deny-by-default on a real host, and `.wash/config.yaml` is read only by
  `wash dev`, so the manifest carries its own copy of the allowlist
  (`dog.ceo`). Without it every outbound request fails.
- **`config.host`** on the `wasi:http` entry. The host routes each request to
  a workload by its `Host` header, so set this to the hostname your ingress
  forwards to the `http-client` Service. Keep it unique per workload: two
  workloads naming the same host become replicas of one route and split its
  traffic.

See the wasmCloud [workload deployment
quickstart](https://wasmcloud.com/docs/quickstart/deploy-a-webassembly-workload/)
for cluster setup.
