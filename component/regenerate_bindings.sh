#!/bin/bash
set -euo pipefail

# Regenerates the committed componentize-go bindings for this SDK.
#
# Uses the componentize-go binary from PATH by default; override with e.g.
#   COMPONENTIZE_GO=~/source/bytecodealliance/componentize-go/target/debug/componentize-go ./regenerate_bindings.sh
COMPONENTIZE_GO="${COMPONENTIZE_GO:-componentize-go}"

MODULE=go.wasmcloud.dev/component

WORLDS=(
  "wasmcloud:component-go/wasip2@0.2.0"
  "wasmcloud:component-go/wasip3@0.2.0"
)

# Remove any previously-generated import bindings.
rm -rf imports

# Generate bindings for all supported worlds. We only keep the imports from
# this pass, discarding the exports.
world_flags=()
for world in "${WORLDS[@]}"; do
  world_flags+=(-w "$world")
done

"$COMPONENTIZE_GO" \
  --ignore-toml-files \
  "${world_flags[@]}" \
  -d wit \
  bindings \
  --format \
  -o imports \
  --pkg-name "$MODULE/imports" \
  --include-versions

rm -r imports/wit_exports

# For each supported world, generate bindings specific to that world. We keep
# only the exports, deferring to the shared imports generated above.
for world in "${WORLDS[@]}"; do
  rm -rf tmp
  dir=exports/$(echo "$world" | sed 's+[:/@.-]+_+g')
  rm -rf "$dir/wit_exports"
  mkdir -p "$dir"
  "$COMPONENTIZE_GO" \
    --ignore-toml-files \
    -w "$world" \
    -d wit \
    bindings \
    --format \
    -o tmp \
    --export-pkg-name "$MODULE/$dir" \
    --pkg-name "$MODULE/imports" \
    --include-versions
  cp -r tmp/wit_exports "$dir/"
  rm -rf tmp
done

go mod tidy
