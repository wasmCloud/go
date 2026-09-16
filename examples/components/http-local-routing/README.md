# HTTP local routing (Go)

Two async WASI P3 components demonstrating **same-host local routing**: when a
caller and a callee run on the same wasmCloud host and the host enables local
routing, the caller's outgoing `wasi:http` request is served by the callee
**in-memory**. It never touches the network, kube-proxy, or an ingress.

- [`caller/`](./caller/) makes an outgoing request with `wasihttp.DefaultClient`
  and reports the result. Target: `?url=`, then `CALLEE_URL`, then
  `http://callee.internal/hello`.
- [`callee/`](./callee/) answers with a greeting echoing the path and `Host`
  it received.

Rust counterpart:
[wasmCloud/examples/local-ingress](https://github.com/wasmCloud/wasmCloud/tree/main/examples/local-ingress).

## How it is switched on

Two keys, and neither works alone:

1. **The host** enables the capability: `wash host --http-local-routing`, or
   `runtime.hostGroups[].http.localBypassRouting: true` in the runtime-operator
   chart.
2. **The callee** offers names to co-located callers with `localRoute` on its
   `wasi:http/incoming-handler` config.

| `localRoute` entry | Serves |
| --- | --- |
| `callee.internal` | every path on that hostname |
| `callee.internal/hello` | `/hello` and below, on `/` segment boundaries |

The caller declares nothing about local routing. It still needs an
`allowedHosts` entry for the authority, which is checked first.

`host` and `localRoute` are separate scopes. A `host` name is never
short-circuited, and a `localRoute` name is never reachable from the network.

> **This is a trust boundary.** A `localRoute` is a claim, not proof of
> ownership: any workload on the host can claim any hostname, including a
> public one, and receive its neighbours' requests to it in plaintext with no
> TLS check. Locally routed calls also skip ingress auth, mesh mTLS and
> NetworkPolicy. Enable it only on a hostgroup whose workloads are equally
> trusted.

## Build

```shell
wash -C callee build
wash -C caller build
```

Both build against the SDK's `wasmcloud:component-go/wasip3@0.2.0` world.
componentize-go downloads a patched Go for async worlds by itself; see
[http-p3-streaming](../http-p3-streaming#toolchain-note).

## Develop with wash dev

`wash dev` runs one component per session, so in dev the caller reaches the
callee over loopback (`CALLEE_URL` in
[caller/.wash/config.yaml](./caller/.wash/config.yaml)). This runs the same
code but not the in-memory short-circuit.

```shell
wash -C callee dev   # terminal 1, :8001
wash -C caller dev   # terminal 2, :8000
```

```shell
$ curl localhost:8000
caller -> http://localhost:8001/hello
upstream status: 200 OK
upstream body: hello from the Go callee! (path: /hello, host: localhost)
```

## Deploy to Kubernetes

The hostgroup needs `http.localBypassRouting: true`, and both workloads must
land on the same host (one replica, or a dedicated hostgroup). Push the
components, replace `ghcr.io/[your-org]` in
[deployment.yaml](./deployment.yaml), and apply:

```shell
wash oci push ghcr.io/<your-org>/http-local-routing-callee:0.1.0 callee/build/http_local_routing_callee.wasm
wash oci push ghcr.io/<your-org>/http-local-routing-caller:0.1.0 caller/build/http_local_routing_caller.wasm
kubectl apply -f deployment.yaml
```

On a kind cluster with Traefik on `:80`, the caller is at
`go-hello.localhost.cosmonic.sh`:

```shell
$ curl http://go-hello.localhost.cosmonic.sh/
caller -> http://callee.internal/hello
upstream status: 200 OK
upstream body: hello from the Go callee! (path: /hello, host: callee.internal)
```

`callee.internal` has no DNS record or Service, so the only way that request
reached the callee is the host's in-memory route. Each of these isolates one
property and should fail:

```shell
# Path outside the prefix: falls through to the network, where nothing answers.
curl 'http://go-hello.localhost.cosmonic.sh/?url=http://callee.internal/'

# Name nobody offers (and not in allowedHosts either).
curl 'http://go-hello.localhost.cosmonic.sh/?url=http://nope.internal/hello'

# A localRoute name is not reachable from the network.
curl -H 'Host: callee.internal' http://localhost/hello
```

Set `localBypassRouting: false` and the first request fails too.

## Required capabilities

- `wasi:http/handler@0.3.0` export (both)
- `wasi:http/client@0.3.0` import (caller)
