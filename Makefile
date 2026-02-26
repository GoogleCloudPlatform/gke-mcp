# Makefile for GKE MCP Server
# Provides convenient shortcuts for common development tasks

.PHONY: help build run install test clean presubmit update-version

# Default target - show help
.DEFAULT_GOAL := help

# Variables
BINARY_NAME := gke-mcp
DOCKER_IMAGE := $(BINARY_NAME)

help: ## Display available commands
	@echo "GKE MCP Server - Available Commands"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) .
	@echo "✓ Built $(BINARY_NAME)"

run: build ## Build and run the server
	./$(BINARY_NAME)

install: ## Install the binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	go install .
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

update-version: ## Update version. Usage: make update-version [BUMP_TYPE=major|minor|patch]
	@echo "Updating version..."
	@./dev/tasks/update_version.sh $(BUMP_TYPE)

docker-build: ## Build the docker image
	@echo "Building docker image $(DOCKER_IMAGE)..."
	docker build -t $(DOCKER_IMAGE) .
	@echo "✓ Built docker image $(DOCKER_IMAGE)"

docker-run: docker-build ## Build and run the docker image
	@echo "Running docker image $(DOCKER_IMAGE)..."
	docker run -it --rm -p 8080:8080 $(DOCKER_IMAGE) --server-mode http --server-host 0.0.0.0
