package wasihttp

import (
	"fmt"
	"net/http"
	"os"

	witTypes "go.bytecodealliance.org/pkg/wit/types"
	incominghandler "go.wasmcloud.dev/component/exports/wasmcloud_component_go_wasip2_0_2_0/export_wasi_http_0_2_8_incoming_handler"
	types "go.wasmcloud.dev/component/imports/wasi_http_0_2_8_types"

	// Pull in the //go:wasmexport glue for the component's exports.
	_ "go.wasmcloud.dev/component/exports/wasmcloud_component_go_wasip2_0_2_0/wit_exports"
)

// handler is the function that will be called by the http server.
var handler = defaultHandler

// defaultHandler is a placeholder for returning a useful error to stderr when
// the handler is not set.
var defaultHandler = func(http.ResponseWriter, *http.Request) {
	fmt.Fprintln(os.Stderr, "http handler undefined")
}

// Handle sets the handler function for the http trigger.
// It must be set in an init() function.
func Handle(h http.Handler) {
	handler = h.ServeHTTP
}

// HandleFunc sets the [net/http.HandlerFunc] that will be called to handle the
// incoming request.
func HandleFunc(h http.HandlerFunc) {
	handler = h
}

func wasiHandle(request *types.IncomingRequest, responseOut *types.ResponseOutparam) {
	httpReq, err := WASItoHTTPRequest(request)
	if err != nil {
		types.ResponseOutparamSet(responseOut, witTypes.Err[*types.OutgoingResponse, types.ErrorCode](
			types.MakeErrorCodeInternalError(witTypes.Some(err.Error()))),
		)
		return
	}
	if httpReq.Body != nil {
		defer httpReq.Body.Close()
	}

	httpRes := WASItoHTTPResponseWriter(responseOut)
	defer httpRes.Close()

	handler(httpRes, httpReq)
}

func init() {
	incominghandler.Exports.Handle = wasiHandle
}
