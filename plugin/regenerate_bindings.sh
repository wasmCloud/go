#!/bin/bash
set -euo pipefail

# Regenerates the committed componentize-go bindings for the
# go.wasmcloud.dev/plugin module.
#
# Import bindings (wasmcloud:host/identity, wasmcloud:host/cancel, and the
# types used by wasmcloud:host/workload-lifecycle) are generated into
# imports/; the export glue for the workload-lifecycle interface is generated
# into exports/. The handwritten trampoline package
# (export_wasmcloud_host_0_1_0_workload_lifecycle) is preserved; only the
# generated wit_exports directory is refreshed.
#
# Uses the componentize-go binary from PATH by default; override with e.g.
#   COMPONENTIZE_GO=~/source/bytecodealliance/componentize-go/target/debug/componentize-go ./regenerate_bindings.sh
COMPONENTIZE_GO="${COMPONENTIZE_GO:-componentize-go}"

MODULE=go.wasmcloud.dev/plugin

WORLDS=(
  "wasmcloud:plugin-go/host-imports@0.1.0"
  "wasmcloud:plugin-go/lifecycle@0.1.0"
)

# Generate the union of import bindings for all worlds, discarding exports.
world_flags=()
for world in "${WORLDS[@]}"; do
  world_flags+=(-w "$world")
done

rm -rf imports
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

# Export glue for the lifecycle world, deferring to the shared imports above.
EXPORT_WORLDS=(
  "wasmcloud:plugin-go/lifecycle@0.1.0"
)

for world in "${EXPORT_WORLDS[@]}"; do
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

  # Allow the package to compile on non-wasm hosts despite bodyless
  # //go:wasmimport declarations (same trick as the generated imports).
  if [ ! -f "$dir/wit_exports/empty.s" ]; then
    cat > "$dir/wit_exports/empty.s" <<'EOF'
// This file exists for testing this package without WebAssembly,
// allowing empty function bodies with a //go:wasmimport directive.
// See https://pkg.go.dev/cmd/compile for more information.
EOF
  fi
done

go mod tidy
