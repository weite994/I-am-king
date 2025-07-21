#!/bin/bash

# Build script for local GitHub MCP Server Docker image

set -e

echo "Building local GitHub MCP Server binary..."
cd ..
GOOS=linux GOARCH=arm64 go build -o docker-local/github-mcp-server cmd/github-mcp-server/main.go

echo "Building Docker image..."
cd docker-local
docker build -t github-mcp-server-local .

echo "Docker image built successfully!"
echo "You can now use it with:"
echo "  docker run --rm -e GITHUB_PERSONAL_ACCESS_TOKEN=your_token github-mcp-server-local stdio" 