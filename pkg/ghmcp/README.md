# GitHub MCP Server Go Library

This package provides a Go library interface to the GitHub MCP Server, allowing you to embed the server functionality in your own Go applications.

## Installation

```bash
go get github.com/github/github-mcp-server/pkg/ghmcp
```

## Usage

### Running a Stdio Server

The most common use case is running the MCP server using stdio for communication:

```go
package main

import (
    "log"
    "os"
    
    "github.com/github/github-mcp-server/pkg/ghmcp"
)

func main() {
    config := ghmcp.StdioServerConfig{
        Version:         "1.0.0",
        Host:            "https://github.com", // or your GitHub Enterprise URL
        Token:           os.Getenv("GITHUB_TOKEN"),
        EnabledToolsets: []string{"repos", "issues", "pulls"},
        ReadOnly:        false,
    }
    
    if err := ghmcp.RunStdioServer(config); err != nil {
        log.Fatal(err)
    }
}
```

### Creating a Custom MCP Server

For more advanced use cases, you can create an MCP server instance directly:

```go
package main

import (
    "log"
    
    "github.com/github/github-mcp-server/pkg/ghmcp"
    "github.com/github/github-mcp-server/pkg/translations"
)

func main() {
    config := ghmcp.MCPServerConfig{
        Version:         "1.0.0",
        Host:            "https://github.com",
        Token:           "your-github-token",
        EnabledToolsets: []string{"repos", "issues"},
        ReadOnly:        true,
        Translator:      translations.NullTranslationHelper,
    }
    
    server, err := ghmcp.NewMCPServer(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Use the server instance as needed
    _ = server
}
```

## Configuration Options

### StdioServerConfig

- `Version`: Version of your server
- `Host`: GitHub API host (e.g., "https://github.com" or "https://github.enterprise.com")
- `Token`: GitHub personal access token
- `EnabledToolsets`: List of toolsets to enable (e.g., "repos", "issues", "pulls", "users", "search")
- `DynamicToolsets`: Enable dynamic toolset discovery
- `ReadOnly`: Restrict to read-only operations
- `ExportTranslations`: Export translations to a JSON file
- `EnableCommandLogging`: Log all command requests and responses
- `LogFilePath`: Path to log file (defaults to stderr)

### MCPServerConfig

- `Version`: Version of your server
- `Host`: GitHub API host
- `Token`: GitHub personal access token
- `EnabledToolsets`: List of toolsets to enable
- `DynamicToolsets`: Enable dynamic toolset discovery
- `ReadOnly`: Restrict to read-only operations
- `Translator`: Translation helper function (use `translations.NullTranslationHelper` for default)

## Available Toolsets

- `repos`: Repository management tools
- `issues`: Issue management tools
- `pulls`: Pull request management tools
- `users`: User management tools
- `search`: Search functionality
- `all`: Enable all available toolsets

## Requirements

- Go 1.21 or later
- Valid GitHub personal access token with appropriate permissions 