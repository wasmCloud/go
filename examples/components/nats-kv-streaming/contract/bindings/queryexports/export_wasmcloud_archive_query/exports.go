// Package export_wasmcloud_archive_query is the hand-written end of the
// generated glue in ../wit_exports, which calls [Run] for every call into
// `wasmcloud:archive/query`. It forwards to [Exports], which the query
// component fills in during init. See
// ../../storeexports/export_wasmcloud_archive_store for why the indirection.
package export_wasmcloud_archive_query

import (
	query "github.com/wasmCloud/go/examples/components/nats-kv-streaming/contract/bindings/queryexports/wasmcloud_archive_query"
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// InputDone is the future the input stream's writer resolves once it has
// finished.
type InputDone = witTypes.FutureReader[witTypes.Result[witTypes.Unit, string]]

// OutputDone is the future the query resolves once its output is complete.
type OutputDone = witTypes.FutureReader[witTypes.Result[query.Summary, string]]

// Exports holds the implementation. A call while Run is nil panics, which
// traps the component.
var Exports struct {
	Run func(input *witTypes.StreamReader[uint8], inputDone *InputDone, filter query.Filter) (*witTypes.StreamReader[uint8], *OutputDone)
}

func Run(input *witTypes.StreamReader[uint8], inputDone *InputDone, filter query.Filter) (*witTypes.StreamReader[uint8], *OutputDone) {
	return Exports.Run(input, inputDone, filter)
}
