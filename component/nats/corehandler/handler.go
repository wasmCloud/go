// Package corehandler wires up the optional wasmcloud:nats/core-handler@0.1.0
// export so a component receives core NATS messages pushed by the host.
//
// Importing this package links the `handle-message` wasm export into the
// component, so import it only from apps whose world exports the interface:
//
//	world app {
//	  include wasmcloud:component-go/headless@0.2.0;
//	  import wasmcloud:nats/types@0.1.0;
//	  export wasmcloud:nats/core-handler@0.1.0;
//	}
//
// Register a callback during init:
//
//	func init() {
//	  corehandler.HandleFunc(func(msg nats.Message) error {
//	    // ...
//	    return nil
//	  })
//	}
//
// The workload manifest must declare a matching hostInterfaces entry
// (namespace `wasmcloud`, package `nats`) whose `core-subscriptions` config
// names the subjects to deliver, and whose `subject-allow` grant covers them.
// Core delivery has no acknowledgement: a returned error is logged by the
// host, and the message is not redelivered.
//
// # Instance reuse
//
// Whether consecutive messages share one component instance is decided by the
// workload manifest (`poolSize` on the component), not by this package. Under
// the default every message gets a fresh instance and package-level state
// resets between messages; with `poolSize` set, state in package-level
// variables survives from one message to the next until the instance is
// recycled (`maxInvocations`). Write handlers that are correct either way:
// treat package-level state as a cache, never as a guarantee of freshness or
// of persistence.
package corehandler

import (
	witTypes "go.bytecodealliance.org/pkg/wit/types"
	export "go.wasmcloud.dev/component/exports/wasmcloud_component_go_wasmcloud_nats_core_handler_0_2_0/export_wasmcloud_nats_0_1_0_core_handler"
	types "go.wasmcloud.dev/component/imports/wasmcloud_nats_0_1_0_types"
	"go.wasmcloud.dev/component/nats"

	// Pull in the //go:wasmexport glue for the handle-message export.
	_ "go.wasmcloud.dev/component/exports/wasmcloud_component_go_wasmcloud_nats_core_handler_0_2_0/wit_exports"
)

// HandleFunc registers fn as the callback invoked for each core NATS message
// delivered to the component. A non-nil error is reported to the host as the
// handler's failure string.
func HandleFunc(fn func(msg nats.Message) error) {
	export.Exports.HandleMessage = func(msg types.NatsMessage) witTypes.Result[witTypes.Unit, string] {
		if err := fn(nats.FromWitMessage(msg)); err != nil {
			return witTypes.Err[witTypes.Unit, string](err.Error())
		}
		return witTypes.Ok[witTypes.Unit, string](witTypes.Unit{})
	}
}
