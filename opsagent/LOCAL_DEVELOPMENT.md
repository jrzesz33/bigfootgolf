# BOAT Agent - Local Development & Testing

## Quick Start

```bash
# Build and run test scenarios
make dev-test

# Start interactive mode
make dev-interactive

# Generate sample event files
make dev-sample

# Quick debug with verbose output
make debug
```

## Available Testing Modes

### 1. **Predefined Test Scenarios** (`make dev-test`)
Runs three built-in scenarios:
- Simple container deployment
- Microservice with database
- Troubleshooting request

### 2. **Interactive Mode** (`make dev-interactive`)
Enter natural language descriptions and see how BOAT processes them:
```
BOAT> Deploy a web application with Redis
BOAT> Scale the user service to 3 replicas
BOAT> quit
```

### 3. **Sample Event Files** (`make dev-sample`)
Generates JSON event files you can modify and test:
- `deploy-webapp.json` - React app deployment
- `deploy-api.json` - Go API with Redis
- `troubleshoot-service.json` - Service troubleshooting

### 4. **Custom Event Processing**
```bash
# Process a specific event file
./build/boat-local event deploy-webapp.json

# Or use the debug script
./debug.sh --mode event deploy-webapp.json
```

## Debug Script Features

The `debug.sh` script provides advanced testing capabilities:

```bash
# Basic usage
./debug.sh --mode test --verbose

# With Claude API key
./debug.sh --mode interactive --claude-key sk-xxx

# With specific AWS profile
./debug.sh --mode test --aws-profile dev

# Show help
./debug.sh --help
```

## Environment Setup

### Required (for full functionality)
```bash
export CLAUDE_API_KEY="your-api-key"
```

### Optional
```bash
export AWS_REGION="us-east-1"
export AWS_PROFILE="default" 
export LOG_LEVEL="debug"
```

### Quick Setup
```bash
# Copy environment template
cp .env.example .env
# Edit .env with your values
# Source it: source .env
```

## Testing With Real AWS Credentials

### ✅ **What WORKS with AWS Credentials:**
- **VPC Discovery** - Finds your actual VPCs and subnets
- **ECS Cluster Listing** - Shows real clusters in your account
- **Security Group Queries** - Reads existing security groups
- **Cloud Control API** - **NOW IMPLEMENTED** for resource creation
- **Cost Estimation** - Real calculations for your resources
- **Resource Status Checking** - Track deployment progress

### ⚠️ **What's PARTIALLY WORKING:**
- **Resource Creation** - Cloud Control API implemented, needs testing
- **Load Balancer Setup** - Basic support (ELBv2 client needs expansion)

### ❌ **What Still Needs Implementation:**
- **Claude AI Integration** - Requires API key for natural language processing
- **Full Deployment Orchestration** - End-to-end workflows

### **Testing with AWS:**
```bash
# Test AWS integration
./test-aws-integration.sh

# Run with AWS credentials
AWS_PROFILE=your-profile make dev-test
```

## Testing Without AWS/Claude

The local development system also works without credentials:

- **AWS calls** are simulated and logged for read operations
- **Claude integration** uses placeholder responses  
- **All validation and logic** is fully functional
- **Structured logging** shows exactly what would happen

## Sample Output

```
[BOAT] Starting BOAT Agent local development environment
[SUCCESS] Build complete
[WARNING] AWS credentials not configured - some features may be limited
[BOAT] Starting BOAT agent in 'test' mode

=== Test Scenario 1: Simple Container Deployment ===
INFO[2025-09-11T21:39:44Z] BOAT Agent: Processing task request
INFO[2025-09-11T21:39:44Z] Parsed task request task_id=test-001
INFO[2025-09-11T21:39:44Z] Would call Claude API here system_prompt="You are BOAT..."
INFO[2025-09-11T21:39:44Z] Starting deployment plan execution
✅ Scenario 1 completed successfully in 45.2ms
```

## File Structure

```
opsagent/
├── cmd/local/main.go       # Local development entry point
├── debug.sh               # Advanced debug script
├── test/unit_test.go      # Unit tests
├── .env.example          # Environment template
├── deploy-*.json         # Generated sample events
└── LOCAL_DEVELOPMENT.md  # This file
```

## Unit Tests

```bash
# Run unit tests
go test ./test/...

# Run benchmarks
go test -bench=. ./test/...

# Test with coverage
go test -cover ./...
```

## Common Use Cases

### 1. Test Container Validation
```bash
# Generate sample, modify it, then test
make dev-sample
# Edit deploy-webapp.json to test validation
./build/boat-local event deploy-webapp.json
```

### 2. Debug Cost Estimation
```bash
# Set verbose logging and test scenarios
LOG_LEVEL=debug make dev-test
```

### 3. Test Error Handling
Create invalid event files and see how BOAT handles them:
```json
{
  "id": "test-error",
  "type": "deploy",
  "description": "Deploy with invalid image",
  "parameters": {
    "image": "invalid-registry/app:latest"
  }
}
```

### 4. Performance Testing
```bash
# Run benchmarks
make benchmark

# Time specific operations
time ./build/boat-local test
```

## Troubleshooting

### Build Issues
```bash
# Clean and rebuild
make clean && make build-local
```

### Import Errors
```bash
# Update dependencies
go mod tidy
```

### Permission Issues
```bash
# Make scripts executable
chmod +x debug.sh
```

## Next Steps

1. **Add Claude API Key** for full AI integration
2. **Configure AWS credentials** for real resource testing
3. **Customize sample events** for your specific use cases
4. **Run unit tests** to ensure everything works
5. **Deploy to Lambda** when ready for production

## Advanced Development

### Custom Test Scenarios
Add your scenarios to `cmd/local/main.go`:
```go
{
    name: "My Custom Scenario",
    event: map[string]any{
        "id": "custom-001",
        "type": "deploy",
        "description": "My custom deployment",
        // ...
    },
},
```

### Mock AWS Services
The local development system includes mocked AWS services for testing without real infrastructure costs.

### Integration Testing
```bash
# Test full workflow
make dev-test && make test && make lint
```