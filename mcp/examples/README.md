# MCP Server Examples

Example clients and test utilities for the MCP server.

## Stateful Client Example

A complete example of using the MCP server with stateful sessions.

```bash
# Start the server first
cd /workspaces/golf_app/mcp
go run .

# In another terminal, run the client
go run examples/client.go
```

This demonstrates:
- Initializing a session
- Capturing and reusing the session ID
- Listing available tools
- Calling tools multiple times with the same session

## Integration Test Example

Manual integration tests using HTTP client utilities.

```bash
# Make sure server is running
go run examples/integration_test.go
```

## Files

- `client.go` - Stateful MCP client example with session management
- `http_client.go` - HTTP client utilities (library)
- `integration_test.go` - Integration test example

## Custom Server URL

Set the `MCP_SERVER_URL` environment variable:

```bash
MCP_SERVER_URL=http://localhost:9000/mcp go run examples/client.go
```
