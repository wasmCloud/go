// Package messaging is an idiomatic Go wrapper for the wasmCloud
// wasmcloud:messaging@0.2.0 host-plugin interfaces (`types` and `consumer`,
// vendored from wasmCloud v2.6.1 under wit/deps/wasmcloud-messaging-0.2.0).
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
//	  import wasmcloud:messaging/consumer@0.2.0;
//	}
//
// and the workload manifest must declare the matching hostInterfaces entry
// (namespace `wasmcloud`, package `messaging`) so the host binds a messaging
// plugin to the component.
package messaging
