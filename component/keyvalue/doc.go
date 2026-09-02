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
//
// # Labeled bindings, and why not wasi:keyvalue
//
// This package wraps `wasmcloud:keyvalue`, and it is deliberate: the host's
// `wasmcloud:keyvalue` plugin binds both routes. A component that imports the
// interface plainly — which is what componentize-go emits for a plain world
// import, and what this package's committed bindings name — is routed to the
// workload's default backend, the unnamed hostInterfaces entry, `backend:`
// and all. An unregistered `backend:` name fails loudly at bind.
//
// `wasi:keyvalue@0.2.0-draft` is not a drop-in swap for reaching a second
// backend. Its multiplexer links `(implements ..)` instances only, so an
// unnamed entry is not served by it at all: the plain import falls through to
// the standalone NATS-backed `wasi:keyvalue` plugin, which has no `backend:`
// key and ignores the one the entry carries — the data lands in a JetStream
// KV bucket with nothing logged. Probe for it by setting
// `backend: bogus-backend`; if the workload still deploys and serves traffic,
// the key is being discarded.
//
// A `name:` on the entry has the same consequence here as everywhere else: it
// selects the `(implements <label>)` instance, leaving the plain instance
// this package's committed bindings ask for unbound, and the workload fails
// to start with "a matching implementation was not found in the linker".
// //go:wasmimport takes its instance name as a compile-time literal, so
// reaching a labeled instance means generating bindings for a world that
// imports `store` under that label:
//
//	world app {
//	  include wasmcloud:component-go/wasip3@0.2.0;
//	  import wasmcloud:keyvalue/types@0.2.0;
//	  import wasmcloud:keyvalue/atomics@0.2.0;
//	  import tenant-a: wasmcloud:keyvalue/store@0.2.0;
//	}
//
// Generate them with this module's import path as --pkg-name
// (`--pkg-name go.wasmcloud.dev/component/imports`) so `types` — where the
// `bucket` resource lives — resolves to this module's committed package
// instead of being regenerated. componentize-go emits one package per labeled
// instance, named for the label, and it satisfies [Store] directly:
//
//	type tenantAStore struct{}
//
//	func (tenantAStore) Open(id string) witTypes.Result[*types.Bucket, types.Error] {
//	  return tenanta.Open(id)
//	}
//
//	bucket, err := keyvalue.OpenFrom(tenantAStore{}, "sessions")
//
// Only `store.open` is routed by label. Atomics, cas, batch, and the bucket's
// own methods are bound standalone and operate on a bucket that already
// carries the backend it was opened through, so the [Bucket] that comes back
// is an ordinary one and the rest of this package works on it unchanged.
package keyvalue
