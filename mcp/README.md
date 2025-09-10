# MCP Server for Golf Booking Application

This module implements a Model Context Protocol (MCP) server that provides AI agents with secure access to reservation management tools.

## Features

- **OAuth Authentication**: Secure token-based authentication for all MCP requests
- **Reservation Management**: Get, book, and cancel golf tee time reservations
- **Tee Time Discovery**: Find available tee times based on date, time range, and party size
- **Weather & Conditions**: Get current weather forecasts and course conditions

## Architecture

The MCP server consists of three main components:

1. **MCP Server** (`server.go`): The main server that handles MCP protocol requests with OAuth authentication
2. **Development Proxy** (`proxy.go`): A proxy server for local development that handles CORS and authentication
3. **MCP Client** (`pkg/models/anthropic/mcp_integration.go`): Client library for the AI agent to communicate with the MCP server

## Setup

### Prerequisites

- Go 1.23.4 or higher
- Neo4j database running on `bolt://localhost:7687`
- JWT secret configured via `JWT_SECRET` environment variable

### Installation

1. Install dependencies:
```bash
cd /workspaces/golf_app
go work sync
```

2. Set environment variables:
```bash
export JWT_SECRET="your-secret-key"
export DB_ADMIN="your-neo4j-password"
export ANTHROPIC_API_KEY="your-anthropic-api-key"
```

## Running the MCP Server

### Production Mode

Run the MCP server directly:

```bash
cd mcp
go run . -mode server
# Or set environment variable
MCP_MODE=server go run .
```

The server will start on port 8081 (or the port specified by `MCP_PORT` environment variable).

### Development Mode with Proxy

For local development with CORS support:

```bash
cd mcp
go run . -mode proxy
# Or set environment variable
MCP_MODE=proxy go run .
```

The proxy server will start on port 8082 (or the port specified by `PROXY_PORT` environment variable).

## Integration with AI Agent

The AI agent in the application can be configured to use the MCP server:

### Enable MCP in the Agent Controller

```go
// In your handler that creates the agent
agent := controllers.NewAgentController()
err := agent.SetUserInfo(userID, userEmail, true) // true enables MCP
if err != nil {
    // Handle error
}
```

### Development Configuration

For local development, set the following environment variables:

```bash
# Enable MCP proxy for development
export MCP_USE_PROXY=true
export MCP_SERVER_URL=http://localhost:8082

# Or connect directly to MCP server
export MCP_USE_PROXY=false
export MCP_SERVER_URL=http://localhost:8081
```

## Available MCP Tools

### 1. manage_reservations

Manage user reservations (get, book, cancel).

**Parameters:**
- `action` (required): "get", "book", or "cancel"
- `user_id` (required): User ID
- `reservation_id`: Required for cancel action
- `tee_time`: ISO format datetime, required for book action
- `players`: Number of players, required for book action

### 2. find_tee_times

Find available tee times.

**Parameters:**
- `date` (required): Date in YYYY-MM-DD format
- `time_range`: "morning", "midday", "afternoon", or "all"
- `players`: Number of players

### 3. get_conditions

Get weather and course conditions.

**Parameters:**
- `date`: Date in YYYY-MM-DD format (optional, defaults to today)

## Testing

### Test the MCP Server

1. Start the MCP server:
```bash
cd mcp
go run . -mode server
```

2. Generate a test token:
```bash
# Use the JWT helper to generate a token
# You can create a test script or use the existing auth endpoints
```

3. Test with curl:
```bash
curl -X POST http://localhost:8081/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "tool": "find_tee_times",
    "params": {
      "date": "2024-09-10",
      "players": 4
    }
  }'
```

### Test with Development Proxy

1. Start the proxy:
```bash
cd mcp
go run . -mode proxy
```

2. Test without authentication (proxy handles dev tokens):
```bash
curl -X POST http://localhost:8082/proxy/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "get_conditions",
    "params": {}
  }'
```

## Security Notes

- All MCP requests require valid JWT tokens
- Tokens are validated against the auth server's secret
- User permissions are enforced at the tool level
- The development proxy should NEVER be used in production

## Troubleshooting

### Common Issues

1. **Authentication failures**: Ensure JWT_SECRET is set correctly
2. **Database connection errors**: Verify Neo4j is running and DB_ADMIN is set
3. **CORS errors in development**: Use the proxy server with `-mode proxy`
4. **Port conflicts**: Change ports using MCP_PORT and PROXY_PORT environment variables

### Logging

The MCP server and proxy include detailed logging for debugging:
- Request/response bodies are logged in development mode
- Authentication failures are logged with details
- Tool execution errors include stack traces

## Future Enhancements

- [ ] Add rate limiting for MCP requests
- [ ] Implement request/response caching
- [ ] Add metrics and monitoring
- [ ] Support for additional golf-specific tools
- [ ] WebSocket support for real-time updates