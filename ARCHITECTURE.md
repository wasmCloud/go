# Architecture

This repository holds the Go SDKs, examples and project templates
for building [wasmCloud](https://github.com/wasmCloud/wasmCloud) components.
It ships no service.
What it releases are Go modules:
`go.wasmcloud.dev/component`, tagged `component/vX.Y.Z`
and distributed through the Go module proxy,
and `go.wasmcloud.dev/plugin`, which hasn't been released yet.
The examples and the template are copied, not imported,
and track the latest SDK release.

This document covers who and what acts on the repository,
how a change becomes a release,
and the external interfaces of the SDKs and the template.
The wasmCloud host that runs the components is covered by wasmCloud's own
[threat model](https://github.com/wasmCloud/wasmCloud/blob/main/THREAT_MODEL.md).

## Actors

- **Contributors** open pull requests, usually from forks.
  Every commit must be signed off (DCO).
- **Maintainers**, listed in [MAINTAINERS.md](MAINTAINERS.md),
  review and merge pull requests, triage issues, start releases
  and manage the repository settings.
  Code owners for each path are in [CODEOWNERS](.github/CODEOWNERS).
- **Dependabot** opens pull requests for Go module and GitHub Actions updates,
  and for security updates.
  They're reviewed and merged like any other pull request.
- **GitHub Actions workflows** in [.github/workflows](.github/workflows)
  build, test, scan and release the repository.
  Each job's token gets only the permissions that job asks for.
- **[octo-sts](https://github.com/octo-sts/app)** gives the release workflows a GitHub App token
  in exchange for the job's OIDC token, so no app key is stored.
  Their pull requests and tags start other workflows,
  which `GITHUB_TOKEN` can't do.
  A trust policy in [.github/chainguard](.github/chainguard/) admits each workflow, on `main` only.
- **The bench workflow**, [bench.yml](.github/workflows/bench.yml),
  is meant to run benchmarks on the wasmCloud organization's self-hosted bench runner,
  a persistent host shared with wasmCloud/wasmCloud's benchmarks.
  Anyone with write access can start it on any ref, including a pull request's,
  and it also starts when a release is published.
  For any ref but a pull request's,
  a separate job assumes an AWS role through OIDC to push the results to the shared benchmark bucket;
  the job that runs the benched code can't request an OIDC token.
  It isn't set up here yet:
  its secrets and variable aren't set,
  and every run so far has been cancelled.
- **GitHub Pages** serves `go.wasmcloud.dev` from [docs/](docs/):
  the `go-import` meta tags that point the module paths at this repository.
- **The Go module proxy and checksum database** (`proxy.golang.org`, `sum.golang.org`)
  serve each tagged module version
  and record its hash the first time it's fetched.
- **Users** add the SDK to their own module with `go get`,
  or copy the template into a new project with `wash new`,
  then build with componentize-go and run with `wash` and a wasmCloud host.

## Actions

### Changing the repository

1. A contributor opens a pull request against `main`.
2. CI runs on the pull request:
   - **Component / Go** compiles and vets the SDK for `GOOS=wasip1`,
     runs its unit tests,
     builds every example into a component with `wash build`,
     and smoke-tests under `wash dev` those that a released `wash` can serve.
   - **Templates** creates a project from the template with `wash new` and builds it.
   - **Air-gapped build**, when the SDK, the toolchain, an example, the template,
     BUILDING.md or the Makefile changes,
     builds every example and the template with `--network none`
     from the versions pinned in [airgap/versions.env](airgap/versions.env).
   - **CodeQL** analyzes the component and plugin SDKs.
   - **zizmor** audits the workflows.
   - **DCO** checks that every commit is signed off.
   - **commitlint** checks that every commit, and the pull request title,
     is a Conventional Commit with a subject under 70 characters.
3. A maintainer who didn't push the last commit approves it,
   including a code owner for each changed path,
   and every review thread is resolved.
4. The `main` ruleset requires the checks to pass.
   It blocks force pushes and deletion of `main`,
   and blocks merging a change that adds a high or critical CodeQL alert.
   Organization admins can bypass it only through a pull request.

### Releasing the SDK

1. A maintainer runs [release-train.yaml](.github/workflows/release-train.yaml)
   with a patch, minor or major bump.
   It opens a `release: component/vX.Y.Z` pull request
   that moves the examples, the template, `airgap/versions.env` and BUILDING.md
   to the new version.
2. The pull request is reviewed and merged like any other change.
3. [release-tag.yaml](.github/workflows/release-tag.yaml) then runs on `main`:
   - re-checks every example's `go.sum` against the merged commit,
     and pushes nothing if `component/` changed since the train ran
   - pushes the annotated `component/vX.Y.Z` tag at the merge commit,
     refusing a tag that already exists elsewhere
   - checks the published module against `proxy.golang.org` and `sum.golang.org`
   - creates the GitHub Release as a draft with generated notes
   - builds the module zip from the tag, checks it matches the proxy's,
     and attaches it with `SHA256SUMS` and a Sigstore bundle of their SLSA build provenance
   - publishes the release only after the air-gapped build on the tag passes
4. The air-gapped build on the tag also publishes the `ghcr.io/wasmcloud/go-airgap`
   toolchain image,
   with signed build provenance and SBOM attestations pushed alongside it.

The tag makes the module public:
the proxy serves any pushed tag,
and the checksum database fixes its hash from then on.

### Using the SDK

1. `go get go.wasmcloud.dev/component@vX.Y.Z`
   resolves the module path through `go.wasmcloud.dev`,
   downloads it from the module proxy,
   and checks it against the checksum database.
2. `go tool componentize-go build`, or `wash build`,
   generates the bindings for the world the component targets,
   compiles it with Go for `wasip1`
   and packs it into a WebAssembly component.
   componentize-go downloads its toolchain binary from its GitHub Releases,
   and for an async WASI P3 world, a patched Go toolchain;
   [BUILDING.md](BUILDING.md) pins both.
3. `wash dev`, or a wasmCloud host, runs the component
   and provides the interfaces it imports.

`wash new https://github.com/wasmCloud/go.git --subfolder templates/http-hello-world`
copies the template into a new directory instead.
From then on the project belongs to the user.

## External interfaces

The SDKs have no interfaces beyond the
[WIT](https://component-model.bytecodealliance.org/design/wit.html)
interfaces a component imports and exports, and the Go packages that bind them.

### go.wasmcloud.dev/component

[component/wit/world.wit](component/wit/world.wit) declares the worlds a component targets:

| World | Exports | Imports |
| --- | --- | --- |
| `wasip2` (default) | `wasi:http/incoming-handler@0.2.8` | `wasi:cli/imports@0.2.8`, `wasi:http/outgoing-handler@0.2.8`, `wasi:config/store@0.2.0-rc.1`, `wasi:logging/logging@0.1.0-draft` |
| `wasip3` | `wasi:http/service@0.3.0` | the HTTP client, CLI, clocks and random imports that `wasi:http/service@0.3.0` includes, `wasi:config/store@0.2.0-rc.1`, `wasi:logging/logging@0.1.0-draft` |
| `headless` | | `wasi:cli/imports@0.2.8`, `wasi:config/store@0.2.0-rc.1`, `wasi:logging/logging@0.1.0-draft` |

Its packages also bind the wasmCloud host plugin interfaces in
[component/wit/capabilities.wit](component/wit/capabilities.wit).
A component imports one only by declaring it in its own world:

| Package | WIT |
| --- | --- |
| `secrets` | `wasmcloud:secrets@2.1.0` |
| `keyvalue` | `wasmcloud:keyvalue@0.2.0` |
| `messaging` | `wasmcloud:messaging@0.3.0`, with an exported handler for incoming messages |
| `postgres` | `wasmcloud:postgres@0.2.0` |
| `blobstore` | `wasmcloud:blobstore@0.1.0` |
| `nats` | `wasmcloud:nats@0.1.0`, with exported handlers for core, JetStream and KV deliveries |

### go.wasmcloud.dev/plugin

[plugin/wit/world.wit](plugin/wit/world.wit) binds the interfaces of a host plugin:
it imports `wasmcloud:host/identity`, `wasmcloud:host/cancel`
and `wasmcloud:host/workload-call` (`0.1.3`),
and can export `wasmcloud:host/workload-lifecycle@0.1.3`.
The plugin's own capability interface is in its own WIT.

### Template and examples

- **HTTP**: `wash dev` serves components that export `wasi:http`
  on `127.0.0.1:8000`.
  [http-hello-world](templates/http-hello-world/) answers every path with a greeting.
- **Outgoing HTTP**: `http-client` and `http-p3-streaming` make outgoing requests.
- **Key-value**: `http-keyvalue-crud` imports `wasi:keyvalue/store@0.2.0-draft`.
- **Tracing**: `http-otel` imports `wasi:otel@0.2.0-rc.2`.
- **NATS**: the `nats-*` examples import `wasmcloud:nats@0.1.0`
  through the host's NATS plugin, and some export its handlers.

Each example declares its interfaces in its `wit/` directory and README.
