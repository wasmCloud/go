// Package handler wires up the optional wasmcloud:messaging/handler@0.2.0
// export so a component can receive messages from a wasmCloud messaging
// plugin.
//
// Importing this package links the `handle-message` wasm export into the
// component, so import it only from apps whose world exports the interface:
//
//	world app {
//	  include wasmcloud:component-go/wasip3@0.2.0;
//	  import wasmcloud:messaging/consumer@0.2.0;
//	  export wasmcloud:messaging/handler@0.2.0;
//	}
//
// Register a callback during init:
//
//	func init() {
//	  handler.HandleFunc(func(msg messaging.BrokerMessage) error {
//	    // ...
//	    return nil
//	  })
//	}
//
// The workload manifest must declare the matching hostInterfaces entry
// (namespace `wasmcloud`, package `messaging`).
package handler

import (
	witTypes "go.bytecodealliance.org/pkg/wit/types"
	export "go.wasmcloud.dev/component/exports/wasmcloud_component_go_wasmcloud_messaging_handler_0_2_0/export_wasmcloud_messaging_0_2_0_handler"
	types "go.wasmcloud.dev/component/imports/wasmcloud_messaging_0_2_0_types"
	"go.wasmcloud.dev/component/messaging"

	// Pull in the //go:wasmexport glue for the handle-message export.
	_ "go.wasmcloud.dev/component/exports/wasmcloud_component_go_wasmcloud_messaging_handler_0_2_0/wit_exports"
)

// HandleFunc registers fn as the callback invoked for each message delivered
// to the component's wasmcloud:messaging/handler export. A non-nil error
// returned by fn is reported to the host as the handler's failure string.
//
// HandleFunc should be called once, during program initialization (an init
// function or the start of main), before any message is delivered.
func HandleFunc(fn func(msg messaging.BrokerMessage) error) {
	export.Exports.HandleMessage = func(msg types.BrokerMessage) witTypes.Result[witTypes.Unit, string] {
		if err := fn(messaging.FromWit(msg)); err != nil {
			return witTypes.Err[witTypes.Unit, string](err.Error())
		}
		return witTypes.Ok[witTypes.Unit, string](witTypes.Unit{})
	}
}
