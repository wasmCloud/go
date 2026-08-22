# JetStream KV watch (Go)

Watches a feature-flag bucket, recomputes the set of enabled flags into a
single derived key, and announces every change on core NATS.

A watcher is not a queue consumer, and the two differences drive the whole
design: there is no acknowledgement and no replay, and writing to the bucket
you are watching feeds your own writes back to you.

## What it demonstrates

| Behaviour | Where |
| --- | --- |
| Receiving KV change events pushed by the host | `handleChange` |
| Rebuilding derived state instead of applying a delta | `rebuildActive` |
| Keeping the derived key outside the watched prefix, so writes do not loop | `activeKey` and `kv-watches` |
| Telling put from delete from purge | `entry.Operation` |
| Compare-and-swap, retried via `errors.As` on `*nats.RevisionMismatchError` | `storeActive` |
| Skipping a write that would not change anything | `storeActive` |
| A denied subject named as the deployment problem it is | the announce branch |

The rebuild is the part worth reading. Watch events carry no
acknowledgement: returning an error from the handler is logged by the host
and the event is gone — there is no redelivery and no dead-letter. A handler
that applied `entry.Value` as a delta would be permanently wrong after one
dropped event. Reading the bucket instead means the derived value is
correct whether the handler ran once, twice, or missed the event before it.

Which also means the event's own value is never used here. The event says
*something changed*; the bucket says *what is true now*, and it may already
be newer than the event that woke the handler. That is fine, and preferring
the newer answer is exactly what you want.

## The write loop

The component writes `active` into the same bucket it watches. Nothing
detects that after the fact — a watcher that re-triggers itself just keeps
going — so it is prevented twice, before it can start:

- The host's `kv-watches` filter is `feature-flags:flag.>`, and the derived
  key is `active`, so the write is never delivered.
- `handleChange` ignores any key outside `flag.` anyway, in case someone
  widens the filter later.

Keeping the derived value in a separate bucket avoids the question
entirely, at the cost of a second bucket grant. Either is defensible; what
is not defensible is a derived key inside the watched prefix.

## Prerequisites

The bucket is not created by the component — that lifecycle is deliberately
outside `wasmcloud:nats`, so a workload cannot provision storage it was not
granted:

```bash
nats-server -js &

nats kv add feature-flags --history 5
```

## Build

```bash
componentize-go -w 'wasmcloud:examples/nats-kv-watch@0.1.0' build
```

The `-w` flag is required: this component serves no HTTP, so it does not use
the SDK's default `wasip2` world (which mandates a
`wasi:http/incoming-handler` export) and declares its own world in
`wit/world.wit` instead.

## Try it

```bash
nats sub flags.changed &

nats kv put feature-flags flag.dark-mode on
nats kv put feature-flags flag.beta-search off
nats kv get feature-flags active     # -> dark-mode

nats kv put feature-flags flag.beta-search true
nats kv get feature-flags active     # -> beta-search,dark-mode

nats kv del feature-flags flag.dark-mode
nats kv get feature-flags active     # -> beta-search
```

`active` stays sorted, so two rebuilds of the same state produce the same
bytes and an unrelated flag change does not burn a revision on it.

## Running it

`wash dev` builds and runs it against the NATS server above:

```bash
wash dev
```

`wasmcloud:nats` is a host plugin, and `wash dev` has no manifest to read
the binding from — so `.wash/config.yaml` carries the same
`hostInterfaces` shape `deployment.yaml` uses. Change a grant in one and
change it in the other.

The two are not quite interchangeable, and the difference is worth knowing
before a component that works in dev fails on deploy: dev derives host
interfaces from the component's imports, so `wasi:logging` binds there
whether or not you declare it. A real host binds only what the manifest
names, which is why `deployment.yaml` lists it and `.wash/config.yaml`
does not.

For a real host, `componentize-go build` and deploy `deployment.yaml` to a
wasmCloud v2 host whose runtime carries the `wasmcloud:nats` plugin.
