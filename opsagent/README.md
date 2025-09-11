# Bigfoot Ops Agent for Technology (BOAT)

## Overview

The BOAT (Bigfoot Ops Agent for Technology) is an AI-powered cloud operations agent that automates AWS infrastructure deployment and management. It uses Claude AI to understand natural language requests and the AWS Cloud Control API via MCP (Model Context Protocol) to execute operations.

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   EventBridge   │───▶│   Lambda (BOAT)  │───▶│  Claude AI API  │
│                 │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                              │
                              ▼
                     ┌──────────────────┐
                     │ AWS Cloud Control│
                     │      API         │
                     └──────────────────┘
                              │
                              ▼
                     ┌──────────────────┐
                     │  AWS Resources   │
                     │ (ECS, ALB, etc.) │
                     └──────────────────┘
```

### Key Components

1. **Lambda Handler** (`main.go`) - Processes EventBridge events
2. **BOAT Agent** (`internal/agent/`) - Core AI-powered agent logic
3. **Claude Integration** - Natural language processing and planning
4. **AWS Manager** (`internal/aws/`) - AWS resource management
5. **MCP Server** - Model Context Protocol for AI tool access
6. **Validation System** (`pkg/validation/`) - Container requirement validation
7. **Deployment Engine** (`pkg/deployment/`) - Resource deployment orchestration
8. **Monitoring & Alerts** (`internal/monitoring/`) - Comprehensive observability

## Technology Preferences

### Cost Optimization
- **Minimal costs**: Stay within AWS free tier whenever possible
- **Fargate over EC2**: Reduce management overhead
- **Default VPC usage**: Minimize networking costs
- **Conservative resource sizing**: Start small, scale as needed

### Container Requirements
All deployments must specify:
- **Ports**: Container and host port mappings
- **Dynamic Secrets**: Secrets to be created during deployment
- **Existing Secrets**: Pre-existing secret references
- **Environment Variables**: Configuration parameters
- **External Egress**: Required internet access (specific URLs/IPs)
- **Public Ports**: Ports to expose via load balancer
- **Health Checks**: Endpoint, protocol, intervals

### Registry Restriction
- **Only** pull containers from `ghcr.io/jrzesz33/` registry
- Validates image sources during deployment planning

## Usage

### Deployment

1. **Set environment variables**:
   ```bash
   export CLAUDE_API_KEY="your-claude-api-key"
   export NOTIFICATION_EMAIL="your-email@example.com"
   ```

2. **Deploy to AWS**:
   ```bash
   make deploy
   ```

3. **Send a task request**:
   ```bash
   make test-event
   ```

### Local Development

1. **Build locally**:
   ```bash
   make build-local
   ```

2. **Run tests**:
   ```bash
   make test
   ```

3. **Lint code**:
   ```bash
   make lint
   ```

### Task Request Format

Send events to EventBridge with source `boat.ops`:

```json
{
  "Source": "boat.ops",
  "DetailType": "Container Deployment Request",
  "Detail": {
    "id": "task-001",
    "type": "deploy",
    "description": "Deploy user authentication service with Redis cache",
    "parameters": {
      "image": "ghcr.io/jrzesz33/auth-service:v1.2.0",
      "replicas": 2,
      "cache_required": true
    },
    "priority": "high",
    "requester": "dev-team"
  }
}
```

## Agent Capabilities

### 1. Container Deployment
- Validates container requirements
- Creates ECS services with Fargate
- Sets up load balancers and security groups
- Configures health checks and monitoring

### 2. Infrastructure Management
- Uses AWS Cloud Control API for resource management
- Tracks resource dependencies
- Estimates costs before deployment
- Ensures free tier compliance

### 3. Intelligent Planning
- Claude AI analyzes natural language requests
- Generates optimized deployment plans
- Identifies missing requirements
- Suggests cost optimizations

### 4. Monitoring & Alerting
- Comprehensive metrics collection
- Cost threshold monitoring
- Error rate tracking
- Performance analytics

## File Structure

```
opsagent/
├── cmd/
│   └── lambda/              # Lambda entry point
├── internal/
│   ├── agent/              # Core BOAT agent logic
│   ├── aws/                # AWS service integrations
│   ├── config/             # Configuration management
│   ├── models/             # Data models
│   ├── monitoring/         # Metrics and alerting
│   └── notifications/      # Notification system
├── pkg/
│   ├── deployment/         # Deployment orchestration
│   └── validation/         # Requirement validation
├── templates/              # CloudFormation templates
│   ├── ecs-service.json    # ECS service template
│   └── lambda-eventbridge.json # Lambda deployment
├── Makefile               # Build and deployment commands
└── main.go               # Alternative entry point
```

## Monitoring

### Metrics Collected
- Task execution duration and success rates
- AWS resource creation metrics
- Claude API usage and response times
- Cost estimation accuracy
- Error rates and failure patterns

### Alerting Thresholds
- **Cost**: $100 monthly threshold
- **Error Rate**: 20% failure rate
- **Duration**: 10-minute task timeout
- **Consecutive Failures**: 3 failures trigger alert

### Log Structure
All logs use structured JSON format:

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "event_type": "task_completed",
  "task_id": "task-001",
  "status": "success",
  "duration_ms": 45000,
  "resources_created": 5,
  "estimated_cost": 16.20
}
```

## Security

### IAM Permissions
The Lambda function requires permissions for:
- CloudFormation stack operations
- ECS service management
- EC2 VPC and security group access
- Elastic Load Balancing operations
- CloudWatch Logs

### Secret Management
- Claude API key stored as environment variable
- AWS credentials via IAM role
- Container secrets managed via AWS Secrets Manager
- No hardcoded credentials in code

## Cost Optimization Features

### Automatic Optimizations
- **Fargate Spot**: Uses Fargate for cost-effective compute
- **Minimal Resources**: Starts with 256 CPU / 512 MB memory
- **Default VPC**: Uses existing networking infrastructure
- **Log Retention**: 7-14 day log retention to minimize storage costs

### Cost Monitoring
- Real-time cost estimation before deployment
- Free tier compliance checking
- Resource usage tracking
- Monthly cost projections

## Troubleshooting

### Common Issues

1. **Missing Requirements Alert**:
   - Check container image registry (`ghcr.io/jrzesz33/`)
   - Verify all required fields in task request
   - Review validation error logs

2. **Cost Threshold Exceeded**:
   - Review resource requirements
   - Consider smaller instance sizes
   - Check if resources can be shared

3. **Deployment Failures**:
   - Check IAM permissions
   - Verify VPC/subnet availability
   - Review CloudWatch logs

### Debugging Commands

```bash
# Check deployment status
make status

# View recent logs
make logs

# Send test event
make test-event

# Clean and rebuild
make clean && make build
```

## Contributing

### Development Setup

1. **Install tools**:
   ```bash
   make install-tools
   ```

2. **Run quality checks**:
   ```bash
   make test lint vet security
   ```

3. **Local testing**:
   ```bash
   make run-local
   ```
    
### Code Standards
- Go 1.23+ required
- Use structured logging (logrus with JSON format)
- Comprehensive error handling
- Unit tests for all public functions
- Security-first development practices

## Future Enhancements

### Planned Features
- Multi-region deployment support
- Advanced cost optimization algorithms
- Integration with more AWS services
- Custom deployment templates
- Web-based management console
- Slack/Teams integration for notifications

### MCP Extensions
- Additional tool integrations
- Custom tool development
- Enhanced AI agent capabilities
- Multi-model AI support

## License

This project is part of the Bigfoot Golf application suite. See the main project license for details.