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
echo "Testing stdio mode..."
# Test stdio mode with a simple initialize request
echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {}, "clientInfo": {"name": "test", "version": "1.0.0"}}}' | timeout 2s ./mcp_server -mode stdio 2>&1 | grep -q "protocolVersion" && echo "✓ Stdio mode test passed" || echo "✗ Stdio mode test failed (may need DB connection)"

echo ""
echo "MCP server is ready to run with:"
echo "  Production mode: ./mcp_server -mode server"
echo "  Development proxy: ./mcp_server -mode proxy"
echo "  Stdio mode: ./mcp_server -mode stdio"
echo ""
echo "Environment variables needed:"
echo "  JWT_SECRET - for token validation"
echo "  DB_ADMIN - Neo4j database password"
echo "  MCP_PORT - Server port (default: 8081)"
echo "  PROXY_PORT - Proxy port (default: 8082)"