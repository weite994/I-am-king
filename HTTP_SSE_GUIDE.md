# GitHub MCP Server Binary: HTTP SSE Transport Guide

> **IMPORTANT NOTE**: This guide specifically covers the HTTP SSE transport for the GitHub MCP Server Binary implementation. While the binary we tested does not natively support HTTP SSE transport (only stdio transport), this guide explains how to create a wrapper to provide HTTP SSE functionality.

This document provides details on creating and using a custom HTTP Server-Sent Events (SSE) wrapper around the GitHub MCP Server Binary implementation. Since the binary itself only supports stdio transport directly, we'll demonstrate how to create a wrapper to provide HTTP SSE functionality, allowing for remote communication over HTTP.

## What is HTTP SSE Transport?

HTTP Server-Sent Events (SSE) is a server push technology enabling a client to receive automatic updates from a server via an HTTP connection. In the context of the GitHub MCP Server:

- It allows remote communication between clients and the server
- It uses standard HTTP protocols making it firewall-friendly
- It supports asynchronous communication patterns
- It can be used for remote model context protocol interactions

## Creating an HTTP SSE Wrapper

Since the GitHub MCP Server Binary we tested only supports stdio transport directly, we need to create a wrapper to provide HTTP SSE functionality. Here's a Python implementation of such a wrapper:

```python
#!/usr/bin/env python3
"""
HTTP Wrapper for GitHub MCP Server.

This script creates a simple HTTP server that forwards requests to the GitHub MCP Server
using stdio transport and returns the responses as SSE events.
"""

import argparse
import json
import os
import subprocess
import threading
import time
from http.server import HTTPServer, BaseHTTPRequestHandler
from token_helper import get_github_token

# Process to communicate with
mcp_process = None
token = None
verbose = False

class MCPRequestHandler(BaseHTTPRequestHandler):
    """HTTP Request Handler for GitHub MCP Server requests."""
    
    def do_GET(self):
        """Handle GET requests."""
        if self.path == "/health":
            self.send_response(200)
            self.send_header("Content-type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps({"status": "ok"}).encode())
        else:
            self.send_response(404)
            self.send_header("Content-type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps({"error": "Not found"}).encode())
    
    def do_POST(self):
        """Handle POST requests."""
        global mcp_process
        
        if self.path == "/sse":
            # Get request body
            content_length = int(self.headers["Content-Length"])
            request_body = self.rfile.read(content_length).decode()
            
            try:
                # Parse request body
                request = json.loads(request_body)
                
                # Send request to MCP process
                request_str = json.dumps(request) + "\n"
                mcp_process.stdin.write(request_str)
                mcp_process.stdin.flush()
                
                # Read response
                response_str = mcp_process.stdout.readline()
                response = json.loads(response_str)
                
                # Send response as SSE
                self.send_response(200)
                self.send_header("Content-type", "text/event-stream")
                self.send_header("Cache-Control", "no-cache")
                self.end_headers()
                
                # Send data event
                event = f"event: data\ndata: {json.dumps(response)}\n\n"
                self.wfile.write(event.encode())
                
            except Exception as e:
                self.send_response(500)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({"error": str(e)}).encode())
        else:
            self.send_response(404)
            self.send_header("Content-type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps({"error": "Not found"}).encode())

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
        while mcp_process.poll() is None:
            line = mcp_process.stderr.readline()
            if line:
                print(f"MCP: {line.strip()}")
    
    stderr_thread = threading.Thread(target=read_stderr)
    stderr_thread.daemon = True
    stderr_thread.start()
    
    return mcp_process.poll() is None

def main():
    """Main function."""
    global token
    
    parser = argparse.ArgumentParser(description="HTTP Wrapper for GitHub MCP Server")
    parser.add_argument("--port", type=int, default=7444, help="HTTP server port (default: 7444)")
    args = parser.parse_args()
    
    # Get GitHub token
    token = get_github_token()
    if not token:
        print("GitHub token not found")
        sys.exit(1)
    
    # Start MCP server
    if not start_mcp_server():
        print("Failed to start GitHub MCP Server")
        sys.exit(1)
    
    # Create HTTP server
    server_address = ("", args.port)
    httpd = HTTPServer(server_address, MCPRequestHandler)
    
    print(f"Starting HTTP server on port {args.port}")
    print("Press Ctrl+C to stop")
    
    try:
        # Start HTTP server
        httpd.serve_forever()
    except KeyboardInterrupt:
        print("Shutting down")
    finally:
        # Stop MCP server
        mcp_process.terminate()
```

Save this as `http_wrapper.py` and make it executable:

```bash
chmod +x http_wrapper.py
```

## Running the HTTP SSE Wrapper

To run the wrapper:

```bash
# Start the HTTP wrapper on port 7444
python http_wrapper.py --port 7444
```

This will:
1. Start the GitHub MCP Server Binary using stdio transport
2. Create an HTTP server that listens on port 7444
3. Forward JSON-RPC requests to the GitHub MCP Server Binary
4. Return responses as SSE events

## HTTP SSE Client Examples

### Python Client Example

```python
import requests
import json
import sseclient

def http_sse_client(server_url, github_token):
    """Simple HTTP SSE client for GitHub MCP Server."""
    
    # Set up headers with authorization
    headers = {
        'Accept': 'text/event-stream',
        'Content-Type': 'application/json',
        'Authorization': f'Bearer {github_token}'
    }
    
    # Create a request to get authenticated user info
    request = {
        "jsonrpc": "2.0",
        "id": "1",
        "method": "tools/call",
        "params": {
            "name": "get_me",
            "arguments": {}
        }
    }
    
    # Convert to JSON
    request_json = json.dumps(request)
    
    # Send the request to the SSE endpoint
    response = requests.post(
        f"{server_url}/sse",
        headers=headers,
        data=request_json,
        stream=True
    )
    
    # Create an SSE client
    client = sseclient.SSEClient(response)
    
    # Process events
    for event in client.events():
        if event.event == "data":
            data = json.loads(event.data)
            print(f"Received data: {json.dumps(data, indent=2)}")
            # Remember to parse the nested content format as described in our findings
            return data
        elif event.event == "error":
            print(f"Error: {event.data}")
            return None

# Example usage
if __name__ == "__main__":
    server_url = "http://localhost:7444"
    github_token = "your_github_token"
    result = http_sse_client(server_url, github_token)
```

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

The HTTP SSE transport provides a way to communicate with the GitHub MCP Server remotely using standard HTTP protocols. It requires the same response parsing logic as the stdio transport but adds the flexibility of remote communication.

For more information on using the GitHub MCP Server Binary implementation, refer to the [TESTING_GUIDE.md](./TESTING_GUIDE.md) and [GITHUB_MCP_FINDINGS.md](./GITHUB_MCP_FINDINGS.md) documents.