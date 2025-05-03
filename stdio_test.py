#!/usr/bin/env python3
"""
Stdio Test for GitHub MCP Server Binary.

This script tests the stdio transport with the GitHub MCP Server Binary.

Usage:
    python3 stdio_test.py [--verbose]
"""

import argparse
import json
import os
import subprocess
import sys
import time
import uuid
from token_helper import get_github_token

def parse_args():
    """Parse command line arguments."""
    parser = argparse.ArgumentParser(description="Test GitHub MCP Server with stdio transport")
    parser.add_argument("--verbose", action="store_true", help="Enable verbose output")
    return parser.parse_args()

def log(message, level="INFO"):
    """Log a message with timestamp."""
    timestamp = time.strftime("%Y-%m-%d %H:%M:%S", time.localtime())
    print(f"[{timestamp}] {level}: {message}")

def debug(message, verbose=True):
    """Log a debug message if verbose is enabled."""
    if verbose:
        log(message, level="DEBUG")

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
        log(f"ERROR: {error_message} (code {error_code})", "ERROR")
        return None
    
    if "result" not in response:
        log("No 'result' field found in response", "WARNING")
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

class GitHubMCPClient:
    """Client for communicating with the GitHub MCP Server."""
    
    def __init__(self, verbose=False):
        """Initialize the client."""
        self.process = None
        self.request_id = 0
        self.verbose = verbose
    
    def start_server(self):
        """Start the GitHub MCP Server."""
        log("Starting GitHub MCP Server")
        
        # Get GitHub token
        github_token = get_github_token()
        if not github_token:
            log("GitHub token not found", "ERROR")
            log("Please set up a token in ~/.github_token or GITHUB_PERSONAL_ACCESS_TOKEN environment variable")
            sys.exit(1)
        
        # Environment variables
        env = os.environ.copy()
        env["GITHUB_PERSONAL_ACCESS_TOKEN"] = github_token
        
        # Start process
        self.process = subprocess.Popen(
            ["./github-mcp-server", "stdio"],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            env=env,
            text=True,
            bufsize=1  # Line buffered
        )
        
        # Wait for process to start
        time.sleep(1)
        
        # Check if process started successfully
        if self.process.poll() is not None:
            error_message = self.process.stderr.read()
            log(f"Failed to start GitHub MCP Server: {error_message}", "ERROR")
            return False
        
        # Start a thread to read stderr
        def read_stderr():
            while self.process.poll() is None:
                line = self.process.stderr.readline()
                if line:
                    debug(f"MCP: {line.strip()}", self.verbose)
        
        import threading
        stderr_thread = threading.Thread(target=read_stderr)
        stderr_thread.daemon = True
        stderr_thread.start()
        
        log("GitHub MCP Server started successfully")
        return True
    
    def stop_server(self):
        """Stop the GitHub MCP Server."""
        if self.process and self.process.poll() is None:
            log("Stopping GitHub MCP Server")
            self.process.terminate()
            try:
                self.process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                log("Process did not terminate, killing", "WARNING")
                self.process.kill()
            log("GitHub MCP Server stopped")
    
    def call_tool(self, name, arguments):
        """Call a tool in the GitHub MCP Server."""
        if not self.process or self.process.poll() is not None:
            if not self.start_server():
                return None
        
        # Create a unique request ID
        self.request_id += 1
        request_id = str(self.request_id)
        
        # Create the JSON-RPC request
        request = {
            "jsonrpc": "2.0",
            "id": request_id,
            "method": "tools/call",
            "params": {
                "name": name,
                "arguments": arguments
            }
        }
        
        # Convert to JSON and add newline
        request_str = json.dumps(request) + "\n"
        debug(f"Sending request: {request_str}", self.verbose)
        
        try:
            # Send the request
            self.process.stdin.write(request_str)
            self.process.stdin.flush()
            
            # Read the response
            response_str = self.process.stdout.readline()
            
            if not response_str:
                log("No response received", "ERROR")
                return None
            
            debug(f"Received response: {response_str}", self.verbose)
            
            # Parse the response
            try:
                response = json.loads(response_str)
            except json.JSONDecodeError as e:
                log(f"Error parsing response: {e}", "ERROR")
                return None
            
            # Parse and return the result
            return parse_response(response)
            
        except Exception as e:
            log(f"Error calling tool: {e}", "ERROR")
            return None
    
    def get_me(self):
        """Get authenticated user information."""
        return self.call_tool("get_me", {})
    
    def search_repositories(self, query):
        """Search for repositories."""
        return self.call_tool("search_repositories", {"query": query})
    
    def list_tools(self):
        """List available tools."""
        request = {
            "jsonrpc": "2.0",
            "id": str(uuid.uuid4()),
            "method": "tools/list",
            "params": {}
        }
        
        # Convert to JSON and add newline
        request_str = json.dumps(request) + "\n"
        debug(f"Sending tools/list request: {request_str}", self.verbose)
        
        try:
            # Send the request
            self.process.stdin.write(request_str)
            self.process.stdin.flush()
            
            # Read the response
            response_str = self.process.stdout.readline()
            
            if not response_str:
                log("No response received for tools/list", "ERROR")
                return None
            
            debug(f"Received tools/list response: {response_str}", self.verbose)
            
            # Parse the response
            try:
                response = json.loads(response_str)
            except json.JSONDecodeError as e:
                log(f"Error parsing tools/list response: {e}", "ERROR")
                return None
            
            # Extract tools from the response
            if "result" in response and "tools" in response["result"]:
                return response["result"]["tools"]
            else:
                log("No tools found in response", "WARNING")
                return []
                
        except Exception as e:
            log(f"Error listing tools: {e}", "ERROR")
            return None
    
    def __enter__(self):
        """Enter context manager."""
        self.start_server()
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        """Exit context manager."""
        self.stop_server()

