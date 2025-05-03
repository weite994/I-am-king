#!/usr/bin/env python3
"""
Simple test script for GitHub MCP Server.

This script tests the GitHub MCP Server by sending a simple request
to get the authenticated user and printing the response.

Usage:
    export GITHUB_PERSONAL_ACCESS_TOKEN=your_token_here
    python3 simple_test.py
"""

import os
import sys
import json
import subprocess
import time

# Check for GitHub token
token = os.environ.get("GITHUB_PERSONAL_ACCESS_TOKEN")
if not token:
    print("ERROR: GITHUB_PERSONAL_ACCESS_TOKEN environment variable not set.")
    print("Please set your GitHub token:")
    print("export GITHUB_PERSONAL_ACCESS_TOKEN=your_token_here")
    sys.exit(1)

# Path to the GitHub MCP Server binary
binary_path = "./github-mcp-server"

# Start the GitHub MCP Server process
print(f"Starting GitHub MCP Server: {binary_path}")
env = os.environ.copy()
process = subprocess.Popen(
    [binary_path, "stdio"],
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE,
    env=env,
    text=True,
    bufsize=1  # Line buffered
)

# Wait a moment to ensure the process is started
time.sleep(0.5)

# Check if the process is still running
if process.poll() is not None:
    # Process exited immediately, read stderr to get error message
    error_message = process.stderr.read()
    print(f"ERROR: Failed to start GitHub MCP Server: {error_message}")
    sys.exit(1)

print("GitHub MCP Server started successfully")

try:
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
    
    print(f"Sending request: {request}")
    
    # Send the request
    process.stdin.write(request_str)
    process.stdin.flush()
    
    # Read the response
    response_str = process.stdout.readline()
    
    if not response_str:
        print("ERROR: No response received")
        sys.exit(1)
    
    # Parse the response
    response = json.loads(response_str)
    print(f"Received response: {json.dumps(response, indent=2)}")
    
    # Check for errors
    if "error" in response:
        error = response["error"]
        error_message = error.get("message", "Unknown error")
        error_code = error.get("code", -1)
        print(f"ERROR: {error_message} (code {error_code})")
        sys.exit(1)
    
    # Print the result
    if "result" in response:
        result = response["result"]
        print(f"Successfully authenticated as: {result.get('login')}")
        print(f"User details: {json.dumps(result, indent=2)}")
    else:
        print("WARNING: Response missing 'result' field")
    
    print("Test completed successfully!")
    
except Exception as e:
    print(f"ERROR: Test failed: {e}")
    sys.exit(1)
    
finally:
    # Terminate the process
    print("Stopping GitHub MCP Server")
    process.terminate()
    try:
        process.wait(timeout=5)
    except subprocess.TimeoutExpired:
        process.kill()