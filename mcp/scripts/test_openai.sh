#!/bin/bash

# Test OpenAI-compatible endpoint

SERVER_URL="http://localhost:8081"

echo "=== OpenAI-Compatible API Test ==="
echo ""

echo "1. Get server info"
echo "---"
curl -s "$SERVER_URL/info" | jq
echo ""

echo "2. List available tools (OpenAI format)"
echo "---"
curl -s -X POST "$SERVER_URL/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "tools": [],
    "messages": []
  }' | jq
echo ""

echo "3. Call weather forecast tool"
echo "---"
curl -s -X POST "$SERVER_URL/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {
        "role": "user",
        "content": "get_weather_forecast with 3 days"
      }
    ],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "get_weather_forecast",
          "description": "Get weather forecast",
          "parameters": {
            "type": "object",
            "properties": {
              "days": {
                "type": "integer"
              }
            }
          }
        }
      }
    ]
  }' | jq
echo ""

echo "4. Call tee times tool"
echo "---"
TOMORROW=$(date -d '+1 day' +%Y-%m-%d 2>/dev/null || date -v+1d +%Y-%m-%d)
curl -s -X POST "$SERVER_URL/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {
        "role": "user",
        "content": "get_available_tee_times for date '$TOMORROW'"
      }
    ],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "get_available_tee_times",
          "description": "Get available tee times",
          "parameters": {
            "type": "object",
            "required": ["date"],
            "properties": {
              "date": {
                "type": "string"
              }
            }
          }
        }
      }
    ]
  }' | jq
echo ""

echo "=== Test Complete ==="
