---
name: wasmcloud-go-examples
description: Conventions for adding and changing Go example components and the wash new template, covering module layout, .wash/config.yaml, wash build and wash dev, sync P2 versus async P3 worlds, the component-go.yaml CI matrix and smoke tests, air-gapped coverage, and README updates. Use when creating or changing anything under examples/components/ or templates/.
---

# Examples and templates

## Structure

Copy the closest existing example rather than starting from scratch;
[`http-server`](../../../examples/components/http-server/) is the simplest sync HTTP example
and [`http-p3-streaming`](../../../examples/components/http-p3-streaming/) the simplest async one.
Each example is a standalone module with:

- `go.mod` requiring a released `go.wasmcloud.dev/component`, never a `replace`.
- `.wash/config.yaml` whose `build.command` is `go tool componentize-go build ...`
  and whose `component_path` matches the `-o` output.
- A `README.md` with prerequisites, how to run it with `wash dev`, and how to exercise it.
- `wit/` when it declares its own world.

Add a new example to the table in the [repository README](../../../README.md),
marking it if it targets the async P3 world.

## Sync and async worlds

- Sync WASI P2 examples build with stock Go.
- Async WASI P3 worlds, and all `nats-*` examples, need the patched Go described in
  [BUILDING.md](../../../BUILDING.md#the-patched-go).
  Copy the toolchain note from an existing async example's README.
- A build that fails with `downloaded Go does not support async` is a toolchain problem,
  not a code problem; see BUILDING.md before changing code.
- On a fresh clone, run `go mod download` before `wash build`:
  componentize-go reads the SDK's WIT from the module cache.

## CI

Add every new example to the matrix in
[`component-go.yaml`](../../../.github/workflows/component-go.yaml), in the same PR:

- `folder`, relative to `examples/components/`.
- `test-command`, a shell command that exits non-zero on failure,
  run against `wash dev` on `127.0.0.1:8000`.
  It is retried, so it must be safe to repeat.
- `skip-dev-test: true` only when no released `wash` can serve the example,
  with a comment naming what is missing.
- `nats-setup` when the example needs streams, consumers, or buckets created first;
  copy it from the README's prerequisites.

Add the module's directory to the `gomod` `directories` list in
[`.github/dependabot.yml`](../../../.github/dependabot.yml);
it lists every module explicitly.

Template changes are tested by [`templates-go.yaml`](../../../.github/workflows/templates-go.yaml),
which runs `wash new` against the template and builds it.

[`airgapped.yaml`](../../../.github/workflows/airgapped.yaml) builds every module under
`examples/components/` and `templates/` offline, discovered automatically.
A new example must build with only the versions pinned in `airgap/versions.env`;
check it with `make airgap-test AIRGAP_MODULES=examples/components/<example>` when Docker is available.

## Checks

From the example directory:

```
- [ ] gofmt -l .                     (no output)
- [ ] wash build
- [ ] wash dev, then the CI test-command from another shell
```

If `wash` or the patched Go is not installed, say which checks could not run.
