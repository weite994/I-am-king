package ghmcp_test

import (
	"fmt"
	"log"
	"sync"

	"github.com/github/github-mcp-server/pkg/ghmcp"
	"github.com/github/github-mcp-server/pkg/translations"
)

func ExampleRunStdioServer() {
	// Example of how to use RunStdioServer from an external module
	config := ghmcp.StdioServerConfig{
		Version:         "1.0.0",
		Host:            "https://github.com",
		Token:           "your-github-token",
		EnabledToolsets: []string{"repos", "issues"},
		ReadOnly:        true,
	}

	// This would normally block and run the server
	// err := ghmcp.RunStdioServer(config)
	// if err != nil {
	//     log.Fatal(err)
	// }

	// Just to use the config variable in the example
	_ = config
	fmt.Println("Server configured")
	// Output: Server configured
}

func ExampleRunStdioServer_tokenProvider() {
	// Example showing how to use a TokenProvider for dynamic token refresh

	// This simulates a token management system
	tokenManager := &TokenManager{
		currentToken: "initial-token",
	}

	// Create a token provider function
	tokenProvider := func() string {
		return tokenManager.GetCurrentToken()
	}

	config := ghmcp.StdioServerConfig{
		Version:         "1.0.0",
		Host:            "https://github.com",
		TokenProvider:   tokenProvider, // Use TokenProvider instead of Token
		EnabledToolsets: []string{"repos", "issues"},
		ReadOnly:        false,
	}

	// In your application, you can update the token at any time:
	// tokenManager.UpdateToken("new-refreshed-token")

	// This would normally block and run the server
	// err := ghmcp.RunStdioServer(config)
	// if err != nil {
	//     log.Fatal(err)
	// }

	// Just to use the config variable in the example
	_ = config
	fmt.Println("Server configured with token provider")
	// Output: Server configured with token provider
}

func ExampleNewMCPServer() {
	// Example of how to use NewMCPServer from an external module
	config := ghmcp.MCPServerConfig{
		Version:         "1.0.0",
		Host:            "https://github.com",
		Token:           "your-github-token",
		EnabledToolsets: []string{"repos", "issues"},
		ReadOnly:        true,
		Translator:      translations.NullTranslationHelper,
	}

	_, err := ghmcp.NewMCPServer(config)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("MCP Server created")
	// Output: MCP Server created
}

// TokenManager is an example of a token management system
type TokenManager struct {
	mu           sync.RWMutex
	currentToken string
}

func (tm *TokenManager) GetCurrentToken() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.currentToken
}

func (tm *TokenManager) UpdateToken(newToken string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.currentToken = newToken
}
