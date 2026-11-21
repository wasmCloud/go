---
name: wasmcloud-go-automation
description: Conventions for CI and developer automation in the wasmCloud Go repository, covering where scripts live, airgap/ scripts and Makefile targets, workflow path filters and gate jobs, action pinning, permissions, zizmor, and workflow YAML style. Use when creating or changing files under .github/workflows/, .github/commitlint/, airgap/, scripts/, or the Makefile; not needed just to run existing checks.
---

# CI and developer automation

Read the sections that match the files being changed.

## Task tooling

- Extend the existing [`Makefile`](../../../Makefile) and `airgap/` scripts before adding another tool.
  Keep `Makefile` targets thin wrappers with a `## description` for `make help`.
- Prefer a script that CI calls, and that is just as easy to run locally, over logic inlined in a workflow.
- Extend the existing bash scripts in `airgap/` in their current style.
  For new logic that is more than glue, add a Go command and run it with `go run`.

## Workflows

- Pin every action to a full commit SHA with the version in a trailing comment.
- Start each workflow with `permissions: {}` and grant each job only what it needs.
- Check out with `persist-credentials: false`.
- Pass `${{ }}` expressions into `run:` steps through `env:`, not by interpolating them into the script.
  [zizmor](../../../.github/workflows/zizmor.yml) checks this on every PR.
- Workflows that filter on changed paths keep a gate job
  (`component-gate`, `templates-gate`, `airgapped-gate`) that always reports,
  because required checks must run even when every other job is skipped.
  A new path-filtered workflow needs the same, and a new path must be added to the filter that should test it.
- Never write YAML lists in the inline `[a, b]` form.
  Write one item per line so that edits produce small diffs:

  ```yaml
  branches:
    - main
  ```
