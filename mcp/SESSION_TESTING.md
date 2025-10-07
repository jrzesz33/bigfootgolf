# Testing MCP Server with Sessions

This guide shows how to test the MCP server with stateful sessions.

## Quick Test with Script

```bash
# Make sure server is running (without stateless mode)
./test_session.sh
```

## Manual Testing with curl

### Step 1: Initialize Session

The session ID is managed via cookies. Use `-c` to save cookies and `-b` to send them.

```bash
# Initialize and save session cookies
curl -v -c /tmp/mcp_cookies.txt -X POST http://localhost:3000/mcp \
  -H "Content-Type: application/json" \
  -H "Accept: application/json, text/event-stream" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {},
      "clientInfo": {
        "name": "test-client",
        "version": "1.0.0"
      }
    }
  }'
```

**Expected Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "tools": {}
    },
    "serverInfo": {
      "name": "golf-booking-server",
      "version": "1.0.0"
    }
  }
}
```

Look for the `Set-Cookie` header containing `Mcp-Session-Id`.

### Step 2: List Available Tools

```bash
# Use the session cookies from step 1
curl -b /tmp/mcp_cookies.txt -X POST http://localhost:8081/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list"
  }' | jq
```

**Expected Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "tools": [
      {
        "name": "get_weather_forecast",
        "description": "Get weather forecast for golf course area",
        "inputSchema": {
          "type": "object",
          "properties": {
            "days": {
              "type": "integer",
              "description": "Number of days to forecast (1-7)",
              "minimum": 1,
              "maximum": 7,
              "default": 3
            }
          }
        }
      },
      {
        "name": "get_available_tee_times",
        "description": "Get available tee times for a specific date",
        "inputSchema": {
          "type": "object",
          "required": ["date"],
          "properties": {
            "date": {
              "type": "string",
              "description": "Date in YYYY-MM-DD format"
            }
          }
        }
      }
    ]
  }
}
```

### Step 3: Call Weather Tool

```bash
curl -b /tmp/mcp_cookies.txt -X POST http://localhost:8081/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "get_weather_forecast",
      "arguments": {
        "days": 3
      }
    }
  }' | jq
```

### Step 4: Call Tee Times Tool

```bash
curl -b /tmp/mcp_cookies.txt -X POST http://localhost:8081/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "tools/call",
    "params": {
      "name": "get_available_tee_times",
      "arguments": {
        "date": "2024-10-07"
      }
    }
  }' | jq
```

### Step 5: Continue Using Same Session

You can make multiple requests using the same session cookies:

```bash
curl -b /tmp/mcp_cookies.txt -X POST http://localhost:8081/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 5,
    "method": "tools/call",
    "params": {
      "name": "get_weather_forecast",
      "arguments": {
        "days": 7
      }
    }
  }' | jq
```

## Using Session ID Header Directly

Alternatively, you can extract and use the session ID directly:

```bash
# Step 1: Initialize and capture session ID
SESSION_ID=$(curl -v -X POST http://localhost:8081/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {},
      "clientInfo": {
        "name": "test-client",
        "version": "1.0.0"
      }
    }
  }' 2>&1 | grep -i "Mcp-Session-Id" | awk '{print $3}' | tr -d '\r')

echo "Session ID: $SESSION_ID"

# Step 2: Use session ID in subsequent requests
curl -X POST http://localhost:8081/mcp \
  -H "Content-Type: application/json" \
  -H "Mcp-Session-Id: $SESSION_ID" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "get_weather_forecast",
      "arguments": {
        "days": 3
      }
    }
  }' | jq
```

## Testing Both Modes

### Stateful Mode (Default)
```bash
# Edit main.go - remove or comment out server.WithStateLess(true)
# Rebuild and run
go build . && ./mcp-server

# Requires initialize call and session cookies
./test_session.sh
```

### Stateless Mode (Easier for Testing)
```bash
# Edit main.go - add server.WithStateLess(true) option
# Rebuild and run
go build . && ./mcp-server

# No initialize needed, direct tool calls work
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

## Cleanup

```bash
# Remove session cookies
rm -f /tmp/mcp_cookies.txt
```

## Troubleshooting

### "Invalid session ID" error
- Make sure you initialized the session first
- Check that cookies are being saved and sent
- Verify the session ID is in the cookies file: `cat /tmp/mcp_cookies.txt`

### Session expires
- Sessions may have a timeout
- Re-run the initialize step to get a new session

### Easier alternative
- Use stateless mode (`server.WithStateLess(true)`) for simpler testing
- Sessions are not required in stateless mode
