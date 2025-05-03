#!/usr/bin/env python3
"""
Tool to list all available tools in GitHub MCP Server.

This script lists all available tools in the GitHub MCP Server
by calling the tools/list method.

Usage:
    python3 list_tools.py
"""

import os
import sys
import json
import subprocess
import time

# Import token helper
from token_helper import ensure_token_exists

# Get GitHub token
token = ensure_token_exists()

# Path to the GitHub MCP Server binary
binary_path = "./github-mcp-server"

print(f"Starting GitHub MCP Server: {binary_path}")
env = os.environ.copy()
env["GITHUB_PERSONAL_ACCESS_TOKEN"] = token
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
    # Create a request to list all available tools
    request = {
        "jsonrpc": "2.0",
        "id": "1",
        "method": "tools/list",
        "params": {}
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
    
    # Check for errors
    if "error" in response:
        error = response["error"]
        error_message = error.get("message", "Unknown error")
        error_code = error.get("code", -1)
        print(f"ERROR: {error_message} (code {error_code})")
        sys.exit(1)
    
    # Parse and print the list of tools
    if "result" in response:
        result = response["result"]
        print("\n=== AVAILABLE TOOLS ===\n")
        
        tools = result.get("tools", [])
        
        if not tools:
            print("No tools available.")
        else:
            # Group tools by their category
            tool_categories = {}
            for tool in tools:
                name = tool.get("name", "")
                # Try to extract category from name (assuming format like "category/toolname")
                parts = name.split("/")
                if len(parts) > 1:
                    category = parts[0]
                    tool_name = parts[1]
                else:
                    category = "Other"
                    tool_name = name
                
                if category not in tool_categories:
                    tool_categories[category] = []
                
                tool_categories[category].append({
                    "name": name,
                    "description": tool.get("description", "No description")
                })
            
            # Print tools by category
            for category, category_tools in sorted(tool_categories.items()):
                print(f"\n## {category.upper()} TOOLS\n")
                for tool in sorted(category_tools, key=lambda x: x["name"]):
                    print(f"- {tool['name']}: {tool['description']}")
            
            print(f"\nTotal tools: {len(tools)}")
    else:
        print("WARNING: Response missing 'result' field")
    
except Exception as e:
    print(f"ERROR: Test failed: {e}")
    sys.exit(1)
    
finally:
    # Terminate the process
    print("\nStopping GitHub MCP Server")
    process.terminate()
    try:
        process.wait(timeout=5)
    except subprocess.TimeoutExpired:
        process.kill()