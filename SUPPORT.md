# Support

## Supported releases

The Go SDK, `go.wasmcloud.dev/component`, supports only its latest release. Today that is
`component/v0.1.6`.

- Bug fixes and security fixes land on `main` and ship in the next `component/vX.Y.Z` release,
  cut by the [release train](.github/workflows/release-train.yaml).
- A fix for a critical vulnerability ships as a patch release as soon as it is ready.
- When a new release ships, the previous release stops receiving any fixes, including security
  fixes. Fixes are not backported. To get a fix, upgrade to the latest release.
- Each release is built and tested with the toolchain pinned in [BUILDING.md](BUILDING.md),
  including the `wash` version.

The examples and templates in this repository aren't released separately. They track the latest
SDK release.

## Unsupported modules

The wasmCloud 1.x provider SDK (`go.wasmcloud.dev/provider`) and the TinyGo component SDK line
(`component/v0.0.x`) are no longer developed and receive no fixes, including security fixes.
Their existing versions keep resolving through the Go module proxy. See
[MIGRATING.md](MIGRATING.md) to move to the current SDK.

## Getting help

- Ask questions in [Slack](https://slack.wasmcloud.com/) or bring them to a
  [community meeting](https://wasmcloud.com/community/).
- Report bugs as [GitHub issues](https://github.com/wasmCloud/go/issues). See
  [CONTRIBUTING.md](CONTRIBUTING.md#reporting-bugs) for what to include.
- Report security vulnerabilities privately, as described in the
  [security policy](https://github.com/wasmCloud/go/security/policy).

Support is provided by the community on a best-effort basis.
