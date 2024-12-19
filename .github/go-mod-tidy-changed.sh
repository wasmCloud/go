#!/usr/bin/env bash

set -eu

if [ $# -eq 0 ]; then
    exit 0
fi

CHANGED_GO_MOD_DIRS="$@"

for mod_dir in ${CHANGED_GO_MOD_DIRS}; do
    pushd "${mod_dir%go.mod}"
    go mod tidy
    popd

    if [[ "${mod_dir}" == component* ]]; then
        example_dirs=$(find examples/component -maxdepth 1 -mindepth 1 -type d)
        for example_dir in ${example_dirs}; do 
            pushd "$example_dir"
            go mod tidy
            popd
        done
    fi

    if [[ "${mod_dir}" == provider* ]]; then
        example_dirs=$(find examples/provider -maxdepth 1 -mindepth 1 -type d)
        for example_dir in ${example_dirs}; do 
            pushd "$example_dir"
            go mod tidy
            popd
        done
    fi
done