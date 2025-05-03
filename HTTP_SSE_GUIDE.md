# GitHub MCP Server Binary: HTTP SSE Transport Guide

> **IMPORTANT NOTE**: This guide specifically covers the HTTP SSE transport for the GitHub MCP Server Binary implementation. While the binary we tested does not natively support HTTP SSE transport (only stdio transport), this guide explains how to create a wrapper to provide HTTP SSE functionality.

This document provides details on creating and using a custom HTTP Server-Sent Events (SSE) wrapper around the GitHub MCP Server Binary implementation. Since the binary itself only supports stdio transport directly, we've created a wrapper to provide HTTP SSE functionality, allowing for remote communication over HTTP.

## What is HTTP SSE Transport?

HTTP Server-Sent Events (SSE) is a server push technology enabling a client to receive automatic updates from a server via an HTTP connection. In the context of the GitHub MCP Server:

- It allows remote communication between clients and the server
- It uses standard HTTP protocols making it firewall-friendly
- It supports asynchronous communication patterns
- It can be used for remote model context protocol interactions

## Creating an HTTP SSE Wrapper

Since the GitHub MCP Server Binary we tested only supports stdio transport directly, we've created a wrapper to provide HTTP SSE functionality. Our implementation is available in the repository as `http_mcp_wrapper.py`. Here's a simplified example of how it works:

```python
#!/usr/bin/env python3
"""
HTTP and SSE Wrapper for GitHub MCP Server.

This script creates an HTTP server that communicates with the GitHub MCP Server
using stdio transport and returns responses via HTTP or Server-Sent Events (SSE).
"""

import argparse
import http.server
import json
import os
import socketserver
import subprocess
import sys
import threading
import time
import uuid
from urllib.parse import parse_qs, urlparse
from token_helper import get_github_token

# Global process to communicate with
mcp_process = None
token = None
verbose = False

def log(message, level="INFO"):
    """Log a message with timestamp."""
    timestamp = time.strftime("%Y-%m-%d %H:%M:%S", time.localtime())
    print(f"[{timestamp}] {level}: {message}", flush=True)

def start_mcp_server():
    """Start the GitHub MCP Server process."""
    global mcp_process, token
    
    # Environment variables
    env = os.environ.copy()
    env["GITHUB_PERSONAL_ACCESS_TOKEN"] = token
    
    # Start process
    mcp_process = subprocess.Popen(
        ["./github-mcp-server", "stdio"],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        env=env,
        text=True,
        bufsize=1  # Line buffered
    )
    
    # Start a thread to read stderr
    def read_stderr():
        while mcp_process and mcp_process.poll() is None:
            line = mcp_process.stderr.readline()
            if line:
                log(f"MCP: {line.strip()}", "MCP")
    
    stderr_thread = threading.Thread(target=read_stderr)
    stderr_thread.daemon = True
    stderr_thread.start()
    
    return True

class MCPRequestHandler(http.server.SimpleHTTPRequestHandler):
    """HTTP Request Handler for GitHub MCP Server requests."""
    
    def do_GET(self):
        """Handle GET requests."""
        # Endpoints for regular HTTP requests like /health, /user, etc.
        # ...
    
    def do_POST(self):
        """Handle POST requests."""
        # Parse URL and body
        url = urlparse(self.path)
        path = url.path
        
        content_length = int(self.headers.get("Content-Length", 0))
        post_data = self.rfile.read(content_length).decode("utf-8")
        data = json.loads(post_data)
        
        # SSE endpoint
        if path == "/sse":
            # Extract request method and params
            if "jsonrpc" not in data or "method" not in data:
                self.send_response(400)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({"error": "Invalid JSON-RPC request"}).encode())
                return
            
            # Set up SSE response headers
            self.send_response(200)
            self.send_header("Content-type", "text/event-stream")
            self.send_header("Cache-Control", "no-cache")
            self.send_header("Connection", "keep-alive")
            self.send_header("X-Accel-Buffering", "no")  # For NGINX
            self.end_headers()
            
            # Process the request and stream the response
            self.handle_sse_request(data)
            return
        
        # Other endpoints...
        
    def handle_sse_request(self, request_data):
        """Handle an SSE request by streaming the response."""
        # Extract method and params
        method = request_data.get("method")
        params = request_data.get("params", {})
        request_id = request_data.get("id", str(uuid.uuid4()))
        
        # Handle different methods (tools/list, tools/call, etc.)
        if method == "tools/call":
            # Create the request
            request = {
                "jsonrpc": "2.0",
                "id": request_id,
                "method": "tools/call",
                "params": params
            }
            
            # Send the request and stream the response
            self.stream_sse_response(request)
            return
        
        # ...
        
    def stream_sse_response(self, request):
        """Stream a response as SSE events."""
        global mcp_process
        
        try:
            # Send the request to the MCP process
            request_str = json.dumps(request) + "\n"
            mcp_process.stdin.write(request_str)
            mcp_process.stdin.flush()
            
            # Read the response
            response_str = mcp_process.stdout.readline()
            response = json.loads(response_str)
            
            # Send the response as an SSE event
            event_data = json.dumps(response)
            self.wfile.write(f"event: data\n".encode())
            self.wfile.write(f"data: {event_data}\n\n".encode())
            self.wfile.flush()
            
            # End the stream
            self.wfile.write(f"event: end\n".encode())
            self.wfile.write(f"data: end\n\n".encode())
            self.wfile.flush()
            
        except Exception as e:
            # Error handling
            self.send_sse_error(str(e), request.get("id", "unknown"))
```

