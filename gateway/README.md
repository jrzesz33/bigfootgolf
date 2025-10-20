# Agent Gateway

This directory contains the configuration and Dockerfile for running [agentgateway](https://agentgateway.dev/) in a Docker container.

## What is agentgateway?

Agentgateway is an open source data plane optimized for agentic AI connectivity that provides drop-in security, observability, and governance for agent-to-agent and agent-to-tool communication.

## Configuration

The `config.yaml` file contains the gateway configuration with:

- **MCP Server** (port 3000): Connects to your MCP server at `http://localhost:8081/mcp`
- **Bedrock AI** (port 3001): Provides access to AWS Bedrock Claude Sonnet 4.5 model
- **Tracing**: OpenTelemetry integration for observability (configured for Phoenix)

## Running with Docker Compose

From the root of the project:

```bash
# Build and start all services including agentgateway
docker-compose up -d

# Or just start agentgateway
docker-compose up -d agentgateway

# View logs
docker-compose logs -f agentgateway
```

## Running Standalone

```bash
# Build the image
docker build -t agentgateway:local .

# Run the container
docker run -d \
  -p 3000:3000 \
  -p 3001:3001 \
  -e AWS_REGION=us-east-1 \
  -e AWS_ACCESS_KEY_ID=your_key \
  -e AWS_SECRET_ACCESS_KEY=your_secret \
  --name agentgateway \
  agentgateway:local
```

## Environment Variables

The following environment variables can be configured:

- `AWS_REGION`: AWS region for Bedrock (default: us-east-1)
- `AWS_ACCESS_KEY_ID`: AWS access key
- `AWS_SECRET_ACCESS_KEY`: AWS secret access key
- `AWS_SESSION_TOKEN`: AWS session token (if using temporary credentials)

## Endpoints

- **MCP Endpoint**: http://localhost:3000
- **Bedrock AI Endpoint**: http://localhost:3001

## Configuration Updates

To modify the gateway configuration:

1. Edit `config.yaml`
2. Rebuild the container: `docker-compose up -d --build agentgateway`

## Troubleshooting

### Check if the container is running
```bash
docker ps | grep agentgateway
```

### View container logs
```bash
docker logs agentgateway
```

### Test MCP endpoint
```bash
curl http://localhost:3000
```

### Test Bedrock endpoint
```bash
curl http://localhost:3001
```

## Additional Resources

- [Agentgateway Documentation](https://agentgateway.dev/docs)
- [GitHub Repository](https://github.com/agentgateway/agentgateway)
