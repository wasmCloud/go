#!/bin/bash
set -euo pipefail

# Regenerates the committed componentize-go export glue for this SDK.
#
# The shared wasi:* import bindings live in go.bytecodealliance.org/pkg
# under imports/ and are regenerated there; this script only generates the
# per-world export glue under exports/, pointing the generated code at the
# go-pkg import bindings.
#
# Uses the componentize-go binary from PATH by default; override with e.g.
#   COMPONENTIZE_GO=~/source/bytecodealliance/componentize-go/target/debug/componentize-go ./regenerate_bindings.sh
COMPONENTIZE_GO="${COMPONENTIZE_GO:-componentize-go}"

MODULE=go.wasmcloud.dev/component
IMPORTS_PKG=go.bytecodealliance.org/pkg/imports

WORLDS=(
  "wasmcloud:component-go/wasip2@0.2.0"
  "wasmcloud:component-go/wasip3@0.2.0"
)

# Import bindings for the wasmCloud capability interfaces (wasmcloud:secrets,
# wasmcloud:keyvalue, wasmcloud:messaging, wasmcloud:postgres,
# wasmcloud:blobstore), generated from the wasmcloud-capabilities world in
# wit/capabilities.wit into imports/. These are this module's own bindings
# (the wasi:* import bindings live in go.bytecodealliance.org/pkg/imports).
CAPABILITIES_WORLD="wasmcloud:component-go/wasmcloud-capabilities@0.2.0"

rm -rf imports
"$COMPONENTIZE_GO" \
  --ignore-toml-files \
  -w "$CAPABILITIES_WORLD" \
  -d wit \
  bindings \
  --format \
  -o imports \
  --pkg-name "$MODULE/imports" \
  --include-versions
rm -r imports/wit_exports

# Export glue for the optional wasmcloud:messaging/handler callback. The
# handwritten trampoline package (export_wasmcloud_messaging_0_2_0_handler)
# is preserved; only the generated wit_exports directory is refreshed.
EXPORT_WORLDS=(
  "wasmcloud:component-go/wasmcloud-messaging-handler@0.2.0"
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

# For each supported world, generate bindings specific to that world. We keep
# only the exports, deferring to the shared import bindings from
# go.bytecodealliance.org/pkg/imports.
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
    --pkg-name "$IMPORTS_PKG" \
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
