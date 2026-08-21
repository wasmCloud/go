// Package secrets is an idiomatic Go wrapper for the wasmCloud
// wasmcloud:secrets@2.0.0 host-plugin interfaces (`store` and `reveal`,
// vendored from wasmCloud v2.6.1 under wit/deps/wasmcloud-secrets-2.0.0).
//
// [Get] looks up a secret and returns an opaque [Secret] handle; the value is
// not transferred until [Secret.Reveal] is called, so a host can gate (and
// audit) reveal independently of lookup.
//
// # Opting in
//
// The capability is not part of the SDK's default worlds. The app's own world
// must import the interfaces:
//
//	world app {
//	  include wasmcloud:component-go/wasip3@0.2.0;
//	  import wasmcloud:secrets/store@2.0.0;
//	  import wasmcloud:secrets/reveal@2.0.0;
//	}
//
// and the workload manifest must declare the matching hostInterfaces entry
// (namespace `wasmcloud`, package `secrets`) so the host binds a secrets
// plugin to the component.
//
// The interfaces are `async func`s: componentize-go builds a world that
// imports them with its async support enabled automatically (the
// componentizego_async build tag and patched Go toolchain). The package
// itself also compiles under default build tags.
package secrets
