# Hardware Resources & Repository Configuration

## System Specifications

This development environment is configured to utilize:

- **CPU**: Full access to 10 physical cores / 16 threads (i7-13620H)
- **GPU**: Maximum NVIDIA RTX 3050 6GB GPU utilization
- **Memory**: Limited to 14GB RAM maximum (to maintain system stability)
- **Storage**: 250GB SSD available for development
- **Current Date**: April 13, 2025

## Repository Configuration

- **Repository URL**: `git@github.com:DeanLuus22021994/vscode-docs.git`
- **Fork Name**: DeanDev
- **SSH Access**: Enabled and configured

## MCP Server Configuration

The GitHub MCP (Model Context Protocol) Server is available locally through Docker with the following setup:

- Full access to all MCP features
- Pre-configured for maximum resource utilization
- Automatically initializes with the devcontainer

## Resource Utilization Guidelines

### For CPU Optimization

- All compilation tasks use 10 physical cores / 16 threads
- Build tools are configured with parallel job execution (`-j16`)
- Node.js and Go processes utilize worker threads efficiently

### For GPU Acceleration

- CUDA 12.2 with cuDNN is available for GPU-accelerated tasks
- Docker is configured with GPU passthrough
- Rendering and compute-intensive tasks are GPU-accelerated

### For Storage Utilization

- Build caches are persisted to improve subsequent build times
- Docker layer caching is enabled to speed up container builds
- Temporary files use tmpfs for faster I/O

## Environment Variables

The following environment variables are available in the devcontainer:

- `GITHUB_PERSONAL_ACCESS_TOKEN`: For GitHub API access
- `DOCKER_ACCESS_TOKEN`: For Docker registry authentication
- `DOCKER_USERNAME`: For Docker Hub access
- `SSH_DEV_CONTAINER_REPO`: Repository SSH URL
- `OWNER`: Repository owner (DeanLuus22021994)
- `FORK_NAME`: Current fork name (DeanDev)
- `DOCKER_REGISTRY`: Docker registry URL
- `DOCKER_HOST`: Docker daemon socket

## Development Guidelines

1. **Maximize hardware usage** for builds, tests, and development tasks
2. **Utilize the MCP server** for all GitHub operations and automation
3. **Keep RAM usage under 14GB** to prevent system slowdowns
4. **Leverage GPU acceleration** for compatible workloads
5. **Use cached storage** where possible to improve performance
