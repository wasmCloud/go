// Package nats provides NATS-native messaging for wasmCloud components:
// core pub/sub and request-reply, JetStream, and JetStream KV.
//
// It wraps `wasmcloud:nats@0.1.0`. Use it over
// [go.wasmcloud.dev/component/messaging] when you need what the portable
// messaging interface cannot express: durable delivery with explicit
// acknowledgement and redelivery, replay from an arbitrary stream position,
// compare-and-swap on a KV revision, and publish deduplication. Reach for
// `messaging` instead when portability across brokers matters more.
//
// # Declaring the capability
//
// A component opts in by importing the interfaces in its own world:
//
//	world app {
//	  include wasmcloud:component-go/wasip2@0.2.0;
//	  import wasmcloud:nats/types@0.1.0;
//	  import wasmcloud:nats/jetstream@0.1.0;
//	  import wasmcloud:nats/kv@0.1.0;
//	}
//
// and the workload manifest must declare the matching hostInterfaces entry.
//
// # Grants
//
// Access is deny-by-default and the grants are separate on purpose:
// `subject-allow` covers publish and request, `stream-allow` covers stream
// reads, and `bucket-allow` covers KV. Permission to publish to a subject
// does not carry permission to read the stream capturing it. A call outside
// its grant fails host-side with a [SubjectDeniedError] and never reaches
// the server.
//
//	hostInterfaces:
//	  - namespace: wasmcloud
//	    package: nats
//	    version: "0.1.0"
//	    interfaces: [types, jetstream, kv, jetstream-handler]
//	    config:
//	      servers: nats://nats.default.svc:4222
//	      subject-allow: orders.processed
//	      stream-allow: ORDERS,PROCESSED
//	      bucket-allow: order-totals
//	      subscriptions: ORDERS:orders.received:all
//	    secretFrom:
//	      - nats-credentials
//
// Credentials never appear in `config`. The host merges
// `config` → `configFrom` → `secretFrom` (later wins) before the plugin sees
// them, and an nkey seed is signed host-side without crossing into the
// component.
//
// # Receiving messages
//
// Publishing and KV work through this package directly. To be *given*
// messages, export a handler interface and register a callback from one of
// the subpackages — [go.wasmcloud.dev/component/nats/corehandler],
// [go.wasmcloud.dev/component/nats/jetstreamhandler], or
// [go.wasmcloud.dev/component/nats/kvhandler]. They are separate packages
// because importing one links its wasm export into the component; a single
// package would force every component to export all three.
package nats
