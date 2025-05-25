// Package ghmcp provides a public API wrapper for the GitHub MCP Server functionality.
// This package exposes the necessary types and functions from the internal implementation
// for use by external Go modules.
//
// Usage example:
//
//	config := ghmcp.StdioServerConfig{
//	    Version:         "1.0.0",
//	    Host:            "https://github.com",
//	    Token:           os.Getenv("GITHUB_TOKEN"),
//	    EnabledToolsets: []string{"repos", "issues"},
//	    ReadOnly:        false,
//	}
//
//	if err := ghmcp.RunStdioServer(config); err != nil {
//	    log.Fatal(err)
//	}
package ghmcp

import (
	"github.com/github/github-mcp-server/internal/ghmcp"
	"github.com/mark3labs/mcp-go/server"
)

// StdioServerConfig contains configuration for running the GitHub MCP Server
// in stdio mode. This is a re-export of the internal type.
type StdioServerConfig = ghmcp.StdioServerConfig

// MCPServerConfig contains configuration for creating a new MCP Server instance.
// This is a re-export of the internal type.
type MCPServerConfig = ghmcp.MCPServerConfig

// RunStdioServer runs the GitHub MCP Server using stdio for communication.
// This function wraps the internal implementation and is not concurrent safe.
func RunStdioServer(cfg StdioServerConfig) error {
	return ghmcp.RunStdioServer(cfg)
}

// NewMCPServer creates a new MCP Server instance with the given configuration.
// This function wraps the internal implementation.
func NewMCPServer(cfg MCPServerConfig) (*server.MCPServer, error) {
	return ghmcp.NewMCPServer(cfg)
}
