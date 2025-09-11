.PHONY: build deploy test clean lint vet

# Build settings
BINARY_NAME=boat-agent
LAMBDA_BINARY=bootstrap
BUILD_DIR=build
TEMPLATE_DIR=templates

# AWS settings
AWS_REGION ?= us-east-1
STACK_NAME ?= boat-ops-agent

# Go settings
GO_VERSION=1.23.4

# Default target
all: build

# Build the Lambda function
build:
	@echo "Building BOAT Agent for AWS Lambda..."
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(LAMBDA_BINARY) cmd/lambda/main.go
	@echo "Build complete: $(BUILD_DIR)/$(LAMBDA_BINARY)"

# Build locally for development
build-local:
	@echo "Building BOAT Agent for local development..."
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	go build -o $(BUILD_DIR)/boat-local cmd/local/main.go
	go build -o $(BUILD_DIR)/boat-simulate cmd/simulate/main.go
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME), $(BUILD_DIR)/boat-local, and $(BUILD_DIR)/boat-simulate"

# Package Lambda function for deployment
package: build
	@echo "Packaging Lambda function..."
	cd $(BUILD_DIR) && zip -r boat-agent.zip $(LAMBDA_BINARY)
	@echo "Package complete: $(BUILD_DIR)/boat-agent.zip"

# Deploy to AWS using CloudFormation
deploy: package
	@echo "Deploying BOAT Agent stack..."
	aws cloudformation deploy \
		--template-file $(TEMPLATE_DIR)/lambda-eventbridge.json \
		--stack-name $(STACK_NAME) \
		--capabilities CAPABILITY_NAMED_IAM \
		--region $(AWS_REGION) \
		--parameter-overrides \
			ClaudeAPIKey=$(CLAUDE_API_KEY) \
			AWSRegion=$(AWS_REGION) \
			NotificationEmail=$(NOTIFICATION_EMAIL)
	
	# Update Lambda function code
	aws lambda update-function-code \
		--function-name boat-ops-agent \
		--zip-file fileb://$(BUILD_DIR)/boat-agent.zip \
		--region $(AWS_REGION)
	
	@echo "Deployment complete!"

# Test the application
test:
	@echo "Running tests..."
	go test -v ./...
	go test -race -v ./...

# Run static analysis
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean -cache
	@echo "Clean complete"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Run local development server (for testing)
run-local: build-local
	@echo "Running BOAT Agent locally..."
	$(BUILD_DIR)/$(BINARY_NAME)

# Local development and testing commands
dev-test: build-local
	@echo "Running BOAT Agent test scenarios..."
	$(BUILD_DIR)/boat-local test

dev-interactive: build-local
	@echo "Starting BOAT Agent interactive mode..."
	$(BUILD_DIR)/boat-local interactive

dev-sample: build-local
	@echo "Generating sample event files..."
	$(BUILD_DIR)/boat-local sample

# Quick debug using the debug script
debug:
	@echo "Starting BOAT Agent debug mode..."
	./debug.sh --mode test --verbose

debug-interactive:
	@echo "Starting BOAT Agent interactive debug mode..."
	./debug.sh --mode interactive --verbose

# Full integration setup and demo
setup-full:
	@echo "Setting up full AWS + Claude integration..."
	./setup-full-integration.sh

demo-full:
	@echo "Running full functionality demo..."
	./demo-full-functionality.sh

# AWS integration testing
test-aws:
	@echo "Testing AWS integration..."
	./test-aws-integration.sh

# AWS simulation with real credentials
simulate-aws: build-local
	@echo "Starting AWS simulation with real credentials..."
	./simulate-with-aws.sh

simulate-discovery: build-local
	@echo "Running AWS infrastructure discovery..."
	$(BUILD_DIR)/boat-simulate aws-discovery

simulate-cost: build-local
	@echo "Running cost analysis simulation..."
	$(BUILD_DIR)/boat-simulate cost-analysis

simulate-deployment: build-local
	@echo "Running deployment simulation..."
	$(BUILD_DIR)/boat-simulate deployment-simulation

simulate-interactive: build-local
	@echo "Starting interactive AWS simulation..."
	$(BUILD_DIR)/boat-simulate interactive

