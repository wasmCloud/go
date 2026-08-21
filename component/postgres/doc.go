// Package postgres is an idiomatic Go wrapper for the wasmCloud
// wasmcloud:postgres@0.2.0 host-plugin interfaces (`types`, `query`, and
// `prepared`, vendored from wasmCloud v2.6.1 under
// wit/deps/wasmcloud-postgres-0.2.0).
//
// [Query] runs a parameterized statement and streams result rows back
// incrementally through a [Rows] iterator; [QueryBatch] runs multi-statement
// (parameter-free) batches such as migrations; [Prepare] creates a
// [PreparedStatement] that can be executed repeatedly with
// [PreparedStatement.Exec].
//
// Values are the wasmcloud:postgres `pg-value` variant, re-exported as
// [Value]; helpers such as [Text], [Int64], [Float64], [JSON], and [UUID]
// build common variants, and everything else is reachable through the
// generated Make* constructors in the imports/wasmcloud_postgres_0_2_0_types
// package.
//
// # Opting in
//
// The capability is not part of the SDK's default worlds. The app's own world
// must import the interfaces:
//
//	world app {
//	  include wasmcloud:component-go/wasip3@0.2.0;
//	  import wasmcloud:postgres/query@0.2.0;
//	  import wasmcloud:postgres/prepared@0.2.0;
//	}
//
// and the workload manifest must declare the matching hostInterfaces entry
// (namespace `wasmcloud`, package `postgres`) so the host binds a postgres
// plugin to the component.
//
// The interfaces are `async func`s, and `query` returns component-model
// stream/future handles: componentize-go builds a world that imports them
// with its async support enabled automatically (the componentizego_async
// build tag and patched Go toolchain). The package itself also compiles
// under default build tags.
package postgres
