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

Just a NATS server — no JetStream, no streams, no buckets:

```bash
nats-server &
```

## Build

```bash
cd service && componentize-go -w 'wasmcloud:examples/nats-reply-service@0.1.0' build
cd ../gateway && componentize-go build
```

The service needs `-w`: it serves no HTTP, so it does not use the SDK's
default `wasip2` world (which mandates a `wasi:http/incoming-handler`
export) and declares its own world in `wit/world.wit` instead. The gateway
does serve HTTP, so its world is merged on top of the SDK's and no flag is
needed.

## Try it

Against the service directly, with the NATS CLI as the requester:

```bash
nats req service.greet 'ada'     # -> hello, ada
nats req service.greet ''        # -> hello, world
nats req service.upper 'shout'   # -> SHOUT
nats req service.upper ''        # -> Nats-Service-Error: empty body (code 400)
nats req service.nope 'x'        # -> Nats-Service-Error: no endpoint named nope (code 404)
```

The CLI publishes its inbox under the default `_INBOX.` prefix, which is why
`subject-allow` in `deployment.yaml` grants `_INBOX.>` alongside the
gateway's own prefix. Drop that entry once you are past experimenting: a
service that can publish to every inbox on the server can answer requests
that were never addressed to it.

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

## Three things the grant has to cover

Deny-by-default bites in more places than it first looks, and each one fails
differently:

- **The subscription pattern.** `core-subscriptions: service.*` is checked
  against `subject-allow` *before the host subscribes*, so a service granted
  only its reply inbox never starts — the workload fails with `core
  subscription 'service.*' is outside this workload's subject grant`. Being
  handed messages on a subject is as much a permission as publishing to it.
- **The reply.** Replies are publishes like any other. Deny this one and the
  handler runs, does the work, and the answer is dropped host-side — the
  requester sees only a timeout, which is the most confusing failure of the
  three.
- **The requester's subjects.** `service.>` on the gateway, and nothing
  else.

The reply grant has a wrinkle worth knowing. A workload's inbox prefix
defaults to one derived from its workload *id*, which is a UUID assigned at
schedule time — nothing a responder's grant can name in advance. So the
gateway pins `inbox-prefix: _INBOX_gateway` and the service grants exactly
that. Keep it unique per workload: a prefix two workloads share lets them
race to consume each other's replies, which is the isolation the derived
default exists to provide.

## Why two workloads

The gateway issues requests and the service answers them, and running both
in one workload would mean a component answering itself. With a bounded
`maxConcurrency`, an instance blocked inside `nats.Request` is an instance
not available to serve the subscription the request is waiting on — so the
call deadlocks until it times out. Separate workloads, separate grants, and
neither can reach the other's inbox prefix.

## Running it

`wasmcloud:nats` is served by a wasmCloud host plugin, not by `wash dev` —
the dev host registers `wasmcloud:messaging` and `wasi:keyvalue` but not
this one yet. So the loop is `componentize-go build`, then deploy
`deployment.yaml` to a wasmCloud v2 host whose runtime carries the
`wasmcloud:nats` plugin, pointed at the NATS server above.
