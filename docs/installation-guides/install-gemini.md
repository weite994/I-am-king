# Install GitHub MCP Server in Gemini CLI

This guide covers installation of the GitHub MCP server for the Gemini CLI.

## Gemini CLI

The Gemini CLI provides command-line access to Gemini with MCP server integration.

### Prerequisites

1. Gemini CLI installed
2. [GitHub Personal Access Token](https://github.com/settings/personal-access-tokens/new) with `repo`, `read:org`, and `gist` scopes.

### Installation

The Gemini CLI can be configured to connect to the remote GitHub MCP server.

#### Configuration

1.  **Locate your settings file**: The Gemini CLI settings are typically located at `~/.config/google/gemini/settings.json`.
2.  **Edit the settings file**: Add the following `mcp_server` configuration to your `settings.json` file.

```json
{
  "mcp_server": {
    "github": {
      "display_name": "GitHub",
      "url": "https://api.githubcopilot.com/mcp/",
      "auth": {
        "github_pat": {
          "token": "YOUR_GITHUB_PAT"
        }
      }
    }
  }
}
```

Replace `YOUR_GITHUB_PAT` with your GitHub Personal Access Token.

#### Using Environment Variables

For better security, you can reference an environment variable instead of hardcoding the token:

```json
{
  "mcp_server": {
    "github": {
      "display_name": "GitHub",
      "url": "https://api.githubcopilot.com/mcp/",
      "auth": {
        "github_pat": {
          "token": "$GITHUB_PAT"
        }
      }
    }
  }
}
```

Then, set the `GITHUB_PAT` environment variable in your shell's configuration file (e.g., `.bashrc`, `.zshrc`):

```bash
export GITHUB_PAT="your_github_pat"
```

### Verification

Run the following command to verify the installation:

```bash
gemini status
```

You should see the GitHub MCP server listed as a connected tool provider.
