# HTTP P3 Streaming

An **async WASI P3** component (`wasi:http/handler@0.3.0`) demonstrating
what the P3 world adds over sync P2:

- `POST /echo` — streams the request body back without buffering; bytes
  flow as they arrive.
- `GET /fanout` — issues concurrent outbound requests with plain
  goroutines, multiplexed natively by the component-model async runtime.

## Toolchain note

Async worlds currently require a patched Go runtime
(`runtime.wasiOnIdle`); componentize-go downloads it automatically. Once
[golang/go#76775](https://github.com/golang/go/pull/76775) merges, stock Go
will work. Sync components (see the other examples) build with stock Go
today.

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
