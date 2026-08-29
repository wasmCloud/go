// Package keyvalue is an idiomatic Go wrapper for the wasmCloud
// wasmcloud:keyvalue@0.2.0 host-plugin interfaces (`types`, `store`,
// `atomics`, `cas`, and `batch`, vendored from wasmCloud v2.8.0 under
// wit/deps/wasmcloud-keyvalue-0.2.0). The `watcher` export and
// `watch-service` world are not wrapped.
//
// [Open] returns a [Bucket] whose methods cover single-key CRUD (Get, Set,
// Delete, Exists, ListKeys), atomic increment, compare-and-swap (Current,
// Swap), and batch operations (GetMany, SetMany, DeleteMany).
//
// # Opting in
//
// The capability is not part of the SDK's default worlds. The app's own world
// must import the interfaces it uses (store is required; atomics, cas, and
// batch are separately grantable):
//
//	world app {
//	  include wasmcloud:component-go/wasip3@0.2.0;
//	  import wasmcloud:keyvalue/store@0.2.0;
//	  import wasmcloud:keyvalue/atomics@0.2.0;
//	  import wasmcloud:keyvalue/cas@0.2.0;
//	  import wasmcloud:keyvalue/batch@0.2.0;
//	}
//
// and the workload manifest must declare the matching hostInterfaces entry
// (namespace `wasmcloud`, package `keyvalue`) so the host binds a keyvalue
// plugin to the component.
//
// The interfaces are `async func`s: componentize-go builds a world that
// imports them with its async support enabled automatically (the
// componentizego_async build tag and patched Go toolchain). The package
// itself also compiles under default build tags.
package keyvalue