Our full implementation in `http_mcp_wrapper.py` includes:

1. A complete HTTP server with multiple endpoints:
   - `GET /health`: Health check endpoint
   - `GET /user`: Get authenticated user information
   - `GET /search/repositories`: Search GitHub repositories
   - `GET /tools`: List all available tools
   - `POST /tools/call`: Generic endpoint to call any GitHub MCP tool
   - `POST /sse`: Server-Sent Events endpoint for streaming responses

2. Robust error handling and process management
3. Support for both regular HTTP and SSE responses
4. Proper parsing of the nested JSON response format

To run the wrapper, use our provided script:

```bash
./start_http_sse_server.sh [--port PORT] [--verbose]
```

## Running the HTTP SSE Wrapper

To run the HTTP SSE wrapper:

```bash
# Start the HTTP wrapper on port 7444 (default)
./start_http_sse_server.sh

# With verbose logging
./start_http_sse_server.sh --verbose

# On a custom port
./start_http_sse_server.sh --port 8000
```

This will:
1. Activate the Python virtual environment if it exists
2. Install required dependencies if needed
3. Start the GitHub MCP Server Binary using stdio transport
4. Create an HTTP server that listens on the specified port
5. Forward JSON-RPC requests to the GitHub MCP Server Binary
6. Return responses as SSE events or regular HTTP responses

To test the HTTP SSE wrapper, you can use our test script:

```bash
./run_http_sse_test.sh

# With verbose logging
./run_http_sse_test.sh --verbose

# On a custom port
./run_http_sse_test.sh --port 8000
```

This test script will verify the following functionality:
1. Connection to the HTTP server
2. Authentication with the GitHub API
3. Getting user information
4. Searching repositories
5. Listing available tools

## HTTP SSE Client Examples

### Python Client Example

Here's a sample Python client based on our `http_sse_test.py` implementation:

```python
import json
import requests
import sys
import uuid
from token_helper import get_github_token  # Our helper for reading tokens

# Make sure you have the SSE client library
try:
    import sseclient
except ImportError:
    print("The 'sseclient' library is required for this script.")
    print("Install it with: pip install sseclient-py")
    sys.exit(1)

def http_sse_client(server_url, github_token, tool_name, arguments, verbose=False):
    """
    Send a request to the GitHub MCP Server using HTTP SSE.
    
    Args:
        server_url (str): Server URL (e.g. http://localhost:7444)
        github_token (str): GitHub personal access token
        tool_name (str): The name of the tool to call (e.g. "get_me")
        arguments (dict): Tool arguments
        verbose (bool): Enable verbose output
        
    Returns:
        The parsed response
    """
    # Set up headers with authorization
    headers = {
        'Accept': 'text/event-stream',
        'Content-Type': 'application/json',
        'Authorization': f'Bearer {github_token}'
    }
    
    # Create a unique request ID
    request_id = str(uuid.uuid4())
    
    # Create the JSON-RPC request
    request = {
        "jsonrpc": "2.0",
        "id": request_id,
        "method": "tools/call",
        "params": {
            "name": tool_name,
            "arguments": arguments
        }
    }
    
    # Convert to JSON
    request_json = json.dumps(request)
    if verbose:
        print(f"Sending request: {request_json}")
    
    try:
        # Send the request to the SSE endpoint
        response = requests.post(
            f"{server_url}/sse",
            headers=headers,
            data=request_json,
            stream=True,
            timeout=10
        )
        
        # Check if the request was successful
        if response.status_code != 200:
            print(f"HTTP error: {response.status_code} - {response.text}")
            return None
        
        # Create an SSE client
        client = sseclient.SSEClient(response)
        
        # Process events
        for event in client.events():
            if event.event == "data":
                try:
                    data = json.loads(event.data)
                    if verbose:
                        print(f"Received data: {json.dumps(data, indent=2)}")
                    
                    # Parse the nested response format
                    parsed_data = parse_response(data)
                    return parsed_data
                except json.JSONDecodeError as e:
                    print(f"Error parsing response: {e}")
                    return None
            elif event.event == "error":
                print(f"Error event received: {event.data}")
                return None
            elif event.event == "end":
                if verbose:
                    print("End of stream")
                break
        
        print("No data events received")
        return None
        
    except requests.exceptions.RequestException as e:
        print(f"Request error: {e}")
        return None
    except Exception as e:
        print(f"Unexpected error: {e}")
        return None

def parse_response(response):
    """
    Parse a response from the GitHub MCP Server.
    
    This function handles the complex nested response format
    sometimes returned by the GitHub MCP Server.
    
    Args:
        response (dict): The JSON-RPC response from the server
        
    Returns:
        The parsed result data, which could be a dict, list, or string
    """
    if "error" in response:
        error = response["error"]
        error_message = error.get("message", "Unknown error")
        error_code = error.get("code", -1)
        print(f"ERROR: {error_message} (code {error_code})")
        return None
    
    if "result" not in response:
        print("No 'result' field found in response")
        return {}
    
    result = response["result"]
    
    # Check if result contains content field (the new format)
    if "content" in result and isinstance(result["content"], list):
        for item in result["content"]:
            if item.get("type") == "text":
                text = item.get("text", "")
                
                # Try to parse the text as JSON
                try:
                    return json.loads(text)
                except json.JSONDecodeError:
                    # If it's not valid JSON, return the text as is
                    return text
    
    # If no content field or parsing failed, return the result as is
    return result

# Example usage
if __name__ == "__main__":
    server_url = "http://localhost:7444"  # Default server URL
    github_token = get_github_token()     # Get token from environment or file
    
    if not github_token:
        print("GitHub token not found!")
        sys.exit(1)
    
    # Test the get_me tool
    print("Testing get_me tool...")
    user_info = http_sse_client(server_url, github_token, "get_me", {}, verbose=True)
    
    if user_info and "login" in user_info:
        print(f"Successfully authenticated as: {user_info['login']}")
    else:
        print("Failed to get user information")
```

