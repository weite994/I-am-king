# GitHub MCP Server Binary Test Scripts

> **IMPORTANT NOTE**: These test scripts are specifically designed for the GitHub MCP Server Binary implementation. Other implementations might have different behaviors, response formats, or tool names.
>
> The GitHub MCP Server Binary can be built from source following the [official instructions](https://github.com/github/github-mcp-server/blob/main/README.md#build-from-source).
>
> These scripts primarily use the **stdio** transport protocol (standard input/output pipes), although the GitHub MCP Server also supports **HTTP SSE** transport (HTTP Server-Sent Events) for remote communication.

This directory contains test scripts for the GitHub MCP Server Binary implementation. These scripts demonstrate how to use the GitHub MCP Server Binary API to perform various operations, including creating pull requests.

## Test Script Overview

| Script | Description |
|--------|-------------|
| `token_helper.py` | Helper module for securely handling GitHub tokens |
| `list_tools.py` | Lists all available tools in the GitHub MCP Server |
| `simple_test.py` | Basic test for GitHub MCP Server API call |
| `simple_test_updated.py` | Updated version with robust response parsing |
| `comprehensive_test.py` | More comprehensive test covering multiple API calls |
| `pr_workflow_test.py` | Initial PR workflow test (note: contains tool name issues) |
| `pr_workflow_fixed.py` | Fixed PR workflow with correct tool names |
| `pr_workflow_fixed2.py` | Final version with robust response parsing |
| `run_tests.sh` | Script to run all tests |
| `run_fixed_pr_test.sh` | Script to run the fixed PR workflow test |
| `run_fixed_pr_test2.sh` | Script to run the final PR workflow test |

## Key Script Features

### token_helper.py

Securely handles GitHub tokens from either environment variables or token files:

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

### list_tools.py

Discovers all available tools in the GitHub MCP Server:

```python
# Create a request to list all available tools
request = {
    "jsonrpc": "2.0",
    "id": "1",
    "method": "tools/list",
    "params": {}
}
```

### pr_workflow_fixed2.py

The most robust and complete implementation with proper response parsing:

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
        # ...implementation with robust response parsing...
```

## Running the Tests

### Prerequisites

1. GitHub Personal Access Token
2. GitHub MCP Server binary or Docker image
3. Python 3.6+

### Setting Up

```bash
# Save your token to a file (recommended)
echo "your_token_here" > ~/.github_token
chmod 600 ~/.github_token

# Or set as environment variable
export GITHUB_PERSONAL_ACCESS_TOKEN=your_token_here
```

### Discover Available Tools

```bash
# Make sure the binary is executable
chmod +x github-mcp-server

# Run the tool discovery script
python list_tools.py
```

### Run the PR Workflow Test

```bash
# Use the run script
./run_fixed_pr_test2.sh your_username your_repo

# Or run directly
python pr_workflow_fixed2.py --owner your_username --repo your_repo
```

## Response Parsing

The GitHub MCP Server returns responses in a complex format that requires careful parsing:

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

## Response Type Handling

Different tools return different response types, requiring flexible handling:

```python
# Handle different response formats for branches
branches = []
if isinstance(branches_response, list):
    branches = branches_response
elif isinstance(branches_response, dict) and "items" in branches_response:
    branches = branches_response.get("items", [])
```

## Troubleshooting

If you encounter issues running the scripts:

1. Verify that your token is valid and has the necessary permissions
2. Make sure the GitHub MCP Server binary is executable
3. Check that the server is running correctly with `./github-mcp-server --verbose stdio`
4. Use `list_tools.py` to verify the available tools and their names
5. Ensure you're using the correct tool names in your scripts

For more detailed information, see the [TESTING_GUIDE.md](./TESTING_GUIDE.md) and [GITHUB_MCP_FINDINGS.md](./GITHUB_MCP_FINDINGS.md) documents.