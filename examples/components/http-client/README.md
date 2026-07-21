# HTTP Client

A WebAssembly component that serves HTTP requests and makes outbound HTTP
calls using `wasihttp.Transport` (an `http.RoundTripper` over
`wasi:http/outgoing-handler`).

## Develop

```shell
wash dev
curl localhost:8000
```

Outbound hosts are declared in [.wash/config.yaml](./.wash/config.yaml)
under `workload.allowedHosts`.

## Build & deploy

```shell
wash build
wash oci push ghcr.io/<your-org>/http-client:0.1.0 build/http_client.wasm
```

On Kubernetes, deploy with a `WorkloadDeployment` declaring both
`wasi:http/incoming-handler` and `wasi:http/outgoing-handler` in
`hostInterfaces` — see the
[http-server example](../http-server/deployment.yaml) and add:

```yaml
- namespace: wasi
  package: http
  interfaces:
    - outgoing-handler
```
