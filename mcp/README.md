# Golf Booking MCP Server

A Model Context Protocol (MCP) server for golf tee time booking and weather forecasts, implemented with StreamableHTTP transport in Go.

## Features

- **StreamableHTTP Transport**: HTTP-based MCP server with streamable responses
- **Real Weather Data**: Fetches live weather forecasts from NOAA Weather API
- **Real Tee Time Data**: Queries actual tee time availability from Neo4j database
- **Database Integration**: Connects to Neo4j for tee time and booking data
- **Health Monitoring**: Built-in health check and server info endpoints
- **Comprehensive Testing**: Full unit test suite and integration test utilities

## Architecture

- **Transport**: StreamableHTTP (via `github.com/mark3labs/mcp-go`)
- **Tools**: Matches the functionality in `pkg/models/bigfoottools.go`
- **Port**: 8081 (configurable via `PORT` environment variable)

## Prerequisites

- Go 1.21 or higher
- Neo4j database running on `bolt://localhost:7687`
- Environment variables:
  - `DB_URI`: Neo4j connection URI (default: `bolt://localhost:7687`)
  - `DB_ADMIN`: Neo4j password
  - `PORT`: Server port (default: 8081)

## Installation

```bash
cd mcp
go mod download
```

## Running the Server

### Standard Mode

```bash
go run .
```

### Custom Port

```bash
PORT=9000 go run .
```

## API Endpoints

### MCP Endpoint

- **URL**: `/mcp`
- **Method**: POST
- **Content-Type**: `application/json`
- **Description**: Main MCP protocol endpoint for tool calls

### Health Check

- **URL**: `/health`
- **Method**: GET
- **Response**: `{"status":"healthy"}`

### Server Info

- **URL**: `/info`
- **Method**: GET
- **Response**: Server metadata including name, version, and available endpoints

## Available Tools

### 1. get_weather_forecast

Get real-time weather forecast from NOAA Weather API for the golf course area.

**Parameters:**
- `days` (integer, optional): Number of days to forecast (1-7, default: 3)

**Example Request:**
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

**Example Response:**
```
Golf course weather forecast:
- This Afternoon: Partly Sunny, 72°F, 5 mph SW
- Tonight: Partly Cloudy, 58°F, 5 mph S
- Monday: Mostly Sunny, 75°F, 5 to 10 mph S
- Monday Night: Partly Cloudy, 60°F, 5 mph S

Perfect for planning your golf outing!
```

**Data Source**: NOAA Weather API (weather.gov)

### 2. get_available_tee_times

Get real tee time availability from Neo4j database for a specific date.

**Parameters:**
- `date` (string, required): Date in YYYY-MM-DD format (e.g., "2024-01-15")

**Example Request:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_available_tee_times",
    "arguments": {
      "date": "2024-10-07"
    }
  }
}
```

**Example Response:**
```
Available tee times for 2024-10-07:
- 7:00 AM | Slot 1 | 2 spots available | $35.00 | Members
- 7:15 AM | Slot 2 | 4 spots available | $35.00 | Public
- 8:00 AM | Slot 5 | 3 spots available | $45.00 | Members
- 9:30 AM | Slot 10 | 1 spots available | $45.00 | Public
```

**Data Source**: Neo4j database with real booking engine data

## Testing

### Run Unit Tests

```bash
go test -v
```

### Test with curl

See [SESSION_TESTING.md](SESSION_TESTING.md) for detailed session testing guide.

#### Quick Test (Stateless Mode)

Enable stateless mode in `main.go` for easier testing:
```go
streamableServer := server.NewStreamableHTTPServer(
    mcpServer.GetServer(),
    server.WithEndpointPath("/mcp"),
    server.WithStateLess(true), // Add this line
)
```

Then test directly:
```bash
curl -X POST http://localhost:8081/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "get_weather_forecast",
      "arguments": {"days": 3}
    }
  }' | jq
```

#### Test with Sessions (Stateful Mode)

Use the provided script:
```bash
./test_session.sh
```

Or manually with cookies:
```bash
# 1. Initialize session
curl -c /tmp/mcp_cookies.txt -X POST http://localhost:8081/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}'

# 2. Call tools using session
curl -b /tmp/mcp_cookies.txt -X POST http://localhost:8081/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_weather_forecast","arguments":{"days":3}}}' | jq
```

### Run Integration Tests

First, start the server in one terminal:

```bash
go run .
```

Then, in another terminal, run the integration test client:

```bash
go run test_main.go test_client.go
```

Or test against a custom server URL:

```bash
MCP_SERVER_URL=http://localhost:9000 go run test_main.go test_client.go
```

### Test Coverage

```bash
go test -cover
```

## Development

### Project Structure

```
mcp/
├── server.go              # MCP server implementation with tool handlers
├── main.go                # HTTP server and endpoint setup
├── server_test.go         # Unit tests for server and tools
├── Makefile               # Build and test automation
├── go.mod                 # Go module dependencies
├── README.md              # Full documentation
├── QUICKSTART.md          # Quick start guide
├── SESSION_TESTING.md     # Detailed session testing guide
├── examples/              # Example clients and integration tests
│   ├── client.go          # Stateful client example
│   ├── http_client.go     # HTTP utilities library
│   ├── integration_test.go# Integration test example
│   └── README.md          # Examples documentation
└── scripts/               # Bash test scripts
    ├── test_session.sh    # Session-based curl tests
    ├── test_simple.sh     # Simple debugging script
    └── README.md          # Scripts documentation
```

### Adding New Tools

1. Define the tool schema in `registerTools()` method in `server.go`
2. Implement the handler function (e.g., `handleToolName()`)
3. Add unit tests in `server_test.go`
4. Update this README with the new tool documentation

### Example: Adding a New Tool

```go
// In server.go - registerTools()
mcpServer.AddTool(mcp.Tool{
    Name:        "your_tool_name",
    Description: "Your tool description",
    InputSchema: mcp.ToolInputSchema{
        Type:     "object",
        Required: []string{"param1"},
        Properties: map[string]interface{}{
            "param1": map[string]interface{}{
                "type":        "string",
                "description": "Parameter description",
            },
        },
    },
}, s.handleYourTool)

// Implement handler
func (s *MCPServer) handleYourTool(args map[string]interface{}) (*mcp.CallToolResult, error) {
    // Extract parameters
    param1, ok := args["param1"].(string)
    if !ok {
        return mcp.NewToolResultError("param1 is required"), nil
    }

    // Process and return result
    result := processData(param1)
    resultJSON, _ := json.Marshal(result)
    return mcp.NewToolResultText(string(resultJSON)), nil
}
```

## Integration with Main Application

To integrate this MCP server with the main golf application, update `pkg/models/bigfoottools.go`:

```go
func GetBigfootTools(useMCP bool) []BigfootTool {
    if useMCP {
        // Connect to MCP server at http://localhost:8081/mcp
        // Use MCP client to discover and call tools
        return getMCPTools("http://localhost:8081/mcp")
    }
    // ... existing direct tools
}
```

## Dependencies

- `github.com/mark3labs/mcp-go` v0.39.1 - MCP protocol implementation

## License

Part of the Bigfoot Golf Application

## Contributing

When contributing, ensure:
1. All tests pass (`go test -v`)
2. Code follows Go conventions (`go fmt`, `go vet`)
3. New features include unit tests
4. README is updated with new functionality
