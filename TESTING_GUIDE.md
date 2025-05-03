# GitHub MCP Server Binary Testing Guide

> **IMPORTANT NOTE**: This testing guide specifically covers the GitHub MCP Server Binary implementation. Other implementations might have different behaviors, response formats, or tool names.

This document explains how to test the GitHub MCP Server Binary implementation using different approaches with the client library in the [github-mcp-client](../github-mcp-client) repository.

## Testing Approaches

There are several ways to test the GitHub MCP Server:

1. **Binary Transport Testing** - Using the GitHub MCP Server binary directly
2. **Docker Transport Testing** - Using the Docker container
3. **HTTP SSE Transport Testing** - Testing remote communication

## Prerequisites

Before testing, you need:

1. **GitHub Personal Access Token** - Create a token with the necessary permissions
2. **GitHub MCP Server** - Either built from source or using the Docker image
3. **Python 3.6+** - For running the client library
4. **Go 1.19+** - If building the server from source

## Setting Up Your Token

Set up your GitHub token in one of these ways:

```bash
# Set as environment variable
export GITHUB_PERSONAL_ACCESS_TOKEN=your_token_here

# OR save to a file (recommended for security)
echo "your_token_here" > ~/.github_token
chmod 600 ~/.github_token
```

## Binary Transport Testing

### Building the GitHub MCP Server Binary

