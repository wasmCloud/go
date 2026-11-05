# Contributing to wasmCloud Go

Thanks for your interest in the wasmCloud Go SDK, examples and templates. wasmCloud is a CNCF
project; the [project governance](https://github.com/wasmCloud/wasmCloud/blob/main/GOVERNANCE.md)
and [code of conduct](https://github.com/cncf/foundation/blob/main/code-of-conduct.md) apply here.

## Reporting Bugs

Report bugs as [GitHub issues](https://github.com/wasmCloud/go/issues). Include what you ran,
what happened, what you expected, and your versions of Go, componentize-go and `wash`.

Don't report security vulnerabilities as issues. Follow the
[security policy](https://github.com/wasmCloud/go/security/policy) to report them privately.

## Building

[BUILDING.md](BUILDING.md) lists the pinned toolchain versions and how to build components,
including offline builds.

## Submitting Changes

- Open a pull request against `main`. For a larger change, open an issue first to discuss it.
- Sign off every commit with `git commit -s` to certify the
  [Developer Certificate of Origin](https://developercertificate.org/). The DCO check runs on every
  pull request.
- [`@wasmCloud/go-maintainers`](MAINTAINERS.md) review every change, as set in
  [`CODEOWNERS`](.github/CODEOWNERS).
- CI must pass before a pull request merges. What it runs is in [Testing](#testing).

## Testing

### When tests run

CI runs these on every pull request,
and again on each push to `main` that changes the files they test:

- **Component / Go** ([component-go.yaml](.github/workflows/component-go.yaml))
  compiles and vets the SDK with `GOOS=wasip1`, the only WASI port Go has,
  runs its unit tests,
  builds each example into a WASI P2 or P3 component with `wash build`,
  and smoke-tests under `wash dev` the examples that a released `wash` can serve.
- **Templates** ([templates-go.yaml](.github/workflows/templates-go.yaml))
  creates a project from the template with `wash new` and builds it.
- **Air-gapped build** ([airgapped.yaml](.github/workflows/airgapped.yaml))
  builds every example and template with no network,
  from the versions pinned in [BUILDING.md](BUILDING.md).

On a pull request, each workflow skips its jobs when none of its paths changed,
and its gate job (`component-gate`, `templates-gate`, `airgapped-gate`) still reports,
so the required checks always run.

### How to run them

From a checkout, with the toolchain in [BUILDING.md](BUILDING.md):

```shell
# SDK unit tests. They cover the Go side of the WIT types and run natively.
cd component && go test ./...

# What CI compiles and vets.
GOOS=wasip1 GOARCH=wasm go build ./...
GOOS=wasip1 GOARCH=wasm go vet $(go list ./... | grep -vE '/(imports|exports)/')

# An example, built and served the way CI smoke-tests it.
cd examples/components/http-server && wash build && wash dev

# Every example and template, offline.
make airgap-test
```

### Tests for new changes

Every change to the SDK's behaviour must add or update the unit tests that cover it,
in the package it changes, and a bug fix must add a test that fails without the fix.
A new example must build in CI, and must pass the `wash dev` smoke test if a released `wash` can serve it.
Reviewers don't approve a change that's missing its tests.

## Dependencies

Before adding a new dependency, check whether the standard library or a module already in use does
the job. A new dependency must be actively maintained and use a license on the
[CNCF allowed license list](https://github.com/cncf/foundation/blob/main/policies-guidance/allowed-third-party-license-policy.md).

- **Obtaining:** Go modules come from the Go module proxy and are pinned in each module's `go.sum`,
  which the Go checksum database verifies. GitHub Actions are pinned to a full commit SHA, with the
  version in a trailing comment.
- **Tracking:** Renovate opens Go module update PRs, including indirect dependencies. Dependabot
  opens GitHub Actions update PRs, and security-update PRs for both. Review them like any other
  change.
- **Fixing:** a known vulnerability in `go.wasmcloud.dev/component` or `go.wasmcloud.dev/plugin`, or
  in a dependency they ship with, is fixed within 7 days for critical, 14 days for high and 90 days
  for medium severity, on the timeline in [Code Scanning](#code-scanning). A release isn't cut
  while a critical or high advisory affecting the SDK is open. Examples and templates get the same
  update PRs, but their findings don't block a release.

## Code Scanning

[CodeQL](.github/workflows/codeql.yml) analyzes the SDK on every pull request, on every push to
`main`, and weekly. Its alerts appear in the repository's Security tab.

- A pull request that adds a high or critical severity CodeQL alert is blocked from merging until
  the alert is fixed or dismissed with a stated reason.
- Alerts open on `main` must be fixed, or dismissed with a stated reason, within 7 days for
  critical, 14 days for high and 90 days for medium severity. Low severity alerts are fixed on a
  best-effort basis.
