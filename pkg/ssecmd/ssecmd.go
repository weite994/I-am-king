// Package ssecmd provides functionality for creating and running an SSE server
// without any dependencies on specific CLI or configuration systems.
package ssecmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/github/github-mcp-server/internal/ghmcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
)

// Config holds all configuration options for the SSE server
type Config struct {
	Token           string
	Host            string
	Address         string
	BasePath        string
	LogFilePath     string
	EnabledToolsets []string
	DynamicToolsets bool
	ReadOnly        bool
	Version         string
}

// DefaultConfig creates a basic Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		Address:  "localhost:8080",
		BasePath: "",
		ReadOnly: false,
	}
}

// Server represents an SSE server that can be started and stopped
type Server struct {
	config    Config
	mcpServer interface{} // Using interface{} since we don't need to access it directly
	sseServer *server.SSEServer
}

// NewServer creates a new SSE server with the provided configuration
func NewServer(config Config) (*Server, error) {
	if config.Token == "" {
		return nil, errors.New("GitHub personal access token not set")
	}

	// Configure logging
	if config.LogFilePath != "" {
		file, err := os.OpenFile(config.LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		logger := logrus.New()
		logger.SetOutput(file)
		logger.SetLevel(logrus.DebugLevel)
		logrus.SetOutput(file) // Set global logger output as well
	}

	// Create a translation helper function
	translator := func(key, defaultValue string) string {
		return defaultValue // Simple implementation just returning the default value
	}

	// Create the MCP server instance with GitHub tools
	mcpServer, err := ghmcp.NewMCPServer(ghmcp.MCPServerConfig{
		Version:         config.Version,
		Host:            config.Host,
		Token:           config.Token,
		EnabledToolsets: config.EnabledToolsets,
		DynamicToolsets: config.DynamicToolsets,
		ReadOnly:        config.ReadOnly,
		Translator:      translator,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP server: %w", err)
	}

	// Create SSE server using built-in functionality from mark3labs/mcp-go
	sseServer := server.NewSSEServer(mcpServer,
		server.WithStaticBasePath(config.BasePath),
	)

	return &Server{
		config:   config,
		mcpServer: mcpServer,
		sseServer: sseServer,
	}, nil
}

// Start starts the SSE server
func (s *Server) Start() error {
	// Print server info
	fmt.Fprintf(os.Stderr, "GitHub MCP Server running in SSE mode on %s with base path %s\n",
		s.config.Address, s.config.BasePath)

	// Start the server
	return s.sseServer.Start(s.config.Address)
}

// RunSSEServer is a convenience function that creates and starts an SSE server in one call
// This is provided for backward compatibility and simple use cases
func RunSSEServer(config Config) error {
	server, err := NewServer(config)
	if err != nil {
		return err
	}
	return server.Start()
}

// ServerOption represents an option for configuring an SSE server
type ServerOption func(*Config)

// WithAddress sets the address for the SSE server
func WithAddress(address string) ServerOption {
	return func(c *Config) {
		c.Address = address
	}
}

// WithBasePath sets the base path for SSE server URLs
func WithBasePath(basePath string) ServerOption {
	return func(c *Config) {
		c.BasePath = basePath
	}
}

// WithLogFilePath sets the log file path for the SSE server
func WithLogFilePath(logFilePath string) ServerOption {
	return func(c *Config) {
		c.LogFilePath = logFilePath
	}
}

// WithReadOnly sets the read-only mode for the SSE server
func WithReadOnly(readOnly bool) ServerOption {
	return func(c *Config) {
		c.ReadOnly = readOnly
	}
}

// WithDynamicToolsets sets whether to use dynamic toolsets for the SSE server
func WithDynamicToolsets(dynamicToolsets bool) ServerOption {
	return func(c *Config) {
		c.DynamicToolsets = dynamicToolsets
	}
}

// WithEnabledToolsets sets the enabled toolsets for the SSE server
func WithEnabledToolsets(enabledToolsets []string) ServerOption {
	return func(c *Config) {
		c.EnabledToolsets = enabledToolsets
	}
}

// WithHost sets the GitHub host for the SSE server
func WithHost(host string) ServerOption {
	return func(c *Config) {
		c.Host = host
	}
}

// WithToken sets the GitHub token for the SSE server
func WithToken(token string) ServerOption {
	return func(c *Config) {
		c.Token = token
	}
}

// WithVersion sets the version for the SSE server
func WithVersion(version string) ServerOption {
	return func(c *Config) {
		c.Version = version
	}
}

// CreateServerWithOptions creates a new SSE server with the provided options
func CreateServerWithOptions(options ...ServerOption) (*Server, error) {
	config := DefaultConfig()
	for _, option := range options {
		option(&config)
	}
	return NewServer(config)
}
