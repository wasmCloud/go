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
componentize-go build
```

No `-w` flag here: this component exports `wasi:http/incoming-handler`, so
its world merges on top of the SDK's default `wasip2` world rather than
replacing it.

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

The drain response names an outcome per message:

```json
{"sequence":3,"subject":"events.three","deliveryCount":1,"outcome":"term"}
```

Run it a second time and only the naked message comes back, with
`deliveryCount` incremented — the acked and termed ones are gone for good.
Run it once more with nothing pending and it answers `204 No Content`.

Asking for a stream outside `stream-allow` is refused host-side:

```bash
curl -i 'localhost:8000/streams/SECRETS/replay'   # -> 403, never reaches the server
```

## Running it

`wasmcloud:nats` is served by a wasmCloud host plugin, not by `wash dev` —
the dev host registers `wasmcloud:messaging` and `wasi:keyvalue` but not
this one yet. So the loop is `componentize-go build`, then deploy
`deployment.yaml` to a wasmCloud v2 host whose runtime carries the
`wasmcloud:nats` plugin, pointed at the NATS server above.
