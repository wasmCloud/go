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

Every `wasmcloud:nats` function is an `async func`, so this builds a **WASI
P3** component. componentize-go notices the async world and fetches a patched
Go toolchain for it on first use; the Go code is unaffected — calls still read
as ordinary blocking Go.

## Running it

`wash dev` builds and runs it against the NATS server above:

```bash
wash dev
```

`wasmcloud:nats` is a host plugin, and `wash dev` has no manifest to read
the binding from — so `.wash/config.yaml` carries the binding whole:
servers, grants, and asks together. It is deliberately *not* a mirror of
`deployment.yaml` any more. Dev runs
`--wasmcloud-nats-workload-config=allow` precisely so a checkout is
runnable on its own, while a real host runs `deny` and keeps the servers,
the credentials, and the grants on its side. So a grant lives in this file
for dev and in the host group's `wasmcloudNats` for a cluster — change it
in both.

## Try it

```bash
nats pub orders.received "order-1:100"
nats pub orders.received "order-1:50"

nats kv get order-totals order-1   # -> 150@2  (total@last-applied-sequence)
nats sub orders.processed          # -> order-1:100, then order-1:150

nats pub orders.received "not-an-order"   # dropped, not retried
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
          subject-allow: orders.processed,orders.received
          stream-allow: ORDERS,PROCESSED
          bucket-allow: order-totals
        # NATS credentials reach the host this way rather than through a
        # workload manifest or a CLI arg — the rendered `wash host` config
        # file never appears in `kubectl describe pod`.
        secretFrom:
          - orders-nats-creds
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
wash oci push ghcr.io/<your-org>/nats-jetstream-orders:0.1.0 build/nats_jetstream_orders.wasm
kubectl apply -f deployment.yaml
```

See [deployment.yaml](./deployment.yaml) for the `WorkloadDeployment`
definition and the wasmCloud [workload deployment
quickstart](https://wasmcloud.com/docs/quickstart/deploy-a-webassembly-workload/)
for cluster setup.
