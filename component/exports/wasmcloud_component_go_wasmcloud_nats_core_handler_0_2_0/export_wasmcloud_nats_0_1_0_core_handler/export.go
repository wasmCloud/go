// Package export_wasmcloud_nats_0_1_0_core_handler is the export trampoline
// for the optional `wasmcloud:nats/core-handler@0.1.0` export. The generated
// wit_exports glue calls HandleMessage; the SDK's nats package assigns
// Exports.HandleMessage when the app registers a callback with
// nats/corehandler.HandleFunc.
package export_wasmcloud_nats_0_1_0_core_handler

import (
	witTypes "go.bytecodealliance.org/pkg/wit/types"
	"go.wasmcloud.dev/component/imports/wasmcloud_nats_0_1_0_types"
)

var Exports struct {
	HandleMessage func(msg wasmcloud_nats_0_1_0_types.NatsMessage) witTypes.Result[witTypes.Unit, string]
}

func HandleMessage(msg wasmcloud_nats_0_1_0_types.NatsMessage) witTypes.Result[witTypes.Unit, string] {
	if Exports.HandleMessage == nil {
		return witTypes.Err[witTypes.Unit, string]("wasmcloud:nats/core-handler export invoked, but no callback was registered; call corehandler.HandleFunc during init")
	}
	return Exports.HandleMessage(msg)
}
