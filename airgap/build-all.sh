#!/usr/bin/env bash
#
# Build every example and template in the repo, with no network.
#
# Runs as the entrypoint of the air-gapped image (see airgap/Dockerfile) with
# the repo bind-mounted at /workspace. Collects failures and reports all of them
# at the end rather than stopping at the first, so one broken example does not
# hide the state of the other ten.
#
# AIRGAP_USE_LOCAL_SDK=1 builds against the SDK in the checkout instead of the
# released module. Off by default: the point of this image is to verify what a
# user cloning the repo actually gets.

set -uo pipefail

WORKSPACE=${WORKSPACE:-/workspace}
USE_LOCAL_SDK=${AIRGAP_USE_LOCAL_SDK:-0}

cd "$WORKSPACE" || { echo "no workspace at $WORKSPACE" >&2; exit 1; }

# Everything that produces a component. The two nats-request-reply halves are
# separate modules, so the glob has to reach a level deeper.
MODULES=()
while IFS= read -r dir; do
  MODULES+=("$dir")
done < <(find examples/components templates -name go.mod 2>/dev/null | xargs -n1 dirname | sort)

if [ ${#MODULES[@]} -eq 0 ]; then
  echo "found no modules under examples/components or templates" >&2
  exit 1
fi

echo "Building ${#MODULES[@]} modules with no network."
echo "  SDK mode:  $([ "$USE_LOCAL_SDK" = "1" ] && echo 'in-repo (replace applied)' || echo 'released')"
echo "  GOPROXY:   ${GOPROXY:-<unset>}"
echo "  GOTOOLCHAIN: ${GOTOOLCHAIN:-<unset>}"
echo

# A filesystem replace needs no go.sum entry of its own, but go.mod has to be
# writable for the edit, so make sure -mod=mod is in play.
if [ "$USE_LOCAL_SDK" = "1" ]; then
  export GOFLAGS="${GOFLAGS:-} -mod=mod"
fi

passed=()
failed=()

# The workspace is a bind mount of a real checkout, so every edit this script
# makes has to be undone on every exit path, including the early ones.
#
# Snapshot and restore rather than trying to reverse each edit: `go mod edit
# -dropreplace` and `go mod tidy` both reformat go.mod, so "undoing" an edit
# still leaves the file different from how it was found.
snapshot() {
  local dir=$1
  cp "$dir/go.mod" "$dir/go.mod.airgap-orig"
  [ -f "$dir/go.sum" ] && cp "$dir/go.sum" "$dir/go.sum.airgap-orig"
  return 0
}

restore() {
  local dir=$1
  [ -f "$dir/go.mod.airgap-orig" ] && mv "$dir/go.mod.airgap-orig" "$dir/go.mod"
  if [ -f "$dir/go.sum.airgap-orig" ]; then
    mv "$dir/go.sum.airgap-orig" "$dir/go.sum"
  else
    # No go.sum before we started, so any that exists now is ours.
    rm -f "$dir/go.sum"
  fi
  return 0
}

for dir in "${MODULES[@]}"; do
  printf '=== %s\n' "$dir"

  # Snapshot before touching anything; restored on every path out.
  snapshot "$dir"

  if [ "$USE_LOCAL_SDK" = "1" ]; then
    (cd "$dir" && go mod edit -replace "go.wasmcloud.dev/component=${WORKSPACE}/component") || {
      failed+=("$dir (could not apply replace)")
      printf '    FAIL\n\n'
      restore "$dir"
      continue
    }
  fi

  # Templates ship without a go.sum: they are consumed through `wash new`,
  # which tidies the generated project before building it. Reproduce that here
  # rather than skipping the template — whether a tidy can be satisfied from
  # the module cache alone is exactly the kind of thing this test should catch.
  if [ ! -f "$dir/go.sum" ]; then
    echo "    no go.sum (template); running go mod tidy from cache"
    if ! (cd "$dir" && go mod tidy); then
      failed+=("$dir (go mod tidy could not be satisfied offline)")
      printf '    FAIL\n\n'
      restore "$dir"
      continue
    fi
  fi

  # --skip-fetch is not optional here: without it wash tries to resolve WIT
  # dependencies from an OCI registry. wit/deps is not checked in; it is
  # fetched into the checkout beforehand by fetch-wit.sh (`make airgap-fetch`),
  # so a missing one is a skipped step, not a build failure worth diagnosing.
  if [ -d "$dir/wit" ] && [ -f "$dir/wkg.lock" ] && [ ! -d "$dir/wit/deps" ]; then
    failed+=("$dir (no wit/deps; run \`make airgap-fetch\` first)")
    printf '    FAIL\n\n'
    restore "$dir"
    continue
  fi

  if (cd "$dir" && wash build --skip-fetch); then
    passed+=("$dir")
    printf '    PASS\n\n'
  else
    failed+=("$dir")
    printf '    FAIL\n\n'
  fi

  restore "$dir"
done

echo "======================================================================"
printf 'passed: %d/%d\n' "${#passed[@]}" "${#MODULES[@]}"

if [ ${#failed[@]} -gt 0 ]; then
  echo
  echo "failed:"
  printf '  - %s\n' "${failed[@]}"
  echo
  echo "A failure here means the build needed something the documented version"
  echo "set does not provide. Either the code grew a new dependency, or"
  echo "BUILDING.md is out of date."
  exit 1
fi

echo "All modules built with no network."
