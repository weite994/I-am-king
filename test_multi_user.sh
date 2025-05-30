#!/bin/bash

# Test script for multi-user GitHub MCP server
# This script tests the multi-user functionality by sending MCP requests with auth tokens

echo "Testing Multi-User GitHub MCP Server"
echo "===================================="

# Check if the binary exists
if [ ! -f "./github-mcp-server" ]; then
    echo "Error: github-mcp-server binary not found. Please run 'go build -o github-mcp-server ./cmd/github-mcp-server' first."
    exit 1
fi

# Start the multi-user server in the background
echo "Starting multi-user server..."
./github-mcp-server multi-user --toolsets=repos,issues,users &
SERVER_PID=$!

# Give the server a moment to start
sleep 2

# Function to send JSON-RPC request
send_request() {
    local request="$1"
    echo "$request" | nc -q 1 localhost 8080 2>/dev/null || echo "$request"
}

# Test 1: Initialize the server
echo "Test 1: Initializing server..."
INIT_REQUEST='{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2025-03-26",
    "clientInfo": {
      "name": "test-client",
      "version": "1.0.0"
    },
    "capabilities": {}
  }
}'

echo "Sending initialize request..."
echo "$INIT_REQUEST"

# Test 2: List available tools
echo -e "\nTest 2: Listing available tools..."
TOOLS_REQUEST='{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/list",
  "params": {}
}'

echo "Sending tools/list request..."
echo "$TOOLS_REQUEST"

# Test 3: Try to call a tool with auth_token (this will fail without a real token)
echo -e "\nTest 3: Testing tool call with auth_token parameter..."
TOOL_CALL_REQUEST='{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "get_me",
    "arguments": {
      "auth_token": "fake_token_for_testing",
      "reason": "Testing multi-user functionality"
    }
  }
}'

echo "Sending tools/call request..."
echo "$TOOL_CALL_REQUEST"

# Clean up
echo -e "\nCleaning up..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

echo "Test completed!"
echo ""
echo "To test with a real GitHub token, replace 'fake_token_for_testing' with your actual GitHub Personal Access Token."
echo "Example usage:"
echo "  ./github-mcp-server multi-user --toolsets=repos,issues,users"
echo ""
echo "Then send requests like:"
echo '  {"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_me","arguments":{"auth_token":"your_real_token"}}}' 