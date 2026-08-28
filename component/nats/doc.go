// Package nats provides NATS-native messaging for wasmCloud components: core
// pub/sub and request-reply, JetStream, and JetStream KV.
//
// It wraps `wasmcloud:nats@0.1.0`, whose every function is an `async func`.
// That makes it a WASI P3 interface: lifting a sync-signature function with
// the async canonical ABI fails component validation, so there is no sync
// revision and a component that imports it targets P3. The Go API is
// unaffected in shape — componentize-go drives the async ABI underneath, so
// calls still read as ordinary blocking Go.
//
// Use it over
// [go.wasmcloud.dev/component/messaging] when you need what the portable
// messaging interface cannot express: durable delivery with explicit
// acknowledgement and redelivery, replay from an arbitrary stream position,
// compare-and-swap on a KV revision, and publish deduplication. Reach for
// `messaging` instead when portability across brokers matters more.
//
// # Declaring the capability
//
// A component opts in by importing the interfaces in its own world. Which
// world it builds on depends on whether it also serves HTTP.
//
// A component that serves no HTTP — which is every pure NATS component —
// must not include `wasip2` or `wasip3`: both worlds mandate an HTTP handler
// export (`wasi:http/incoming-handler@0.2.8` and `wasi:http/service@0.3.0`
// respectively), so a component that never wrote one fails to build with
//
//	failed to decode world from module
//	  module was not valid
//	  failed to find export of interface `wasi:http/incoming-handler@0.2.8` function `handle`
//
// naming an export it never asked for. Include the CLI imports directly
// instead:
//
//	world app {
//	  include wasi:cli/imports@0.2.8;
//
//	  import wasi:logging/logging@0.1.0-draft;
//	  import wasmcloud:nats/types@0.1.0;
//	  import wasmcloud:nats/jetstream@0.1.0;
//	  import wasmcloud:nats/kv@0.1.0;
//	}
//
// The SDK ships that shape as a world of its own — `headless` — for a project
// that resolves the SDK's wit/ directory:
//
//	world app {
//	  include wasmcloud:component-go/headless@0.2.0;
//	  import wasmcloud:nats/types@0.1.0;
//	}
//
// A component that *does* serve HTTP as well as NATS includes the HTTP world
// it wants and adds the NATS imports to it:
//
//	world app {
//	  include wasmcloud:component-go/wasip3@0.2.0;
//	  import wasmcloud:nats/types@0.1.0;
//	  import wasmcloud:nats/jetstream@0.1.0;
//	  import wasmcloud:nats/kv@0.1.0;
//	}
//
// In both cases the workload manifest must declare the matching
// hostInterfaces entry.
//
// # Timers
//
// Do not call [time.Sleep] (or anything that parks a goroutine on a Go
// runtime timer — [time.After], tickers, timer-based context deadlines)
// inside a handler: an async-lifted export whose goroutine parks on the Go
// runtime traps with "async-lifted export failed to produce a result", on
// every delivery. Use [go.wasmcloud.dev/component/sleep.Sleep] instead, which
// awaits the host's clock through the same async-import mechanism every
// wasmcloud:nats call uses. The `wasmcloud:nats-guest@0.1.0` worlds import
// the clock it needs already.
//
// # Generating bindings
//
// Use componentize-go, not wit-bindgen-go. Every function in
// `wasmcloud:nats@0.1.0` is an `async func`, and wit-bindgen-go does not
// support the component-model async ABI: it exits 0 and generates a package
// with all 30 async functions silently missing.
//
// # Grants
//
// Access is deny-by-default and the grants are separate on purpose:
// `subject-allow` covers publish and request, `stream-allow` covers stream
// reads, and `bucket-allow` covers KV. Permission to publish to a subject
// does not carry permission to read the stream capturing it. A call outside
// its grant fails host-side with a [DeniedError] and never reaches the
// server. `stream-allow` alone reaches nothing readable: every stored message
// is checked against `subject-allow` too, so grant the subjects a stream
// stores alongside the stream.
//
// Where those grants are *written* is the host's business, not the
// component's. A `wash host` declares each binding under
// `host.wasmcloudNats` — the servers, the credentials, and the three grants —
// and refuses a manifest that tries to set any of them. What a workload
// declares is only what it wants delivered:
//
//	hostInterfaces:
//	  - namespace: wasmcloud
//	    package: nats
//	    version: "0.1.0"
//	    # `(implements orders)`: the binding the host declared.
//	    name: orders
//	    interfaces: [types, jetstream, kv, jetstream-handler]
//	    config:
//	      subscriptions: ORDERS:orders.received:all
//	      ack-mode: auto
//
// Under `wash dev`, which runs `--wasmcloud-nats-workload-config=allow`, the
// same manifest may carry `servers` and the grants itself, so a project stays
// runnable on its own.
//
// Credentials never appear in a manifest. The host merges
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
