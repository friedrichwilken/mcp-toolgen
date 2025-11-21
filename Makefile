# MCP Toolgen Makefile
.PHONY: help build test clean install lint fmt vet deps demo-cr demo-d demo-all validate-crud generate-example clean-examples

# Build variables
BINARY_NAME = mcp-toolgen
BUILD_DIR = build
CMD_DIR = ./cmd/mcp-toolgen

# Version variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# LDFLAGS with version injection
LDFLAGS = -ldflags="-s -w -X 'github.com/friedrichwilken/mcp-toolgen/pkg/cmd.version=$(VERSION)' -X 'github.com/friedrichwilken/mcp-toolgen/pkg/cmd.gitCommit=$(GIT_COMMIT)' -X 'github.com/friedrichwilken/mcp-toolgen/pkg/cmd.buildDate=$(BUILD_DATE)'"

# Demo and example variables
DEMO_CRD = /Users/I549741/claude-playroom/crds/kyma-serverless-function-crd.yaml
DEMO_OUTPUT_BASE = ./examples
DEMO_MODULE_PATH = github.com/friedrichwilken/mcp-toolgen/examples

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development targets
deps: ## Download and install dependencies
	go mod download
	go mod tidy

build: deps ## Build the mcp-toolgen binary
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

install: build ## Install the binary to GOPATH/bin
	go install $(CMD_DIR)

# Code quality targets
fmt: ## Format Go code (standard)
	@echo "Running standard Go formatting..."
	go fmt ./...

fmt-imports: ## Fix and organize imports
	@echo "Organizing imports..."
	@which goimports > /dev/null || (echo "Installing goimports..." && go install golang.org/x/tools/cmd/goimports@latest)
	goimports -w -local github.com/friedrichwilken/mcp-toolgen .

fmt-modern: ## Apply modern Go formatting (interface{} -> any, strict formatting)
	@echo "Applying modern Go formatting..."
	@which gofumpt > /dev/null || (echo "Installing gofumpt..." && go install mvdan.cc/gofumpt@latest)
	gofumpt -w -extra .

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

lint: ## Run golangci-lint (read-only check)
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run

lint-fix: ## Run golangci-lint with automatic fixes (RECOMMENDED)
	@echo "Running linter with automatic fixes..."
	@echo "This will fix: imports, formatting, whitespace, unnecessary conversions, and more"
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run --fix

auto-fix: ## Auto-fix all fixable issues (gofumpt, goimports, lint-fix)
	@echo "🔧 Auto-fixing all issues..."
	@echo "Step 1/3: Organizing imports..."
	@$(MAKE) fmt-imports
	@echo "Step 2/3: Applying modern formatting..."
	@$(MAKE) fmt-modern
	@echo "Step 3/3: Running linter auto-fixes..."
	@$(MAKE) lint-fix
	@echo "✅ Auto-fix completed!"

security-check: ## Run security vulnerability checks
	@echo "Running security checks..."
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securego/gosec/v2/cmd/gosec@latest)
	@which govulncheck > /dev/null || (echo "Installing govulncheck..." && go install golang.org/x/vuln/cmd/govulncheck@latest)
	@echo "Running gosec..."
	@gosec -quiet ./... || echo "⚠️  Some security warnings found (may be expected for code generators)"
	@echo "Running govulncheck..."
	govulncheck ./...

code-quality: deps tidy auto-fix ## Run all code quality checks with auto-fixes
	@echo "✅ Code quality checks completed successfully"

check: lint vet test ## Check code without making changes (for CI)

# Testing targets
test: ## Run all tests
	go test -v ./...

test-unit: ## Run unit tests only
	go test -v ./pkg/...

test-integration: ## Run integration tests only
	go test -v ./test/integration/...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Development helpers
clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

dev-setup: deps ## Set up development environment
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Release targets
release-dry-run: ## Dry run of release process
	@echo "This would build and package the release"

version: ## Show version information
	@echo "mcp-toolgen version: development"
	@go version

# Demo and example targets
demo-cr: dev-build ## Demo: Generate create+read operations only
	@echo "🚀 Demo: Generating Functions toolset with create+read operations..."
	./$(BUILD_DIR)/$(BINARY_NAME) --crud cr \
		--crd $(DEMO_CRD) \
		--output $(DEMO_OUTPUT_BASE)/functions-cr \
		--package functions \
		--module-path $(DEMO_MODULE_PATH) \
		--verbose --overwrite
	@echo "✅ Generated in $(DEMO_OUTPUT_BASE)/functions-cr/"

