# Security self-assessment

Last reviewed: 2026-11-05

This assessment covers the wasmCloud/go repository:
its SDKs, examples and template, its CI and its releases.
How the repository works, and the interfaces of the SDKs,
are in [ARCHITECTURE.md](ARCHITECTURE.md).

The wasmCloud host that runs the components,
and the sandbox it gives them,
are covered by the wasmCloud
[self-assessment](https://github.com/wasmCloud/wasmCloud/blob/main/SELF_ASSESSMENT.md)
and [threat model](https://github.com/wasmCloud/wasmCloud/blob/main/THREAT_MODEL.md).

## What is at stake

`go.wasmcloud.dev/component` is compiled into every component built with it.
So the main risk is a supply-chain one:
anything wrong in the SDK at a release tag
ends up in every component that depends on that version.
The Go checksum database fixes each version's hash the first time it's fetched,
so a version can't be changed after release, only replaced by a new one.

Project/repository assets that should be watched closely:

- the SDK sources, their generated bindings and the WIT they're generated from
- the `component/vX.Y.Z` tags
- the `go.wasmcloud.dev` vanity import pages in [docs/](docs/)
- the toolchain versions pinned in [airgap/versions.env](airgap/versions.env)
- the GitHub Actions workflows, the tokens they get and the octo-sts trust policies in [.github/chainguard](.github/chainguard/)
- the repository settings and the `main` ruleset

## Threats and mitigations

### Malicious or mistaken code merged into the SDK

- Every change goes through a pull request:
  the `main` ruleset blocks direct pushes, force pushes and deletion,
  and organization admins can bypass it only through a pull request.
- A maintainer who didn't push the last commit must approve,
  including a code owner for each changed path,
  and every review thread must be resolved.
- CodeQL analyzes both SDKs on every pull request,
  and a change that adds a high or critical alert can't merge.
- The SDK, every example and the template are built on every pull request that changes them.

### A compromised or vulnerable dependency

- Each module commits its `go.sum`,
  and the Go checksum database verifies every module download.
- Renovate and Dependabot open update pull requests,
  reviewed like any other change.
  Dependabot waits out a cooldown before proposing a new version.
- Dependency review and osv-scanner block a pull request that adds
  a vulnerable, malicious or disallowed-license dependency.
- Vulnerabilities affecting the SDKs have fix deadlines, in
  [CONTRIBUTING.md](CONTRIBUTING.md#dependencies),
  and a release isn't cut while a critical or high one is open.

### A compromised CI workflow

- No workflow runs on `pull_request_target` or `workflow_run`,
  so code from a fork never runs with write access or secrets.
- Every workflow grants no token permissions at the top level;
  each job asks for only what it uses.
- Checkouts don't persist credentials.
- Every action is pinned to a full commit SHA,
  except the OSPS Baseline publishing workflow, which grc.store requires by tag.
- Untrusted values reach shell steps through environment variables, not `${{ }}` expansion.
- zizmor audits the workflows on every pull request.
- The release and Renovate workflows get their app tokens from octo-sts, not a stored key.
  Each trust policy admits only its own workflow on `main`,
  with only the permissions that workflow needs.
- [bench.yml](.github/workflows/bench.yml) runs the ref it's given,
  including a pull request's, on a persistent self-hosted runner.
  Only someone with write access can start it.
  The job that runs the benched code can't request an OIDC token,
  and its checkout doesn't persist the job's token.
  A separate job, which runs only the workflow's own push script,
  assumes the AWS role, and is skipped for a pull request's ref.

### A tampered release

- [release-tag.yaml](.github/workflows/release-tag.yaml) tags only a merged release commit on `main`,
  after checking the examples' `go.sum` against it,
  and refuses to move a tag that already exists.
- It checks the published module against `proxy.golang.org` and `sum.golang.org`,
  and publishes the GitHub Release only after the offline build on the tag passes.
- Once fetched, a version's hash is fixed in the checksum database,
  so `go` refuses a module that was changed after release.
- Each release attaches the module zip and `SHA256SUMS` with Sigstore-signed SLSA provenance.
  [README.md](README.md#verifying-a-release) shows how to verify it,
  including that `release-tag.yaml` signed it.

### A redirected module path

- `go.wasmcloud.dev` is served by GitHub Pages from `docs/` on `main`,
  so its `go-import` pages change only through a reviewed pull request.
- A redirect to another repository can't change a version that's already in the checksum database.

### Secrets committed to the repository

- The repository stores no secrets of its own.
  A few organization secrets are visible to it, but no workflow here references them.
  [bench.yml](.github/workflows/bench.yml) names four `WASMCLOUD_BENCH_*` secrets,
  which aren't set here.
- Secret scanning and push protection are on,
  so a push with a detected secret is blocked.
- How the secrets are stored, accessed and rotated is in [SECURITY.md](SECURITY.md#secrets-and-credentials).

## Threat model and attack surface

The sections above cover the repository and its releases.
This one covers the SDK code that runs inside every component built with it.
The host's side of each boundary, the WebAssembly sandbox,
is covered by wasmCloud's [threat model](https://github.com/wasmCloud/wasmCloud/blob/main/THREAT_MODEL.md).

### Assets

- The data a component handles:
  HTTP requests and responses, messages, key-value and blob contents, and query results.
- Secrets a component reveals through `wasmcloud:secrets`.
- The component's own behaviour,
  which runs with whatever capabilities the host grants it.

### Entry points and trust boundaries

Everything the SDK receives crosses one boundary, the WIT interface between host and component:

- **Host to component, through exports:**
  incoming HTTP requests, through `go.bytecodealliance.org/wasi`,
  and messages delivered to the `messaging` and `nats` handler exports.
- **Host to component, through the results of imports:**
  values returned by key-value, blobstore, Postgres, NATS and secrets calls,
  and the errors those calls return.

The host lifts every value into the canonical ABI, so its shape is trusted:
the generated bindings never parse raw bytes from the network.
Its content isn't trusted:
request and message bodies, stored values and error strings come from outside parties,
through the host.

The component reaches nothing the host doesn't grant.
A package's imports enter a component only when its own world declares them,
since [capabilities.wit](component/wit/capabilities.wit) is not a build target.

### Critical code paths

- **Generated bindings** in `component/imports` and `component/exports`
  move every value across the boundary.
  They come from the pinned wit-bindgen release, are never edited by hand,
  and `make airgap-versions` fails if their generator drifts from the pin.
- **Handwritten conversions** turn WIT types into Go:
  the error mappings in each package, the Postgres value encodings, and message dispositions.
  They have unit tests,
  which found and fixed a bug where `postgres.DecodeFloat64` returned 0 for the smallest subnormal values.
- **Handler exports** in `messaging/handler` and `nats/*handler`
  call the component's handler with each delivered message.
  `messaging/handler` closes the message body once the handler returns.
- **Secret reveal** in `secrets` returns a secret's plaintext to the component.

### Threats and mitigations

| Threat | Path | Mitigation |
| --- | --- | --- |
| A value is converted wrongly, so the component acts on different data than the host sent | Handwritten conversions | Unit tests in CI; CodeQL on every pull request |
| Generated bindings are edited or go stale | Generated bindings | Regenerated only by `regenerate_bindings.sh`; generator version pinned and checked; code owner review |
| A revealed secret leaks into logs | Secret reveal | `secrets.Value` has a `String` method that returns the plaintext, so it isn't safe to print; see [Risks users inherit](#risks-users-inherit) |
| A handler's error text, which may hold message data, is passed to the host and broker | Handler exports | Only the disposition (retry, reject, other) is meaningful, so keep message data out of error text; not yet documented on `HandleFunc` |
| A message body is read after its handler returns | `messaging/handler` | The SDK closes the body when the handler returns, and `HandleFunc` documents that reading it later isn't valid |
| A handler panics on hostile input | Handler exports | The SDK doesn't recover panics, so the instance traps and the host fails that delivery; nothing is half-processed silently |
| A caller isn't who it claims to be | HTTP and message entry points | Out of the SDK's scope: authentication belongs to the component and the host; the examples don't authenticate, as noted below |

## Risks users inherit

- **componentize-go downloads part of the toolchain at build time.**
  Its toolchain binary comes from its GitHub Releases,
  and for an async WASI P3 world, a patched Go from `dicej/go`.
  Neither goes through the Go checksum database.
  [BUILDING.md](BUILDING.md) pins the versions.
- **A revealed secret prints as plaintext.**
  `secrets.Value` implements `fmt.Stringer`,
  so `fmt.Print(v)`, `%v` and most structured loggers write the secret itself.
  Keep `Value` out of logs and error messages.
- **The examples don't authenticate requests.**
  Every HTTP route answers anyone who can reach it.
  `http-keyvalue-crud` lets any caller read, write and delete keys.
- **The template ships a `go.mod` without a `go.sum`.**
  `wash new` runs `go mod tidy`,
  which resolves the versions in `go.mod` and checks them against the checksum database.

## Known gaps

- No ruleset protects the `component/v*` tags;
  a repository admin could push or delete one outside the release workflow.
  The checksum database still rejects a changed version.
- Releases published before 2026-11-05, when releases became immutable,
  carry no release attestation or provenance.
  Later releases, and the `go-airgap` image, are attested.
- The `dicej/go` and componentize-go downloads in the air-gapped image build
  aren't pinned by SHA256, as BUILDING.md notes.
- Renovate's Go module updates have no cooldown.
- Unit tests cover only the component SDK's conversions between WIT and Go types;
  the plugin SDK has none,
  and calls into the host are exercised only by the examples' smoke tests.
- Once bench.yml is set up, a pull request benched on the self-hosted runner
  runs on a persistent host that later benchmark runs reuse,
  including the job that assumes the AWS role.
- The async WASI P3 examples build with `dicej/go` `go1.27.1-wasi-on-idle`,
  a patched Go 1.27.1 without the standard library fixes in 1.27.2 and 1.27.3.
  [vex/](vex/) records why the SDK isn't affected; the examples' own code isn't covered.

## Vulnerabilities

- Report a vulnerability as described in [SECURITY.md](SECURITY.md).
- Fixed vulnerabilities are published as
  [GitHub Security Advisories](https://github.com/wasmCloud/go/security/advisories)
  in this repository.
