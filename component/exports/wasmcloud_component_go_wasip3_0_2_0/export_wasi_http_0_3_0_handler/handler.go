// Package export_wasi_http_0_3_0_handler is the export trampoline for the
// async `wasip3` world's `wasi:http/handler@0.3.0` export. The generated
// wit_exports glue calls Handle; the SDK's net/wasihttp3 package assigns
// Exports.Handle at init time.
package export_wasi_http_0_3_0_handler

import (
	witTypes "go.bytecodealliance.org/pkg/wit/types"
	"go.wasmcloud.dev/component/imports/wasi_http_0_3_0_types"
)

var Exports struct {
	Handle func(request *wasi_http_0_3_0_types.Request) witTypes.Result[*wasi_http_0_3_0_types.Response, wasi_http_0_3_0_types.ErrorCode]
}

func Handle(request *wasi_http_0_3_0_types.Request) witTypes.Result[*wasi_http_0_3_0_types.Response, wasi_http_0_3_0_types.ErrorCode] {
	return Exports.Handle(request)
}
