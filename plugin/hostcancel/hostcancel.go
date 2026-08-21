// Package hostcancel is an idiomatic Go wrapper for the
// wasmcloud:host/cancel@0.1.0 interface (vendored from wasmCloud v2.6.1
// under wit/deps/wasmcloud-host-0.1.0), provided by the wasmCloud host so a
// host component plugin can cooperatively cancel one of its own in-flight
// invocations.
//
// Job ids are host-minted and opaque. A long-running invocation stashes
// [CurrentJob] somewhere another invocation can find it; that other
// invocation calls [RequestCancel], and the long-running one observes the
// mark via [IsCancelled] (or a dropped stream reader) and unwinds itself —
// without disturbing the store's other tenants.
//
// The plugin's world must import the interface:
//
//	world plugin {
//	  import wasmcloud:host/cancel@0.1.0;
//	  export acme:kv/store@0.1.0; // the plugin's own capability
//	}
//
// The import is host-provided to component plugins loaded via the wash
// `dev.host_plugins` config key; workloads cannot import it.
package hostcancel

import (
	cancel "go.wasmcloud.dev/plugin/imports/wasmcloud_host_0_1_0_cancel"
)

// JobID is an opaque, host-minted identifier for one in-flight invocation.
type JobID = uint64

// CurrentJob returns the id of the invocation the caller is currently
// running under, or 0 if none.
func CurrentJob() JobID {
	return cancel.CurrentJob()
}

// RequestCancel asks the invocation with the given job id to cancel. It
// returns whether the request was accepted (not whether the invocation has
// stopped); a request is accepted only for a caller in the same workload as
// the job's owner.
func RequestCancel(job JobID) bool {
	return cancel.RequestCancel(job)
}

// IsCancelled reports whether the invocation the caller is currently running
// under has been asked to cancel. A long-running invocation polls this and
// returns early.
func IsCancelled() bool {
	return cancel.IsCancelled()
}
