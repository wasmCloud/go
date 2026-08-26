# NATS request-reply (Go)

Two components that make up one round trip: `service` answers requests on
`service.*`, and `gateway` turns an HTTP request into a NATS request and the
reply back into an HTTP response.

This is the core NATS half of `wasmcloud:nats` — no JetStream, no
durability, nothing stored. A request either gets an answer inside its
timeout or it does not, and nothing is retried for you.

## What it demonstrates

| Behaviour | Where |
| --- | --- |
| Answering a request by publishing to its reply subject | `reply` in `service` |
| Reporting failure *to the caller*, because core NATS never retries | the `Nats-Service-Error` headers |
| One `service.*` subscription serving several endpoints | `dispatch` |
| Queue groups load-balancing across replicas | `core-subscriptions` in `deployment.yaml` |
| Issuing a request with a timeout | `nats.Request` in `gateway` |
| Mapping NATS failures onto HTTP status codes, by type | `statusFor` |
| Recovering from `MaxPayloadExceededError` using the limit it carries | `reply` in `service` |

The error handling is the part worth reading. `nats.Request` fails in four
ways that mean genuinely different things — nobody is listening
(`ErrNoResponders`, 503), the subject is outside the workload's grant
(`SubjectDeniedError`, 403), the payload is too big
(`MaxPayloadExceededError`, 413), and nobody answered in time (504) — and
the SDK reports each as a distinct Go error rather than a string to match
on. Only the first three are safe to retry blindly; a timeout leaves the
request in an unknown state, because the responder may well have done the
work.

## Prerequisites

The server and the CLI are separate packages, and the examples use
both — `brew install nats-server nats` on macOS.

Just a NATS server — no JetStream, no streams, no buckets:

```bash
nats-server &
```

## Build

```bash
cd service && componentize-go -w 'wasmcloud:examples/nats-reply-service@0.1.0' build
cd ../gateway && GOFLAGS=-tags=componentizego_async componentize-go \
  -w 'wasmcloud:component-go/wasip3@0.2.0' \
  -w 'wasmcloud:examples/nats-request-gateway@0.1.0' build
```

Both are **WASI P3** components: every `wasmcloud:nats` function is an `async
func`. The service serves no HTTP, so it declares its own world and needs only
that one. The gateway does serve HTTP, so it names the SDK's `wasip3` world
(`wasi:http/handler@0.3.0`) alongside its own — the default `wasip2` world
would put it on P2 `wasi:http` and fail to link. The `GOFLAGS` prefix selects
`wasihttp`'s async P3 implementation; componentize-go sets that tag on its own
from the release after v0.4.1.

## Running it

Two components, two dev sessions — one per directory, in two terminals:

```bash
cd service && wash dev     # binds 8001; serves no HTTP, the port just avoids a clash
cd gateway && wash dev     # binds 8000; this is the one you curl
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

The two are not quite interchangeable, and the difference is worth knowing
before a component that works in dev fails on deploy: dev derives host
interfaces from the component's imports, so `wasi:logging` binds there
whether or not you declare it. A real host binds only what the manifest
names, which is why `deployment.yaml` lists it and `.wash/config.yaml`
does not.

## Try it

Against the service directly, with the NATS CLI as the requester:

```bash
nats req service.greet 'ada'     # -> hello, ada
nats req service.greet ''        # -> hello, world
nats req service.upper 'shout'   # -> SHOUT
nats req service.upper ''        # -> Nats-Service-Error: empty body (code 400)
nats req service.nope 'x'        # -> Nats-Service-Error: no endpoint named nope (code 404)
```

The CLI publishes its inbox under the default `_INBOX.` prefix, and nothing
has to grant it: the host authorizes the service's reply to whatever
`reply-to` it just delivered, once, whoever the requester was. So `nats req`
reaches the service without widening anyone's `subject-allow`.

And through the gateway, where the path becomes the subject:

```bash
curl -X POST --data 'ada' localhost:8000/ask/service/greet     # -> hello, ada
curl -i -X POST --data '' localhost:8000/ask/service/upper     # -> 400, from the service
curl -i -X POST --data 'x' localhost:8000/ask/nothing/here     # -> 403, refused by the grant
```

Stop the service and try again to see the difference between "nobody is
listening" and "nobody answered":

```bash
curl -i -X POST --data 'ada' localhost:8000/ask/service/greet  # -> 503, immediately
```

## What the grant has to cover

Deny-by-default bites in more places than it first looks, and each one fails
differently:

- **The subscription pattern.** `core-subscriptions: service.*` is checked
  against the host's `subject-allow` *before the host subscribes*, so a
  service whose host granted it nothing under `service.` never starts — the
  workload fails with `core subscription 'service.*' is outside this
  workload's subject grant`. Being handed messages on a subject is as much a
  permission as publishing to it.
- **The requester's subjects.** `service.>`, and nothing else. A subject
  outside it fails host-side with subject-denied, which the gateway maps to
  a 403 without the request reaching the server.

The reply is the interesting one, because it needs no grant at all. A
workload's inbox prefix is derived from its workload *id* — a UUID assigned
at schedule time — so no responder's grant could name it in advance. Rather
than making responders ask for `_INBOX.>`, which would let them read and
answer every other reply on the connection, the host authorizes the reply
itself: when it hands a request to the guest it records that one `reply-to`
as publishable, single use, for 30 seconds. Answer it once and the grant is
spent.

That is also why neither workload pins `inbox-prefix`. It is a connection
key, so under `deny` only the host could set one — and the host's declaration
covers every workload on the group, which would hand the gateway and the
service the same prefix and let them race for each other's replies. Left
alone, each derives its own, which is the isolation the default exists to
provide.

## Why two workloads

The gateway issues requests and the service answers them, and running both
in one workload would mean a component answering itself. With a bounded
`maxConcurrency`, an instance blocked inside `nats.Request` is an instance
not available to serve the subscription the request is waiting on — so the
call deadlocks until it times out. Separate workloads, separate grants, and
neither can reach the other's inbox prefix.

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
          # Both workloads want the same thing, so one unnamed
          # binding serves the gateway and the service alike. No
          # `_INBOX` grant: the host authorizes each reply itself.
          subject-allow: service.>
        # NATS credentials reach the host this way rather than through a
        # workload manifest or a CLI arg — the rendered `wash host` config
        # file never appears in `kubectl describe pod`.
        secretFrom:
          - service-nats-creds
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
wash oci push ghcr.io/<your-org>/nats-reply-service:0.1.0 service/build/nats_reply_service.wasm
wash oci push ghcr.io/<your-org>/nats-request-gateway:0.1.0 gateway/build/nats_request_gateway.wasm
kubectl apply -f deployment.yaml
```

See [deployment.yaml](./deployment.yaml) for the `WorkloadDeployment`
definition and the wasmCloud [workload deployment
quickstart](https://wasmcloud.com/docs/quickstart/deploy-a-webassembly-workload/)
for cluster setup.
