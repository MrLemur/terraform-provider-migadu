default: install

GO ?= go
GOBIN ?= $(shell $(GO) env GOBIN)
ifeq ($(strip $(GOBIN)),)
GOBIN := $(shell $(GO) env GOPATH)/bin
endif

GOOS ?= $(shell $(GO) env GOOS)
GOARCH ?= $(shell $(GO) env GOARCH)
PLUGIN_VERSION ?= 0.1.0
PLUGIN_DIR ?= $(HOME)/.terraform.d/plugins/registry.terraform.io/MrLemur/migadu/$(PLUGIN_VERSION)/$(GOOS)_$(GOARCH)

# Build and install the provider locally
.PHONY: install
install: build
	mkdir -p $(PLUGIN_DIR)
	cp $(GOBIN)/terraform-provider-migadu $(PLUGIN_DIR)/

# Build the provider
.PHONY: build
build:
	$(GO) build -o $(GOBIN)/terraform-provider-migadu

# Run unit tests
.PHONY: test
test:
	$(GO) test -count=1 ./...

# Run acceptance tests
.PHONY: testacc
testacc:
	@test -n "$(MIGADU_USERNAME)" || (echo "MIGADU_USERNAME is required for acceptance tests" && exit 1)
	@test -n "$(MIGADU_API_KEY)" || (echo "MIGADU_API_KEY is required for acceptance tests" && exit 1)
	@test -n "$(MIGADU_TEST_DOMAIN)" || (echo "MIGADU_TEST_DOMAIN is required for acceptance tests" && exit 1)
	TF_ACC=1 $(GO) test ./internal/provider -run '^TestAcc' -count=1 -v $(TESTARGS) -timeout 120m

# Format code
.PHONY: fmt
fmt:
	gofmt -s -w -e .
	terraform fmt -recursive ./examples

# Generate documentation
.PHONY: docs
docs:
	go generate ./...

# Clean build artifacts
.PHONY: clean
clean:
	rm -f $(GOBIN)/terraform-provider-migadu
	rm -rf $(HOME)/.terraform.d/plugins/registry.terraform.io/MrLemur/migadu

# Lint code
.PHONY: lint
lint:
	golangci-lint run

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  install  - Build and install the provider locally"
	@echo "  build    - Build the provider binary"
	@echo "  test     - Run unit tests"
	@echo "  testacc  - Run acceptance tests"
	@echo "  fmt      - Format code and Terraform files"
	@echo "  docs     - Generate documentation"
	@echo "  clean    - Remove build artifacts"
	@echo "  lint     - Run linter"
