#!/bin/bash

# Test GitHub MCP Server Connection
# This script tests if the server can connect to GitHub with your PAT

set -e

echo "🧪 Testing GitHub MCP Server Connection"
echo "======================================="

# Check if .env exists
if [ ! -f ".env" ]; then
    echo "❌ .env file not found. Run ./setup.sh first."
    exit 1
fi

# Load environment variables
source .env

# Check if PAT is configured
if [ -z "$GITHUB_PERSONAL_ACCESS_TOKEN" ] || [ "$GITHUB_PERSONAL_ACCESS_TOKEN" = "your_github_pat_here" ]; then
    echo "❌ GitHub Personal Access Token not configured in .env"
    echo "   Please edit .env and set a valid GITHUB_PERSONAL_ACCESS_TOKEN"
    exit 1
fi

echo "✅ Environment configured"
echo "🔍 Testing GitHub API access..."

# Test GitHub API access using curl
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Authorization: Bearer $GITHUB_PERSONAL_ACCESS_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    https://api.github.com/user)

if [ "$HTTP_STATUS" -eq 200 ]; then
    echo "✅ GitHub API connection successful"
    
    # Get user info
    USER_INFO=$(curl -s \
        -H "Authorization: Bearer $GITHUB_PERSONAL_ACCESS_TOKEN" \
        -H "Accept: application/vnd.github.v3+json" \
        https://api.github.com/user)
    
    USERNAME=$(echo "$USER_INFO" | grep -o '"login":"[^"]*"' | cut -d'"' -f4)
    echo "👤 Authenticated as: $USERNAME"
    
    echo ""
    echo "🎉 GitHub MCP Server is ready!"
    echo ""
    echo "Test the MCP server:"
    echo "  ./github-mcp-server stdio --toolsets context,repos"
    echo ""
    echo "Or with read-only mode:"
    echo "  ./github-mcp-server stdio --read-only"
    
elif [ "$HTTP_STATUS" -eq 401 ]; then
    echo "❌ Authentication failed (HTTP 401)"
    echo "   Check your Personal Access Token in .env"
    exit 1
elif [ "$HTTP_STATUS" -eq 403 ]; then
    echo "❌ Access forbidden (HTTP 403)"
    echo "   Your token may lack required permissions"
    echo "   Required scopes: repo, read:packages"
    exit 1
else
    echo "❌ GitHub API request failed (HTTP $HTTP_STATUS)"
    echo "   Check your internet connection and token"
    exit 1
fi