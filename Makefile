.PHONY: all build test coverage lint fmt clean install-tools generate docs validate

# Variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=gofmt
GOLINT=golangci-lint
GOMOD=$(GOCMD) mod

# Build variables
BINARY_NAME=monarch
BINARY_DIR=bin
PKG_DIR=./pkg/...
CMD_DIR=./cmd/...
EXAMPLES_DIR=./examples/...

# Test variables  
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
NC=\033[0m # No Color

all: clean fmt lint test build

## help: Display this help message
help:
	@echo "Available targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /'

## build: Build the binary
build:
	@echo "$(GREEN)Building binary...$(NC)"
	@mkdir -p $(BINARY_DIR)
	@$(GOBUILD) -v -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/monarch
	@echo "$(GREEN)Build complete: $(BINARY_DIR)/$(BINARY_NAME)$(NC)"

## test: Run all tests
test:
	@echo "$(GREEN)Running tests...$(NC)"
	@$(GOTEST) -v -race $(PKG_DIR)
	@echo "$(GREEN)Tests complete$(NC)"

## test-short: Run tests in short mode (skip integration tests)
test-short:
	@echo "$(GREEN)Running short tests...$(NC)"
	@$(GOTEST) -v -short $(PKG_DIR)

## coverage: Run tests with coverage
coverage:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	@$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic $(PKG_DIR)
	@$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "$(GREEN)Coverage report: $(COVERAGE_HTML)$(NC)"
	@echo "$(GREEN)Coverage summary:$(NC)"
	@$(GOCMD) tool cover -func=$(COVERAGE_FILE)

## benchmark: Run benchmarks
benchmark:
	@echo "$(GREEN)Running benchmarks...$(NC)"
	@$(GOTEST) -bench=. -benchmem $(PKG_DIR)

## lint: Run linter
lint:
	@echo "$(GREEN)Running linter...$(NC)"
	@if ! which $(GOLINT) > /dev/null; then \
		echo "$(YELLOW)golangci-lint not found, installing...$(NC)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@$(GOLINT) run --timeout 5m ./...

## fmt: Format code
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	@$(GOFMT) -w -s .
	@$(GOCMD) fmt ./...
	@echo "$(GREEN)Code formatted$(NC)"

## vet: Run go vet
vet:
	@echo "$(GREEN)Running go vet...$(NC)"
	@$(GOCMD) vet ./...

## clean: Clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BINARY_DIR)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@rm -f coverage.txt
	@rm -f mcp-monarch
	@rm -f cmd/mcp-server/mcp-monarch
	@rm -f test_*.go
	@rm -f security-report.json
	@find . -name "*.bak" -type f -delete
	@echo "$(GREEN)Clean complete$(NC)"

## deps: Download dependencies
deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "$(GREEN)Dependencies updated$(NC)"

## install-tools: Install development tools
install-tools:
	@echo "$(GREEN)Installing development tools...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/Khan/genqlient@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/stretchr/testify@latest
	@echo "$(GREEN)Tools installed$(NC)"

## generate: Run code generation (GraphQL client, mocks, etc.)
generate:
	@echo "$(GREEN)Running code generation...$(NC)"
	@$(GOCMD) generate ./...
	@echo "$(GREEN)Code generation complete$(NC)"

## validate: Run Python compatibility validator
validate:
	@echo "$(GREEN)Running Python compatibility validator...$(NC)"
	@$(GOBUILD) -v -o $(BINARY_DIR)/validator ./cmd/validator
	@$(BINARY_DIR)/validator
	@echo "$(GREEN)Validation complete$(NC)"

## examples: Build all examples
examples:
	@echo "$(GREEN)Building examples...$(NC)"
	@for example in $(shell ls examples/*.go); do \
		name=$$(basename $$example .go); \
		echo "  Building $$name..."; \
		$(GOBUILD) -o $(BINARY_DIR)/example_$$name $$example; \
	done
	@echo "$(GREEN)Examples built$(NC)"

## docs: Generate documentation
docs:
	@echo "$(GREEN)Generating documentation...$(NC)"
	@godoc -http=:6060 &
	@echo "$(GREEN)Documentation server started at http://localhost:6060$(NC)"

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "$(GREEN)All checks passed!$(NC)"

## ci: Run CI pipeline
ci: deps check coverage
	@echo "$(GREEN)CI pipeline complete$(NC)"

## mod-update: Update all dependencies to latest versions
mod-update:
	@echo "$(GREEN)Updating dependencies...$(NC)"
	@$(GOCMD) get -u ./...
	@$(GOMOD) tidy
	@echo "$(GREEN)Dependencies updated to latest versions$(NC)"

## security: Run security scanner
security:
	@echo "$(GREEN)Running security scan...$(NC)"
	@if ! which gosec > /dev/null; then \
		echo "$(YELLOW)gosec not found, installing...$(NC)"; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	fi
	@gosec -fmt json -out security-report.json ./...
	@echo "$(GREEN)Security scan complete: security-report.json$(NC)"

## integration-test: Run integration tests (requires Monarch API access)
integration-test:
	@echo "$(GREEN)Running integration tests...$(NC)"
	@if [ -z "$$MONARCH_TOKEN" ]; then \
		echo "$(RED)Error: MONARCH_TOKEN environment variable not set$(NC)"; \
		exit 1; \
	fi
	@$(GOTEST) -v -tags=integration ./test/integration/...

## docker-build: Build Docker image
docker-build:
	@echo "$(GREEN)Building Docker image...$(NC)"
	@docker build -t monarch-go:latest .
	@echo "$(GREEN)Docker image built: monarch-go:latest$(NC)"

## release: Create a new release
release: clean check
	@echo "$(GREEN)Creating release...$(NC)"
	@read -p "Enter version (e.g., v1.0.0): " version; \
	git tag $$version; \
	git push origin $$version; \
	echo "$(GREEN)Release $$version created$(NC)"

# Default target
.DEFAULT_GOAL := help