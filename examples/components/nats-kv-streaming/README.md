# Large objects on NATS KV, streamed between components (Go)

An archive of sensor-reading exports, stored in a NATS KV bucket. A
long-running service writes a fresh ~10 MB export every few minutes. An HTTP
API uploads, downloads and queries objects. A query component filters the
readings as they stream past. No component ever holds more than one 512 KiB
chunk of an object, whatever its size.

It is four components in **one workload**. That matters: inside a workload the
host links one component's imports to another's exports, so they call each
other directly, and the `stream<u8>` handles they pass go straight from one
component to the next.

```
                      wasmcloud:archive/store            wasmcloud:nats/kv
 recorder ──── put(name, stream) ───────▶ store ─────── 512 KiB chunks ───────▶ NATS KV "archive"
 (service)                                 ▲ │
                                           │ │ get(name) → stream
 HTTP ──▶ api ─── put / get / list ────────┘ │
           │                                 ▼
           └── query.run(that same stream) ──▶ query ──▶ stream of matches ──▶ api ──▶ HTTP
```

| Directory | What it is | Built from |
| --- | --- | --- |
| [`recorder/`](./recorder) | The workload's **service**: started with the workload, runs until it stops. Every interval it generates 100,000 readings (~10.9 MB of NDJSON) and streams them into the store. | SDK `headless` + `archive/client` + its own `wasi:cli/run@0.3.0` world |
| [`store/`](./store) | Exports `wasmcloud:archive/store`. The only component that talks to NATS. | SDK `headless` + `archive/store-exports` + its own `wasmcloud:nats` world |
| [`query/`](./query) | Exports `wasmcloud:archive/query`. Pure compute: filters and re-encodes NDJSON as it streams. Imports nothing but logging. | SDK `headless` + `archive/query-exports` |
| [`api/`](./api) | `net/http` on `wasi:http/handler@0.3.0`, calling the store and query components. | SDK `wasip3` + `archive/client` |
| [`contract/`](./contract) | The shared WIT package, `wasmcloud:archive@0.1.0`, which `wash build` encodes into `archive-0.1.0.wasm`. Also the Go bindings generated from it (one package per side), and `streamio`, which turns a stream and its future into an `io.Reader`/`io.Writer`. | `wash wit build` |

Measured on a laptop against a stock `nats-server -js` (1 MiB `max_payload`):
the recorder stores 100,000 readings (10.9 MB, 21 chunks) in about 350 ms, a
10 MiB `PUT` takes about 110 ms, and a query scans all 100,000 readings in
under a second.

## What it demonstrates

| Behaviour | Where |
| --- | --- |
| Moving a 10 MB object as `stream<u8>`, a chunk at a time, never as one value | `store.put`, `store.get` |
| Pairing every stream with a `future<result>` so the reader knows whether it got everything | [`contract/wit/world.wit`](./contract/wit/world.wit), `streamio` |
| Handing a stream from one component to another without reading it | `queryObject` in `api` |
| Backpressure across three components: a slow HTTP client slows the KV reads | `copyChunks` in `store` |
| Stopping early: the query's `limit` drops its input, which stops the store reading chunks | `filterReadings` in `query` |
| An aborted upload stores nothing and leaves the previous version untouched | `put` in `store`, `putObject` in `api` |
| Atomic replace: chunks under a fresh upload id, manifest swapped last with compare-and-swap | `commit` in `store` |
| A reader mid-download survives a replace | the previous generation's chunks, kept one more put |
| A long-running Go service that sleeps on the host clock | `recorder` (`sleep.Sleep`, never `time.Sleep`) |
| Results whose counts are only known at the end, as HTTP trailers | `queryObject` in `api` |

## Large values and Go components: read this first

This example exists because of two failure reports from Go components built
with componentize-go:

- a component that receives a `list<record>` of 20,000 records works, and one
  that receives 40,000 records (under 3 MB in total) traps;
- a component that reads a single 8.8 MiB value from NATS KV traps, with
  `gcStart` and `cabi_realloc` in the backtrace.

