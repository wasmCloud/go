# Air-gapped build image.
#
# Stage 1 has a network and downloads every pinned artifact explicitly; stage 2
# is what ships and must never need one. The air gap is enforced when the image
# is *run* (`docker run --network none`, one container per example), which is
# the honest test: building the image online is fine, building components
# inside it offline is the thing under test.
#
# Versions come from airgap/versions.env — see BUILDING.md.

ARG GO_VERSION=1.27.1

# =============================================================================
# Stage 1: seed
# =============================================================================
FROM golang:${GO_VERSION}-bookworm AS seed

ARG WASH_VERSION=v2.10.3
ARG COMPONENTIZE_GO_BINARY=v0.4.3
ARG PATCHED_GO_TAG=go1.27.1-wasi-on-idle
# Populated automatically by BuildKit. Must be declared with NO default: a
# default shadows the real value, and the image then gets a wrong-architecture
# wash and a second, unused patched Go tree.
ARG TARGETARCH

# Fixed cache locations so lookups do not depend on $HOME. componentize-go
# resolves its cache through XDG on Linux, so both the wrapper binary and the
# patched Go land under $XDG_CACHE_HOME/componentize-go.
ENV XDG_CACHE_HOME=/opt/cache \
    GOMODCACHE=/opt/gomodcache \
    GOCACHE=/opt/gobuild \
    GOTOOLCHAIN=local

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl bzip2 \
    && rm -rf /var/lib/apt/lists/*

# wash ships as a static musl binary from the wasmCloud/wasmCloud releases.
# Pinned rather than piped through the install script, which resolves "latest".
RUN set -eux; \
    case "${TARGETARCH}" in \
      amd64) washarch=x86_64 ;; \
      arm64) washarch=aarch64 ;; \
      *) echo "unsupported TARGETARCH: ${TARGETARCH}" >&2; exit 1 ;; \
    esac; \
    curl -sSfL -o /usr/local/bin/wash \
      "https://github.com/wasmCloud/wasmCloud/releases/download/${WASH_VERSION}/wash-${washarch}-unknown-linux-musl"; \
    chmod +x /usr/local/bin/wash; \
    wash --version

WORKDIR /src
COPY . .

# Fill GOMODCACHE for every module in the repo. Done before any build so a
# failure here is unambiguously a module-resolution problem.
RUN set -eux; \
    for mod in $(find examples/components templates -name go.mod); do \
      dir=$(dirname "$mod"); \
      echo "==> downloading deps for ${dir}"; \
      (cd "$dir" && go mod download all); \
    done

# The componentize-go Go module is only a wrapper: on first run it downloads a
# Rust binary named by the `release :=` string in its main.go, and an async
# world makes that binary download a patched Go. Neither is left to the tools.
# Both are fetched here, explicitly, at exactly the versions in versions.env,
# into the cache layout the tools look in (see BUILDING.md, "Building offline").
#
# The wrapper re-downloads whenever version.txt differs from its own `release`,
# so a COMPONENTIZE_GO_BINARY that disagrees with the module would be thrown
# away on first use. Catch that here rather than in the offline run.
RUN set -eux; \
    mod=$(cd examples/components/http-server && \
      go list -m -f '{{.Dir}}' github.com/bytecodealliance/componentize-go); \
    grep -q "release := \"${COMPONENTIZE_GO_BINARY}\"" "${mod}/main.go" || { \
      echo "COMPONENTIZE_GO_BINARY=${COMPONENTIZE_GO_BINARY} but ${mod}/main.go fetches:" >&2; \
      grep 'release :=' "${mod}/main.go" >&2; exit 1; }

RUN set -eux; \
    case "${TARGETARCH}" in \
      amd64) goarch=amd64 ;; \
      arm64) goarch=arm64 ;; \
      *) echo "unsupported TARGETARCH: ${TARGETARCH}" >&2; exit 1 ;; \
    esac; \
    cg="${XDG_CACHE_HOME}/componentize-go"; \
    mkdir -p "${cg}/bin" "${cg}/v2"; \
    \
    curl -sSfL "https://github.com/bytecodealliance/componentize-go/releases/download/${COMPONENTIZE_GO_BINARY}/componentize-go-linux-${goarch}.tar.gz" \
      | tar -xz -C "${cg}/bin"; \
    chmod 0755 "${cg}/bin/componentize-go"; \
    printf '%s' "${COMPONENTIZE_GO_BINARY}" > "${cg}/version.txt"; \
    "${cg}/bin/componentize-go" --version | grep -F "${COMPONENTIZE_GO_BINARY#v}"; \
    \
    name="go-linux-${goarch}-bootstrap"; \
    curl -sSfL "https://github.com/dicej/go/releases/download/${PATCHED_GO_TAG}/${name}.tbz" \
      | tar -xj -C "${cg}/v2"; \
    want="${PATCHED_GO_TAG#go}"; want="${want%%-*}"; \
    have=$(head -n1 "${cg}/v2/${name}/VERSION" | sed 's/^go//'); \
    [ "${have}" = "${want}" ] || { echo "patched Go VERSION '${have}' != '${want}'" >&2; exit 1; }; \
    grep -q wasiOnIdle "${cg}/v2/${name}/src/runtime/lock_wasip1.go"

# Warm GOCACHE with one sync and one async build — and, since these run with no
# network, prove the seeded caches are complete before the image ships. If
# either tool still wanted to download something, the image build fails here.
RUN --network=none set -eux; \
    export GOPROXY=off GOSUMDB=off; \
    cd /src/examples/components/http-server && wash build --skip-fetch

RUN --network=none set -eux; \
    export GOPROXY=off GOSUMDB=off; \
    cd /src/examples/components/http-p3-streaming && wash build --skip-fetch

# The builds above wrote artifacts into the source tree. They are throwaway —
# the real run bind-mounts a clean checkout — so drop them rather than baking
# stale .wasm files into the image.
RUN rm -rf /src

# =============================================================================
# Stage 2: runtime — no network from here on
# =============================================================================
FROM golang:${GO_VERSION}-bookworm AS runtime

ENV XDG_CACHE_HOME=/opt/cache \
    GOMODCACHE=/opt/gomodcache \
    GOCACHE=/opt/gobuild \
    GOTOOLCHAIN=local \
    GOPROXY=off \
    GOFLAGS="-mod=mod -buildvcs=false" \
    CGO_ENABLED=0 \
    GOSUMDB=off

# -buildvcs=false because the workspace is a bind-mounted git checkout. Go
# otherwise shells out to `git` to stamp version info, and that fails with
# "error obtaining VCS status: exit status 128" whenever git refuses the
# repository — on Linux CI the checkout is owned by the runner's UID while the
# container runs as root, which trips git's dubious-ownership check. It does
# not reproduce on macOS, where the bind mount reports everything as root-owned.
#
# Disabling it is right on its own terms: a hermetic build should not depend on
# the checkout's git state, and the stamp would make output non-reproducible.

# GOPROXY=off alone is not enough: writing a go.sum entry, which any module
# lacking one does (the templates), verifies the hash against sum.golang.org —
# a separate network call from the module fetch. The modules were verified when
# they were downloaded into the cache in the seed stage, so GOSUMDB=off above
# skips a check that has already happened rather than skipping it outright.

COPY --from=seed /usr/local/bin/wash /usr/local/bin/wash
COPY --from=seed /opt/cache /opt/cache
COPY --from=seed /opt/gomodcache /opt/gomodcache
COPY --from=seed /opt/gobuild /opt/gobuild
COPY airgap/build-all.sh /usr/local/bin/build-all.sh

# The repo is bind-mounted here at run time.
WORKDIR /workspace

ENTRYPOINT ["/usr/local/bin/build-all.sh"]
