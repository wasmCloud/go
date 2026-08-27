// Package kvhandler wires up the optional wasmcloud:nats/kv-handler@0.1.0
// export so a component receives KV change events pushed by the host.
//
// Importing this package links the `handle-event` wasm export into the
// component, so import it only from apps whose world exports the interface:
//
//	world app {
//	  include wasmcloud:component-go/wasip3@0.2.0;
//	  import wasmcloud:nats/types@0.1.0;
//	  import wasmcloud:nats/kv@0.1.0;
//	  export wasmcloud:nats/kv-handler@0.1.0;
//	}
//
// Register a callback during init:
//
//	func init() {
//	  kvhandler.HandleFunc(func(bucket string, e nats.Entry) error {
//	    if e.Operation == nats.OperationDelete {
//	      return evict(e.Key)
//	    }
//	    return apply(e.Key, e.Value)
//	  })
//	}
//
// The workload manifest must declare a matching hostInterfaces entry
// (namespace `wasmcloud`, package `nats`) whose `kv-watches` config names the
// bucket and key filter, and whose `bucket-allow` grant covers the bucket.
// Watch events carry no acknowledgement: a returned error is logged by the
// host and the event is not replayed.
package kvhandler

import (
	witTypes "go.bytecodealliance.org/pkg/wit/types"
	export "go.wasmcloud.dev/component/exports/wasmcloud_component_go_wasmcloud_nats_kv_handler_0_2_0/export_wasmcloud_nats_0_1_0_kv_handler"
	kv "go.wasmcloud.dev/component/imports/wasmcloud_nats_0_1_0_kv"
	"go.wasmcloud.dev/component/nats"

	// Pull in the //go:wasmexport glue for the handle-event export.
	_ "go.wasmcloud.dev/component/exports/wasmcloud_component_go_wasmcloud_nats_kv_handler_0_2_0/wit_exports"
)

// HandleFunc registers fn as the callback invoked for each KV change event
// delivered to the component. A non-nil error is reported to the host as the
// handler's failure string.
func HandleFunc(fn func(bucket string, entry nats.Entry) error) {
	export.Exports.HandleEvent = func(bucket string, entry kv.Entry) witTypes.Result[witTypes.Unit, string] {
		if err := fn(bucket, nats.FromWitEntry(entry)); err != nil {
			return witTypes.Err[witTypes.Unit, string](err.Error())
		}
		return witTypes.Ok[witTypes.Unit, string](witTypes.Unit{})
	}
}
