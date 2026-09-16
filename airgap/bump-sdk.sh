#!/usr/bin/env bash
#
# Move the examples, the template and the pinned-version docs to a released SDK.
#
#   airgap/bump-sdk.sh v0.1.6
#
# Run after `component/<version>` is tagged and pushed — never before, since the
# modules have to resolve the release. Touches exactly what BUILDING.md's
# "Keeping this current" lists for an SDK release:
#
#   - SDK_VERSION and the "Verified against" line in airgap/versions.env
#   - the release link and the version table in BUILDING.md
#   - the go.wasmcloud.dev/component requirement in every example and template
#
# then runs check-versions.sh, so a bump that leaves anything disagreeing fails
# here rather than in the PR. Toolchain bumps (componentize-go, wash, the
# patched Go) are deliberately out of scope: they move several coupled values
# and can require regenerated bindings, which is a job for a person.

set -euo pipefail

REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$REPO_ROOT"

version=${1:-}
if ! [[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "usage: $0 vX.Y.Z" >&2
  exit 2
fi

tag="component/${version}"
if ! commit=$(git rev-parse --verify --quiet "${tag}^{commit}"); then
  echo "tag ${tag} not found; fetch tags first (git fetch --tags)" >&2
  exit 1
fi
short=$(git rev-parse --short=7 "$commit")

# shellcheck source=versions.env
old=$(. airgap/versions.env && echo "$SDK_VERSION")
echo "SDK ${old} -> ${version} (${tag} at ${short})"

# Portable in-place edit: BSD and GNU sed disagree on -i, so write and move.
# Versions are validated above, so they contain nothing sed treats as special
# beyond the dots, which are escaped.
subst() {
  local file=$1 expr=$2 tmp
  tmp=$(mktemp)
  sed -E "$expr" "$file" >"$tmp"
  cat "$tmp" >"$file"
  rm -f "$tmp"
}
old_re=${old//./\\.}

# --- airgap/versions.env ----------------------------------------------------
subst airgap/versions.env "s|^SDK_VERSION=.*|SDK_VERSION=${version}|"
subst airgap/versions.env "s|^# Verified against tag component/v[^ ]* \\([^)]*\\)\\.|# Verified against tag ${tag} (${short}).|"

# --- BUILDING.md ------------------------------------------------------------
# The release link, both its text (`component/vX`) and URL (component%2FvX).
subst BUILDING.md "s#component(/|%2F)${old_re}([\`)])#component\\1${version}\\2#g"
subst BUILDING.md "s#^(\\| \`go\\.wasmcloud\\.dev/component\` \\| )${old_re}( \\|)#\\1${version}\\2#"

# --- modules ----------------------------------------------------------------
#
# The proxy can briefly 404 a tag pushed moments ago, and this runs straight off
# the tag push, so retry before giving up.
with_retry() {
  local attempt
  for attempt in 1 2 3 4 5; do
    if "$@"; then return 0; fi
    [ "$attempt" -lt 5 ] || break
    echo "    attempt ${attempt}/5 failed; retrying in 30s" >&2
    sleep 30
  done
  return 1
}

while IFS= read -r mod; do
  dir=$(dirname "$mod")
  echo "=== ${dir}"
  if [ -f "${dir}/go.sum" ]; then
    (cd "$dir" && with_retry go get "go.wasmcloud.dev/component@${version}" && with_retry go mod tidy)
  else
    # Templates ship without a go.sum and are tidied by `wash new`; `go get`
    # would create one. `go mod edit` would also reformat the file and fold the
    # commented require block, so move just the one line.
    subst "${dir}/go.mod" "s|^([[:space:]]*go\\.wasmcloud\\.dev/component )v[^ ]*|\\1${version}|"
  fi
done < <(find examples/components templates -name go.mod | sort)

echo
bash airgap/check-versions.sh
