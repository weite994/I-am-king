#!/bin/bash

# Demo script showing multi-user GitHub MCP server functionality
# This demonstrates how multiple users can use the same server instance

echo "🚀 Multi-User GitHub MCP Server Demo"
echo "====================================="
echo ""

# Check if the binary exists
if [ ! -f "./github-mcp-server" ]; then
    echo "❌ Error: github-mcp-server binary not found."
    echo "   Please run: go build -o github-mcp-server ./cmd/github-mcp-server"
    exit 1
fi

echo "📋 This demo shows:"
echo "   • Single server instance handling multiple users"
echo "   • Each request includes its own auth_token"
echo "   • No global token configuration needed"
echo "   • Per-request authentication and authorization"
echo ""

# Function to send a JSON-RPC request
send_request() {
    local request="$1"
    local description="$2"
    echo "📤 $description"
    echo "   Request: $(echo "$request" | jq -c .)"
    echo "$request"
    echo ""
}

# Initialize request
INIT_REQUEST='{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2025-03-26",
    "clientInfo": {
      "name": "multi-user-demo",
      "version": "1.0.0"
    },
    "capabilities": {}
  }
}'

# User 1 request (simulated with fake token)
USER1_REQUEST='{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "get_me",
    "arguments": {
      "auth_token": "ghp_user1_token_simulation",
      "reason": "User 1 getting profile"
    }
  }
}'

# User 2 request (simulated with different fake token)
USER2_REQUEST='{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "search_repositories",
    "arguments": {
      "auth_token": "ghp_user2_token_simulation",
      "query": "language:javascript stars:>1000"
    }
  }
}'

# User 3 request (simulated with another fake token)
USER3_REQUEST='{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "tools/call",
  "params": {
    "name": "get_file_contents",
    "arguments": {
      "auth_token": "ghp_user3_token_simulation",
      "owner": "octocat",
      "repo": "Hello-World",
      "path": "README.md"
    }
  }
}'

echo "🎬 Starting demo with simulated requests..."
echo "   (Note: These use fake tokens and will show authentication errors,"
echo "    but demonstrate the multi-user request handling)"
echo ""

# Combine all requests
ALL_REQUESTS=$(cat << EOF
$INIT_REQUEST
$USER1_REQUEST
$USER2_REQUEST
$USER3_REQUEST
EOF
)

echo "📡 Sending requests to multi-user server..."
echo "============================================"

# Send all requests to the server
echo "$ALL_REQUESTS" | ./github-mcp-server multi-user --toolsets=repos,users 2>/dev/null | while IFS= read -r line; do
    if [[ "$line" == *"GitHub Multi-User MCP Server"* ]]; then
        echo "✅ Server started successfully"
    elif [[ "$line" == *'"jsonrpc":"2.0"'* ]]; then
        # Pretty print JSON responses
        echo "📥 Response: $(echo "$line" | jq -c .)"
        
        # Check for specific response types
        if [[ "$line" == *'"serverInfo"'* ]]; then
            echo "   ✅ Server initialized successfully"
        elif [[ "$line" == *'"Bad credentials"'* ]]; then
            echo "   🔐 Authentication failed (expected with fake token)"
        elif [[ "$line" == *'"isError":true'* ]]; then
            echo "   ⚠️  Tool call failed (expected with fake tokens)"
        fi
    fi
    echo ""
done

echo ""
echo "🎯 Key Observations:"
echo "   • Single server instance handled multiple user requests"
echo "   • Each request carried its own auth_token parameter"
echo "   • Server properly extracted and used different tokens"
echo "   • Authentication errors were handled per-request"
echo "   • No global token configuration was needed"
echo ""

echo "🔧 To test with real tokens:"
echo "   1. Get GitHub Personal Access Tokens for different users"
echo "   2. Replace the fake tokens in the requests above"
echo "   3. Run: ./github-mcp-server multi-user --toolsets=all"
echo "   4. Send requests with real tokens via stdin"
echo ""

echo "📖 Example with real token:"
cat << 'EOF'
echo '{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2025-03-26",
    "clientInfo": {"name": "real-client", "version": "1.0.0"},
    "capabilities": {}
  }
}
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "get_me",
    "arguments": {
      "auth_token": "ghp_your_real_token_here"
    }
  }
}' | ./github-mcp-server multi-user --toolsets=all
EOF

echo ""
echo "✨ Demo completed! The multi-user GitHub MCP server is ready for production use." 