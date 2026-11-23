---
name: wasmcloud-go-code
description: Code reuse, readability, and Go style conventions for the wasmCloud Go repository, covering reusing existing helpers and dependencies, keeping abstractions small, comment style, gofmt, and the golangci-lint configuration. Use when implementing, refactoring, or reviewing Go code in component/, plugin/, examples/, templates/, or scripts/.
---

# Code changes

Follow the core rules in [AGENTS.md](../../../AGENTS.md), including searching for existing code first.

## Reuse and structure

- Prefer the standard library and existing dependencies before adding a dependency.
  Follow the [dependency policy](../../../CONTRIBUTING.md#dependencies) when a new one is needed.
- Extend existing scripts, Makefile targets, and test harnesses before introducing new infrastructure.
  Keep additions focused on the current requirement;
  avoid parallel implementations, generic frameworks, and configuration layers for hypothetical future needs.
- Extract shared code when it removes meaningful duplication between existing callers.
  Keep the abstraction small and preserve package boundaries;
  do not force unrelated code through a shared helper just because it looks similar.
- Examples are standalone modules meant to be copied.
  Do not make one example import another, or add shared example helpers;
  a little duplication between examples is expected.

## Comments

- Do not leave maintainer- or reviewer-facing comments in the code.
  Comments describe what the code does and why,
  not the history of the change, the review that prompted it, or the task you were given.
- Give new exported identifiers in `component/` and `plugin/` a doc comment that starts with their name.
  A new package gets a `// Package <name> ...` comment, in a `doc.go` when the package has more than one file.
- Deprecate an exported API with a `// Deprecated: use X instead.` paragraph
  for at least one release before removing it.
- Avoid run-on blocks of code.
  Group statements semantically, separated by blank lines.
- Put the explanation of a dense chunk of code before it, not inside or after it.

## Go style

- Format with `gofmt`; a file that `gofmt -l` lists is not ready.
- [`.golangci.yml`](../../../.golangci.yml) is the lint configuration.
  CI does not run it, but run `golangci-lint run ./...` in the module you changed when it is installed,
  and fix new findings in code you touched.
- Use MixedCaps names, and keep initialisms in one case (`ID`, `URL`, `HTTP`, not `Id` or `Url`).
  Do not repeat the package name in exported names: `keyvalue.Store`, not `keyvalue.KeyValueStore`.
  Keep receiver names short and the same across a type's methods.
- Return errors rather than panicking, and wrap them with `fmt.Errorf("...: %w", err)` when adding context.
  Error strings start lowercase, do not end with punctuation,
  and in the SDK start with the package name, as in `keyvalue: no such store`.
- Name sentinel errors `ErrXxx` and error types `XxxError`, as in [`keyvalue/errors.go`](../../../component/keyvalue/errors.go).
  Callers check errors with `errors.Is` and `errors.As`, never by comparing strings.
- Take `context.Context` as the first parameter, except after `*testing.T`.
  Pass it explicitly; do not store it in a struct.
- Write tests with the standard `testing` package; do not add an assertion library.
  Prefer table-driven tests with `t.Run` subtests,
  and report failures as `t.Errorf("F(%v) = %v, want %v", in, got, want)`.
- Mark test helpers with `t.Helper()`.
- Use `net/http` constants such as `http.MethodPost` and `http.StatusOK` rather than literals.
