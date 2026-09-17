# HTTP P3 Streaming

An **async WASI P3** component (`wasi:http/handler@0.3.0`) demonstrating
what the P3 world adds over sync P2:

- `POST /echo` — streams the request body back without buffering; bytes
  flow as they arrive.
- `GET /fanout` — issues concurrent outbound requests with plain
  goroutines, multiplexed natively by the component-model async runtime.

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

### Toolchain note

Async worlds require a Go runtime carrying `runtime.wasiOnIdle`
([golang/go#76775](https://github.com/golang/go/pull/76775)).
componentize-go uses the `go` on your `PATH` when it already has the patch
and downloads a patched toolchain otherwise, so no setup is needed either
way. Sync components (see the other examples) build with stock Go today.

## Develop

```shell
wash dev
curl -X POST --data-binary @somefile localhost:8000/echo
curl localhost:8000/fanout
```

Outbound hosts for `/fanout` are declared in
[.wash/config.yaml](./.wash/config.yaml) under `workload.allowedHosts`.

## Build

```shell
wash build
```

This runs `go tool componentize-go -w wasmcloud:component-go/wasip3@0.2.0 build`:
the `-w` selects the SDK's P3 world, and componentize-go sets the
`componentizego_async` build tag itself, which is what switches `wasihttp` to
its P3 implementation.

## Deploy to wasmCloud on Kubernetes

Push the component to an OCI registry, point `image` in
[deployment.yaml](./deployment.yaml) at it, and apply the manifest:

```shell
wash oci push ghcr.io/<your-org>/http-p3-streaming:0.1.0 build/http_p3_streaming.wasm
kubectl apply -f deployment.yaml
```

The manifest has the same shape as a P2 one — the host serves
`wasi:http/handler@0.3.0` behind the same `incoming-handler` entry. Two
things in it have to match your cluster:

- **`localResources.allowedHosts`** on the component. Outbound HTTP is
  deny-by-default on a real host, and `.wash/config.yaml` is read only by
  `wash dev`, so the manifest carries its own copy of the allowlist
  (`dog.ceo`). Without it `/fanout` fails.
- **`config.host`** on the `wasi:http` entry. The host routes each request to
  a workload by its `Host` header, so set this to the hostname your ingress
  forwards to the `http-p3-streaming` Service. Keep it unique per workload: two
  workloads naming the same host become replicas of one route and split its
  traffic.

See the wasmCloud [workload deployment
quickstart](https://wasmcloud.com/docs/quickstart/deploy-a-webassembly-workload/)
for cluster setup.