For a more comprehensive example, see our full `http_sse_test.py` implementation in the repository.

## Response Format Considerations

The HTTP SSE transport uses the same response format as the stdio transport, with the nested content structure described in our findings. Make sure to apply the same parsing logic:

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

## Authentication with HTTP SSE

When using HTTP SSE transport, you can authenticate in multiple ways:

1. **Bearer Token in Authorization Header** (preferred method)
   ```
   Authorization: Bearer your_github_token
   ```

2. **Token Query Parameter** (less secure)
   ```
   http://localhost:7444/sse?token=your_github_token
   ```

3. **Environment Variable** (when running the server)
   ```bash
   export GITHUB_PERSONAL_ACCESS_TOKEN=your_token
   ./github-mcp-server http --port 7444
   ```

## Troubleshooting HTTP SSE

Common issues when working with HTTP SSE transport:

1. **Connection Refused**
   - Ensure the server is running and the port is correct
   - Check for firewall blocking the connection

2. **Authentication Errors**
   - Verify your token is correctly included in the headers
   - Check token permissions

3. **Timeout Issues**
   - HTTP SSE maintains a long-lived connection; some proxies might disconnect after inactivity
   - Implement reconnection logic in your client

4. **Response Parsing Issues**
   - The same nested response format applies to HTTP SSE
   - Use the same parsing logic as described in our findings

## HTTP SSE vs. stdio Transport

| Feature | HTTP SSE | stdio |
|---------|----------|-------|
| Remote Communication | Yes | No (local only) |
| Connection Type | Long-lived HTTP | Pipe |
| Firewall Considerations | Standard HTTP (usually allowed) | N/A |
| Authentication | HTTP headers, query params | Environment variables |
| Response Format | Same nested format | Same nested format |
| Typical Use Case | Remote services | Local development |

## Conclusion

The HTTP SSE transport provides a way to communicate with the GitHub MCP Server remotely using standard HTTP protocols. Our implementation demonstrates how to create a wrapper around the GitHub MCP Server Binary to provide both regular HTTP endpoints and Server-Sent Events for streaming responses.

We've successfully implemented and tested:
1. A complete HTTP server with multiple endpoints
2. SSE streaming for GitHub MCP Server responses
3. Robust error handling and process management
4. Proper parsing of the nested JSON response format

To use our implementation, simply:
1. Run `./start_http_sse_server.sh` to start the HTTP SSE server
2. Test it with `./run_http_sse_test.sh` to verify functionality
3. Integrate with your own applications using the client example provided

For more information on using the GitHub MCP Server Binary implementation, refer to the [TESTING_GUIDE.md](./TESTING_GUIDE.md) and [GITHUB_MCP_FINDINGS.md](./GITHUB_MCP_FINDINGS.md) documents.