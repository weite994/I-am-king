#!/usr/bin/env python3
"""
GitHub MCP Server JSON-RPC Test Script

This script tests communication with the GitHub MCP Server using the correct JSON-RPC format
discovered from examining the mcpcurl tool implementation.
"""

import os
import sys
import json
import random
import logging
import subprocess
from typing import Dict, Any, Optional

# Set up logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger("mcp-jsonrpc-test")

# Get GitHub token from environment or token file
def get_github_token():
    """Get GitHub token from environment or token file."""
    token = os.environ.get("GITHUB_PERSONAL_ACCESS_TOKEN")
    if token:
        return token
    
    # Try to get from token file
    token_file = os.path.expanduser("~/.github_token")
    if os.path.exists(token_file):
        with open(token_file, "r") as f:
            return f.read().strip()
    
    return None

def build_jsonrpc_request(method: str, tool_name: str, arguments: Optional[Dict[str, Any]] = None) -> Dict:
    """Build a JSON-RPC request using the correct format.
    
    Args:
        method: The JSON-RPC method (e.g., "tools/call")
        tool_name: The name of the tool to call
        arguments: The arguments for the tool
        
    Returns:
        A JSON-RPC request dictionary
    """
    return {
        "jsonrpc": "2.0",
        "id": random.randint(1, 10000),
        "method": method,
        "params": {
            "name": tool_name,
            "arguments": arguments or {}
        }
    }

def run_command(cmd: str, input_data: Optional[Dict] = None) -> Dict:
    """Run a command and send input data to it.
    
    Args:
        cmd: The command to run
        input_data: The data to send to the command's stdin
        
    Returns:
        The parsed JSON response
    """
    process = subprocess.Popen(
        cmd.split(),
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        universal_newlines=True
    )
    
    if input_data:
        input_str = json.dumps(input_data) + "\n"
        stdout, stderr = process.communicate(input_str)
    else:
        stdout, stderr = process.communicate()
    
    if process.returncode != 0:
        logger.error(f"Command failed with exit code {process.returncode}")
        logger.error(f"stderr: {stderr}")
        raise RuntimeError(f"Command failed: {stderr}")
    
    if stderr:
        logger.info(f"stderr: {stderr}")
    
    try:
        return json.loads(stdout.strip())
    except json.JSONDecodeError:
        logger.error(f"Failed to parse JSON response: {stdout}")
        raise

def test_get_authenticated_user(server_cmd: str) -> None:
    """Test getting the authenticated user.
    
    Args:
        server_cmd: The command to run the GitHub MCP Server
    """
    logger.info("Testing get_authenticated_user...")
    
    # Build request using proper JSON-RPC format
    request = build_jsonrpc_request("tools/call", "get_me", {})
    
    # Run command and get response
    try:
        response = run_command(server_cmd, request)
        logger.info(f"Response: {json.dumps(response, indent=2)}")
        
        # Check if we got a result
        if "result" in response:
            logger.info("✅ get_authenticated_user test successful")
        else:
            logger.error("❌ get_authenticated_user test failed: No result in response")
    except Exception as e:
        logger.error(f"❌ get_authenticated_user test failed: {e}")

def test_list_repos(server_cmd: str) -> None:
    """Test listing repositories.
    
    Args:
        server_cmd: The command to run the GitHub MCP Server
    """
    logger.info("Testing list_repositories...")
    
    # Build request using proper JSON-RPC format
    request = build_jsonrpc_request("tools/call", "search_repos", {
        "query": "user:github",
        "sort": "updated",
        "per_page": 5
    })
    
    # Run command and get response
    try:
        response = run_command(server_cmd, request)
        logger.info(f"Response: {json.dumps(response, indent=2)}")
        
        # Check if we got a result
        if "result" in response:
            logger.info("✅ list_repositories test successful")
        else:
            logger.error("❌ list_repositories test failed: No result in response")
    except Exception as e:
        logger.error(f"❌ list_repositories test failed: {e}")

def test_get_file(server_cmd: str) -> None:
    """Test getting a file from a repository.
    
    Args:
        server_cmd: The command to run the GitHub MCP Server
    """
    logger.info("Testing get_file_contents...")
    
    # Build request using proper JSON-RPC format
    request = build_jsonrpc_request("tools/call", "get_file_contents", {
        "owner": "github",
        "repo": "docs",
        "path": "README.md"
    })
    
    # Run command and get response
    try:
        response = run_command(server_cmd, request)
        logger.info(f"Response: {json.dumps(response, indent=2)}")
        
        # Check if we got a result
        if "result" in response:
            logger.info("✅ get_file_contents test successful")
        else:
            logger.error("❌ get_file_contents test failed: No result in response")
    except Exception as e:
        logger.error(f"❌ get_file_contents test failed: {e}")

def main():
    """Main entry point."""
    # Check for GitHub token
    token = get_github_token()
    if not token:
        logger.error("GitHub token not found. Please set GITHUB_PERSONAL_ACCESS_TOKEN environment variable or create ~/.github_token")
        return 1
    
    # Check if we have the GitHub MCP Server binary
    server_path = "./github-mcp-server"
    if not os.path.exists(server_path) or not os.access(server_path, os.X_OK):
        logger.error(f"GitHub MCP Server binary not found at {server_path} or not executable")
        return 1
    
    # Construct server command
    server_cmd = f"{server_path} stdio"
    
    # Run tests
    try:
        # Run the tests
        test_get_authenticated_user(server_cmd)
        test_list_repos(server_cmd)
        test_get_file(server_cmd)
        
        logger.info("All tests complete")
        return 0
    except Exception as e:
        logger.error(f"Test failed: {e}")
        return 1

if __name__ == "__main__":
    sys.exit(main())