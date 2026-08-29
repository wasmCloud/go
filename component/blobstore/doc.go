// Package blobstore is an idiomatic Go wrapper for the wasmCloud
// wasmcloud:blobstore@0.1.0 host-plugin interfaces (`types`, `container`,
// and `blobstore`, vendored from wasmCloud v2.8.0 under
// wit/deps/wasmcloud-blobstore-0.1.0).
//
// Containers are collections of byte objects. [CreateContainer] and
// [GetContainer] return a [Container]; object bodies are native
// component-model `stream<u8>` values, surfaced here as io.ReadCloser
// ([Container.GetData]) and io.Reader ([Container.WriteData]).
//
// # Opting in
//
// The capability is not part of the SDK's default worlds. The app's own world
// must import the interfaces:
//
//	world app {
//	  include wasmcloud:component-go/wasip3@0.2.0;
//	  import wasmcloud:blobstore/blobstore@0.1.0;
//	  import wasmcloud:blobstore/container@0.1.0;
//	}
//
// and the workload manifest must declare the matching hostInterfaces entry
// (namespace `wasmcloud`, package `blobstore`) so the host binds a blobstore
// plugin to the component.
//
// The interfaces are `async func`s and object bodies are stream handles
// pumped concurrently with the call: componentize-go builds a world that
// imports them with its async support enabled automatically (the
// componentizego_async build tag and patched Go toolchain). The package
// itself also compiles under default build tags, but [Container.WriteData]
// in particular relies on the async runtime to drive its body-writing
// goroutine while the host call is in flight.
package blobstore
