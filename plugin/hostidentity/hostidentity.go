// Package hostidentity is an idiomatic Go wrapper for the
// wasmcloud:host/identity@0.1.3 interface (vendored from wasmCloud v2.8.0
// under wit/deps/wasmcloud-host-0.1.3), provided by the wasmCloud host to a
// host component plugin so it can partition state by the workload currently
// calling it.
//
// The host answers based on the in-flight capability call, resolved from the
// caller's own guest task, so the values are exact under concurrency no
// matter when in the call the plugin reads them.
//
// The plugin's world must import the interface:
//
//	world plugin {
//	  import wasmcloud:host/identity@0.1.3;
//	  export acme:kv/store@0.1.0; // the plugin's own capability
//	}
//
// The import is host-provided to component plugins loaded via the wash
// `dev.host_plugins` config key; workloads cannot import it.
//
// Inside a workload-lifecycle hook, [WorkloadID] reports the workload the
// hook concerns, but [ComponentID] is not meaningful there — read ids from
// the hook's WorkloadInfo argument instead.
package hostidentity

import (
	identity "go.wasmcloud.dev/plugin/imports/wasmcloud_host_0_1_3_identity"
)

// WorkloadID returns the workload id of the caller currently invoking this
// plugin.
func WorkloadID() string {
	return identity.GetWorkloadId()
}

// ComponentID returns the component id of the caller currently invoking
// this plugin.
func ComponentID() string {
	return identity.GetComponentId()
}
