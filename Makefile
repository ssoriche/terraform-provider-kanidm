.PHONY: default build install test testacc generate clean fmt lint help

default: build

# Build the provider
build:
	go build -o terraform-provider-kanidm

# Version published into the local filesystem mirror. Must satisfy the
# `version` constraint in the consuming config (homelab: terraform/kanidm).
PROVIDER_VERSION ?= 0.1.0
MIRROR_DIR ?= $(HOME)/.terraform.d/plugins/registry.terraform.io/ssoriche/kanidm/$(PROVIDER_VERSION)
# linux_amd64 is built alongside the host platform so .terraform.lock.hcl can
# carry CI checksums as well as local ones. Keep in sync with the platforms
# passed to `terraform providers lock -platform=...`.
MIRROR_PLATFORMS ?= darwin_arm64 linux_amd64

# Install the provider into the local filesystem mirror for every platform in
# MIRROR_PLATFORMS. Re-run this after any change, or the mirror will serve a
# stale binary and `terraform init` will fail on a lock file checksum mismatch.
install:
	@for platform in $(MIRROR_PLATFORMS); do \
		goos="$${platform%_*}"; goarch="$${platform#*_}"; \
		echo "installing $$goos/$$goarch -> $(MIRROR_DIR)/$$platform"; \
		mkdir -p "$(MIRROR_DIR)/$$platform"; \
		CGO_ENABLED=0 GOOS="$$goos" GOARCH="$$goarch" go build -trimpath -buildvcs=false \
			-ldflags "-s -w -X main.version=$(PROVIDER_VERSION)" \
			-o "$(MIRROR_DIR)/$$platform/terraform-provider-kanidm" . ; \
	done

# Run unit tests
test:
	go test -v ./...

# Run acceptance tests
testacc:
	TF_ACC=1 go test -v -timeout 30m ./internal/provider/

# Generate provider code from OpenAPI schema
generate:
	@echo "Generating provider code from OpenAPI schema..."
	tfplugingen-openapi generate \
		--config internal/spec/generator_config.yml \
		--output internal/spec/provider_code_spec.json \
		internal/spec/kanidm-openapi.json
	tfplugingen-framework generate all \
		--input internal/spec/provider_code_spec.json \
		--output internal/provider

# Generate documentation
docs:
	tfplugindocs generate --provider-name kanidm

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -f terraform-provider-kanidm
	rm -f internal/spec/provider_code_spec.json
	rm -rf dist/

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build the provider binary"
	@echo "  install    - Install the provider locally for testing"
	@echo "  test       - Run unit tests"
	@echo "  testacc    - Run acceptance tests (requires KANIDM_URL and KANIDM_TOKEN)"
	@echo "  generate   - Regenerate provider code from OpenAPI schema"
	@echo "  docs       - Generate documentation"
	@echo "  fmt        - Format code"
	@echo "  lint       - Run linter"
	@echo "  clean      - Remove build artifacts"
	@echo "  help       - Show this help message"
