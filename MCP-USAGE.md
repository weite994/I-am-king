# GitHub MCP Server Usage Guide

This document provides instructions for using the GitHub MCP Server in the VS Code Docs devcontainer environment.

## Environment Specifications

This development environment is optimized for maximum performance with:
- **10 physical cores / 16 threads** (i7-13620H) - fully utilized
- **NVIDIA RTX 3050 6GB GPU** - maximum utilization enabled
- **14GB RAM** limit (preserving system stability)
- **250GB SSD** storage available for development
- Repository: `git@github.com:DeanLuus22021994/vscode-docs.git` (Fork: DeanDev)

## Overview

The GitHub MCP (Model Context Protocol) Server provides tools to interact with GitHub from within VS Code and AI integrations like GitHub Copilot. These tools allow for:

- Repository management (create, fork, search)
- Pull request operations
- Issue management
- Code searches
- File operations
- And more

## Setup Instructions

The MCP server has been automatically configured in your devcontainer to utilize maximum system resources while staying within specified limits. It requires a GitHub Personal Access Token to operate.

### Setting up your GitHub Token

1. Create a Personal Access Token (PAT) on GitHub:
   - Go to GitHub Settings > Developer settings > Personal access tokens
   - Create a new token with the following permissions:
     - `repo` (Full control of private repositories)
     - `workflow` (Update GitHub Action workflows)
     - `read:org` (Read organization membership)

2. Set the token in your environment:
   - Option 1: Add to your local environment before starting VS Code:
     ```bash
     export GITHUB_PERSONAL_ACCESS_TOKEN=your_token_here
     ```
   - Option 2: Add to VS Code settings (settings.json):
     ```json
     {
       "terminal.integrated.env.linux": {
         "GITHUB_PERSONAL_ACCESS_TOKEN": "your_token_here"
       }
     }
     ```

## Performance Optimization

To take advantage of the system's powerful hardware:

### For CPU-intensive Tasks
```bash
# For Node.js
NODE_OPTIONS="--max-old-space-size=12288 --max-http-header-size=16384" npm run build

# For multi-threaded builds
make -j16 build

# For Go builds with max parallelism
GOMAXPROCS=16 go build ./...
```

### For GPU-accelerated Tasks
```bash
# Run GPU-accelerated tests
GPU_ENABLED=1 npm run test-gpu

# Enable GPU for image processing
CUDA_VISIBLE_DEVICES=0 node scripts/process-images.js
```

### For Storage Optimization
```bash
# Use temporary storage for large operations
TMPDIR=/tmp node scripts/generate-docs.js

# Utilize persistent cache for builds
npm run build -- --cache-dir=/home/vscode/.cache/build-cache
```

## Verifying the Server is Running

The server should start automatically when your devcontainer launches. To verify:

1. Open a terminal in VS Code and run:
   ```bash
   if [ -f "/tmp/mcp-server.pid" ] && ps -p "$(cat /tmp/mcp-server.pid)" > /dev/null; then
     echo "MCP Server is running";
   else
     echo "MCP Server is not running";
   fi
   ```

2. Check the server logs:
   ```bash
   cat /tmp/mcp-server.log
   ```

3. Verify GPU utilization:
   ```bash
   nvidia-smi
   ```

## Manually Starting the Server

If the server is not running, you can start it manually:

```bash
/workspace/.github/start-mcp-server.sh
```

## Tools Available

The GitHub MCP Server provides the following tools:

### Repository Management
- Create repository
- Fork repository
- Search repositories
- Get file contents
- Push files

### Issues
- Create issues
- Search issues
- List issues
- Update issues
- Add issue comments

### Pull Requests
- Create pull requests
- Review pull requests
- Merge pull requests
- Update pull request branches

### Search
- Search code
- Search users

## Using with GitHub Copilot

When using GitHub Copilot in VS Code, the MCP server is automatically detected, enabling Copilot to use GitHub operations directly. Examples:

- "Create a pull request for the current branch"
- "Search for issues about CSS in this repository"
- "Create a fork of microsoft/vscode"
- "Update the documentation for the new API"

### Hardware-accelerated Copilot

This environment is configured to use hardware acceleration for GitHub Copilot, providing faster responses and more efficient processing. The GPU is utilized for local model inference when available.

## Fork-specific Information

This environment is configured for work on the DeanDev fork of the vscode-docs repository. All operations through the MCP server will use:

- Repository: `git@github.com:DeanLuus22021994/vscode-docs.git`
- Owner: DeanLuus22021994
- Fork: DeanDev

## Troubleshooting

### Server Won't Start

1. Ensure your GitHub token is set:
   ```bash
   echo $GITHUB_PERSONAL_ACCESS_TOKEN | wc -c
   ```
   (Should return a number greater than 1)

2. Check for build errors:
   ```bash
   cd /workspace/.github/cmd/github-mcp-server
   go build -v
   ```

3. Check logs for errors:
   ```bash
   cat /tmp/mcp-server.log
   ```

4. Verify GPU availability:
   ```bash
   nvidia-smi
   ```

### Server Crashes or Returns Errors

1. Restart the server:
   ```bash
   kill $(cat /tmp/mcp-server.pid)
   /workspace/.github/start-mcp-server.sh
   ```

2. Check if your token has the required permissions or has expired

3. Check system resource usage:
   ```bash
   free -h  # Check memory usage
   htop     # Check CPU usage
   ```

## Additional Resources

- [GitHub MCP Server Repository](https://github.com/github/github-mcp-server)
- [Model Context Protocol Documentation](https://containers.dev)
- [System Resource Details](./.github/RESOURCES.md)