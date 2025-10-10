#!/bin/bash

# Simple MCP Server Test - Shows raw responses for debugging

SERVER_URL="http://localhost:8081/mcp"

echo "=== Simple MCP Server Test ==="
echo ""

echo "Test 1: Initialize (with verbose output)"
echo "---"
curl -v -X POST "$SERVER_URL" \
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
  }' 2>&1 | head -50

echo ""
echo "---"
echo ""

echo "Test 2: Tools list (raw output)"
echo "---"
curl -X POST "$SERVER_URL" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list"
  }'

echo ""
echo "---"
echo ""

echo "Test 3: Call weather tool (raw output)"
echo "---"
curl -X POST "$SERVER_URL" \
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
  }'

echo ""
echo "---"
