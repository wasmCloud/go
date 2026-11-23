---
name: wasmcloud-go-dependencies
description: Dependency selection, update, and advisory handling for the wasmCloud Go repository, covering per-module go.mod and go.sum, Dependabot PRs, dependency review, osv-scanner exceptions with expiry, allow-ghsas, and OpenVEX statements in vex/. Use when adding, removing, or bumping a Go module or GitHub Action, reviewing a Dependabot PR, or responding to a vulnerability alert.
---

# Dependencies and advisories

The [dependency policy](../../../CONTRIBUTING.md#dependencies) is the source of truth.

## Adding or bumping

- Check the standard library and the modules already in use first.
- Change only the module that needs it, and run `go mod tidy` in that module.
  Every module has its own `go.mod` and `go.sum`.
- Do not bump `go.wasmcloud.dev/component` in examples by hand;
  see the toolchain skill.
- Bumping `go.bytecodealliance.org/pkg` or `/wasi` changes what the generated bindings link against.
  Check BUILDING.md's go-pkg section and rebuild the SDK and an async example.
- Licenses must be Apache-2.0 or on the
  [CNCF allowlist](https://github.com/cncf/foundation/blob/main/policies-guidance/allowed-third-party-license-policy.md).

## Advisories

Fix by upgrading whenever possible.
Only a finding that cannot be reached from code in this repository may be suppressed,
and only with a written justification that another maintainer approves:

- osv-scanner: add an entry to [`osv-scanner.toml`](../../../osv-scanner.toml)
  with the reason and an `ignoreUntil` expiry no more than 90 days out.
- Dependency review: add the advisory ID to the `allow-ghsas` input of the
  `dependency-review-action` step in
  [`dependency-review.yml`](../../../.github/workflows/dependency-review.yml),
  creating the input if it is not there yet, with the reason in a comment.
- If the advisory affects a released SDK version,
  also add a `not_affected` statement with a justification to the OpenVEX document in [`vex/`](../../../vex/),
  and bump its `version` and `timestamp`.

Never suppress a finding just to make CI pass.
