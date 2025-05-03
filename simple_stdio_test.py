#!/usr/bin/env python3
"""
Simple Stdio Test for GitHub MCP Server Binary.

This script tests the direct stdio transport with the GitHub MCP Server Binary.
"""

import json
import os
import subprocess
import sys
import time
from token_helper import get_github_token

# Get GitHub token
token = get_github_token()
if not token:
    print("GitHub token not found")
    print("Please set up a token in ~/.github_token or GITHUB_PERSONAL_ACCESS_TOKEN environment variable")
    sys.exit(1)

# Set up environment with token
env = os.environ.copy()
env["GITHUB_PERSONAL_ACCESS_TOKEN"] = token

print("Starting GitHub MCP Server process...")
try:
    # Start the GitHub MCP Server process
    process = subprocess.Popen(
        ["./github-mcp-server", "stdio"],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        env=env,
        text=True,
        bufsize=1  # Line buffered
    )

    # Wait a moment to ensure the process is started
    time.sleep(1)

    # Check if the process is still running
    if process.poll() is not None:
        stderr_output = process.stderr.read()
        print(f"Failed to start GitHub MCP Server: {stderr_output}")
        sys.exit(1)

    print("GitHub MCP Server started successfully")

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

    print(f"Sending request: {request_str}")

    # Send the request
    process.stdin.write(request_str)
    process.stdin.flush()

    # Read the response
    print("Reading response...")
    response_str = process.stdout.readline()

    if not response_str:
        print("No response received")
        sys.exit(1)

    print(f"Received response: {response_str}")

    # Parse the response
    response = json.loads(response_str)

    # Check for errors
    if "error" in response:
        error = response["error"]
        error_message = error.get("message", "Unknown error")
        error_code = error.get("code", -1)
        print(f"Error: {error_message} (code {error_code})")
        sys.exit(1)

    # Extract the result
    if "result" in response:
        result = response["result"]
        print(f"Raw result: {json.dumps(result, indent=2)}")

        # Check if result contains content field (the new format)
        if "content" in result and isinstance(result["content"], list):
            for item in result["content"]:
                if item.get("type") == "text":
                    text = item.get("text", "")
                    
                    # Try to parse the text as JSON
                    try:
                        user_data = json.loads(text)
                        print(f"User data: {json.dumps(user_data, indent=2)}")
                        print(f"Successfully authenticated as: {user_data.get('login')}")
                    except json.JSONDecodeError:
                        print(f"Text content (not JSON): {text}")
        else:
            print("No content field found in result")
    else:
        print("No result field found in response")

    print("Test completed successfully!")

except Exception as e:
    print(f"Error: {e}")
    sys.exit(1)

finally:
    # Terminate the process
    if 'process' in locals() and process.poll() is None:
        print("Stopping GitHub MCP Server")
        process.terminate()
        try:
            process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            process.kill()
            print("Process had to be killed")