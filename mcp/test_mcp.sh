#!/bin/bash

echo "Testing MCP Server compilation and startup..."

# Test compilation
echo "Building MCP server..."
cd /workspaces/golf_app/mcp
if go build -o mcp_server .; then
    echo "✓ MCP server compiled successfully"
else
    echo "✗ MCP server compilation failed"
    exit 1
fi

echo ""
echo "MCP server is ready to run with:"
echo "  Production mode: ./mcp_server -mode server"
echo "  Development proxy: ./mcp_server -mode proxy"
echo ""
echo "Environment variables needed:"
echo "  JWT_SECRET - for token validation"
echo "  DB_ADMIN - Neo4j database password"
echo "  MCP_PORT - Server port (default: 8081)"
echo "  PROXY_PORT - Proxy port (default: 8082)"