# Testing the MCP Server

Quick reference for testing the MCP server with stateful sessions.

## Start the Server

```bash
cd /workspaces/golf_app/mcp
go run .
```

Or build and run:
```bash
go build -o mcp-server .
./mcp-server
```

## Test with Go Client (Recommended)

The easiest way to test stateful sessions:

```bash
go run examples/client.go
```

This automatically:
- Initializes the session
- Lists available tools
- Calls both tools (weather and tee times)
- Reuses the session across multiple calls

**Expected output:**
```
=== MCP Stateful Client Test ===
Server: http://localhost:8081/mcp

🔧 Initializing MCP session...
📌 Session ID: abc123...
✅ Session initialized

📋 Listing available tools...
{
  "tools": [...]
}

🌤️  Calling weather forecast tool (days: 3)...
{
  "content": [...]
}

⛳ Calling tee times tool (date: 2024-10-07)...
{
  "content": [...]
}

✅ All tests completed successfully!
```

## Test with Bash Scripts

### Full Session Test
```bash
./scripts/test_session.sh
```

### Simple Debug Test
```bash
./scripts/test_simple.sh
```

## Manual curl Testing

See [SESSION_TESTING.md](SESSION_TESTING.md) for detailed curl examples.

### Quick curl Test (requires session)

```bash
# 1. Initialize session and save session ID
SESSION_ID=$(curl -v -X POST http://localhost:8081/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  2>&1 | grep -i "Mcp-Session-Id" | awk '{print $3}' | tr -d '\r')

echo "Session ID: $SESSION_ID"

# 2. Call tools with session ID
curl -X POST http://localhost:8081/mcp \
  -H "Content-Type: application/json" \
  -H "Mcp-Session-Id: $SESSION_ID" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_weather_forecast","arguments":{"days":3}}}' | jq
```

## Run Unit Tests

```bash
go test -v
```

## Troubleshooting

### "Invalid session ID" error
- Make sure you initialized the session first
- Verify the session ID header is being sent
- Use the Go client (`examples/client.go`) which handles this automatically

### Server not responding
```bash
# Check if server is running
curl http://localhost:8081/health

# Check server logs
go run . 2>&1 | tee server.log
```

### Connection refused
- Ensure server is running on port 8081
- Check if another process is using the port: `lsof -i :8081`

## Quick Reference

| Test Method | Complexity | Session Handling |
|-------------|-----------|------------------|
| `go run examples/client.go` | ⭐ Easy | Automatic |
| `./scripts/test_session.sh` | ⭐⭐ Medium | Automatic (cookies) |
| Manual curl | ⭐⭐⭐ Hard | Manual |

**Recommendation**: Use `go run examples/client.go` for most testing.
