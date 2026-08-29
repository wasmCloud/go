// Package messaging is an idiomatic Go wrapper for the wasmCloud
// wasmcloud:messaging@0.3.0 host-plugin interfaces (`types` and `consumer`,
// vendored from wasmCloud v2.8.0 under wit/deps/wasmcloud-messaging-0.3.0).
//
// [Publish] sends a message to a subject without awaiting a response;
// [Request] performs a request/reply round trip. To receive messages, see
// the messaging/handler subpackage, which wires up the optional
// wasmcloud:messaging/handler export.
//
// # Opting in
//
// The capability is not part of the SDK's default worlds. The app's own world
// must import the consumer interface:
//
//	world app {
//	  include wasmcloud:component-go/wasip3@0.2.0;
//	  import wasmcloud:messaging/consumer@0.3.0;
//	}
//
// and the workload manifest must declare the matching hostInterfaces entry
// (namespace `wasmcloud`, package `messaging`) so the host binds a messaging
// plugin to the component.
//
// # Async and streaming
//
// Every operation is an `async func` as of 0.3.0, and message bodies are
// stream handles rather than one buffered `list<u8>`, so a component can keep
// many publishes and requests in flight and a payload can flow incrementally.
// A component importing this package therefore targets WASI P3, which is why
// the world above includes wasip3 rather than wasip2.
//
// [BrokerMessage.Body] is an io.Reader on the way out and a stream from the
// host on the way in. Publishing pumps the body concurrently with the host
// call, which relies on the async runtime componentize-go enables for this
// world; a received body must be drained or closed, which the messaging/handler
// subpackage does for you once the callback returns.
package messaging
