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

The server and the CLI are separate packages, and the examples use
both — `brew install nats-server nats` on macOS.

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

Every `wasmcloud:nats` function is an `async func`, so this builds a **WASI
P3** component. componentize-go notices the async world and uses the `go` on
your `PATH` if it already carries the `runtime.wasiOnIdle` patch, downloading
a patched toolchain on first use otherwise; the Go code is unaffected.

## Running it

`wash dev` builds and runs it against the NATS server above:

```bash
wash dev
```

`wasmcloud:nats` is a host plugin, and `wash dev` has no manifest to read
the binding from — so `.wash/config.yaml` carries the binding whole:
servers, grants, and asks together. It is deliberately *not* a mirror of
`deployment.yaml` any more. A `wash dev` plugin entry defaults to
`workloadConfig: allow` precisely so a checkout is runnable on its own,
while a real host defaults to `deny` and keeps the servers, the
credentials, and the grants on its side. So a grant lives in this file for
dev and on the host group's `wasmcloud-nats` plugin entry for a cluster —
change it in both.

The two are not quite interchangeable, and the difference is worth knowing
before a component that works in dev fails on deploy: dev derives host
interfaces from the component's imports, so `wasi:logging` binds there
whether or not you declare it. A real host binds only what the manifest
names, which is why `deployment.yaml` lists it and `.wash/config.yaml`
does not.

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

## Declaring the binding host-side

`wasmcloud:nats` is configured like every other plugin, as a `host.plugins`
entry with `id: wasmcloud-nats`. Under `wash host` that entry defaults to
`workloadConfig: deny`, which splits a binding between two people. The host declares what the
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
      # Omit `servers` to take the group's `wasmcloudNatsUrl` (which falls
      # back to `dataNatsUrl`), so the same manifest runs under `wash dev`
      # and here.
      plugins:
        - id: wasmcloud-nats
          config:
            bucket-allow: feature-flags
            subject-allow: flags.changed
          # NATS credentials reach the host this way rather than through a
          # workload manifest or a CLI arg — the rendered `wash host` config
          # file never appears in `kubectl describe pod`.
          secretFrom:
            - feature-flags-nats-creds
```

`wash dev` defaults the other way round (`workloadConfig: allow`): the
host's declaration is a default a manifest may override, so
`.wash/config.yaml` still carries the whole binding and a checkout stays
runnable on its own.

## Deploy to wasmCloud on Kubernetes

The target is a wasmCloud v2 host whose runtime carries the `wasmcloud:nats`
plugin — the plugin ships with the host, not with the component, so a host
built without it rejects the binding at placement — and whose host group
declares the binding above.

Push the component and apply the manifest:

```shell
wash oci push ghcr.io/<your-org>/nats-kv-watch:0.1.0 build/nats_kv_watch.wasm
kubectl apply -f deployment.yaml
```

See [deployment.yaml](./deployment.yaml) for the `WorkloadDeployment`
definition and the wasmCloud [workload deployment
quickstart](https://wasmcloud.com/docs/quickstart/deploy-a-webassembly-workload/)
for cluster setup.
