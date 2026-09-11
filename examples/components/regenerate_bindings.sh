#!/usr/bin/env bash
#
# Regenerates the committed wit-bindgen bindings for the examples that carry
# their own.
#
# Most examples need nothing here: they import only what the SDK's world
# provides, and componentize-go generates that glue at build time. Two examples
# declare an extra world of their own (wasi:keyvalue, wasi:otel) and commit the
# resulting import bindings so the package is importable from handwritten code.
# Those committed files are what this script refreshes.
#
# Run it after bumping componentize-go, since the wit-bindgen version travels
# with the toolchain — see BUILDING.md. `make airgap-versions` fails if the
# committed headers fall behind the pinned toolchain.
#
# Uses `go tool componentize-go` from each module, so the version is whatever
# that module pins.

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

# dir : world : go package path
TARGETS=(
  "http-keyvalue-crud:wasmcloud:examples/http-keyvalue-crud@0.1.0:github.com/wasmCloud/go/examples/components/http-keyvalue-crud"
  "http-otel:wasmcloud:examples/http-otel@0.1.0:github.com/wasmCloud/go/examples/components/http-otel"
)

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

for target in "${TARGETS[@]}"; do
  dir=${target%%:*}
  rest=${target#*:}
  world=${rest%:*}
  pkg=${rest##*:}

  echo "==> ${dir} (${world})"
  out="$tmp/$dir"
  rm -rf "$out"

  (
    cd "$dir"
    # --ignore-toml-files so only this example's own world is generated; without
    # it the SDK's wasip2 world is merged in and the output balloons to bindings
    # the SDK already ships.
    go tool componentize-go \
      --ignore-toml-files \
      -w "$world" \
      -d wit \
      bindings \
      --format \
      -o "$out" \
      --pkg-name "$pkg" >/dev/null
  )

  # Overwrite only files that are already committed. The generator emits more
  # than these examples keep (unused export glue, for instance), and deciding
  # what to commit is a separate question from refreshing what is committed.
  changed=0
  while IFS= read -r rel; do
    src="$out/$rel"
    if [ ! -f "$src" ]; then
      echo "    WARNING: generator no longer produces ${rel}" >&2
      continue
    fi
    if ! cmp -s "$src" "$dir/$rel"; then
      cp "$src" "$dir/$rel"
      echo "    updated ${rel}"
      changed=$((changed + 1))
    fi
  done < <(
    cd "$dir" && find . -path ./wit -prune -o \
      \( -name 'wit_bindings.go' -o -name 'wit_exports.go' -o -name 'empty.s' \) -print \
      | sed 's|^\./||' | sort
  )

  [ "$changed" -eq 0 ] && echo "    already up to date"
done

echo
echo "Done. Verify with:"
echo "  GOOS=wasip1 GOARCH=wasm go vet -unsafeptr=false ./..."
