FROM ghcr.io/github/github-mcp-server:latest

# Copy our locally built executable and set permissions in one step
COPY --chmod=755 github-mcp-server /server/github-mcp-server

# The base image already has the correct ENTRYPOINT and CMD 