# Generate event for testing
test-event:
	@echo "Sending test event to EventBridge..."
	aws events put-events \
		--entries '[{
			"Source": "boat.ops",
			"DetailType": "Container Deployment Request",
			"Detail": "{\"id\":\"test-001\",\"type\":\"deploy\",\"description\":\"Deploy test application\",\"parameters\":{\"image\":\"ghcr.io/jrzesz33/test-app:latest\"},\"priority\":\"medium\",\"requester\":\"developer\"}"
		}]' \
		--region $(AWS_REGION)

# Show deployment status
status:
	@echo "Checking BOAT Agent status..."
	aws cloudformation describe-stacks \
		--stack-name $(STACK_NAME) \
		--region $(AWS_REGION) \
		--query 'Stacks[0].StackStatus'

# Show logs
logs:
	@echo "Showing recent logs..."
	aws logs tail /aws/lambda/boat-ops-agent \
		--region $(AWS_REGION) \
		--follow

# Remove deployment
destroy:
	@echo "Destroying BOAT Agent stack..."
	aws cloudformation delete-stack \
		--stack-name $(STACK_NAME) \
		--region $(AWS_REGION)
	@echo "Stack deletion initiated. Check AWS Console for progress."

# Security scan
security:
	@echo "Running security scan..."
	gosec ./...

# Performance benchmark
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Generate documentation
docs:
	@echo "Generating documentation..."
	godoc -http=:6060
	@echo "Documentation server running at http://localhost:6060"

# Check for updates
update:
	@echo "Checking for dependency updates..."
	go list -u -m all

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

help:
	@echo "BOAT Agent Makefile Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  build       - Build Lambda function binary"
	@echo "  build-local - Build for local development"
	@echo "  package     - Package Lambda function for deployment"
	@echo ""
	@echo "Deployment Commands:"
	@echo "  deploy      - Deploy to AWS (requires CLAUDE_API_KEY and NOTIFICATION_EMAIL env vars)"
	@echo "  status      - Check deployment status"
	@echo "  destroy     - Remove AWS deployment"
	@echo ""
	@echo "Development Commands:"
	@echo "  test        - Run all tests"
	@echo "  lint        - Run linter"
	@echo "  vet         - Run go vet"
	@echo "  run-local   - Run locally for testing"
	@echo ""
	@echo "Local Testing Commands:"
	@echo "  dev-test    - Run predefined test scenarios locally"
	@echo "  dev-interactive - Start interactive testing mode"
	@echo "  dev-sample  - Generate sample event files"
	@echo "  debug       - Quick debug with verbose output"
	@echo "  debug-interactive - Interactive debug mode"
	@echo ""
	@echo "Full Integration Commands:"
	@echo "  setup-full  - Set up AWS + Claude integration"
	@echo "  demo-full   - Demo full functionality"
	@echo "  test-aws    - Test AWS integration specifically"
	@echo ""
	@echo "AWS Simulation Commands (with real credentials):"
	@echo "  simulate-aws - Interactive AWS credential simulation"
	@echo "  simulate-discovery - Discover your AWS infrastructure"
	@echo "  simulate-cost - Analyze real deployment costs"
	@echo "  simulate-deployment - Simulate complex deployments"
	@echo "  simulate-interactive - Interactive AWS exploration"
	@echo ""
	@echo "Testing Commands:"
	@echo "  test-event  - Send test event to EventBridge"
	@echo "  logs        - Show Lambda logs"
	@echo ""
	@echo "Utility Commands:"
	@echo "  clean       - Clean build artifacts"
	@echo "  deps        - Download dependencies"
	@echo "  security    - Run security scan"
	@echo "  benchmark   - Run performance benchmarks"
	@echo "  install-tools - Install development tools"
	@echo ""
	@echo "Environment Variables:"
	@echo "  CLAUDE_API_KEY     - Required for deployment"
	@echo "  NOTIFICATION_EMAIL - Required for deployment"
	@echo "  AWS_REGION        - AWS region (default: us-east-1)"
	@echo "  STACK_NAME        - CloudFormation stack name (default: boat-ops-agent)"