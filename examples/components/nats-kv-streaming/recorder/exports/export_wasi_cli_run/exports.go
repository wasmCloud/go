// Package export_wasi_cli_run is the hand-written end of the generated glue
// in ../wit_exports, which calls [Run] when the host starts the workload's
// service through `wasi:cli/run@0.3.0`. It forwards to [Exports], which the
// recorder fills in during init.
package export_wasi_cli_run

import (
	witTypes "go.bytecodealliance.org/pkg/wit/types"
)

// Exports holds the implementation. A service's run is expected to last as
// long as the workload does; returning ends the service, and the host
// restarts it up to the workload's `maxRestarts`.
var Exports struct {
	Run func() witTypes.Result[witTypes.Unit, witTypes.Unit]
}

func Run() witTypes.Result[witTypes.Unit, witTypes.Unit] {
	return Exports.Run()
}
