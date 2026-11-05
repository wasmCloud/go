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
- the GitHub Actions workflows, the tokens they get and the bot app's credentials
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
- The release workflows ask the bot app for only the permissions each needs.
  Renovate's token keeps the app's full set.
- [bench.yml](.github/workflows/bench.yml) runs the ref it's given,
  including a pull request's, on a persistent self-hosted runner.
  Only someone with write access can start it.
  For a pull request's ref it skips the AWS role,
  so the code under test never runs alongside AWS credentials,
  and its checkout doesn't persist the job's token.

### A tampered release

- [release-tag.yaml](.github/workflows/release-tag.yaml) tags only a merged release commit on `main`,
  after checking the examples' `go.sum` against it,
  and refuses to move a tag that already exists.
- It checks the published module against `proxy.golang.org` and `sum.golang.org`,
  and publishes the GitHub Release only after the offline build on the tag passes.
- Once fetched, a version's hash is fixed in the checksum database,
  so `go` refuses a module that was changed after release.

### A redirected module path

- `go.wasmcloud.dev` is served by GitHub Pages from `docs/` on `main`,
  so its `go-import` pages change only through a reviewed pull request.
- A redirect to another repository can't change a version that's already in the checksum database.

### Secrets committed to the repository

- The repository's only secrets are the bot app's ID and private key,
  used by the release and Renovate workflows, which pull requests can't trigger.
  A few organization secrets are visible to it, but no workflow here references them.
  [bench.yml](.github/workflows/bench.yml) names four `WASMCLOUD_BENCH_*` secrets,
  which aren't set here.
- Secret scanning and push protection are on,
  so a push with a detected secret is blocked.

## Risks users inherit

- **componentize-go downloads part of the toolchain at build time.**
  Its toolchain binary comes from its GitHub Releases,
  and for an async WASI P3 world, a patched Go from `dicej/go`.
  Neither goes through the Go checksum database.
  [BUILDING.md](BUILDING.md) pins the versions.
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
- The `go-airgap` image and the GitHub Releases carry no signature or provenance.
- The `dicej/go` and componentize-go downloads in the air-gapped image build
  aren't pinned by SHA256, as BUILDING.md notes.
- Renovate's Go module updates have no cooldown.
- The SDKs have no unit tests;
  CI builds them and smoke-tests the examples that a released `wash` can serve.
- Once bench.yml is set up, a pull request benched on the self-hosted runner
  runs on a persistent host that later benchmark runs reuse,
  in a job that can request an OIDC token.
- Pull requests aren't blocked on vulnerable dependencies;
  Dependabot alerts are handled under the fix deadlines instead.

## Vulnerabilities

- Report a vulnerability as described in the wasmCloud
  [security policy](https://github.com/wasmCloud/go/security/policy).
- Fixed vulnerabilities are published as
  [GitHub Security Advisories](https://github.com/wasmCloud/go/security/advisories)
  in this repository.
