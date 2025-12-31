.PHONY: default build install test testacc generate clean fmt lint help

default: build

# Build the provider
build:
	go build -o terraform-provider-kanidm

# Install the provider locally for testing
install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/ssoriche/kanidm/0.1.0/darwin_arm64/
	cp terraform-provider-kanidm ~/.terraform.d/plugins/registry.terraform.io/ssoriche/kanidm/0.1.0/darwin_arm64/

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
