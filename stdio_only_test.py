#!/usr/bin/env python3
"""
Stdio-Only Test for GitHub MCP Server Binary.

This script sends a single command over stdio to test the GitHub MCP Server Binary.
"""

import json
import os
import subprocess
import sys
from token_helper import get_github_token

def main():
    """Run a simple test with the GitHub MCP Server using stdio transport."""
    # Get GitHub token
    token = get_github_token()
    if not token:
        print("GitHub token not found")
        print("Please set up a token in ~/.github_token or GITHUB_PERSONAL_ACCESS_TOKEN environment variable")
        return 1
    
    # Set up environment
    env = os.environ.copy()
    env["GITHUB_PERSONAL_ACCESS_TOKEN"] = token
    
    # Start the GitHub MCP Server process
    print("Starting GitHub MCP Server...")
    process = subprocess.Popen(
        ["./github-mcp-server", "stdio"],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        env=env,
        text=True,
        bufsize=1
    )
    
    # Create a request to get the authenticated user
    request = {
        "jsonrpc": "2.0",
        "id": "1",
        "method": "tools/call",
        "params": {
            "name": "get_me",
            "arguments": {}
        }
    }
    
    # Convert to JSON and add newline
    request_str = json.dumps(request) + "\n"
    
    try:
        # Send the request
        print(f"Sending request: {request_str}")
        process.stdin.write(request_str)
        process.stdin.flush()
        
        # Read the response
        print("Reading response...")
        response_str = process.stdout.readline()
        
        if not response_str:
            print("No response received")
            return 1
        
        print(f"Received response: {response_str}")
        
        # Parse the response
        response = json.loads(response_str)
        
        # Check for errors
        if "error" in response:
            error = response["error"]
            error_message = error.get("message", "Unknown error")
            error_code = error.get("code", -1)
            print(f"Error: {error_message} (code {error_code})")
            return 1
        
        # Extract the result
        if "result" in response:
            result = response["result"]
            # Check for content field
            if "content" in result and isinstance(result["content"], list):
                for item in result["content"]:
                    if item.get("type") == "text":
                        text = item.get("text", "")
                        # Try to parse as JSON
                        try:
                            user_data = json.loads(text)
                            print(f"Successfully authenticated as: {user_data.get('login')}")
                        except json.JSONDecodeError:
                            print(f"Text content: {text}")
        
        print("Test completed successfully!")
        return 0
        
    except Exception as e:
        print(f"Error: {e}")
        return 1
    
    finally:
        # Terminate the process
        process.terminate()
        try:
            process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            process.kill()
        print("GitHub MCP Server stopped")

if __name__ == "__main__":
    sys.exit(main())