#!/bin/bash

# Test MCP Server with Session Management
# This script demonstrates how to use the MCP server with stateful sessions

SERVER_URL="http://localhost:8081/mcp"
COOKIE_FILE="/tmp/mcp_session_cookies.txt"

echo "=== MCP Server Session Test ==="
echo ""

# Clean up old cookies
rm -f "$COOKIE_FILE"

echo "Step 1: Initialize session"
echo "---"
INIT_RESPONSE=$(curl -s -c "$COOKIE_FILE" -X POST "$SERVER_URL" \
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
  }')

# Check if response is valid JSON
if echo "$INIT_RESPONSE" | jq empty 2>/dev/null; then
    echo "$INIT_RESPONSE" | jq
else
    echo "ERROR: Initialize failed. Raw response:"
    echo "$INIT_RESPONSE"
    echo ""
    echo "Make sure the server is running: go run ."
    exit 1
fi
echo ""

# Extract session ID from cookies if needed
if [ -f "$COOKIE_FILE" ]; then
    SESSION_ID=$(grep "Mcp-Session-Id" "$COOKIE_FILE" | awk '{print $7}')
    if [ -n "$SESSION_ID" ]; then
        echo "Session ID: $SESSION_ID"
        echo ""
    fi
fi

echo "Step 2: List available tools"
echo "---"
LIST_RESPONSE=$(curl -s -b "$COOKIE_FILE" -X POST "$SERVER_URL" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list"
  }')

# Check if response is valid JSON before piping to jq
if echo "$LIST_RESPONSE" | jq empty 2>/dev/null; then
    echo "$LIST_RESPONSE" | jq
else
    echo "Raw response (not JSON):"
    echo "$LIST_RESPONSE"
fi
echo ""

echo "Step 3: Call get_weather_forecast tool"
echo "---"
WEATHER_RESPONSE=$(curl -s -b "$COOKIE_FILE" -X POST "$SERVER_URL" \
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
  }')

if echo "$WEATHER_RESPONSE" | jq empty 2>/dev/null; then
    echo "$WEATHER_RESPONSE" | jq
else
    echo "Raw response (not JSON):"
    echo "$WEATHER_RESPONSE"
fi
echo ""

echo "Step 4: Call get_available_tee_times tool"
echo "---"
TOMORROW=$(date -d '+1 day' +%Y-%m-%d 2>/dev/null || date -v+1d +%Y-%m-%d)
TEETIMES_RESPONSE=$(curl -s -b "$COOKIE_FILE" -X POST "$SERVER_URL" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "tools/call",
    "params": {
      "name": "get_available_tee_times",
      "arguments": {
        "date": "'"$TOMORROW"'"
      }
    }
  }')

if echo "$TEETIMES_RESPONSE" | jq empty 2>/dev/null; then
    echo "$TEETIMES_RESPONSE" | jq
else
    echo "Raw response (not JSON):"
    echo "$TEETIMES_RESPONSE"
fi
echo ""

echo "Step 5: Another weather call (reusing session)"
echo "---"
WEATHER2_RESPONSE=$(curl -s -b "$COOKIE_FILE" -X POST "$SERVER_URL" \
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
  }')

if echo "$WEATHER2_RESPONSE" | jq empty 2>/dev/null; then
    echo "$WEATHER2_RESPONSE" | jq
else
    echo "Raw response (not JSON):"
    echo "$WEATHER2_RESPONSE"
fi
echo ""

echo "=== Test Complete ==="
echo "Session cookies saved to: $COOKIE_FILE"

# Clean up
# rm -f "$COOKIE_FILE"
