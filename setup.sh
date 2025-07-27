#!/bin/bash

# GitHub MCP Server Setup Script
# This script helps you set up the GitHub MCP server with proper security

set -e

echo "🔧 GitHub MCP Server Setup"
echo "=========================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.23+ first."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✅ Go version: $GO_VERSION"

# Build the server
echo "🔨 Building GitHub MCP server..."
go build -o github-mcp-server ./cmd/github-mcp-server
if [ -f "./github-mcp-server" ]; then
    echo "✅ Build successful!"
    echo "📍 Binary location: $(pwd)/github-mcp-server"
else
    echo "❌ Build failed!"
    exit 1
fi

# Create .env file if it doesn't exist
if [ ! -f ".env" ]; then
    echo "📝 Creating .env file..."
    cp .env.example .env
    echo "✅ .env file created from example"
    echo "⚠️  Please edit .env and add your GitHub Personal Access Token"
    echo "   Create a token at: https://github.com/settings/personal-access-tokens/new"
    echo ""
else
    echo "✅ .env file already exists"
fi

# Check if PAT is set
if [ -f ".env" ]; then
    source .env
    if [ -z "$GITHUB_PERSONAL_ACCESS_TOKEN" ] || [ "$GITHUB_PERSONAL_ACCESS_TOKEN" = "your_github_pat_here" ]; then
        echo "⚠️  GitHub Personal Access Token not configured"
        echo "   Please edit .env and set GITHUB_PERSONAL_ACCESS_TOKEN"
        echo ""
    else
        echo "✅ GitHub PAT configured"
    fi
fi

# Test the server
echo "🧪 Testing server..."
if ./github-mcp-server --help > /dev/null 2>&1; then
    echo "✅ Server executable works correctly"
else
    echo "❌ Server test failed"
    exit 1
fi

echo ""
echo "🎉 Setup complete!"
echo ""
echo "Next steps:"
echo "1. Edit .env file and add your GitHub Personal Access Token"
echo "2. Test the server: ./github-mcp-server stdio"
echo "3. Configure your MCP host (Claude, VS Code, etc.)"
echo ""
echo "For MCP host configuration examples, see:"
echo "- docs/installation-guides/"
echo "- README.md"
echo ""
echo "Security reminders:"
echo "- Never commit your .env file to version control"
echo "- Use minimum required scopes for your PAT"
echo "- Rotate tokens regularly"