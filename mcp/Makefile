.PHONY: build run test test-integration clean help

# Default port
PORT ?= 8081

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the MCP server binary
	go build -o mcp-server .

run: ## Run the MCP server
	go run .

run-custom-port: ## Run with custom port (use PORT=9000 make run-custom-port)
	PORT=$(PORT) go run .

test: ## Run unit tests
	go test -v -cover

test-race: ## Run unit tests with race detector
	go test -v -race

test-integration: ## Run integration tests (requires server to be running)
	@echo "Make sure the server is running on port $(PORT)"
	@MCP_SERVER_URL=http://localhost:$(PORT) go run test_main.go test_client.go

test-integration-auto: ## Start server and run integration tests
	@echo "Starting server..."
	@PORT=$(PORT) go run . & \
	SERVER_PID=$$!; \
	sleep 2; \
	echo "Running integration tests..."; \
	MCP_SERVER_URL=http://localhost:$(PORT) go run test_main.go test_client.go; \
	TEST_EXIT=$$?; \
	echo "Stopping server..."; \
	kill $$SERVER_PID; \
	exit $$TEST_EXIT

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

lint: fmt vet ## Run all linters

clean: ## Clean build artifacts
	rm -f mcp-server
	go clean

deps: ## Download dependencies
	go mod download
	go mod tidy

check: lint test ## Run all checks (lint + test)

all: clean deps check build ## Run clean, deps, check, and build
