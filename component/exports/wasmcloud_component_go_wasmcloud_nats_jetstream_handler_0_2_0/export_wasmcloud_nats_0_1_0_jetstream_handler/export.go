// Package export_wasmcloud_nats_0_1_0_jetstream_handler is the export
// trampoline for the optional `wasmcloud:nats/jetstream-handler@0.1.0`
// export. The generated wit_exports glue calls HandleMessage; the SDK's
// nats package assigns Exports.HandleMessage when the app registers a
// callback with nats/jetstreamhandler.HandleFunc.
package export_wasmcloud_nats_0_1_0_jetstream_handler

import (
	witTypes "go.bytecodealliance.org/pkg/wit/types"
	"go.wasmcloud.dev/component/imports/wasmcloud_nats_0_1_0_jetstream"
)

var Exports struct {
	HandleMessage func(handle *wasmcloud_nats_0_1_0_jetstream.MessageHandle) witTypes.Result[witTypes.Unit, string]
}

func HandleMessage(handle *wasmcloud_nats_0_1_0_jetstream.MessageHandle) witTypes.Result[witTypes.Unit, string] {
	if Exports.HandleMessage == nil {
		return witTypes.Err[witTypes.Unit, string]("wasmcloud:nats/jetstream-handler export invoked, but no callback was registered; call jetstreamhandler.HandleFunc during init")
	}
	return Exports.HandleMessage(handle)
}
