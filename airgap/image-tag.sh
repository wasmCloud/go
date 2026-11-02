#!/usr/bin/env bash
#
# Print the air-gapped image tag: a hash of every file the image's seed stage
# resolves dependencies from.
#
#   airgap/image-tag.sh
#
# The seed stage pins the toolchain from versions.env and fills GOMODCACHE from
# every example's and template's go.mod and go.sum. Hashing only versions.env
# meant a dependency bump in an example kept the old tag, so CI pulled an image
# whose module cache predated the bump and the offline build failed, with
# nothing ever rebuilding it. Covering the module files makes a cache miss mean
# the image would differ.
#
# The Makefile and airgapped.yaml both call this, so the tag a build writes and
# the tag CI pulls and pushes can't disagree.

set -euo pipefail

REPO_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$REPO_ROOT"

if command -v sha256sum >/dev/null 2>&1; then
  sum=(sha256sum)
else
  sum=(shasum -a 256)
fi

# One "<hash>  <path>" line per file, in a fixed order, hashed again. Both tools
# print the same format, so Linux and macOS agree on the tag.
{
  echo airgap/versions.env
  find examples/components templates -name go.mod -o -name go.sum
} | LC_ALL=C sort | xargs "${sum[@]}" | "${sum[@]}" | cut -c1-12
