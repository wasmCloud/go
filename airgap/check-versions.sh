#!/usr/bin/env bash
#
# Re-derive every pinned version from its real source and compare against
# airgap/versions.env and BUILDING.md.
#
# This is the thing that stops the documentation rotting. Each value is read
# from wherever it actually lives — a go.mod, a git tag, a Rust manifest inside
# a module in GOMODCACHE — never copied from the file being checked.
#
# Needs a network on first run (to populate GOMODCACHE); offline afterwards.

set -uo pipefail

REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$REPO_ROOT" || exit 1

# shellcheck source=versions.env
set -a; . airgap/versions.env; set +a

problems=0

# check <label> <expected> <actual>
check() {
  local label=$1 expected=$2 actual=$3
  if [ "$expected" = "$actual" ]; then
    printf '  ok    %-28s %s\n' "$label" "$actual"
  else
    printf '  DRIFT %-28s pinned=%s actual=%s\n' "$label" "$expected" "$actual"
    problems=$((problems + 1))
  fi
}

note() { printf '  --    %-28s %s\n' "$1" "$2"; }

echo "Resolving pinned versions from their sources..."
echo

# --- SDK: the newest component/* tag in the repo ---------------------------
actual_sdk=$(git tag -l 'component/v*' --sort=-v:refname | head -n1 | sed 's|^component/||')
check "SDK (latest tag)" "$SDK_VERSION" "$actual_sdk"

# --- every consumer module must require exactly that ------------------------
# Plain word-splitting rather than mapfile: this script runs on developer
# machines, and macOS still ships bash 3.2.
# shellcheck disable=SC2207
sdk_pins=($(grep -rhoE 'go\.wasmcloud\.dev/component v[0-9][^ ]*' \
  --include=go.mod examples templates | awk '{print $2}' | sort -u))
if [ "${#sdk_pins[@]}" -eq 1 ] && [ "${sdk_pins[0]}" = "$SDK_VERSION" ]; then
  printf '  ok    %-28s all modules at %s\n' "SDK (module pins)" "$SDK_VERSION"
else
  printf '  DRIFT %-28s expected all at %s, found: %s\n' \
    "SDK (module pins)" "$SDK_VERSION" "${sdk_pins[*]}"
  problems=$((problems + 1))
fi

# --- go-pkg -----------------------------------------------------------------
actual_gopkg=$(grep -oE 'go\.bytecodealliance\.org/pkg v[^ ]*' component/go.mod | awk '{print $2}')
check "go-pkg" "$GO_PKG" "$actual_gopkg"

# --- componentize-go: module vs binary --------------------------------------
actual_cg_mod=$(grep -rhoE 'github\.com/bytecodealliance/componentize-go v[^ ]*' \
  --include=go.mod examples | awk '{print $2}' | sort -u | head -n1)
check "componentize-go (module)" "$COMPONENTIZE_GO_MODULE" "$actual_cg_mod"

# The binary is named by a string hardcoded in the module's main.go. This is the
# whole point of the check: nothing in the module graph reveals it.
modcache=$(go env GOMODCACHE)
cg_mod_dir="${modcache}/github.com/bytecodealliance/componentize-go@${COMPONENTIZE_GO_MODULE}"
if [ ! -d "$cg_mod_dir" ]; then
  go mod download "github.com/bytecodealliance/componentize-go@${COMPONENTIZE_GO_MODULE}" 2>/dev/null
fi
if [ -f "${cg_mod_dir}/main.go" ]; then
  actual_cg_bin=$(grep -oE 'release := "[^"]+"' "${cg_mod_dir}/main.go" | sed 's/.*"\(.*\)"/\1/')
  check "componentize-go (binary)" "$COMPONENTIZE_GO_BINARY" "$actual_cg_bin"

  # The invariant BUILDING.md documents: pin a tag, and the module and the
  # toolchain agree. A pseudo-version names the release *before* it, so this
  # catches a well-meaning bump to a `main` snapshot, which would silently
  # build everything with an older wit-bindgen.
  if [ "$actual_cg_mod" = "$actual_cg_bin" ]; then
    printf '  ok    %-28s module == toolchain (%s)\n' "componentize-go (pinned tag)" "$actual_cg_bin"
  else
    printf '  DRIFT %-28s module %s fetches toolchain %s — pin a tag, not a pseudo-version\n' \
      "componentize-go (pinned tag)" "$actual_cg_mod" "$actual_cg_bin"
    problems=$((problems + 1))
  fi
else
  note "componentize-go (binary)" "SKIPPED — module not in GOMODCACHE"
fi

# --- wit-bindgen / wit-component / patched Go -------------------------------
#
# All three live in the *binary's* source tree, not the module we depend on.
# Reading them from the wrong version is exactly the mistake this script exists
# to catch, so resolve COMPONENTIZE_GO_BINARY explicitly.
cg_bin_dir="${modcache}/github.com/bytecodealliance/componentize-go@${COMPONENTIZE_GO_BINARY}"
if [ ! -d "$cg_bin_dir" ]; then
  go mod download "github.com/bytecodealliance/componentize-go@${COMPONENTIZE_GO_BINARY}" 2>/dev/null
