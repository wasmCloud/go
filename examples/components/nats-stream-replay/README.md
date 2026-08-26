# JetStream replay and pull consumer (Go)

Three ways to read a JetStream stream from a component, behind HTTP so you
can drive them by hand:

```
GET  /streams/{stream}/messages/{seq}          one message, by sequence
GET  /streams/{stream}/replay?from=&count=     a range, consuming nothing
POST /streams/{stream}/consumers/{name}/drain  a batch from a durable consumer
```

The first two read the stream directly and create no consumer, so they do
not move anyone else's position — that is what makes them safe to point at a
live stream, and it is the thing `wasmcloud:messaging` cannot express. The
third does the opposite on purpose: it takes messages from a durable
consumer and settles them, so what it reads, nobody else gets.

## What it demonstrates

| Behaviour | Where |
| --- | --- |
| Replaying an arbitrary range without a consumer | `nats.Scan` in `replay` |
| Paginating a short read rather than assuming the stream ended | the `next` field |
| Reading a single message by sequence | `nats.GetBySequence` |
| Attaching to a consumer that already exists | `nats.OpenPullConsumer` |
| An empty fetch as an idle result, not a failure | `ErrNoMessages` -> 204 |
| Telling a short batch from a drained consumer | `FetchedBatch.Stop` |
| Manual settling: ack, nak with backoff, term | `settle` |
| Extending ack-wait for slow work | `h.InProgress()` |
| Mapping JetStream failures onto HTTP status by error type | `writeNatsError` |

`settle` is the part worth reading. Which of the three outcomes a message
gets is the whole design of a worker: an ack that should have been a nak
drops work silently, a nak that should have been a term retries a
never-succeeding message forever, and a term that should have been a nak
throws away a message that only needed a moment. The nak backoff is
multiplied by `DeliveryCount`, so a message that keeps failing keeps
failing more slowly.

Note that a short `Scan` is normal. The host caps both the number of
messages and the time a single scan may take, so a call asking for 500 can
come back with 40 — resume from `next` rather than concluding you reached
the end.

## Prerequisites

The server and the CLI are separate packages, and the examples use
both — `brew install nats-server nats` on macOS.

Neither stream nor consumer is created by the component — that lifecycle is
deliberately outside `wasmcloud:nats`, so a workload cannot provision
durable state it was not granted:

```bash
nats-server -js &

nats stream add EVENTS --subjects 'events.>' \
  --storage file --retention limits --discard old \
  --max-msgs=-1 --max-bytes=-1 --max-age=24h --replicas 1 --defaults

# Explicit ack policy, because the drain endpoint settles each message itself.
nats consumer add EVENTS workers --pull --ack explicit \
  --deliver all --max-deliver 5 --wait 30s --defaults
```

## Build

```bash
GOFLAGS=-tags=componentizego_async componentize-go \
  -w 'wasmcloud:component-go/wasip3@0.2.0' \
  -w 'wasmcloud:examples/nats-stream-replay@0.1.0' build
```

Both worlds are named because this is a **WASI P3** component: every
`wasmcloud:nats` function is an `async func`, so it takes the SDK's `wasip3`
world (`wasi:http/handler@0.3.0`) rather than the default `wasip2` one, and
its own world adds the JetStream imports on top. The `GOFLAGS` prefix selects
`wasihttp`'s async P3 implementation; componentize-go sets that tag on its own
for async worlds, but the `go tool` wrapper still fetches the v0.4.1 binary,
which predates that. The prefix can go once a later release ships.

## Running it

`wash dev` builds and runs it against the NATS server above:

```bash
wash dev
```

It serves the three routes on `localhost:8000`, which is where the calls
below are pointed.

`wasmcloud:nats` is a host plugin, and `wash dev` has no manifest to read
the binding from — so `.wash/config.yaml` carries the binding whole:
servers, grants, and asks together. It is deliberately *not* a mirror of
`deployment.yaml` any more. Dev runs
`--wasmcloud-nats-workload-config=allow` precisely so a checkout is
runnable on its own, while a real host runs `deny` and keeps the servers,
the credentials, and the grants on its side. So a grant lives in this file
for dev and in the host group's `wasmcloudNats` for a cluster — change it
in both.

The two are not quite interchangeable, and the difference is worth knowing
before a component that works in dev fails on deploy: dev derives host
interfaces from the component's imports, so `wasi:logging` binds there
whether or not you declare it. A real host binds only what the manifest
names, which is why `deployment.yaml` lists it and `.wash/config.yaml`
does not.

