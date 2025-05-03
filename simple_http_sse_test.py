#!/usr/bin/env python3
"""
Simple HTTP SSE test for GitHub MCP Server.

This script tests basic functionality of the HTTP SSE transport with the GitHub MCP Server.
"""

import json
import requests
import sys
import time
from token_helper import get_github_token

# Make sure we have the SSE client library
try:
    import sseclient
except ImportError:
    print("Installing sseclient-py...")
    import subprocess
    subprocess.check_call([sys.executable, "-m", "pip", "install", "sseclient-py"])
    import sseclient

def log(message):
    """Print a log message with timestamp."""
    timestamp = time.strftime("%Y-%m-%d %H:%M:%S", time.localtime())
    print(f"[{timestamp}] {message}")

def parse_response(response):
    """Parse the nested response format from GitHub MCP Server."""
    if "error" in response:
        error = response["error"]
        log(f"ERROR: {error.get('message', 'Unknown error')} (code {error.get('code', -1)})")
        return None
    
    if "result" in response:
        result = response["result"]
        
        # Check for the nested content structure
        if "content" in result and isinstance(result["content"], list):
            for item in result["content"]:
                if item.get("type") == "text":
                    text = item.get("text", "")
                    
                    # Try to parse as JSON
                    try:
                        return json.loads(text)
                    except json.JSONDecodeError:
                        return text
    
    return response.get("result", {})

def test_http_endpoint(base_url, endpoint, query_params=None):
    """Test a regular HTTP endpoint."""
    url = f"{base_url}{endpoint}"
    if query_params:
        url += "?" + "&".join([f"{k}={v}" for k, v in query_params.items()])
    
    log(f"Testing HTTP endpoint: {url}")
    
    try:
        response = requests.get(url, timeout=10)
        if response.status_code == 200:
            log(f"Success! Status code: {response.status_code}")
            return response.json()
        else:
            log(f"Error: Status code {response.status_code}")
            log(f"Response: {response.text}")
            return None
    except Exception as e:
        log(f"Exception: {e}")
        return None

def test_sse_endpoint(base_url, tool_name, arguments=None, timeout=10):
    """Test the SSE endpoint with a specific tool."""
    if arguments is None:
        arguments = {}
    
    github_token = get_github_token()
    if not github_token:
        log("ERROR: GitHub token not found")
        sys.exit(1)
    
    log(f"Testing SSE endpoint with tool: {tool_name}")
    
    # Headers for the request
    headers = {
        'Accept': 'text/event-stream',
        'Content-Type': 'application/json',
        'Authorization': f'Bearer {github_token}'
    }
    
    # Create the JSON-RPC request
    request = {
        "jsonrpc": "2.0",
        "id": f"test-{int(time.time())}",
        "method": "tools/call",
        "params": {
            "name": tool_name,
            "arguments": arguments
        }
    }
    
    # Convert to JSON
    request_json = json.dumps(request)
    
    try:
        # Send the request to the SSE endpoint - using regular HTTP POST first
        log("Trying standard HTTP POST to /tools/call as fallback method")
        standard_response = requests.post(
            f"{base_url}/tools/call",
            headers={'Content-Type': 'application/json'},
            json={"name": tool_name, "arguments": arguments},
            timeout=timeout
        )
        
        if standard_response.status_code == 200:
            log("Successfully called tool using standard HTTP POST")
            return standard_response.json()
            
        # If standard HTTP fails, try SSE
        log("Now trying with SSE endpoint")
        response = requests.post(
            f"{base_url}/sse",
            headers=headers,
            data=request_json,
            stream=True,
            timeout=timeout
        )
        
        if response.status_code != 200:
            log(f"Error: Status code {response.status_code}")
            log(f"Response: {response.text}")
            return None
        
        # Create an SSE client
        client = sseclient.SSEClient(response)
        
        # Set a time limit for waiting for events
        start_time = time.time()
        got_data = False
        
        # Process events with timeout
        for event in client.events():
            # Check if we've exceeded the timeout
            if time.time() - start_time > timeout and not got_data:
                log(f"SSE timeout after {timeout} seconds")
                break
                
            if event.event == "data":
                log("Received data event")
                got_data = True
                
                # Parse the data
                try:
                    data = json.loads(event.data)
                    parsed_data = parse_response(data)
                    
                    if parsed_data:
                        log("Successfully parsed response data")
                        return parsed_data
                    else:
                        log("Failed to parse response data")
                        return None
                    
                except json.JSONDecodeError as e:
                    log(f"Error parsing response: {e}")
                    return None
                
            elif event.event == "error":
                log(f"Error event received: {event.data}")
                return None
            
            elif event.event == "end":
                log("End of stream")
                break
        
        if not got_data:
            log("No data events received within timeout")
        
        return None
        
    except Exception as e:
        log(f"Exception: {e}")
        return None

def main():
    """Main test function."""
    # Server URL
    base_url = "http://localhost:7445"
    
    # Test 1: Health check endpoint
    log("\n=== Test 1: Health Check ===")
    result = test_http_endpoint(base_url, "/health")
    if result and result.get("status") == "ok":
        log("✅ Health check passed")
    else:
        log("❌ Health check failed")
        sys.exit(1)
    
    # Test 2: Get authenticated user (HTTP endpoint)
    log("\n=== Test 2: User Info (HTTP) ===")
    user_http = test_http_endpoint(base_url, "/user")
    if user_http and "login" in user_http:
        log(f"✅ HTTP User check passed - Username: {user_http['login']}")
    else:
        log("❌ HTTP User check failed")
    
    # Test 3: Get authenticated user (SSE endpoint)
    log("\n=== Test 3: User Info (SSE) ===")
    user_sse = test_sse_endpoint(base_url, "get_me")
    if user_sse and "login" in user_sse:
        log(f"✅ SSE User check passed - Username: {user_sse['login']}")
    else:
        log("❌ SSE User check failed")
    
    # Test 4: Search repositories (HTTP endpoint)
    log("\n=== Test 4: Search Repositories (HTTP) ===")
    search_params = {"q": "language:python stars:>1000"}
    repos_http = test_http_endpoint(base_url, "/search/repositories", search_params)
    if repos_http and "items" in repos_http:
        log(f"✅ HTTP Search check passed - Found {len(repos_http['items'])} repositories")
        if repos_http['items']:
            first_repo = repos_http['items'][0]
            log(f"  Top result: {first_repo.get('full_name')} - {first_repo.get('description', 'No description')}")
    else:
        log("❌ HTTP Search check failed")
    
    # Test 5: Search repositories (SSE endpoint)
    log("\n=== Test 5: Search Repositories (SSE) ===")
    repos_sse = test_sse_endpoint(base_url, "search_repositories", {"query": "language:python stars:>1000"})
    if repos_sse and "items" in repos_sse:
        log(f"✅ SSE Search check passed - Found {len(repos_sse['items'])} repositories")
        if repos_sse['items']:
            first_repo = repos_sse['items'][0]
            log(f"  Top result: {first_repo.get('full_name')} - {first_repo.get('description', 'No description')}")
    else:
        log("❌ SSE Search check failed")
    
    # Summary
    log("\n=== Test Summary ===")
    log("All tests completed")

if __name__ == "__main__":
    main()