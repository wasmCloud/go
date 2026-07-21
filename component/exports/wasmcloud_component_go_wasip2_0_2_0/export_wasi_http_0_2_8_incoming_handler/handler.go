// Package export_wasi_http_0_2_8_incoming_handler is the export trampoline
// for the sync `wasip2` world's `wasi:http/incoming-handler@0.2.8` export.
// The generated wit_exports glue calls Handle; the SDK's net/wasihttp
// package assigns Exports.Handle at init time.
package export_wasi_http_0_2_8_incoming_handler

import (
	"go.wasmcloud.dev/component/imports/wasi_http_0_2_8_types"
)

var Exports struct {
	Handle func(request *wasi_http_0_2_8_types.IncomingRequest, responseOut *wasi_http_0_2_8_types.ResponseOutparam)
}

func Handle(request *wasi_http_0_2_8_types.IncomingRequest, responseOut *wasi_http_0_2_8_types.ResponseOutparam) {
	Exports.Handle(request, responseOut)
}
