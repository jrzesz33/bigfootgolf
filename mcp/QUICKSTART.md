# MCP Server Quick Start Guide

## Prerequisites

- Go 1.21 or higher
- Port 8081 available (or configure a custom port)

## Installation

```bash
cd /workspaces/golf_app/mcp
go mod download
```

## Running the Server

### Option 1: Using `go run`

```bash
go run .
```

### Option 2: Build and run binary

```bash
go build -o mcp-server .
./mcp-server
```

### Option 3: Custom port

```bash
PORT=9000 go run .
```

### Option 4: Using Makefile

```bash
make run           # Run on default port 8081
make build         # Build binary
make test          # Run unit tests
make help          # See all available commands
```

## Testing the Server

### 1. Health Check

```bash
curl http://localhost:8081/health
```

Expected response:
```json
{"status":"healthy"}
```

### 2. Server Info

```bash
curl http://localhost:8081/info
```

Expected response:
```json
{
  "name": "golf-booking-server",
  "version": "1.0.0",
  "description": "MCP server for golf tee time booking and weather forecasts",
  "transport": "streamable-http",
  "endpoint": "/mcp"
}
```

### 3. Run Unit Tests

```bash
go test -v
```

### 4. Run Integration Tests

Start the server first:
```bash
go run .
```

In another terminal:
```bash
go run test_main.go test_client.go
```

## Using the MCP Tools

The server provides two tools via the MCP protocol:

### Tool 1: get_weather_forecast

Get weather forecast for the golf course area.

**Example MCP Request:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_weather_forecast",
    "arguments": {
      "days": 5
    }
  }
}
```

### Tool 2: get_available_tee_times

Get available tee times for a specific date.

**Example MCP Request:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_available_tee_times",
    "arguments": {
      "date": "2024-12-25"
    }
  }
}
```

## Architecture

```
┌─────────────────┐
│  MCP Client     │
│  (AI Agent)     │
└────────┬────────┘
         │ HTTP POST
         │ /mcp
         ▼
┌─────────────────────────────┐
│  StreamableHTTP Server      │
│  (Port 8081)                │
├─────────────────────────────┤
│  ├─ /mcp                    │
│  ├─ /health                 │
│  └─ /info                   │
└─────────┬───────────────────┘
          │
          ▼
┌─────────────────────────────┐
│  MCP Server                 │
│  ├─ get_weather_forecast    │
│  └─ get_available_tee_times │
└─────────────────────────────┘
```

## Development

### Project Structure

```
mcp/
├── server.go          # Core MCP server logic and tool handlers
├── main.go            # HTTP server setup with StreamableHTTP
├── server_test.go     # Unit tests
├── test_client.go     # Integration test client
├── test_main.go       # Integration test runner
├── Makefile           # Build and test automation
├── README.md          # Full documentation
├── QUICKSTART.md      # This file
└── go.mod             # Dependencies
```

### Adding a New Tool

1. Add tool definition in `registerTools()` in `server.go`
2. Implement handler function `handleYourTool(ctx, request)`
3. Add tests in `server_test.go`
4. Update README.md

### Running with Debug Output

```bash
go run . 2>&1 | tee server.log
```

## Troubleshooting

### Port Already in Use

```bash
# Check what's using the port
lsof -i :8081

# Use a different port
PORT=9000 go run .
```

### Build Errors

```bash
# Clean and rebuild
go clean
go mod tidy
go build .
```

### Test Failures

```bash
# Run tests with verbose output
go test -v

# Run specific test
go test -v -run TestHandleGetWeatherForecast
```

## Next Steps

- See [README.md](README.md) for full documentation
- Check `test_main.go` for integration test examples
- Review `server_test.go` for unit test examples
- Integrate with main application via `pkg/models/bigfoottools.go`
