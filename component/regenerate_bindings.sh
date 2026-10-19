#!/bin/bash
set -euo pipefail

# Regenerates the committed wit-bindgen bindings for this SDK.
#
# The shared wasi:* bindings, and the wasi:http export glue behind wasihttp,
# live in go.bytecodealliance.org/wasi and are regenerated there. This script
# only generates what is wasmCloud-specific: the import bindings for the
# wasmCloud capability interfaces under imports/, and the export glue for the
# optional handler callbacks under exports/.
#
# Mirrors go.bytecodealliance.org/wasi's regenerate_bindings.sh. Run it from
# this directory with the wit-bindgen release pinned as SDK_WIT_BINDGEN in
# airgap/versions.env (see BUILDING.md).
#
# Uses the wit-bindgen binary from PATH by default; override with e.g.
#   WIT_BINDGEN=~/source/bytecodealliance/wit-bindgen/target/debug/wit-bindgen ./regenerate_bindings.sh
WIT_BINDGEN="${WIT_BINDGEN:-wit-bindgen}"

MODULE=go.wasmcloud.dev/component

# Import bindings for the wasmCloud capability interfaces (wasmcloud:secrets,
# wasmcloud:keyvalue, wasmcloud:messaging, wasmcloud:postgres,
# wasmcloud:blobstore, wasmcloud:nats), from the wasmcloud-capabilities world
# in wit/capabilities.wit.
CAPABILITIES_WORLD="wasmcloud:component-go/wasmcloud-capabilities@0.2.0"

# Export glue for the optional handler callbacks. The handwritten trampoline
# packages (export_*) next to each wit_exports directory are preserved.
EXPORT_WORLDS=(
  "wasmcloud:component-go/wasmcloud-messaging-handler@0.2.0"
  "wasmcloud:component-go/wasmcloud-nats-core-handler@0.2.0"
  "wasmcloud:component-go/wasmcloud-nats-jetstream-handler@0.2.0"
  "wasmcloud:component-go/wasmcloud-nats-kv-handler@0.2.0"
)

generate() {
  "$WIT_BINDGEN" go \
    -w "$1" \
    --pkg-name "$MODULE" \
    --include-versions \
    --format \
    --quiet \
    --out-dir "$2" \
    wit >/dev/null
}

# Allow a package to compile on non-wasm hosts despite bodyless
# //go:wasmimport declarations (same trick as the generated imports).
add_empty_s() {
  if [ ! -f "$1/empty.s" ]; then
    cat > "$1/empty.s" <<'EOF'
// This file exists for testing this package without WebAssembly,
// allowing empty function bodies with a //go:wasmimport directive.
// See https://pkg.go.dev/cmd/compile for more information.
EOF
  fi
}

rm -rf imports tmp
mkdir -p imports

generate "$CAPABILITIES_WORLD" "tmp/$CAPABILITIES_WORLD"
cp -r "tmp/$CAPABILITIES_WORLD"/imports/* imports/

for world in "${EXPORT_WORLDS[@]}"; do
  dir=exports/$(echo "$world" | sed 's+[:/@.-]+_+g')
  out="tmp/$world"

  generate "$world" "$out"

  # Merge in the import packages the export glue references. Packages the
  # capabilities world already generated must be identical. A handler world
  # also emits alias-only packages for its own interface; those are skipped
  # unless the glue actually imports them (an interface with a `stream`
  # parameter, such as wasmcloud:messaging/handler@0.3.0, carries the stream
  # ABI intrinsics there).
  for name in $(grep -hoE "\"$MODULE/imports/[a-z0-9_]+\"" "$out/exports/wit_exports/wit_exports.go" \
      | sed -e "s|.*/imports/||" -e 's|"$||' | sort -u); do
    if [ -e "imports/$name" ]; then
      if ! diff -r "imports/$name" "$out/imports/$name" >/dev/null; then
        echo "error: imports/$name differs between worlds" >&2
        exit 1
      fi
    else
      cp -r "$out/imports/$name" imports/
    fi
  done

  # Keep only the generated export glue; the export_* trampolines next to it
  # are hand-written. Point the glue at the per-world trampoline packages.
  rm -rf "$dir/wit_exports"
  mkdir -p "$dir"
  cp -r "$out/exports/wit_exports" "$dir/"
  sed "s+\"$MODULE/exports/export_+\"$MODULE/$dir/export_+g" \
    "$out/exports/wit_exports/wit_exports.go" > "$dir/wit_exports/wit_exports.go"
  add_empty_s "$dir/wit_exports"
done

rm -rf tmp

go mod tidy