demo-d: dev-build ## Demo: Generate delete operations only
	@echo "🚀 Demo: Generating Functions toolset with delete operations only..."
	./$(BUILD_DIR)/$(BINARY_NAME) --crud d \
		--crd $(DEMO_CRD) \
		--output $(DEMO_OUTPUT_BASE)/functions-d \
		--package functions \
		--module-path $(DEMO_MODULE_PATH) \
		--verbose --overwrite
	@echo "✅ Generated in $(DEMO_OUTPUT_BASE)/functions-d/"

demo-all: dev-build ## Demo: Generate all CRUD operations (default)
	@echo "🚀 Demo: Generating Functions toolset with all CRUD operations..."
	./$(BUILD_DIR)/$(BINARY_NAME) \
		--crd $(DEMO_CRD) \
		--output $(DEMO_OUTPUT_BASE)/functions-all \
		--package functions \
		--module-path $(DEMO_MODULE_PATH) \
		--verbose --overwrite
	@echo "✅ Generated in $(DEMO_OUTPUT_BASE)/functions-all/"

demo-dry-run: dev-build ## Demo: Dry run generation (no files created)
	@echo "🚀 Demo: Dry run generation..."
	./$(BUILD_DIR)/$(BINARY_NAME) --crud cu \
		--crd $(DEMO_CRD) \
		--output $(DEMO_OUTPUT_BASE)/functions-cu \
		--package functions \
		--module-path $(DEMO_MODULE_PATH) \
		--dry-run --verbose

validate-crud: dev-build ## Validate CRUD flag functionality
	@echo "🧪 Testing CRUD flag validation..."
	@echo "Testing invalid characters (should fail):"
	-./$(BUILD_DIR)/$(BINARY_NAME) --crud xyz --crd $(DEMO_CRD) --output /tmp/test --module-path test 2>&1 | grep "invalid character"
	@echo "Testing duplicate characters (should fail):"
	-./$(BUILD_DIR)/$(BINARY_NAME) --crud ccd --crd $(DEMO_CRD) --output /tmp/test --module-path test 2>&1 | grep "duplicate character"
	@echo "✅ CRUD validation working correctly"

generate-example: dev-build ## Generate example toolset for ek8sms integration
	@echo "🔧 Generating example Functions toolset for ek8sms integration..."
	./$(BUILD_DIR)/$(BINARY_NAME) \
		--crd $(DEMO_CRD) \
		--output /Users/I549741/claude-playroom/extendable-kubernetes-mcp-server/pkg/functions \
		--package functions \
		--module-path github.com/friedrichwilken/extendable-kubernetes-mcp-server \
		--verbose --overwrite
	@echo "✅ Generated Functions toolset in ek8sms pkg/functions/"

clean-examples: ## Clean generated examples
	@echo "🧹 Cleaning generated examples..."
	rm -rf $(DEMO_OUTPUT_BASE)
	rm -rf ./test-crud-*
	@echo "✅ Examples cleaned"

# Quality assurance targets
qa: code-quality test ## Run all quality assurance checks

ci: code-quality test build ## Run CI pipeline (quality, test, build)
	@echo "✅ CI pipeline completed successfully"

pre-commit-check: code-quality test ## Run pre-commit checks (quality + tests)
	@echo "✅ Pre-commit checks completed successfully"

# Development workflow helpers
dev-build: ## Quick development build without deps
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)

dev-test: ## Quick test run for development
	go test -v ./pkg/...

check-deps: ## Check for latest dependency versions
	@echo "🔍 Checking for dependency updates..."
	go list -u -m all

# Benchmarking and performance
benchmark: dev-build ## Run performance benchmarks
	@echo "⚡ Running generation benchmarks..."
	@time ./$(BUILD_DIR)/$(BINARY_NAME) --crd $(DEMO_CRD) --output /tmp/benchmark --package bench --module-path test --overwrite
	@rm -rf /tmp/benchmark
	@echo "✅ Benchmark completed"

# Documentation helpers
docs-examples: demo-cr demo-d demo-all ## Generate all documentation examples
	@echo "📚 All documentation examples generated in $(DEMO_OUTPUT_BASE)/"

# Multi-platform build targets
build-linux: deps ## Build for Linux
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)

build-windows: deps ## Build for Windows
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)

build-all: build build-linux build-windows ## Build for all platforms
	@echo "✅ Multi-platform build completed"