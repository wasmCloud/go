// Package export_wasmcloud_host_0_1_3_workload_lifecycle is the export
// trampoline for the optional `wasmcloud:host/workload-lifecycle@0.1.3`
// export. The generated wit_exports glue calls OnWorkloadBind and
// OnWorkloadUnbind; the SDK's lifecycle package assigns the Exports fields
// when the plugin registers hooks.
package export_wasmcloud_host_0_1_3_workload_lifecycle

import (
	witTypes "go.bytecodealliance.org/pkg/wit/types"
	"go.wasmcloud.dev/plugin/imports/wasmcloud_host_0_1_3_workload_lifecycle"
)

var Exports struct {
	OnWorkloadBind   func(workload wasmcloud_host_0_1_3_workload_lifecycle.WorkloadInfo) witTypes.Result[witTypes.Unit, string]
	OnWorkloadUnbind func(id string)
}

func OnWorkloadBind(workload wasmcloud_host_0_1_3_workload_lifecycle.WorkloadInfo) witTypes.Result[witTypes.Unit, string] {
	if Exports.OnWorkloadBind == nil {
		// No hook registered: accept the bind.
		return witTypes.Ok[witTypes.Unit, string](witTypes.Unit{})
	}
	return Exports.OnWorkloadBind(workload)
}

func OnWorkloadUnbind(id string) {
	if Exports.OnWorkloadUnbind != nil {
		Exports.OnWorkloadUnbind(id)
	}
}
