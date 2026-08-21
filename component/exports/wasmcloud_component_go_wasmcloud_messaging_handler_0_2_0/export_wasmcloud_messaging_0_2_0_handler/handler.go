// Package export_wasmcloud_messaging_0_2_0_handler is the export trampoline
// for the optional `wasmcloud:messaging/handler@0.2.0` export. The generated
// wit_exports glue calls HandleMessage; the SDK's messaging package assigns
// Exports.HandleMessage when the app registers a callback with
// messaging.HandleFunc.
package export_wasmcloud_messaging_0_2_0_handler

import (
	witTypes "go.bytecodealliance.org/pkg/wit/types"
	"go.wasmcloud.dev/component/imports/wasmcloud_messaging_0_2_0_types"
)

var Exports struct {
	HandleMessage func(msg wasmcloud_messaging_0_2_0_types.BrokerMessage) witTypes.Result[witTypes.Unit, string]
}

func HandleMessage(msg wasmcloud_messaging_0_2_0_types.BrokerMessage) witTypes.Result[witTypes.Unit, string] {
	if Exports.HandleMessage == nil {
		return witTypes.Err[witTypes.Unit, string]("wasmcloud:messaging/handler export invoked, but no callback was registered; call messaging.HandleFunc during init")
	}
	return Exports.HandleMessage(msg)
}
