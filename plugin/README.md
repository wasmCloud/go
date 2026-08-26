# go.wasmcloud.dev/plugin

Go SDK for authoring **wasmCloud host component plugins**: WebAssembly
components that *export* capability interfaces and are run by the wasmCloud
host in a long-lived, supervised store. One plugin instance serves concurrent
capability calls from every workload that imports the interfaces it exports.

The module vendors the `wasmcloud:host@0.1.0` WIT (from wasmCloud v2.6.1)
under `wit/` and provides:

| Package | WIT | Direction | Purpose |
|---|---|---|---|
| `hostidentity` | `wasmcloud:host/identity@0.1.0` | import | `WorkloadID()` / `ComponentID()` of the caller currently invoking the plugin |
| `hostcancel` | `wasmcloud:host/cancel@0.1.0` | import | `CurrentJob()`, `RequestCancel(job)`, `IsCancelled()` — cooperative cancellation of the plugin's own in-flight invocations |
| `lifecycle` | `wasmcloud:host/workload-lifecycle@0.1.0` | **export** | `OnWorkloadBind` / `OnWorkloadUnbind` hook registration; the host invokes the hooks as workloads bind/unbind |
| `imports/`, `exports/` | — | — | generated bindings (`./regenerate_bindings.sh`), do not edit |

## Authoring a host component in Go

1. **Define a world** that exports your capability interface. Identity and
   cancel are available as imports; workload-lifecycle is an optional export.
   `wasmcloud:host` exports are reserved host-invoked contracts — your
   workload-matchable capability must live in your own namespace:

   ```wit
   // wit/world.wit
   package acme:kv-plugin@0.1.0;

   world plugin {
     import wasmcloud:host/identity@0.1.0;
     import wasmcloud:host/cancel@0.1.0;
     export wasmcloud:host/workload-lifecycle@0.1.0;
     export acme:kv/store@0.1.0;
   }
   ```

   Copy `wit/deps/wasmcloud-host-0.1.0/` from this module (or the embedded
   `plugin.Wit` FS) into your project's `wit/deps/`, alongside the WIT for
   your own capability. Prefer `async func`s in your capability interface:
   the host dispatches cross-store capability calls asynchronously, and an
   async world lets one plugin store serve many calls concurrently.

2. **Implement it.** Generate export glue for your world with
   `componentize-go bindings` (see `regenerate_bindings.sh` in this module
   for the pattern: a generated `wit_exports` package plus a handwritten
   trampoline package your code registers into). Use this module's packages
   for the host-side interfaces:

   ```go
   import (
     "go.wasmcloud.dev/plugin/hostidentity"
     "go.wasmcloud.dev/plugin/lifecycle"
   )

   func init() {
     lifecycle.OnWorkloadBind(func(w lifecycle.WorkloadInfo) error {
       // provision per-workload state; must be idempotent (replayed after
       // a plugin restart). Returning an error rejects the deploy.
       return provision(w.ID, w.Interfaces)
     })
     lifecycle.OnWorkloadUnbind(func(id string) { reclaim(id) })
   }

   // inside a capability export:
   //   tenant := hostidentity.WorkloadID()
   ```

   Until go-pkg#10 merges, mirror this module's `go.mod` pin:

   ```
   require go.bytecodealliance.org/pkg v0.2.2
   replace go.bytecodealliance.org/pkg => github.com/jfleitz/go-pkg v0.2.4-0.20260731175613-c7a085937f13
   ```

3. **Build** with componentize-go (async support is enabled automatically
   for a world that uses async features; the patched Go toolchain is
   downloaded on demand):

   ```
   componentize-go -w acme:kv-plugin/plugin -d wit build -o kv_plugin.wasm
   ```

## Loading the plugin into a host

`wash dev` loads host component plugins from the `dev.host_plugins` key of
the wash config (project `.wash/config.yaml`, or global
`~/.config/wash/config.yaml`). Each entry has a host-unique `id` and exactly
one of `file` (local path) or `image` (OCI reference); nested fields are
camelCase. Requires a wash build with the `host-component-plugins` feature.

```yaml
dev:
  host_plugins:
    - id: acme-kv
      file: ./build/kv_plugin.wasm
      maxRestarts: 3            # optional: supervised restarts before the plugin is declared dead
    - id: acme-widgets
      image: ghcr.io/acme/widgets:1.2.0
      pullPolicy: ifNotPresent  # image sources only
      # expectedDigest: sha256:...   # optional OCI digest pin
```

## Calling a host component from an app

From the app's point of view a host component is just another capability
provider: the app's world declares an ordinary WIT import of the plugin's
exported interface,

```wit
world app {
  include wasmcloud:component-go/wasip3@0.2.0;
  import acme:kv/store@0.1.0;
}
```

and the workload manifest declares the matching `hostInterfaces` entry so
the host binds the workload to the plugin:

```yaml
spec:
  hostInterfaces:
    - namespace: acme
      package: kv
      interfaces: [store]
      version: 0.1.0
      config:
        bucket: orders
```

Arguments and results cross the store boundary transparently: plain values
are copied, `stream`/`future` handles are pumped, and resources are proxied.
