// Package jetstreamhandler wires up the optional
// wasmcloud:nats/jetstream-handler@0.1.0 export so a component receives
// JetStream deliveries pushed by the host.
//
// Importing this package links the `handle-message` wasm export into the
// component, so import it only from apps whose world exports the interface:
//
//	world app {
//	  include wasmcloud:component-go/wasip3@0.2.0;
//	  import wasmcloud:nats/types@0.1.0;
//	  import wasmcloud:nats/jetstream@0.1.0;
//	  export wasmcloud:nats/jetstream-handler@0.1.0;
//	}
//
// Register a callback during init:
//
//	func init() {
//	  jetstreamhandler.HandleFunc(func(h *nats.MessageHandle) error {
//	    if h.DeliveryCount() > 1 {
//	      // a previous attempt did not ack; this work must be idempotent
//	    }
//	    return process(h.Message())
//	  })
//	}
//
// The workload manifest must declare a matching hostInterfaces entry
// (namespace `wasmcloud`, package `nats`) whose `subscriptions` config names
// the stream and filter, and whose `stream-allow` grant covers the stream.
//
// # Who acknowledges
//
// Under the binding's default `ack-mode: auto` the host settles the message
// from the callback's outcome: returning nil acks, and returning an error
// naks with a backoff that grows by delivery count. Do not settle the handle
// yourself in that mode — it reports an error, because the host owns the
// acknowledgement.
//
// Under `ack-mode: manual` the host settles nothing and the callback must
// call Ack, Nak, or Term on the handle. Returning without settling stalls
// the consumer until ack-wait expires.
package jetstreamhandler

import (
	witTypes "go.bytecodealliance.org/pkg/wit/types"
	export "go.wasmcloud.dev/component/exports/wasmcloud_component_go_wasmcloud_nats_jetstream_handler_0_2_0/export_wasmcloud_nats_0_1_0_jetstream_handler"
	js "go.wasmcloud.dev/component/imports/wasmcloud_nats_0_1_0_jetstream"
	"go.wasmcloud.dev/component/nats"

	// Pull in the //go:wasmexport glue for the handle-message export.
	_ "go.wasmcloud.dev/component/exports/wasmcloud_component_go_wasmcloud_nats_jetstream_handler_0_2_0/wit_exports"
)

// HandleFunc registers fn as the callback invoked for each JetStream message
// delivered to the component. A non-nil error is reported to the host as the
// handler's failure string, and under `ack-mode: auto` naks the message.
func HandleFunc(fn func(handle *nats.MessageHandle) error) {
	export.Exports.HandleMessage = func(handle *js.MessageHandle) witTypes.Result[witTypes.Unit, string] {
		if err := fn(nats.NewMessageHandle(handle)); err != nil {
			return witTypes.Err[witTypes.Unit, string](err.Error())
		}
		return witTypes.Ok[witTypes.Unit, string](witTypes.Unit{})
	}
}
