---
name: wasmcloud-go-toolchain
description: How pinned toolchain and SDK versions work in the wasmCloud Go repository, covering airgap/versions.env as the single source, BUILDING.md, make airgap-versions, the componentize-go release-string trap, the patched Go for async worlds, bump-sdk.sh, the release train, and offline builds. Use when bumping Go, componentize-go, wit-bindgen, go.bytecodealliance.org/pkg or /wasi, wash, or the SDK version in examples, or when debugging air-gapped builds.
---

# Toolchain and versions

[BUILDING.md](../../../BUILDING.md) explains every pinned version and why;
read the section for the tool you are changing before changing it.

## Where versions live

- [`airgap/versions.env`](../../../airgap/versions.env) is the single source.
  The Dockerfile, CI, and `make` read it.
- BUILDING.md's tables describe the same values in prose.
- `make airgap-versions` re-derives each value from upstream and fails if anything disagrees.
  Run it after any version change.

## componentize-go

- Pin a tagged release, never a pseudo-version.
  The Go module is a downloader whose `main.go` hardcodes `release := "vX.Y.Z"`,
  and that string decides which Rust toolchain actually runs.
- Read `wit-bindgen` and `wit-component` versions from the release named by `release :=`,
  not from the module version or `main`'s `Cargo.toml`.
- A tag whose `release :=` names a different version (such as v0.5.0) is not a safe bump.

## Bumping the SDK in examples

- Do not hand-edit `go.wasmcloud.dev/component` versions across examples.
  Run `airgap/bump-sdk.sh vX.Y.Z` once `component/vX.Y.Z` is tagged;
  it updates `versions.env`, BUILDING.md, and every example and template, then checks them.
- The [release train](../../../.github/workflows/release-train.yaml) runs
  `airgap/bump-sdk.sh --provisional` to pin examples to a not-yet-pushed tag;
  see `airgap/local-sdk-env.sh` for how that resolves.

## Bumping a toolchain

- Toolchain bumps (componentize-go, `wash`, the patched Go) move several coupled values,
  and may require regenerating the SDK bindings.
  `bump-sdk.sh` does not cover them.
- Change `versions.env`, then BUILDING.md, then any example READMEs that quote the version,
  then run `make airgap-versions`.
- If `SDK_WIT_BINDGEN` changes, regenerate bindings as described in the SDK skill.
- Verify with `make airgap-test`, which builds every module offline in a fresh container.
  Narrow it with `AIRGAP_MODULES=<path>` while iterating.
  It needs Docker; say so if it could not run.
