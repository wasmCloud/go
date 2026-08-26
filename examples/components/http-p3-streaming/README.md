# HTTP P3 Streaming

An **async WASI P3** component (`wasi:http/handler@0.3.0`) demonstrating
what the P3 world adds over sync P2:

- `POST /echo` — streams the request body back without buffering; bytes
  flow as they arrive.
- `GET /fanout` — issues concurrent outbound requests with plain
  goroutines, multiplexed natively by the component-model async runtime.

## Toolchain note

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

## Build & deploy

```shell
wash build
wash oci push ghcr.io/<your-org>/http-p3-streaming:0.1.0 build/http_p3_streaming.wasm
```
