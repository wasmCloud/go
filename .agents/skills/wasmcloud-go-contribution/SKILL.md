---
name: wasmcloud-go-contribution
description: Commit, pull request, release, and documentation conventions for the wasmCloud Go repository, covering Conventional Commits checked by commitlint, the 69-character header limit including the squash suffix, DCO sign-off, breaking-change markers, draft PRs, rebasing, and semantic line breaks. Use when writing a commit message, preparing or reviewing a pull request, checking PR readiness, or editing Markdown docs such as README.md, CONTRIBUTING.md, BUILDING.md, or example READMEs.
---

# Contribution conventions

Read only the sections relevant to the current task.
These conventions describe how to prepare a contribution; they do not authorize creating a PR or publishing changes.

## Commits

- Write titles as [Conventional Commits](https://www.conventionalcommits.org),
  for example `fix(keyvalue): return ErrNoSuchStore for a missing bucket`.
  [commitlint](../../../.github/commitlint/commitlint.config.mjs) checks every commit and the PR title.
- Headers are limited to 69 characters,
  and a squash merge appends ` (#NNN)` to the PR title,
  so keep PR titles to about 60 characters.
- Use a scope that names the package, example, or area changed,
  such as `keyvalue`, `plugin`, `http-server`, `airgap`, or `ci`.
  The `release` type is reserved for the release train.
- Keep commits small, and organize them around semantically distinct changes.
- Add a body only when the title and diff do not make the reason for the change clear.
  Keep it to a few lines on why, and wrap it at 70 characters.
- Sign off every commit with `git commit -s`.
  The DCO `Signed-off-by:` trailer and any required `BREAKING CHANGE:` footer
  do not count as a body.
- Never merge `main` into a branch.
  Rebase the branch onto `main` instead, including to resolve conflicts.

## Releases

SDK releases are cut by the [release train](../../../.github/workflows/release-train.yaml),
which opens a `release: component/vX.Y.Z` PR; do not hand-write release commits or tags.

Mark every commit that introduces a breaking change to `component/` or `plugin/` in both places:

- A `!` in the title, for example `feat(messaging)!: drop the legacy handler`.
- A `BREAKING CHANGE: <what breaks and how to migrate>` footer.

## Pull requests

- After local tests and applicable lint and format checks pass, open a pull request as a draft,
  and iterate on it until CI is green.
  Re-run the relevant local checks after each fix before pushing it.
- If a personal fork is configured as a remote, open the draft against the fork first.
  Open the pull request against upstream once it is ready.
- Keep the body succinct and human readable.
  A few sentences on what changed and why is usually enough.
- Do not use an excessive number of sections.
- Do not include verification methodology
  (test commands, "Validation" sections, lint results)
  unless explicitly asked to.
  The [test policy](../../../CONTRIBUTING.md#tests-for-new-changes) still applies:
  SDK behavior changes need unit tests,
  bug fixes need a test that fails without the fix,
  and new examples must build in CI.
  Explain why a change needs no new tests, or why a local check could not run.

## Documentation

- Use [semantic line breaks](https://sembr.org) instead of run-on paragraphs.
- Prefer a list to a dense paragraph that covers several facets in prose.
- Keep pinned versions in [`airgap/versions.env`](../../../airgap/versions.env);
  BUILDING.md describes them, and `make airgap-versions` checks the two agree.
