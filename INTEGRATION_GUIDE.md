# GitHub MCP Server Integration Guide

This guide covers setting up the GitHub MCP Server securely and integrating it with various MCP hosts.

## 🔐 Security First Setup

### 1. Initial Setup

```bash
# Clone or navigate to the server directory
cd /Users/shunsuke/Dev/organized/mcp-servers/github-mcp-server

# Run setup script
./setup.sh
```

### 2. Create GitHub Personal Access Token

1. Go to [GitHub PAT Settings](https://github.com/settings/personal-access-tokens/new)
2. Create a new token with these **minimum scopes**:
   - `repo` - Repository operations
   - `read:packages` - For Docker image access (if using Docker)
   - `read:org` - For organization operations (optional)
   - `read:user` - For user profile access

### 3. Configure Environment

Edit `.env` file:
```bash
# Required
GITHUB_PERSONAL_ACCESS_TOKEN=github_pat_11ABC...

# Optional configurations
GITHUB_TOOLSETS=context,repos,issues,pull_requests,actions
GITHUB_READ_ONLY=0
GITHUB_DYNAMIC_TOOLSETS=0
```

### 4. Test Connection

```bash
./test-connection.sh
```

## 🔧 MCP Host Integration

### Claude Code CLI

Add to your `~/.claude_config/config.json`:

```json
{
  "mcpServers": {
    "github": {
      "command": "/Users/shunsuke/Dev/organized/mcp-servers/github-mcp-server/github-mcp-server",
      "args": ["stdio"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_PAT}"
      }
    }
  }
}
```

Set environment variable:
```bash
export GITHUB_PAT="your_token_here"
```

### VS Code with Agent Mode

Create `.vscode/mcp.json` in your workspace:

```json
{
  "inputs": [
    {
      "type": "promptString",
      "id": "github_token",
      "description": "GitHub Personal Access Token",
      "password": true
    }
  ],
  "servers": {
    "github": {
      "command": "/Users/shunsuke/Dev/organized/mcp-servers/github-mcp-server/github-mcp-server",
      "args": ["stdio"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "${input:github_token}"
      }
    }
  }
}
```

### Docker Configuration

```json
{
  "servers": {
    "github": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-e", "GITHUB_PERSONAL_ACCESS_TOKEN",
        "ghcr.io/github/github-mcp-server"
      ],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_PAT}"
      }
    }
  }
}
```

## 🛠 Available Toolsets

| Toolset | Description | Recommended |
|---------|-------------|-------------|
| `context` | User and GitHub context | ✅ Always enable |
| `repos` | Repository operations | ✅ Core functionality |
| `issues` | Issue management | ✅ Common use case |
| `pull_requests` | PR operations | ✅ Code review workflows |
| `actions` | GitHub Actions/CI | 📋 For DevOps tasks |
| `code_security` | Security scanning | 🔒 For security analysis |
| `dependabot` | Dependency alerts | 📦 For dependency management |
| `discussions` | GitHub Discussions | 💬 For community features |
| `notifications` | Notification management | 📢 For workflow automation |
| `secret_protection` | Secret scanning | 🔐 For security monitoring |

### Common Configurations

**Developer Setup** (recommended):
```bash
GITHUB_TOOLSETS=context,repos,issues,pull_requests,actions
```

**Security Auditor**:
```bash
GITHUB_TOOLSETS=context,repos,code_security,dependabot,secret_protection
GITHUB_READ_ONLY=1
```

**Project Manager**:
```bash
GITHUB_TOOLSETS=context,issues,pull_requests,discussions,notifications
```

## 🧪 Testing Your Setup

### Basic Test
```bash
# Test server startup
./github-mcp-server stdio --toolsets context

# Test with limited toolsets
./github-mcp-server stdio --toolsets context,repos --read-only
```

### Test GitHub Connection
```bash
# Test API access
./test-connection.sh

# Test specific functionality
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "get_me"}}' | ./github-mcp-server stdio --toolsets context
```

## 🔒 Security Best Practices

### Token Management
- **Rotate tokens regularly** (monthly recommended)
- **Use fine-grained PATs** when possible
- **Limit scopes** to minimum required
- **Never commit tokens** to version control
- **Use environment variables** for configuration

### File Permissions
```bash
# Secure your configuration files
chmod 600 .env
chmod 600 claude-config.json
```

### Network Security
- Use **read-only mode** when possible
- Enable only **required toolsets**
- Monitor token usage in GitHub settings
- Use **organization PAT policies** in enterprise environments

## 🔧 Troubleshooting

### Common Issues

**Authentication Errors (401)**:
- Check token validity: `./test-connection.sh`
- Verify token hasn't expired
- Ensure token has required scopes

**Permission Errors (403)**:
- Check repository access permissions
- Verify organization policies
- Ensure PAT has necessary scopes

**Connection Errors**:
- Check internet connectivity
- Verify GitHub API status
- Test with curl: `curl -H "Authorization: Bearer $TOKEN" https://api.github.com/user`

**MCP Server Errors**:
- Check server logs: `--log-file /tmp/github-mcp.log`
- Enable debug logging: `--enable-command-logging`
- Test stdio mode: `./github-mcp-server stdio --help`

### Debug Mode

Enable verbose logging:
```bash
./github-mcp-server stdio \
  --log-file /tmp/github-mcp.log \
  --enable-command-logging \
  --toolsets context,repos
```

## 📚 Advanced Configuration

### Custom GitHub Enterprise

```bash
# .env configuration
GITHUB_HOST=https://github.your-company.com
GITHUB_PERSONAL_ACCESS_TOKEN=your_enterprise_token
```

### Dynamic Toolsets

Enable on-demand tool discovery:
```bash
GITHUB_DYNAMIC_TOOLSETS=1
```

### Tool Description Overrides

Create `github-mcp-server-config.json`:
```json
{
  "TOOL_GET_ME_DESCRIPTION": "Get current user profile information",
  "TOOL_CREATE_ISSUE_DESCRIPTION": "Create a new issue in a repository"
}
```

## 🤝 Integration with Shared Utilities

For future migrations to our shared MCP utilities:

1. **Configuration Management**: Use our `ConfigLoader` pattern
2. **Logging**: Integrate `structlog` for JSON logging
3. **HTTP Client**: Replace with our retry-enabled client
4. **Health Checks**: Add comprehensive monitoring
5. **Rate Limiting**: Implement tiered rate limiting

See `../shared/mcp-utils/` for reference implementations.

## 📖 Additional Resources

- [GitHub MCP Server README](./README.md)
- [Installation Guides](./docs/installation-guides/)
- [Security Documentation](./SECURITY.md)
- [Contributing Guide](./CONTRIBUTING.md)