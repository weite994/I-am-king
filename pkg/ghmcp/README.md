# GitHub MCP Server Go Library

This package provides a Go library interface to the GitHub MCP Server, allowing you to embed the server functionality in your own Go applications.

## Installation

```bash
go get github.com/github/github-mcp-server/pkg/ghmcp
```

## Usage

### Running a Stdio Server with Static Token

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

### Running a Stdio Server with Dynamic Token Provider

For applications that need to refresh tokens without restarting the server:

```go
package main

import (
    "log"
    "sync"
    "time"
    
    "github.com/github/github-mcp-server/pkg/ghmcp"
)

// TokenManager manages dynamic token updates
type TokenManager struct {
    mu           sync.RWMutex
    currentToken string
}

func (tm *TokenManager) GetToken() string {
    tm.mu.RLock()
    defer tm.mu.RUnlock()
    return tm.currentToken
}

func (tm *TokenManager) UpdateToken(newToken string) {
    tm.mu.Lock()
    defer tm.mu.Unlock()
    tm.currentToken = newToken
}

func main() {
    tokenManager := &TokenManager{
        currentToken: getInitialToken(), // Your initial token
    }
    
    // The token provider will be called on each API request
    tokenProvider := func() string {
        return tokenManager.GetToken()
    }
    
    config := ghmcp.StdioServerConfig{
        Version:         "1.0.0",
        Host:            "https://github.com",
        TokenProvider:   tokenProvider, // Use TokenProvider instead of Token
        EnabledToolsets: []string{"repos", "issues", "pulls"},
        ReadOnly:        false,
    }
    
    // Start a goroutine to refresh the token periodically
    go func() {
        for {
            time.Sleep(30 * time.Minute)
            newToken := refreshTokenFromAuthService() // Your token refresh logic
            tokenManager.UpdateToken(newToken)
        }
    }()
    
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
- `Token`: GitHub personal access token (static token, use TokenProvider for dynamic tokens)
- `TokenProvider`: Function that returns the current GitHub token (takes precedence over Token)
- `EnabledToolsets`: List of toolsets to enable (e.g., "repos", "issues", "pulls", "users", "search")
- `DynamicToolsets`: Enable dynamic toolset discovery
- `ReadOnly`: Restrict to read-only operations
- `ExportTranslations`: Export translations to a JSON file
- `EnableCommandLogging`: Log all command requests and responses
- `LogFilePath`: Path to log file (defaults to stderr)

### MCPServerConfig

- `Version`: Version of your server
- `Host`: GitHub API host
- `Token`: GitHub personal access token (static token, use TokenProvider for dynamic tokens)
- `TokenProvider`: Function that returns the current GitHub token (takes precedence over Token)
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

## Token Provider Best Practices

When using a `TokenProvider`:

1. **Thread Safety**: Ensure your token provider is thread-safe as it will be called concurrently from multiple goroutines.
2. **Performance**: The token provider is called on each API request, so it should be fast. Consider caching the token.
3. **Error Handling**: The token provider should always return a valid token. Handle errors internally and fall back to a cached token if necessary.
4. **Logging**: Be careful not to log the full token. Log only the last few characters for debugging.
5. **Graceful Updates**: When updating tokens, ensure there's no downtime. The old token should remain valid until the new one is ready.

Example of a production-ready token provider:

```go
type TokenCache struct {
    mu           sync.RWMutex
    token        string
    expiry       time.Time
    refreshFunc  func() (string, time.Time, error)
}

func (tc *TokenCache) GetToken() string {
    tc.mu.RLock()
    if time.Now().Before(tc.expiry) {
        defer tc.mu.RUnlock()
        return tc.token
    }
    tc.mu.RUnlock()
    
    // Token expired, refresh it
    tc.mu.Lock()
    defer tc.mu.Unlock()
    
    // Double-check after acquiring write lock
    if time.Now().Before(tc.expiry) {
        return tc.token
    }
    
    newToken, newExpiry, err := tc.refreshFunc()
    if err != nil {
        // Log error and return cached token
        log.Printf("Failed to refresh token: %v", err)
        return tc.token
    }
    
    tc.token = newToken
    tc.expiry = newExpiry
    return tc.token
}
```

## Requirements

- Go 1.21 or later
- Valid GitHub personal access token with appropriate permissions 