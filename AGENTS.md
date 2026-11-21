# Agent instructions

## Core rules

- Follow [CONTRIBUTING.md](CONTRIBUTING.md) and [BUILDING.md](BUILDING.md); read the sections relevant to the task.
  Check for more specific `AGENTS.md` files in the directories you change.
- Do development work on a feature branch, never directly on `main`.
- Search for existing implementations, helpers, test utilities, dependencies, and tooling before adding code.
  Reuse or extend them where their behavior and ownership fit,
  and keep new abstractions focused on the current requirement.
- Keep conversation, prompt, and iteration history out of code comments, commits, and PR descriptions.
  Comments explain the code and its rationale.
- Run relevant local tests and applicable lint and format checks before opening a PR.
  Re-run relevant checks after fixes, and report checks that could not run.
- Never hand-edit generated bindings under `component/` or `plugin/` `imports/` and `exports/wit_exports/`.
  Regenerate them as described in the SDK skill.
- A more specific `AGENTS.md` takes precedence over the skills below for the directory it covers.
  Repository instructions and checked-in lint and format configuration take precedence over external skills.

## Repository layout

Every directory with a `go.mod` is its own module; there is no `go.work`.
Run Go commands from inside the module you are changing.

| Path | What it is |
| --- | --- |
| `component/` | `go.wasmcloud.dev/component`, the SDK for wasmCloud components |
| `plugin/` | `go.wasmcloud.dev/plugin`, the SDK for host component plugins |
| `examples/components/*/` | Example components, each pinned to the released SDK |
| `templates/` | Templates for `wash new` |
| `airgap/` | Pinned toolchain versions, offline builds, and SDK release scripts |

## Task-specific guidance

Read only the skills relevant to the current task, before doing that work.
Combine skills when needed, such as code and SDK guidance for a change under `component/`.
If the harness does not discover local skills, open the linked `SKILL.md` files directly.
Follow their references only when the referenced topic is needed.

| Task | Skill |
| --- | --- |
| Implementing, refactoring, or reviewing Go code | [Code reuse and Go style](.agents/skills/wasmcloud-go-code/SKILL.md) |
| Preparing commits or PRs, checking PR readiness, or editing docs | [Contribution conventions](.agents/skills/wasmcloud-go-contribution/SKILL.md) |
| Changing `component/` or `plugin/`, their WIT, or generated bindings | [SDK changes](.agents/skills/wasmcloud-go-sdk/SKILL.md) |
| Adding or changing examples or templates | [Examples and templates](.agents/skills/wasmcloud-go-examples/SKILL.md) |
| Bumping Go, componentize-go, wit-bindgen, `wash`, or the SDK version; air-gapped builds | [Toolchain and versions](.agents/skills/wasmcloud-go-toolchain/SKILL.md) |
| CI workflows, `airgap/` scripts, or `Makefile` targets | [CI and developer automation](.agents/skills/wasmcloud-go-automation/SKILL.md) |
| Adding or updating dependencies, Dependabot PRs, advisories, or VEX | [Dependencies and advisories](.agents/skills/wasmcloud-go-dependencies/SKILL.md) |
| Designing WIT or building, composing, and running Wasm components | [Component guidance](.agents/skills/wasmcloud-go-components/SKILL.md) |
