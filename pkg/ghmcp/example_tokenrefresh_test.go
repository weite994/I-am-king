//go:build example
// +build example

package ghmcp_test

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/github/github-mcp-server/pkg/ghmcp"
)

// This example demonstrates how to implement a wrapper around ghmcp.RunStdioServer
// that supports dynamic token refresh without restarting the server.
func Example_tokenRefreshWrapper() {
	// Create a token manager
	tokenManager := NewTokenManager()

	// Initialize with the first token
	initialToken := os.Getenv("GITHUB_TOKEN")
	if initialToken == "" {
		log.Fatal("GITHUB_TOKEN environment variable is required")
	}
	tokenManager.UpdateToken(initialToken)

	// Create the server configuration with a token provider
	config := ghmcp.StdioServerConfig{
		Version: "1.0.0",
		Host:    "https://github.com",
		TokenProvider: func() string {
			token := tokenManager.GetCurrentToken()
			log.Printf("Token provider called, returning token ending with: ...%s",
				token[len(token)-4:])
			return token
		},
		EnabledToolsets:      []string{"repos", "issues", "pulls"},
		ReadOnly:             false,
		EnableCommandLogging: true,
		LogFilePath:          "github-mcp-server.log",
	}

	// Start a goroutine that refreshes the token periodically
	stopRefresh := make(chan struct{})
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// In a real application, this would call your auth service
				newToken := refreshTokenFromAuthService()
				if newToken != "" {
					log.Println("Refreshing GitHub token...")
					tokenManager.UpdateToken(newToken)
				}
			case <-stopRefresh:
				return
			}
		}
	}()

	// Run the server (this blocks)
	log.Println("Starting GitHub MCP Server with dynamic token support...")
	if err := ghmcp.RunStdioServer(config); err != nil {
		close(stopRefresh)
		log.Fatal(err)
	}
}

// TokenManager provides thread-safe token management
type TokenManager struct {
	mu           sync.RWMutex
	currentToken string
	lastUpdated  time.Time
}

func NewTokenManager() *TokenManager {
	return &TokenManager{}
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
	tm.lastUpdated = time.Now()
	log.Printf("Token updated at %v", tm.lastUpdated)
}

func (tm *TokenManager) GetLastUpdated() time.Time {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.lastUpdated
}

// refreshTokenFromAuthService simulates fetching a new token from an auth service
func refreshTokenFromAuthService() string {
	// In a real implementation, this would:
	// 1. Call your authentication service
	// 2. Exchange refresh tokens
	// 3. Return the new access token

	// For this example, we'll just return the current token from env
	// In production, you'd implement actual token refresh logic here
	return os.Getenv("GITHUB_TOKEN_REFRESHED")
}

// Example of a more advanced token provider with caching and validation
type AdvancedTokenProvider struct {
	mu             sync.RWMutex
	currentToken   string
	tokenExpiry    time.Time
	refreshToken   string
	authServiceURL string
	minRefreshTime time.Duration
}

func NewAdvancedTokenProvider(authServiceURL, refreshToken string) *AdvancedTokenProvider {
	return &AdvancedTokenProvider{
		authServiceURL: authServiceURL,
		refreshToken:   refreshToken,
		minRefreshTime: 5 * time.Minute, // Don't refresh more often than every 5 minutes
	}
}

func (atp *AdvancedTokenProvider) GetToken() string {
	atp.mu.RLock()

	// Check if token needs refresh
	if time.Now().After(atp.tokenExpiry.Add(-atp.minRefreshTime)) {
		atp.mu.RUnlock()
		return atp.refreshTokenIfNeeded()
	}

	token := atp.currentToken
	atp.mu.RUnlock()
	return token
}

func (atp *AdvancedTokenProvider) refreshTokenIfNeeded() string {
	atp.mu.Lock()
	defer atp.mu.Unlock()

	// Double-check after acquiring write lock
	if time.Now().Before(atp.tokenExpiry.Add(-atp.minRefreshTime)) {
		return atp.currentToken
	}

	// In a real implementation, you would call your auth service here
	// newToken, newExpiry := callAuthService(atp.authServiceURL, atp.refreshToken)
	// atp.currentToken = newToken
	// atp.tokenExpiry = newExpiry

	log.Println("Token refreshed via advanced provider")
	return atp.currentToken
}
