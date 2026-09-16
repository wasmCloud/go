# Sourced, not run. Makes `go` resolve go.wasmcloud.dev/component from this
# checkout's git tags instead of from proxy.golang.org.
#
# That is what lets a release PR pin the examples to a version that is not
# tagged upstream yet. The module's go.sum hash is a function of the files under
# component/ alone, and resolving the vanity path in direct mode builds the
# module zip from git exactly as the proxy does, so the hash written here is the
# one the proxy and sum.golang.org will report once the tag is pushed.
# release-tag.yaml checks that before tagging, and again after.
#
# Only the git remote is redirected. The go-get lookup on go.wasmcloud.dev still
# needs a network, and a tag must exist *locally* for the version asked for.
#
# GIT_CONFIG_COUNT adds to the git config rather than replacing it, so global
# settings and the checkout's credentials are left alone.

local_sdk_env() {
  local root=$1
  export GIT_CONFIG_COUNT=1
  export GIT_CONFIG_KEY_0="url.file://${root}.insteadOf"
  export GIT_CONFIG_VALUE_0="https://github.com/wasmCloud/go"
  # Direct mode and no checksum-database lookup, for this one module only:
  # sum.golang.org has never seen an unpushed tag.
  export GOPRIVATE=go.wasmcloud.dev/component
}