def main():
    """Main function to run the tests."""
    args = parse_args()
    verbose = args.verbose
    
    with GitHubMCPClient(verbose=verbose) as client:
        # Test 1: Get authenticated user
        log("Test 1: Get authenticated user")
        user = client.get_me()
        if user and isinstance(user, dict) and "login" in user:
            log(f"✅ Successfully authenticated as: {user['login']}")
            if verbose:
                log(f"User details: {json.dumps(user, indent=2)}")
        else:
            log("❌ Failed to get authenticated user", "ERROR")
            return False
        
        # Test 2: Search repositories
        log("Test 2: Search repositories")
        query = "language:go stars:>1000"
        log(f"Searching for: {query}")
        repos = client.search_repositories(query)
        if repos and isinstance(repos, dict) and "items" in repos:
            items = repos.get("items", [])
            log(f"✅ Found {len(items)} repositories")
            if verbose and items:
                for repo in items[:3]:  # Show first 3 repos
                    log(f"  - {repo.get('full_name')}: {repo.get('description', 'No description')}")
        else:
            log("❌ Failed to search repositories", "ERROR")
            return False
        
        # Test 3: List tools
        log("Test 3: List tools")
        tools = client.list_tools()
        if tools and isinstance(tools, list):
            log(f"✅ Found {len(tools)} tools")
            
            if verbose:
                # Group by category
                categories = {}
                for tool in tools:
                    name = tool.get("name", "")
                    parts = name.split("/")
                    if len(parts) > 1:
                        category = parts[0]
                    else:
                        category = "Other"
                    
                    if category not in categories:
                        categories[category] = []
                    
                    categories[category].append(name)
                
                # Display categories and counts
                for category, category_tools in sorted(categories.items()):
                    log(f"  - {category}: {len(category_tools)} tools")
        else:
            log("❌ Failed to list tools", "ERROR")
            return False
    
    log("🎉 All tests passed!")
    return True

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)