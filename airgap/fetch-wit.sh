#!/usr/bin/env bash
#
# Fetch WIT dependencies for every example and template, WITH network.
#
# The one step of the air-gapped suite that is allowed a network. wit/deps is
# not checked in — it is resolved from wkg.lock, the same way go.sum is resolved
# into GOMODCACHE — so it has to land in the checkout before build-all.sh runs
# `wash build --skip-fetch` with `--network none`.
#
# It runs in the checkout rather than being baked into the image: the image is
# tagged by a hash of airgap/versions.env, which does not cover wkg.lock, so
# deps baked in there would go stale without the tag ever changing.
#
# Run from the bind mount (not COPY'd into the image) for the same reason — a
# pulled image must not pin an old copy of this script.

set -uo pipefail

WORKSPACE=${WORKSPACE:-/workspace}

cd "$WORKSPACE" || { echo "no workspace at $WORKSPACE" >&2; exit 1; }

# Only projects with their own wit/ have anything to fetch; the rest take their
# worlds from the SDK module, which GOMODCACHE already covers.
DIRS=()
while IFS= read -r dir; do
  [ -d "$dir/wit" ] && DIRS+=("$dir")
done < <(find examples/components templates -name go.mod 2>/dev/null | xargs -n1 dirname | sort)

echo "Fetching WIT dependencies for ${#DIRS[@]} projects."
echo

failed=()
for dir in "${DIRS[@]}"; do
  printf '=== %s\n' "$dir"
  if (cd "$dir" && wash wit fetch); then
    printf '    OK\n\n'
  else
    failed+=("$dir")
    printf '    FAIL\n\n'
  fi
done

if [ ${#failed[@]} -gt 0 ]; then
  echo "failed:"
  printf '  - %s\n' "${failed[@]}"
  exit 1
fi
