#!/usr/bin/env bash
#
# Check that every example's go.sum carries the real checksum of an SDK release.
#
#   airgap/verify-sdk-sums.sh --local vX.Y.Z      # before the tag is pushed
#   airgap/verify-sdk-sums.sh --published vX.Y.Z  # after
#
# A release PR writes go.sum entries for a version that is not tagged yet (see
# local-sdk-env.sh). Those entries are only right if component/ is byte-for-byte
# what gets tagged, so release-tag.yaml runs both modes around the tag push:
#
#   --local      Resolve from this checkout at HEAD (tagging HEAD locally if
#                the tag is missing). A mismatch means component/ changed on
#                main after the release PR was prepared: nothing is pushed.
#   --published  Resolve through proxy.golang.org, which also checks
#                sum.golang.org. A mismatch here would be a real integrity
#                failure and should never happen after --local passed.

set -euo pipefail

REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$REPO_ROOT"

mode=${1:-}
version=${2:-}
if ! [[ "$mode" =~ ^--(local|published)$ && "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "usage: $0 --local|--published vX.Y.Z" >&2
  exit 2
fi

module=go.wasmcloud.dev/component
tag="component/${version}"

# A throwaway module cache, so nothing already cached can answer for the proxy
# or for the checkout. -modcacherw keeps it deletable.
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT
export GOMODCACHE="${scratch}/mod" GOFLAGS=-modcacherw

if [ "$mode" = "--local" ]; then
  if ! git rev-parse --verify --quiet "refs/tags/${tag}" >/dev/null; then
    git tag "$tag" HEAD
    echo "created local-only tag ${tag} at $(git rev-parse --short HEAD)"
  fi
  # shellcheck source=local-sdk-env.sh
  . airgap/local-sdk-env.sh
  local_sdk_env "$REPO_ROOT"
else
  export GOPROXY=https://proxy.golang.org GOSUMDB=sum.golang.org GOPRIVATE='' GONOSUMDB='' GONOPROXY=''
fi

# Outside the repo, so no go.mod is in play.
json=""
for attempt in 1 2 3 4 5 6; do
  if json=$(cd "$scratch" && go mod download -json "${module}@${version}"); then
    break
  fi
  json=""
  [ "$attempt" -lt 6 ] || break
  echo "download attempt ${attempt}/6 failed; retrying in 30s" >&2
  sleep 30
done
if [ -z "$json" ]; then
  echo "could not resolve ${module}@${version} (${mode#--})" >&2
  exit 1
fi

field() { printf '%s\n' "$json" | sed -nE "s/^[[:space:]]*\"$1\": \"([^\"]*)\".*/\\1/p" | head -n1; }
sum=$(field Sum)
gomodsum=$(field GoModSum)
echo "${module}@${version} (${mode#--}): ${sum}, go.mod ${gomodsum}"

problems=0
while IFS= read -r gosum; do
  if grep -qxF "${module} ${version} ${sum}" "$gosum" &&
     grep -qxF "${module} ${version}/go.mod ${gomodsum}" "$gosum"; then
    echo "  ok    ${gosum}"
  else
    echo "  WRONG ${gosum}"
    grep -F "${module} ${version}" "$gosum" | sed 's/^/          has: /' || echo "          has: no entry for ${version}"
    problems=$((problems + 1))
  fi
done < <(find examples/components -name go.sum | sort)

if [ "$problems" -gt 0 ]; then
  echo "${problems} go.sum file(s) disagree with ${mode#--} ${module}@${version}" >&2
  exit 1
fi
