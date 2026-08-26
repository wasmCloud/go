// Package lifecycle wires up the optional
// wasmcloud:host/workload-lifecycle@0.1.0 export (vendored from wasmCloud
// v2.6.1 under wit/deps/wasmcloud-host-0.1.0), through which the wasmCloud
// host tells a host component plugin about workloads binding to and
// unbinding from it, so the plugin can provision and reclaim per-workload
// state eagerly rather than lazily on first call.
//
// The interface is exported by the plugin (hooks the host invokes), not
// imported: importing this package links the export glue into the
// component, so import it only from plugins whose world exports the
// interface:
//
//	world plugin {
//	  import wasmcloud:host/identity@0.1.0;
//	  export wasmcloud:host/workload-lifecycle@0.1.0;
//	  export acme:kv/store@0.1.0; // the plugin's own capability
//	}
//
// Register hooks during init:
//
//	func init() {
//	  lifecycle.OnWorkloadBind(func(w lifecycle.WorkloadInfo) error { ... })
//	  lifecycle.OnWorkloadUnbind(func(id string) { ... })
//	}
//
// Semantics (see the vendored WIT for the full contract):
//
//   - The bind hook runs to completion before any capability call from that
//     workload is delivered; returning an error rejects the bind and the
//     workload fails to deploy.
//   - Bind MUST be idempotent: after a plugin restart the host replays a
//     bind for every workload still bound.
//   - Unbind is best-effort and must tolerate ids never bound or already
//     unbound.
//   - During a hook, hostidentity.WorkloadID reports the workload the hook
//     concerns; hostidentity.ComponentID is not meaningful — every id a hook
//     needs is on the [WorkloadInfo] argument.
//   - Hooks must not invoke the plugin's own exported capabilities.
package lifecycle

import (
	witTypes "go.bytecodealliance.org/pkg/wit/types"
	export "go.wasmcloud.dev/plugin/exports/wasmcloud_plugin_go_lifecycle_0_1_0/export_wasmcloud_host_0_1_0_workload_lifecycle"
	wit "go.wasmcloud.dev/plugin/imports/wasmcloud_host_0_1_0_workload_lifecycle"

	// Pull in the //go:wasmexport glue for the workload-lifecycle export.
	_ "go.wasmcloud.dev/plugin/exports/wasmcloud_plugin_go_lifecycle_0_1_0/wit_exports"
)

// Version is the semantic version of an interface instance.
type Version struct {
	Major uint64
	Minor uint64
	Patch uint64
	// Pre is the pre-release identifier, if any (e.g. "draft", "rc.1").
	Pre string
	// Build is the build metadata, if any.
	Build string
}

// InterfaceBinding is one interface instance a workload matched on this
// plugin — e.g. the `store` interface of `acme:kv@0.1.0`, labeled `cache`,
// with its manifest config.
type InterfaceBinding struct {
	// Namespace is the interface namespace (e.g. "acme").
	Namespace string
	// Package is the package within the namespace (e.g. "kv").
	Package string
	// Interfaces are the matched interface names within the package.
	Interfaces []string
	// Version is the version of the instance, if one was requested.
	Version *Version
	// Name is the instance label when the workload names multiple instances
	// of the same namespace:package ("" if none) — what the component calls
	// this import.
	Name string
	// ExternalID is the platform-facing id the binding was declared with
	// ("" if none). It is not an identity: it carries no uniqueness
	// requirement.
	ExternalID string
	// Config is the interface-level configuration from the workload
	// manifest, sorted by key.
	Config map[string]string
}

// WorkloadInfo is the identity and configuration of a workload binding to
// this plugin.
type WorkloadInfo struct {
	// ID is the unique id of the workload instance; it matches
	// hostidentity.WorkloadID on later capability calls from the workload.
	ID string
	// Name is the human-readable workload name.
	Name string
	// Namespace is the namespace the workload was deployed under.
	Namespace string
	// Service is the id of the workload's service, "" if it has none.
	Service string
	// Components are the ids of the workload's components, excluding the
	// service.
	Components []string
	// Interfaces are the interface instances the workload matched on this
	// plugin.
	Interfaces []InterfaceBinding
}

// OnWorkloadBind registers fn to run as a workload importing one of this
// plugin's exported capabilities is bound. A non-nil error rejects the bind
// and the workload fails to deploy. fn must be idempotent: after a plugin
// restart the host replays a bind for every workload still bound.
//
// OnWorkloadBind should be called once, during program initialization.
func OnWorkloadBind(fn func(workload WorkloadInfo) error) {
	export.Exports.OnWorkloadBind = func(w wit.WorkloadInfo) witTypes.Result[witTypes.Unit, string] {
		if err := fn(fromWit(w)); err != nil {
			return witTypes.Err[witTypes.Unit, string](err.Error())
		}
		return witTypes.Ok[witTypes.Unit, string](witTypes.Unit{})
	}
}

// OnWorkloadUnbind registers fn to run when a workload is gone — a normal
// stop, or cleanup after a failed bind. fn must tolerate ids never bound or
// already unbound.
//
// OnWorkloadUnbind should be called once, during program initialization.
func OnWorkloadUnbind(fn func(id string)) {
	export.Exports.OnWorkloadUnbind = fn
}

func fromWit(w wit.WorkloadInfo) WorkloadInfo {
	out := WorkloadInfo{
		ID:         w.Id,
		Name:       w.Name,
		Namespace:  w.Namespace,
		Service:    w.Service.SomeOr(""),
		Components: w.Components,
	}
	for _, ib := range w.Interfaces {
		binding := InterfaceBinding{
			Namespace:  ib.Namespace,
			Package:    ib.Package,
			Interfaces: ib.Interfaces,
			Name:       ib.Name.SomeOr(""),
			ExternalID: ib.ExternalId.SomeOr(""),
			Config:     make(map[string]string, len(ib.Config)),
		}
		if ib.Version.IsSome() {
			v := ib.Version.Some()
			binding.Version = &Version{
				Major: v.Major,
				Minor: v.Minor,
				Patch: v.Patch,
				Pre:   v.Pre.SomeOr(""),
				Build: v.Build.SomeOr(""),
			}
		}
		for _, kv := range ib.Config {
			binding.Config[kv.F0] = kv.F1
		}
		out.Interfaces = append(out.Interfaces, binding)
	}
	return out
}
