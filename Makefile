# Makefile for GKE MCP Server
# Provides convenient shortcuts for common development tasks

.PHONY: help build install test clean presubmit

# Default target - show help
.DEFAULT_GOAL := help

# Variables
BINARY_NAME := gke-mcp
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

help: ## Display available commands
	@echo "GKE MCP Server - Available Commands"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "✓ Built $(BINARY_NAME)"

install: ## Install the binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) .
	@echo "✓ Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

clean: ## Remove build artifacts
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@rm -rf dist/
	@echo "✓ Cleaned"

presubmit: ## Run all presubmit checks (build, test, vet, format)
	@echo "Running presubmit checks..."
	@./dev/tasks/presubmit.sh
	@echo "✓ All presubmit checks passed"
