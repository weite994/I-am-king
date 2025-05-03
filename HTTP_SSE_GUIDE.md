# GitHub MCP Server Binary: HTTP SSE Transport Guide

> **IMPORTANT NOTE**: This guide specifically covers the HTTP SSE transport for the GitHub MCP Server Binary implementation. Other implementations might have different behaviors, response formats, or tool names.

This document provides details on using the HTTP Server-Sent Events (SSE) transport with the GitHub MCP Server Binary implementation. The HTTP SSE transport allows for remote communication with the GitHub MCP Server over HTTP.

## What is HTTP SSE Transport?

HTTP Server-Sent Events (SSE) is a server push technology enabling a client to receive automatic updates from a server via an HTTP connection. In the context of the GitHub MCP Server:

- It allows remote communication between clients and the server
- It uses standard HTTP protocols making it firewall-friendly
- It supports asynchronous communication patterns
- It can be used for remote model context protocol interactions

## Setting Up HTTP SSE Server

To run the GitHub MCP Server with HTTP SSE transport:

```bash
# Start the server with HTTP SSE transport on port 7444
./github-mcp-server http --port 7444
```

You can also run the server with Docker:

```bash
# Using Docker with HTTP SSE transport
docker run -p 7444:7444 -e GITHUB_PERSONAL_ACCESS_TOKEN=your_token ghcr.io/github/github-mcp-server http --port 7444
```

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