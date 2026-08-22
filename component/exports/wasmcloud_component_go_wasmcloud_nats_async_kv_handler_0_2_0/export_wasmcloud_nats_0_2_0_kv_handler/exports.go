// Package export_wasmcloud_nats_0_2_0_kv_handler is the export trampoline
// for the optional `wasmcloud:nats/kv-handler@0.2.0` export — the async
// (WASI P3) revision. The generated wit_exports glue calls HandleEvent; the
// SDK's natsp3 package assigns Exports.HandleEvent when the app registers a
// callback with natsp3/kvhandler.HandleFunc.
package export_wasmcloud_nats_0_2_0_kv_handler

import (
	witTypes "go.bytecodealliance.org/pkg/wit/types"
	"go.wasmcloud.dev/component/imports/wasmcloud_nats_0_2_0_kv"
)

var Exports struct {
	HandleEvent func(bucket string, entry wasmcloud_nats_0_2_0_kv.Entry) witTypes.Result[witTypes.Unit, string]
}

func HandleEvent(bucket string, entry wasmcloud_nats_0_2_0_kv.Entry) witTypes.Result[witTypes.Unit, string] {
	if Exports.HandleEvent == nil {
		return witTypes.Err[witTypes.Unit, string]("wasmcloud:nats/kv-handler export invoked, but no callback was registered; call kvhandler.HandleFunc during init")
	}
	return Exports.HandleEvent(bucket, entry)
}
