// Package export_wasmcloud_messaging_0_3_0_handler is the export trampoline
// for the optional `wasmcloud:messaging/handler@0.3.0` export. The generated
// wit_exports glue calls HandleMessage; the SDK's messaging package assigns
// Exports.HandleMessage when the app registers a callback with
// messaging.HandleFunc.
package export_wasmcloud_messaging_0_3_0_handler

import (
	witTypes "go.bytecodealliance.org/pkg/wit/types"
	"go.wasmcloud.dev/component/imports/wasmcloud_messaging_0_3_0_types"
)

var Exports struct {
	HandleMessage func(msg wasmcloud_messaging_0_3_0_types.BrokerMessage) witTypes.Result[witTypes.Unit, wasmcloud_messaging_0_3_0_types.HandleMessageError]
}

func HandleMessage(msg wasmcloud_messaging_0_3_0_types.BrokerMessage) witTypes.Result[witTypes.Unit, wasmcloud_messaging_0_3_0_types.HandleMessageError] {
	if Exports.HandleMessage == nil {
		// Redelivery cannot help a component that registered no callback, so
		// this is a reject rather than a retry.
		return witTypes.Err[witTypes.Unit, wasmcloud_messaging_0_3_0_types.HandleMessageError](
			wasmcloud_messaging_0_3_0_types.MakeHandleMessageErrorOther("wasmcloud:messaging/handler export invoked, but no callback was registered; call messaging.HandleFunc during init"),
		)
	}
	return Exports.HandleMessage(msg)
}
