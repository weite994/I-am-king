# Local GitHub MCP Server Docker Image

This directory contains a Docker setup to build a local version of the GitHub MCP Server using the official image as a base and overriding it with your locally built executable.

## Quick Start

1. **Build the image:**
   ```bash
   ./build.sh
   ```

2. **Use the image:**
   ```bash
   docker run --rm -e GITHUB_PERSONAL_ACCESS_TOKEN=your_token github-mcp-server-local stdio
   ```

## Manual Build

If you prefer to build manually:

1. **Build the binary for Linux ARM64:**
   ```bash
   cd ..
   GOOS=linux GOARCH=arm64 go build -o docker-local/github-mcp-server cmd/github-mcp-server/main.go
   ```

2. **Build the Docker image:**
   ```bash
   cd docker-local
   docker build -t github-mcp-server-local .
   ```

## Usage

The image works exactly like the official `ghcr.io/github/github-mcp-server` image, but uses your locally built binary instead.

### Environment Variables

- `GITHUB_PERSONAL_ACCESS_TOKEN`: Your GitHub Personal Access Token (required)

### Example with VS Code

Add this to your VS Code MCP configuration:

```json
{
  "servers": {
    "github": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-e",
        "GITHUB_PERSONAL_ACCESS_TOKEN",
        "github-mcp-server-local"
      ],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "${input:github_token}"
      }
    }
  },
  "inputs": [
    {
      "type": "promptString",
      "id": "github_token",
      "description": "GitHub Personal Access Token",
      "password": true
    }
  ]
}
```

## Advantages

- **Fast builds**: Uses the official image as base, only rebuilds your local changes
- **No build issues**: Avoids SSL certificate and dependency issues from building from scratch
- **Consistent environment**: Uses the same runtime environment as the official image
- **Easy updates**: Just rebuild the binary and Docker image when you make changes 