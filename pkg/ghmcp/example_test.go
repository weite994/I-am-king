package ghmcp_test

import (
	"fmt"
	"log"

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
