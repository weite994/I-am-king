# Multi-User GitHub MCP Server

This is a modified version of GitHub's official MCP server that supports multiple users with a single server instance, instead of requiring separate Docker instances per user.

## Key Changes

### Original Architecture
- Single user per server instance
- GitHub Personal Access Token (PAT) provided via `GITHUB_PERSONAL_ACCESS_TOKEN` environment variable
- Token set at server startup and used for all requests

### New Multi-User Architecture
- Multiple users per server instance
- GitHub PAT provided with each individual request via `auth_token` parameter
- No global token required at server startup
- Each tool request creates a new GitHub client with the provided token

## Usage

### Building
```bash
go build -o github-mcp-server ./cmd/github-mcp-server
```

### Running Multi-User Server
```bash
./github-mcp-server multi-user --toolsets=repos,issues,users,pull_requests
```

Available flags:
- `--toolsets`: Comma-separated list of toolsets to enable (default: all)
- `--read-only`: Restrict to read-only operations
- `--dynamic-toolsets`: Enable dynamic toolset discovery
- `--gh-host`: GitHub hostname (for GitHub Enterprise)

### Tool Usage

All tools now require an `auth_token` parameter containing a valid GitHub Personal Access Token.

#### Example: Get User Information
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "get_me",
    "arguments": {
      "auth_token": "ghp_your_github_token_here",
      "reason": "Getting user profile"
    }
  }
}
```

#### Example: List Repository Contents
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "get_file_contents",
    "arguments": {
      "auth_token": "ghp_your_github_token_here",
      "owner": "octocat",
      "repo": "Hello-World",
      "path": "README.md"
    }
  }
}
```

#### Example: Create an Issue
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "create_issue",
    "arguments": {
      "auth_token": "ghp_your_github_token_here",
      "owner": "octocat",
      "repo": "Hello-World",
      "title": "Bug report",
      "body": "Found a bug in the application"
    }
  }
}
```

## Testing

### Quick Test
```bash
# Start the server
./github-mcp-server multi-user --toolsets=repos

# In another terminal, test with a real GitHub token
echo '{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2025-03-26",
    "clientInfo": {"name": "test-client", "version": "1.0.0"},
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
      "auth_token": "your_real_github_token_here"
    }
  }
}' | ./github-mcp-server multi-user --toolsets=repos
```

### Test Script
Run the included test script:
```bash
chmod +x test_multi_user.sh
./test_multi_user.sh
```

## Implementation Details

### Code Changes

1. **New Server Configuration** (`internal/ghmcp/server.go`):
   - `MultiUserMCPServerConfig`: Configuration without global token
   - `NewMultiUserMCPServer()`: Creates server with per-request authentication
   - `RunMultiUserStdioServer()`: Runs multi-user server via stdio

2. **Multi-User Tools** (`pkg/github/tools.go`):
   - `InitMultiUserToolsets()`: Creates toolsets with auth token support
   - `createMultiUserTool()`: Wraps tools to add auth_token parameter
   - `wrapToolHandlerWithAuth()`: Extracts auth tokens from requests
   - `extractAuthTokenFromRequest()`: Helper for token extraction

3. **Command Line Interface** (`cmd/github-mcp-server/main.go`):
   - New `multi-user` subcommand
   - Uses `RunMultiUserStdioServer()` instead of `RunStdioServer()`

### Authentication Flow

1. Client sends tool request with `auth_token` parameter
2. `wrapToolHandlerWithAuth()` extracts token from request
3. Token is injected into request context
4. Tool handler retrieves token from context
5. New GitHub client created with the token for this request
6. API call made with user-specific authentication

### Security Considerations

- Each request uses its own authentication token
- No shared state between different users' requests
- Tokens are not logged or persisted
- Failed authentication returns proper error responses

## Compatibility

- **Backward Compatible**: Original single-user mode still available via `stdio` command
- **API Compatible**: All existing tools work the same way, just with additional `auth_token` parameter
- **MCP Protocol**: Fully compliant with MCP protocol specifications

## Benefits

1. **Resource Efficiency**: Single server instance handles multiple users
2. **Simplified Deployment**: No need for per-user Docker containers
3. **Better Scalability**: Reduced memory and CPU overhead
4. **Easier Management**: Single process to monitor and maintain
5. **Security**: Per-request authentication prevents token sharing

## Migration from Single-User

To migrate from the original single-user setup:

1. Replace `./github-mcp-server stdio` with `./github-mcp-server multi-user`
2. Remove `GITHUB_PERSONAL_ACCESS_TOKEN` environment variable
3. Update client code to include `auth_token` parameter in all tool requests
4. Test with your existing GitHub tokens

## Troubleshooting

### Common Issues

1. **Missing auth_token**: All tools require the `auth_token` parameter
   ```json
   {"error": "authentication error: missing required parameter: auth_token"}
   ```

2. **Invalid token**: GitHub returns 401 for invalid tokens
   ```json
   {"error": "failed to get user: GET https://api.github.com/user: 401 Bad credentials"}
   ```

3. **Insufficient permissions**: Token lacks required scopes
   ```json
   {"error": "403 Forbidden"}
   ```

### Debug Mode
Enable command logging to see all requests:
```bash
./github-mcp-server multi-user --enable-command-logging --log-file=debug.log
```

## Contributing

This modification maintains the original codebase structure while adding multi-user support. When contributing:

1. Ensure both single-user and multi-user modes continue to work
2. Add tests for new multi-user functionality
3. Update documentation for any new features
4. Follow the existing code style and patterns 