fi

if [ -f "${cg_bin_dir}/Cargo.toml" ]; then
  actual_rev=$(grep -oE 'wit-bindgen-go = \{ git = "[^"]*", rev = "[^"]*"' "${cg_bin_dir}/Cargo.toml" \
    | grep -oE 'rev = "[^"]*"' | sed 's/.*"\(.*\)"/\1/')
  check "wit-bindgen rev" "$WIT_BINDGEN_REV" "$actual_rev"

  actual_witcomp=$(grep -oE '^wit-component = "[^"]*"' "${cg_bin_dir}/Cargo.toml" | sed 's/.*"\(.*\)"/\1/')
  check "wit-component" "$WIT_COMPONENT" "$actual_witcomp"

  actual_wbg=$(grep -A2 'name = "wit-bindgen-go"' "${cg_bin_dir}/Cargo.lock" \
    | grep -oE '^version = "[^"]*"' | sed 's/.*"\(.*\)"/\1/')
  check "wit-bindgen-go" "$WIT_BINDGEN_GO" "$actual_wbg"

  actual_patched=$(grep -oE 'dicej/go/releases/download/[^/]+/' "${cg_bin_dir}/src/utils.rs" \
    | head -n1 | sed 's|.*/download/\(.*\)/|\1|')
  check "patched Go tag" "$PATCHED_GO_TAG" "$actual_patched"
else
  note "wit-bindgen / patched Go" "SKIPPED — ${COMPONENTIZE_GO_BINARY} not in GOMODCACHE"
fi

# --- Go directive -----------------------------------------------------------
actual_go_directive=$(grep -oE '^go [0-9]+\.[0-9]+' component/go.mod | awk '{print $2}')
case "$GO_VERSION" in
  "$actual_go_directive".*|"$actual_go_directive")
    printf '  ok    %-28s %s satisfies directive go %s\n' \
      "Go" "$GO_VERSION" "$actual_go_directive" ;;
  *)
    printf '  DRIFT %-28s pinned %s does not match directive go %s\n' \
      "Go" "$GO_VERSION" "$actual_go_directive"
    problems=$((problems + 1)) ;;
esac

# --- committed bindings must match the pinned wit-bindgen -------------------
#
# Two examples commit generated bindings. The wit-bindgen version travels with
# the componentize-go release, so bumping the toolchain silently leaves these
# stale until someone reruns examples/components/regenerate_bindings.sh.
stale_bindings=""
while IFS= read -r f; do
  # Header reads: // Generated by `wit-bindgen` 0.61.1. DO NOT EDIT!
  # Anchor on the trailing ". DO NOT" so the version's own dots are not eaten.
  header_version=$(sed -n '1s/.*`wit-bindgen` \(.*\)\. DO NOT.*/\1/p' "$f")
  [ -z "$header_version" ] && continue
  if [ "$header_version" != "$WIT_BINDGEN_GO" ]; then
    stale_bindings="${stale_bindings}\n      ${f} (${header_version})"
  fi
done < <(git ls-files 'examples/**/wit_bindings.go' 'examples/**/wit_exports.go')

if [ -z "$stale_bindings" ]; then
  printf '  ok    %-28s all committed headers at %s\n' "generated bindings" "$WIT_BINDGEN_GO"
else
  printf '  DRIFT %-28s expected %s; run examples/components/regenerate_bindings.sh%b\n' \
    "generated bindings" "$WIT_BINDGEN_GO" "$stale_bindings"
  problems=$((problems + 1))
fi

# --- BUILDING.md must repeat the same numbers -------------------------------
echo
echo "Checking BUILDING.md agrees..."
echo
for pair in \
  "SDK:${SDK_VERSION}" \
  "go-pkg:${GO_PKG}" \
  "componentize-go module:${COMPONENTIZE_GO_MODULE}" \
  "componentize-go binary:${COMPONENTIZE_GO_BINARY}" \
  "wit-component:${WIT_COMPONENT}" \
  "wit-bindgen-go:${WIT_BINDGEN_GO}" \
  "patched Go:${PATCHED_GO_TAG}" \
  "wash:${WASH_VERSION}" \
  "Go:${GO_VERSION}"
do
  label=${pair%%:*}
  value=${pair#*:}
  if grep -qF "$value" BUILDING.md; then
    printf '  ok    %-28s %s\n' "$label" "$value"
  else
    printf '  DRIFT %-28s %s not found in BUILDING.md\n' "$label" "$value"
    problems=$((problems + 1))
  fi
done

echo
if [ "$problems" -gt 0 ]; then
  echo "${problems} version(s) drifted. Update airgap/versions.env and BUILDING.md."
  exit 1
fi
echo "All pinned versions match their sources."
