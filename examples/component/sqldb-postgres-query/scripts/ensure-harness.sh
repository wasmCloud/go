#!/bin/bash
set -euo pipefail

if [ ! -f "build/test-harness.wasm" ]; then
  wash build -p test-harness && cp test-harness/build/test-harness.wasm ./build/
else 
  echo 'Harness already exists. To rebuild, run `rm build/test-harness.wasm`'
fi
