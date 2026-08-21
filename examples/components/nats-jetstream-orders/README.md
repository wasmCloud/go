# NATS JetStream order processor (Go)

Consumes orders from a JetStream stream, accumulates a per-order total in
JetStream KV, and publishes a processed-order notification — using
`wasmcloud:nats` from Go rather than `wasmcloud:messaging`.

It is the Go counterpart of the Rust `nats-jetstream-replay` example in the
wasmCloud repo, and shows what the portable messaging interface cannot
express: durable delivery with acknowledgement, redelivery, and
compare-and-swap on a KV revision.

## What it demonstrates

| Behaviour | Where |
| --- | --- |
| At-least-once delivery, and how to actually be idempotent | `accumulate` |
| Auto-ack: returning nil acks, returning an error naks | `handleOrder` |
| Dropping a poison message rather than retrying it forever | the malformed-body branch |
| CAS on a KV revision, retried via `errors.As` on `*nats.RevisionMismatchError` | `accumulate` |
| Publish deduplication via `nats.MsgIDHeader` | the notification publish |
| Typed Go errors instead of string matching | `nats/errors.go` in the SDK |

Idempotency is the part worth reading. Delivery is at-least-once, so a bare
`total += amount` double-counts every redelivery. The stored value is
`total@last-applied-sequence`, and a sequence already counted is skipped.

## Prerequisites

The stream and bucket are not created by the component — that lifecycle is
deliberately outside `wasmcloud:nats`, so a workload cannot provision storage
it was not granted:

```bash
nats-server -js &

nats stream add ORDERS --subjects 'orders.received' \
  --storage file --retention limits --discard old \
  --max-msgs=-1 --max-bytes=-1 --max-age=24h \
  --dupe-window=2m --replicas 1 --defaults

nats stream add PROCESSED --subjects 'orders.processed' \
  --storage file --dupe-window=5m --replicas 1 --defaults

nats kv add order-totals --history 5
```

The `--dupe-window` on `PROCESSED` is what makes `nats.MsgIDHeader` do
anything; without it a redelivery publishes a second notification.

## Build

```bash
componentize-go -w 'wasmcloud:examples/nats-jetstream-orders@0.1.0' build
```

The `-w` flag is required: this component serves no HTTP, so it does not use
the SDK's default `wasip2` world (which mandates a
`wasi:http/incoming-handler` export) and declares its own world in
`wit/world.wit` instead.

## Try it

```bash
nats pub orders.received "order-1:100"
nats pub orders.received "order-1:50"

nats kv get order-totals order-1   # -> 150@2  (total@last-applied-sequence)
nats sub orders.processed          # -> order-1:100, then order-1:150

nats pub orders.received "not-an-order"   # dropped, not retried
```