Follow the official build instructions from the [GitHub MCP Server repository](https://github.com/github/github-mcp-server/blob/main/README.md#build-from-source):

```bash
# Clone the repository
git clone https://github.com/github/github-mcp-server.git
cd github-mcp-server

# Build the binary
go build -o github-mcp-server ./cmd/github-mcp-server

# Verify the binary works
./github-mcp-server --help
```

The GitHub MCP Server Binary supports multiple transport protocols:

1. **stdio** - Standard input/output pipes (used in most of our tests)
2. **HTTP SSE** - HTTP Server-Sent Events for remote communication

This testing guide covers both transport methods, with a focus on the stdio transport which is most commonly used for local development and testing.

### Running Tests with Binary Transport

Once you have the binary, you can run the tests:

```bash
# Make sure the binary is executable
chmod +x github-mcp-server

# Run the JSON-RPC test script
python ../github-mcp-client/test_jsonrpc.py --binary-path ./github-mcp-server

# Run the binary client test
python ../github-mcp-client/test_binary_client.py --binary-path ./github-mcp-server

# Run a complete PR workflow
python ../github-mcp-client/mcp_workflow.py --binary-path ./github-mcp-server --no-docker --owner your_username --repo your_repo
```

## Docker Transport Testing

### Running Tests with Docker Transport

```bash
cd ../github-mcp-client

# Make sure you have docker installed and running
docker --version

# Pull the GitHub MCP Server image
docker pull ghcr.io/github/github-mcp-server:latest

# Test the client with Docker transport
python stdio_client_test.py

# Run a complete PR workflow
python mcp_workflow.py --owner your_username --repo your_repo
```

## HTTP SSE Transport Testing

HTTP SSE allows you to connect to a remote GitHub MCP Server over HTTP.

### Running the HTTP SSE Server

```bash
cd ../github-mcp-client

# Start the HTTP SSE server
./run_http_mcp_server.sh

# In another terminal, run the HTTP SSE client
python docs/examples/http_sse_example.py --server-url http://localhost:7444
```

## Understanding the JSON-RPC Communication Format

The GitHub MCP Server uses JSON-RPC 2.0 for communication. Here's the general format for requests:

```json
{
  "jsonrpc": "2.0",
  "id": "request-id",
  "method": "tools/call",
  "params": {
    "name": "tool-name",
    "arguments": {
      "arg1": "value1",
      "arg2": "value2"
    }
  }
}
```

### Response Format Complexity

The server response format is more complex than documented and requires careful parsing:

```json
{
  "jsonrpc": "2.0",
  "id": "request-id",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"actual_json_data_as_string\"}"
      }
    ]
  }
}
```

**Important Notes on Response Parsing:**
- The actual data is often nested as a JSON string within `content[0].text`
- This string must be parsed separately to access the actual data
- Different tools may return different response structures
- Some tools return direct results, while others use the nested content format
- A robust parser must handle all these variations

Example response parser:

```python
def parse_response(response):
    """Parse a response from the GitHub MCP Server."""
    if "result" in response:
        result = response["result"]
        # Check if result contains content field (the new format)
        if "content" in result and isinstance(result["content"], list):
            for item in result["content"]:
                if item.get("type") == "text":
                    text = item.get("text", "")
                    # Try to parse the text as JSON
                    try:
                        return json.loads(text)
                    except:
                        # If it's not valid JSON, return the text as is
                        return text
        # If no content field or parsing failed, return the result as is
        return result
    
    # Return empty dict if no result found
    return {}
```

## Key Test Files

- `test_jsonrpc.py` - Low-level JSON-RPC communication test
- `test_binary_client.py` - Test using the binary directly
- `mcp_workflow.py` - Complete PR workflow demo with either binary or Docker transport
- `docs/examples/http_sse_example.py` - Example of using HTTP SSE transport
- `list_tools.py` - Discovers all available tools in the server
- `pr_workflow_fixed2.py` - Robust implementation of PR workflow with proper response parsing

## Available Tools

The GitHub MCP Server provides various tools organized into toolsets. Use the `list_tools.py` script to discover all available tools:

```bash
python list_tools.py
```

### Context Tools

- `get_me` - Get information about the authenticated user

### Repository Tools

- `search_repositories` - Search for repositories (**NOTE:** Use this instead of `get_repo`)
- `list_branches` - List branches in a repository
- `create_branch` - Create a new branch
- `get_file_contents` - Get file contents from a repository
- `create_or_update_file` - Create or update a file in a repository

### Pull Request Tools

- `create_pull_request` - Create a new pull request
- `merge_pull_request` - Merge a pull request

## Response Type Variations

Different tools return different response types:

1. **List Responses**: Some tools (e.g., `list_branches`) may return direct arrays
2. **Dictionary with Items**: Some tools (e.g., `search_repositories`) return dictionaries with an `items` array
3. **Single Object**: Some tools (e.g., `get_me`) return a single object
4. **Nested Content**: Many responses embed the result in the `content[0].text` field as a string

Example of handling different response types:

```python
# Handle different response formats for branches
branches = []
if isinstance(branches_response, list):
    branches = branches_response
elif isinstance(branches_response, dict) and "items" in branches_response:
    branches = branches_response.get("items", [])
```

## Troubleshooting

### Common Issues

1. **Authentication Errors**
   - Verify that your token has the necessary permissions
   - Make sure your token is not expired
   - Check that your token is being passed correctly

2. **Binary Executable Permission Issues**
   - Make sure the binary has execute permissions: `chmod +x github-mcp-server`

3. **Docker Issues**
   - Ensure Docker is installed and running
   - Make sure you have pulled the GitHub MCP Server image
   - Check for port conflicts if using HTTP SSE

4. **JSON-RPC Errors**
   - Verify that your request follows the correct JSON-RPC 2.0 format
   - Check that you're using the correct tool name and parameters

5. **Response Parsing Issues**
   - Check for the nested content structure in responses
   - Handle different response formats for different tools
   - Use type checking to safely handle different response structures

### Debugging Tools

- Run the server with verbose logging: `./github-mcp-server --verbose stdio`
- Examine stderr output for error messages
- Use the `mcpcurl` tool to send test requests: `./github-mcp-server/cmd/mcpcurl/mcpcurl --stdio-server-cmd "./github-mcp-server stdio" tools get_me`
- Use the `list_tools.py` script to discover available tools and their correct names

## Advanced Testing

### Testing Different Toolsets

You can test specific toolsets by setting the `GITHUB_TOOLSETS` environment variable:

```bash
export GITHUB_TOOLSETS="repos,issues,pull_requests"
./github-mcp-server stdio
```

### Testing with GitHub Enterprise

You can test with GitHub Enterprise by setting the `GITHUB_HOST` environment variable:

```bash
export GITHUB_HOST="github.mycompany.com"
./github-mcp-server stdio
```

## Complete Pull Request Workflow Example

For a complete example of creating a pull request, see the `pr_workflow_fixed2.py` script:

```bash
python pr_workflow_fixed2.py --owner your_username --repo your_repo
```

This script demonstrates:
- Secure token handling
- Robust response parsing
- Type-safe operations
- Error handling
- Process management
- Detailed logging

## Best Practices for MCP Server Integration

Based on extensive testing, we recommend these best practices:

1. **Tool Discovery**: Always use `tools/list` to discover available tools
2. **Response Parsing**: Implement robust response parsing logic
3. **Error Handling**: Add specific error handling for each tool
4. **Type Safety**: Use type checking and defensive programming
5. **Authentication**: Implement secure token handling
6. **Process Management**: Properly manage server processes
7. **Logging**: Add detailed logging for debugging
8. **Timeouts**: Add timeouts for all operations
9. **Clean Up**: Always clean up resources when done

For more detailed findings, see the [GITHUB_MCP_FINDINGS.md](./GITHUB_MCP_FINDINGS.md) document.