## Try it

```bash
nats pub events.one 'colour=blue'
nats pub events.two 'size=large'
nats pub events.three 'nonsense'          # no '=', so the drain terms it
nats pub events.four 'retry=fail'         # naks with a backoff

curl 'localhost:8000/streams/EVENTS/messages/1'
curl 'localhost:8000/streams/EVENTS/replay?from=1&count=2'   # note "next": 3
curl 'localhost:8000/streams/EVENTS/replay?from=3&count=2'

curl -i -X POST 'localhost:8000/streams/EVENTS/consumers/workers/drain?batch=10'
```

The drain response names an outcome per message, under a header saying why
the batch ended:

```json
{
  "stream": "EVENTS",
  "consumer": "workers",
  "stop": "drained",
  "results": [
    {"sequence":1,"subject":"events.one","deliveryCount":1,"outcome":"ack"},
    {"sequence":2,"subject":"events.two","deliveryCount":1,"outcome":"ack"},
    {"sequence":3,"subject":"events.three","deliveryCount":1,"outcome":"term"},
    {"sequence":4,"subject":"events.four","deliveryCount":1,"outcome":"nak",
     "detail":"downstream unavailable"}
  ]
}
```

`stop` is the field a real worker loops on. `drained` means the consumer had
nothing left to give, so four results out of a batch of ten is the whole
backlog; `batch-filled` or `byte-limit` mean a bound cut the batch short with
messages still waiting, and the next fetch picks up where this one stopped.
Without it a short batch and an empty consumer look identical, and a reader
that stops on a short batch leaves messages behind.

`detail` appears only when there is something to say: the failure that caused
a nak, or — for any outcome — a settling call the host refused, which is what
you get under a binding configured `ack-mode: auto`.

Run it a second time and only the naked message comes back, with
`deliveryCount` incremented — the acked and termed ones are gone for good.
Run it once more with nothing pending and it answers `204 No Content`.

Asking for a stream outside `stream-allow` is refused host-side:

```bash
curl -i 'localhost:8000/streams/SECRETS/replay'   # -> 403, never reaches the server
```

## Declaring the binding host-side

`wash host` runs `--wasmcloud-nats-workload-config=deny`, which splits a
`wasmcloud:nats` binding between two people. The host declares what the
binding *is* — the servers it dials, the credentials it dials them with, and
the subject, stream, and bucket grants it carries. The workload declares only
what it wants delivered within that grant. A manifest that sets `servers`,
`creds`, `jwt`/`nkey-seed`, `username`/`password`, `token`, `tls-*`,
`jetstream-domain`, `inbox-prefix`, or any `*-allow` is refused at bind:

```
wasmcloud:nats binding `<unnamed>` sets `subject-allow`, which this host does
not accept from a workload.
```

That refusal is the feature. A workload can ask for a capability; it can
never widen one.

In the chart, the declaration goes on the host group:

```yaml
runtime:
  hostGroups:
    - name: default
      # Omit `servers` to take the cluster's own NATS (`dataNatsUrl`), so the
      # same manifest runs under `wash dev` and here.
      wasmcloudNats:
        config:
          stream-allow: ORDERS,EVENTS
          # Reads are filtered by the subject grant too: `scan` and
          # `get-by-sequence` only return a message whose stored
          # subject falls inside it.
          subject-allow: events.>,orders.>
        # NATS credentials reach the host this way rather than through a
        # workload manifest or a CLI arg — the rendered `wash host` config
        # file never appears in `kubectl describe pod`.
        secretFrom:
          - events-nats-creds
```

`wash dev` runs the other way round (`allow`): the host's declaration is a
default a manifest may override, so `.wash/config.yaml` still carries the
whole binding and a checkout stays runnable on its own.

## Deploy to wasmCloud on Kubernetes

The target is a wasmCloud v2 host whose runtime carries the `wasmcloud:nats`
plugin — the plugin ships with the host, not with the component, so a host
built without it rejects the binding at placement — and whose host group
declares the binding above.

Push the component and apply the manifest:

```shell
wash oci push ghcr.io/<your-org>/nats-stream-replay:0.1.0 build/nats_stream_replay.wasm
kubectl apply -f deployment.yaml
```

See [deployment.yaml](./deployment.yaml) for the `WorkloadDeployment`
definition and the wasmCloud [workload deployment
quickstart](https://wasmcloud.com/docs/quickstart/deploy-a-webassembly-workload/)
for cluster setup.
