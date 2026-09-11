# wasmCloud Go — developer entry points.
#
# Thin wrappers around the scripts in airgap/. The build system for components
# is `wash build`; this is a front door, not a replacement for it.

SHELL := /bin/bash

# Versions live in one file so the Dockerfile, CI and BUILDING.md cannot drift
# apart. Every non-comment line becomes a --build-arg.
VERSIONS_FILE := airgap/versions.env
AIRGAP_BUILD_ARGS := $(shell grep -vE '^[[:space:]]*(\#|$$)' $(VERSIONS_FILE) | sed 's/^/--build-arg /')

# Tagging by a hash of the pinned version set means a cache miss *is* the signal
# that something moved, rather than a silent rebuild against new versions.
AIRGAP_IMAGE ?= wasmcloud-go-airgap
AIRGAP_TAG   ?= $(shell (sha256sum $(VERSIONS_FILE) 2>/dev/null || shasum -a 256 $(VERSIONS_FILE)) | cut -c1-12)
AIRGAP_REF   := $(AIRGAP_IMAGE):$(AIRGAP_TAG)

# Passed through to the container; see airgap/build-all.sh.
AIRGAP_USE_LOCAL_SDK ?= 0

DOCKER ?= docker

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

.PHONY: airgap-image
airgap-image: ## Build the air-gapped toolchain image
	$(DOCKER) build -f airgap/Dockerfile -t $(AIRGAP_REF) $(AIRGAP_BUILD_ARGS) .

.PHONY: airgap-test
airgap-test: airgap-image ## Build every example with no network
	$(DOCKER) run --rm --network none \
		-e AIRGAP_USE_LOCAL_SDK=$(AIRGAP_USE_LOCAL_SDK) \
		-v "$(CURDIR):/workspace" \
		$(AIRGAP_REF)

.PHONY: airgap-shell
airgap-shell: airgap-image ## Interactive shell in the air-gapped image
	$(DOCKER) run --rm -it --network none \
		-e AIRGAP_USE_LOCAL_SDK=$(AIRGAP_USE_LOCAL_SDK) \
		-v "$(CURDIR):/workspace" \
		--entrypoint /bin/bash \
		$(AIRGAP_REF)

.PHONY: airgap-versions
airgap-versions: ## Re-derive every pinned version and check BUILDING.md
	@bash airgap/check-versions.sh
