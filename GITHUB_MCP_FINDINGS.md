# GitHub MCP Server Binary: Detailed Findings and Implementation Guide

> **IMPORTANT NOTE**: All findings in this document specifically relate to the GitHub MCP Server Binary implementation. Other implementations might have different behaviors, response formats, or tool names.
>
> The GitHub MCP Server Binary can be built from source following the [official instructions](https://github.com/github/github-mcp-server/blob/main/README.md#build-from-source).
>
> This document covers findings related to both **stdio** transport (standard input/output pipes) and **HTTP SSE** transport (HTTP Server-Sent Events), with a primary focus on the stdio transport which is most commonly used for local development and testing.

This document contains comprehensive findings and insights discovered during testing of the GitHub MCP Server Binary implementation, with particular focus on response format parsing, tool name discovery, and authentication methods.

## 1. Response Format Structure

One of the most significant discoveries during our testing was the unexpected and complex response format from the GitHub MCP Server Binary implementation.

### Expected vs. Actual Response Format

**Expected Format (Based on Documentation):**
```json
{
  "jsonrpc": "2.0",
  "id": "request-id",
  "result": {
    // Direct tool result data
  }
}
```

**Actual Format (Discovered During Testing):**
```json
{
  "jsonrpc": "2.0",
  "id": "request-id",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"full_json_response_as_string\"}"
      }
    ]
  }
}
```

### Key Insights:
- The actual result data is nested in a JSON string within `content[0].text`
- This string needs to be parsed as JSON to access the actual data
- Not all responses follow this format - some tools return direct JSON results
- A robust parser must handle both formats gracefully

### Implementation Solution

We developed a flexible response parser that handles the different formats:

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

## 2. Tool Names and Discoverability

Another significant discovery was the discrepancy between documented tool names and actual tool names recognized by the server.

### Tool Name Discrepancies

| Documentation | Actual Tool Name | Notes |
|---------------|-----------------|-------|
| `get_repo` | Not available | Use `search_repositories` instead |
| `list_branches` | Available | Works as documented |
| `create_branch` | Available | Works as documented |
| `get_file_contents` | Available | Works as documented |
| `create_or_update_file` | Available | Works as documented |
| `create_pull_request` | Available | Works as documented |

### Tool Discovery Method

We created a dedicated script `list_tools.py` to discover all available tools:

```python
# Create a request to list all available tools
request = {
    "jsonrpc": "2.0",
    "id": "1",
    "method": "tools/list",
    "params": {}
}
```

This approach revealed the complete list of available tools, their descriptions, and parameters.

### Best Practices for Tool Discovery

1. **Always use `tools/list` first**: Before integrating with the GitHub MCP Server, always query the available tools
2. **Document tool parameters**: Document the required and optional parameters for each tool
3. **Create helper functions**: Create wrapper functions for each tool to simplify their use
4. **Handle errors gracefully**: Add specific error handling for each tool

## 3. Handling Various Response Types

Different endpoints return responses in different formats, requiring flexible parsing logic.

### Response Type Variations

1. **List Responses**: Some tools (e.g., `list_branches`) return direct arrays
2. **Dictionary with Items**: Others (e.g., `search_repositories`) return dictionaries with an `items` array
3. **Single Object**: Some (e.g., `get_me`) return a single object
4. **Nested Content**: Many responses embed the result in the `content[0].text` field as a string

### Example of Type-Safe Handling

```python
# Handle different response formats for branches
branches = []
if isinstance(branches_response, list):
    branches = branches_response
elif isinstance(branches_response, dict) and "items" in branches_response:
    branches = branches_response.get("items", [])
```

### Response Type Best Practices

1. **Always use type checking**: Never assume a response is a particular type
2. **Use defensive programming**: Handle missing fields gracefully
3. **Provide fallbacks**: Always have default values if data extraction fails
4. **Log unexpected formats**: Log unusual response formats to assist debugging

## 4. Authentication Best Practices

We developed secure token handling approaches for the GitHub MCP Server.

### Token Storage Methods

1. **Environment Variables**: Less secure but simpler for one-time use
   ```bash
   export GITHUB_PERSONAL_ACCESS_TOKEN=your_token_here
   ```

2. **Token File**: More secure for persistent use
   ```bash
   echo "your_token_here" > ~/.github_token
   chmod 600 ~/.github_token  # Restrict permissions
   ```

### Secure Token Retrieval

We created a dedicated token helper module that follows security best practices:

```python
def get_github_token():
    """Get GitHub token from token file or environment variable."""
    # Try token file first
    if os.path.exists(TOKEN_FILE):
        with open(TOKEN_FILE, "r") as f:
            token = f.read().strip()
            if token:
                return token
    
    # Try environment variable as fallback
    token = os.environ.get("GITHUB_PERSONAL_ACCESS_TOKEN")
    if token:
        return token
    
    # No token found
    return None
```

### Authentication Best Practices

1. **Never hardcode tokens**: Always use environment variables or secure token files
2. **Check file permissions**: Ensure token files have restricted permissions (0600)
3. **Use library-based auth**: Prefer to use established libraries for authentication
4. **Prompt when missing**: Provide clear instructions when authentication fails
5. **Validate tokens early**: Verify tokens work before attempting multiple operations

## 5. Pull Request Workflow Implementation

We implemented a complete pull request workflow that demonstrates the full capabilities of the GitHub MCP Server.

### Complete PR Workflow Steps

1. **Get Authenticated User**: Verify authentication and get user information
2. **Search Repository**: Locate the specific repository using `search_repositories`
3. **List Branches**: Get the default branch and its SHA
4. **Create Branch**: Create a new branch from the default branch
5. **Create File**: Add a new file to the branch
6. **Create PR**: Create a pull request from the new branch to the default branch

### Key Workflow Components

The workflow is encapsulated in a client class that handles server communication:

```python
class GitHubMCPClient:
    """Client for communicating with the GitHub MCP Server."""
    
    def __init__(self, binary_path):
        """Initialize the client."""
        self.binary_path = binary_path
        self.process = None
        self.request_id = 0
    
    # ...methods for server management...
    
    def call_tool(self, name, arguments):
        """Call a tool in the GitHub MCP Server."""
        # ...implementation...
    
    # ...high-level API methods...
    
    def create_pull_request(self, owner, repo, title, head, base, body, draft=False):
        """Create a new pull request."""
        return self.call_tool("create_pull_request", {
            "owner": owner,
            "repo": repo,
            "title": title,
            "head": head,
            "base": base,
            "body": body,
            "draft": draft
        })
```

## 6. Common Issues and Troubleshooting

Based on our testing experience, we documented common issues and solutions.

### Authentication Issues

- **Token Not Found**: Ensure the token exists and is accessible
- **Invalid Token**: Verify the token has the required scopes
- **Token File Permissions**: Ensure token file has correct permissions (0600)

### Response Parsing Issues

- **Missing Result**: Some responses may not include a result field
- **Content Type**: Check for the correct content type
- **Nested Structure**: Handle deeply nested response structures
- **List vs. Dict**: Be prepared for different response data structures

### Tool Name Issues

- **Incorrect Tool Name**: Verify tool names with `tools/list`
- **Missing Parameters**: Check required parameters for each tool
- **Parameter Format**: Ensure parameters are in the correct format

### Server Communication Issues

- **Server Not Starting**: Check for binary executable permissions
- **Request Format**: Ensure requests follow the JSON-RPC 2.0 format
- **Response Timeout**: Handle timeouts gracefully
- **Process Management**: Properly terminate processes when done

## 7. Best Practices for MCP Server Integration

Based on our testing, we recommend these best practices for integrating with the GitHub MCP Server:

1. **Tool Discovery**: Always use `tools/list` to discover available tools
2. **Response Parsing**: Implement robust response parsing logic
3. **Error Handling**: Add specific error handling for each tool
4. **Type Safety**: Use type checking and defensive programming
5. **Authentication**: Implement secure token handling
6. **Process Management**: Properly manage server processes
7. **Logging**: Add detailed logging for debugging
8. **Timeouts**: Add timeouts for all operations
9. **Clean Up**: Always clean up resources when done
10. **Documentation**: Document all tools, parameters, and response formats

## 8. Example Implementation

The provided `pr_workflow_fixed2.py` script demonstrates a complete implementation of a GitHub MCP Server client that follows all best practices:

- Secure token handling
- Robust response parsing
- Type-safe operations
- Error handling
- Process management
- Detailed logging

This script can be used as a reference for implementing your own GitHub MCP Server client.

## 9. Conclusion

The GitHub MCP Server is a powerful tool for interacting with GitHub, but it requires careful handling of response formats and tool names. By following the best practices outlined in this document, you can build robust integrations with the GitHub MCP Server.