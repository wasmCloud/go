// Package handler wires up the optional wasmcloud:messaging/handler@0.3.0
// export so a component can receive messages from a wasmCloud messaging
// plugin.
//
// Importing this package links the `handle-message` wasm export into the
// component, so import it only from apps whose world exports the interface:
//
//	world app {
//	  include wasmcloud:component-go/wasip3@0.2.0;
//	  import wasmcloud:messaging/consumer@0.3.0;
//	  export wasmcloud:messaging/handler@0.3.0;
//	}
//
// Register a callback during init:
//
//	func init() {
//	  handler.HandleFunc(func(msg messaging.BrokerMessage) error {
//	    body, err := io.ReadAll(msg.Body)
//	    if err != nil {
//	      return fmt.Errorf("%w: read body: %w", messaging.ErrRetry, err)
//	    }
//	    // ...
//	    return nil
//	  })
//	}
//
// The workload manifest must declare the matching hostInterfaces entry
// (namespace `wasmcloud`, package `messaging`).
//
// The callback is an `async func` as of 0.3.0: it may await other imports —
// including messaging.Publish to reply on the message's ReplyTo subject —
// while the host goes on delivering further messages to the same instance.
package handler

import (
	"io"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
	export "go.wasmcloud.dev/component/exports/wasmcloud_component_go_wasmcloud_messaging_handler_0_2_0/export_wasmcloud_messaging_0_3_0_handler"
	types "go.wasmcloud.dev/component/imports/wasmcloud_messaging_0_3_0_types"
	"go.wasmcloud.dev/component/messaging"

	// Pull in the //go:wasmexport glue for the handle-message export.
	_ "go.wasmcloud.dev/component/exports/wasmcloud_component_go_wasmcloud_messaging_handler_0_2_0/wit_exports"
)

// HandleFunc registers fn as the callback invoked for each message delivered
// to the component's wasmcloud:messaging/handler export.
//
// What fn returns is a per-delivery disposition, not a transport failure: a
// handler cannot report a broker error about a message it has already
// received, only whether redelivery could help. Return nil on success,
// [messaging.ErrRetry] (on its own or wrapped with %w) for a transient
// failure worth redelivering, and [messaging.ErrReject] for one that no retry
// can fix. Any other error is reported as a handler-specific failure, which
// hosts treat as a reject.
//
// The message body is a stream from the host. It is closed once fn returns,
// so a handler that does not need the payload may simply ignore it; reading
// it after returning is not valid.
//
// HandleFunc should be called once, during program initialization (an init
// function or the start of main), before any message is delivered.
func HandleFunc(fn func(msg messaging.BrokerMessage) error) {
	export.Exports.HandleMessage = func(msg types.BrokerMessage) witTypes.Result[witTypes.Unit, types.HandleMessageError] {
		delivered := messaging.FromWit(msg)
		if body, ok := delivered.Body.(io.Closer); ok {
			defer body.Close()
		}
		if err := fn(delivered); err != nil {
			return witTypes.Err[witTypes.Unit, types.HandleMessageError](messaging.ToDisposition(err))
		}
		return witTypes.Ok[witTypes.Unit, types.HandleMessageError](witTypes.Unit{})
	}
}