**Both are the same bug, fixed in `go.bytecodealliance.org/pkg` v0.4.1.**
When the host passes a value into a component, it calls the component's
`cabi_realloc` to allocate memory for it: once for a byte list, and once per
string or nested list inside a list of records. A Go allocation can start a
garbage collection cycle. In v0.4.0 and earlier, that cycle could call a host
import while the host was still inside `cabi_realloc`, which the component
model forbids, so the instance trapped. The more allocations a value needs,
the likelier it is: about 80,000 `cabi_realloc` calls for 40,000 records, or
one very large one for an 8.8 MiB blob. The fix
([bytecodealliance/go-pkg#13](https://github.com/bytecodealliance/go-pkg/pull/13))
defers GC work while `cabi_realloc` runs.

To pick it up in your own component:

```bash
go get go.bytecodealliance.org/pkg@v0.4.1
```

The wasmCloud SDK requires v0.4.1 from `go.wasmcloud.dev/component` v0.2.0,
so moving to that release picks it up too. On an older SDK, the `go get`
above is enough: Go's minimum version selection takes the higher version.

The fix stops the trap. The design here is still the one to copy for large
data, for three reasons:

1. **NATS caps a message at `max_payload`, 1 MiB by default.** A KV value is
   one message, so an 8.8 MiB value needs a server reconfigured for it, and
   so does every client, mirror and leaf node that ever carries it. Chunks of
   512 KiB fit any default server.
2. **A value arrives in one allocation.** Reading a 10 MB value means a 10 MB
   allocation in the guest, and a 10 MB `list<u8>` crossing a component
   boundary means another one on the far side. `stream<u8>` reads go into a
   buffer the reader already owns, so a 10 MB object costs each component one
   chunk-sized buffer.
3. **A stream starts delivering right away.** The api sends the first bytes
   of a download while the store is still reading the second chunk from NATS,
   and a query streams results while its input is still arriving.

## The contract

[`contract/wit/world.wit`](./contract/wit/world.wit) defines two interfaces:

```wit
interface store {
  put: async func(name: string, content-type: string,
                  data: stream<u8>, done: future<result<_, string>>)
       -> result<object-info, error>;
  get: async func(name: string)
       -> result<tuple<object-info, stream<u8>, future<result<_, string>>>, error>;
  stat: async func(name: string) -> result<object-info, error>;
  list-objects: async func() -> result<list<object-info>, error>;
  delete: async func(name: string) -> result<_, error>;
}

interface query {
  run: async func(input: stream<u8>, input-done: future<result<_, string>>, filter: filter)
       -> tuple<stream<u8>, future<result<summary, string>>>;
}
```

**Every stream travels with a future.** A stream that ends says nothing about
why it ended: the writer may have finished, or failed half way through. So
the writer resolves the future after its last byte, with `ok` when everything
arrived and `err` when what came before is a fragment. This is the same
pairing `wasi:http` uses for a body and its trailers. `streamio` turns the
pair into the Go interfaces you already use:

```go
// Writing: the writer side of a stream and its future.
out := streamio.NewWriter(data, done)
_, err := io.Copy(out, src)
streamio.Close(out, err) // drops the stream, then resolves the future with err (or ok)

// Reading: io.EOF only once the future says ok.
in := streamio.NewReader(stream, done)
defer in.Close()
_, err = io.Copy(dst, in) // a writer's err comes back as *streamio.RemoteError
```

The order in `Close` matters: drop the stream **before** resolving the
future. A reader blocked in `Read` only reaches the future once the stream
ends, so the other order deadlocks both sides.

**`get` returns its stream before the stream is full.** The export creates
the stream, starts a goroutine to fill it, and returns. The export call is
over, but the goroutine keeps running on the same instance, writing each
chunk as the reader takes the last one.

**The api never reads the object it queries.** It calls `store.get`, passes
the stream and future it got to `query.run` unopened, and copies only the
query's output into the response. Ownership of a stream handle moves with the
call, so the bytes go from the store component to the query component
directly.

The worlds at the bottom of the file (`client`, `store-exports`,
`query-exports`) name one side of the interfaces each. Every component's
`.wash/config.yaml` merges one of them with an SDK world (`-w` once per
world), and the generated packages in `contract/bindings/` are what the Go
code imports. Because the package imports nothing, it has no WIT
dependencies to fetch.

`wash build` in `contract/` encodes the package into `archive-0.1.0.wasm`.
That binary is the contract in a form any toolchain can consume: publish it
to a registry and a component in another language can depend on
`wasmcloud:archive` through `wash wit fetch`, the same way these components
depend on `wasmcloud:nats`.

## How objects are stored

```
meta.<name>                     the manifest: size, chunk count, SHA-256, upload id, …
chunk.<name>.<upload>.<n>       chunk n of one version of the object
```

This is the same layout NATS's own object store uses, on a plain KV bucket.

- **Put** writes chunks under a fresh upload id, hashing as it goes, then
  swaps the manifest in with compare-and-swap. The manifest swap is the
  commit. A failure before it purges the new chunks and leaves the old
  version exactly as it was.
- **Get** reads the manifest once, then streams that version's chunks in
  order and checks the SHA-256 at the end. A reader that started before a
  replace keeps reading the version it started with.
- **The replaced version's chunks are kept until the put after next.** That
  gives a download that was under way when its object was replaced one full
  generation to finish. A reader that is still going two replaces later fails
  with an `err` on its future rather than mixing versions.
- **Delete** purges the manifest first, then the chunks. A failure part way
  through leaves unreachable chunks, never a half-deleted object that still
  lists.

## Prerequisites

On a fresh clone, download the Go modules in each component's directory:

```bash
for d in store query recorder api; do (cd $d && go mod download); done
```

componentize-go reads the SDK's WIT from the module cache and does not fetch
it itself, so until the modules are downloaded the build fails with
`failed to read path for WIT [wit]`.

A NATS server with JetStream, and the bucket. The bucket keeps one revision
per key (`--history 1`): chunks are written once and never updated, and the
store keeps its own previous generation.

```bash
nats-server -js &
nats kv add archive --history 1
```

The server and the CLI are separate packages: `brew install nats-server nats`
on macOS.

`wash` with the `wasmcloud:nats` host plugin. Tested with wash 2.10.3 and
2.12.0.

## Build

```bash
for d in contract store query recorder api; do (cd $d && wash build); done
```

Each runs the recipe in its own `.wash/config.yaml`. The contract's is
`wash wit build`; the components don't need its output, since they compile
the WIT through the committed Go bindings. The store's is:

```bash
go tool componentize-go \
  -w wasmcloud:component-go/headless@0.2.0 \
  -w wasmcloud:archive/store-exports@0.1.0 \
  -w wasmcloud:examples/nats-kv-store@0.1.0 \
  build -o build/store.wasm
```

All four are **WASI P3** components. `wasmcloud:nats` and the archive
interfaces are async throughout, so componentize-go selects a Go toolchain
with async support. The note `does not support async operation; will use
downloaded version` is expected on the first build.

## Run it

```bash
cd api && wash dev
```

`api/.wash/config.yaml` describes the whole workload. It names the store and
query components (`dev.components`) and the recorder (`dev.service_file`) by
their built files. `wash dev` only builds the directory it runs in, so build
the other three first and rebuild them yourself after changing them.

Within a few seconds the recorder writes its first export:

```
recorder started  object=readings records=100000 interval=5m0s
stored object     name=readings bytes=10888011 chunks=21 sha256=6b1002bc… elapsed=342ms
export stored     object=readings records=100000 bytes=10888011 chunks=21 …
```

## Try it

List what is stored:

```bash
curl -s localhost:8000/objects
```

```json
[
  {
    "name": "readings",
    "contentType": "application/x-ndjson",
    "size": 10888011,
    "chunkSize": 524288,
    "chunks": 21,
    "sha256": "6b1002bc61fe6d4776d8294da01a093202ac4cc930db79c071ff76e058e22a51",
    "createdAt": "2026-11-07T20:18:31.756Z"
  }
]
```

Download it and check it against the manifest. `ETag` and `X-Archive-Sha256`
carry the SHA-256, and `Content-Length` the size, both known before the first
byte:

```bash
curl -sI localhost:8000/objects/readings
curl -s  localhost:8000/objects/readings | shasum -a 256
```

Query it. The store streams the export into the query component, and the
response carries only the matches. These are the readings at or above 95 °C,
as CSV:

```bash
curl -s 'localhost:8000/objects/readings/query?min=95&format=csv' | head
```

```
ts,site,sensor,metric,value
2026-11-07T19:18:31.909Z,plant-01,plant-01-s014,temp_c,95.52
2026-11-07T19:18:46.669Z,plant-07,plant-07-s040,temp_c,98.37
…
```

Filters are `site`, `sensor`, `min` and `limit`, and `format` is `ndjson`
(the default) or `csv`. Every export has a few hot sensors, chosen at random,
so `min=95` always finds something.

How many lines were scanned and matched is only known once the query is done,
so it comes in HTTP trailers. curl only shows trailers in `--raw` mode, and
the host only sends them to a client that asks with `TE: trailers`:

```bash
curl -s --raw -H 'TE: trailers' 'localhost:8000/objects/readings/query?min=95&format=csv' | tail -n 5
```

```
…
0
x-query-scanned: 100000
x-query-matched: 352
x-query-malformed: 0
```

A query that matches nothing writes no body, so the headers only go out
once it is done. The counts are known by then and arrive as ordinary
response headers too.

Stop early. With `limit` the query component stops after that many matches
and drops its input, which tells the store to stop reading chunks:

```bash
curl -s 'localhost:8000/objects/readings/query?sensor=plant-03-s007&limit=3'
```

Upload something of your own. The body streams into the store; nothing is
buffered:

```bash
head -c 10485760 /dev/urandom > /tmp/blob.bin
curl -s -T /tmp/blob.bin -H 'Content-Type: application/octet-stream' localhost:8000/objects/blob
curl -s localhost:8000/objects/blob | cmp - /tmp/blob.bin && echo identical
```

Hang up half way through an upload, and nothing is stored:

```bash
(head -c 3000000 /tmp/blob.bin; sleep 5) | curl -s -T - localhost:8000/objects/partial &
sleep 1; kill %1
curl -s -i localhost:8000/objects/partial      # 404
```

The store logs `put aborted name=partial after_bytes=3000000` and purges the
chunks it wrote. A replace that is aborted the same way leaves the previous
version in place.

Replace an object while it is being downloaded, and the download still gets
the version it started with:

```bash
curl -s --limit-rate 2M localhost:8000/objects/readings | shasum -a 256 &
sleep 1; curl -s -T /tmp/blob.bin localhost:8000/objects/readings >/dev/null
wait   # prints the old SHA-256, not the new one
```

Delete:

```bash
curl -s -X DELETE -i localhost:8000/objects/blob   # 204
```

Errors map to status codes: `404` for an unknown object, `400` for an
invalid name (names are 1 to 64 characters of `[A-Za-z0-9_-]`, since they
become KV key tokens), `400` for an aborted upload, and `502` when a NATS
call fails.

## Operating notes

- **Purges leave markers.** Purging a key in a JetStream KV bucket writes a
  small marker in its place, so a bucket that is replaced every five minutes
  slowly accumulates them. `nats kv compact archive` removes them.
- **Listing is capped.** The host caps a key listing at 1,000 keys. The store
  lists `meta.*` only, so the cap counts objects, not chunks, and it logs a
  warning when a listing is cut short.
- **Chunk size is a constant.** `chunkSize` in `store/main.go` is 512 KiB,
  which leaves room under the default 1 MiB `max_payload` for message
  headers. Larger chunks mean fewer KV round trips and a larger buffer per
  read. The manifest records the chunk size, so changing it does not break
  objects that are already stored.
- **The recorder is configured through its environment**:
  `RECORDER_OBJECT` (default `readings`), `RECORDER_RECORDS` (default
  `100000`) and `RECORDER_INTERVAL` (default `5m`). After a failed export,
  usually because the bucket does not exist yet, it retries after 10 seconds.

## Declaring the binding host-side

`wasmcloud:nats` is configured like every other plugin, as a `host.plugins`
entry with `id: wasmcloud-nats`. Under `wash host` that entry defaults to
`workloadConfig: deny`. The host declares the servers, credentials and
grants, and a workload manifest only asks for the interfaces. A manifest that
sets `servers`, `creds` or any `*-allow` is refused at bind.

This workload needs one grant, the bucket:

```yaml
runtime:
  hostGroups:
    - name: default
      plugins:
        - id: wasmcloud-nats
          config:
            bucket-allow: archive
```

`wash dev` defaults the other way round (`workloadConfig: allow`), so
`api/.wash/config.yaml` carries the whole binding, servers and grant
included, and a checkout runs on its own.

## Deploy to wasmCloud on Kubernetes

The target is a wasmCloud v2 host whose runtime carries the `wasmcloud:nats`
plugin, with the binding above declared on its host group. Create the bucket,
push the four components, point each `image` in
[deploy/deployment.yaml](./deploy/deployment.yaml) at them, and apply the
manifest:

```shell
nats kv add archive --history 1
for c in api store query recorder; do
  wash oci push ghcr.io/<your-org>/nats-kv-streaming-$c:0.1.0 $c/build/$c.wasm
done
kubectl apply -f deploy/deployment.yaml
```

The manifest is one `WorkloadDeployment`: the api, store and query under
`components`, and the recorder under `service`, with its environment under
`localResources`. The HTTP side is routed by `Host` header: set `config.host`
on the `wasi:http` entry to the hostname your ingress forwards to the
`nats-kv-streaming` Service.

## Regenerating the bindings

The packages under `contract/bindings/` and `recorder/wit_exports/` are
generated and committed. The `export_*` packages next to them are written by
hand: each is the small trampoline the generated glue calls, and forwards to
the function the component registers in `init`. After changing the WIT, or
bumping componentize-go, regenerate with:

```bash
../regenerate_bindings.sh
```
