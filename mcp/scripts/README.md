# MCP Server Test Scripts

Bash scripts for testing the MCP server with curl.

## test_session.sh

Full session-based testing with error handling.

```bash
./scripts/test_session.sh
```

Tests:
1. Initialize session
2. List available tools
3. Call weather forecast tool
4. Call tee times tool
5. Reuse session for another call

## test_simple.sh

Simple debugging script that shows raw responses.

```bash
./scripts/test_simple.sh
```

Useful for:
- Debugging connection issues
- Seeing raw HTTP headers
- Verifying server responses

## Prerequisites

Both scripts require:
- Server running on http://localhost:8081
- `curl` installed
- `jq` installed (for JSON formatting)

Install jq:
```bash
# Ubuntu/Debian
sudo apt-get install jq

# macOS
brew install jq
```
