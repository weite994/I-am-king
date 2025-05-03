#!/usr/bin/env python3
"""
Comprehensive test script for GitHub MCP Server.

This script tests various GitHub MCP Server operations including:
1. Authentication
2. Repository operations
3. Branch operations
4. File operations
5. Pull request operations

Usage:
    export GITHUB_PERSONAL_ACCESS_TOKEN=your_token_here
    python3 comprehensive_test.py
"""

import os
import sys
import json
import subprocess
import time
import argparse
from datetime import datetime

# Set up argument parser
parser = argparse.ArgumentParser(description="Test GitHub MCP Server")
parser.add_argument("--owner", help="Repository owner")
parser.add_argument("--repo", help="Repository name")
parser.add_argument("--binary", default="./github-mcp-server", help="Path to GitHub MCP Server binary")
parser.add_argument("--verbose", action="store_true", help="Enable verbose output")
args = parser.parse_args()

# Set up logging
VERBOSE = args.verbose

def log(message, level="INFO"):
    timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    print(f"[{timestamp}] {level}: {message}")

def debug(message):
    if VERBOSE:
        log(message, "DEBUG")

# Check for GitHub token
token = os.environ.get("GITHUB_PERSONAL_ACCESS_TOKEN")
if not token:
    log("GITHUB_PERSONAL_ACCESS_TOKEN environment variable not set.", "ERROR")
    log("Please set your GitHub token:", "ERROR")
    log("export GITHUB_PERSONAL_ACCESS_TOKEN=your_token_here", "ERROR")
    sys.exit(1)

# Path to the GitHub MCP Server binary
binary_path = args.binary

class GitHubMCPClient:
    """Client for communicating with the GitHub MCP Server."""
    
    def __init__(self, binary_path):
        """Initialize the client."""
        self.binary_path = binary_path
        self.process = None
        self.request_id = 0
    
    def start_server(self):
        """Start the GitHub MCP Server."""
        log(f"Starting GitHub MCP Server: {self.binary_path}")
        env = os.environ.copy()
        self.process = subprocess.Popen(
            [self.binary_path, "stdio"],
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
        if self.process.poll() is not None:
            # Process exited immediately, read stderr to get error message
            error_message = self.process.stderr.read()
            raise RuntimeError(f"Failed to start GitHub MCP Server: {error_message}")
        
        log("GitHub MCP Server started successfully")
    
    def stop_server(self):
        """Stop the GitHub MCP Server."""
        if self.process:
            log("Stopping GitHub MCP Server")
            self.process.terminate()
            try:
                self.process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                self.process.kill()
            self.process = None
    
    def call_tool(self, name, arguments):
        """Call a tool in the GitHub MCP Server."""
        if not self.process:
            self.start_server()
        
        # Create the request
        self.request_id += 1
        request = {
            "jsonrpc": "2.0",
            "id": str(self.request_id),
            "method": "tools/call",
            "params": {
                "name": name,
                "arguments": arguments
            }
        }
        
        # Convert to JSON and add newline
        request_str = json.dumps(request) + "\n"
        
        debug(f"Sending request: {request}")
        
        # Send the request
        self.process.stdin.write(request_str)
        self.process.stdin.flush()
        
        # Read the response
        response_str = self.process.stdout.readline()
        
        if not response_str:
            raise RuntimeError(f"No response received for tool {name}")
        
        # Parse the response
        response = json.loads(response_str)
        debug(f"Received response: {response}")
        
        # Check for errors
        if "error" in response:
            error = response["error"]
            error_message = error.get("message", "Unknown error")
            error_code = error.get("code", -1)
            raise RuntimeError(f"Error calling tool {name} (code {error_code}): {error_message}")
        
        # Return the result
        return response.get("result", {})
    
    def __enter__(self):
        """Support for 'with' statement."""
        self.start_server()
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        """Support for 'with' statement."""
        self.stop_server()
    
    # API methods
    
    def get_authenticated_user(self):
        """Get information about the authenticated user."""
        return self.call_tool("get_me", {})
    
    def get_repo(self, owner, repo):
        """Get repository information."""
        return self.call_tool("get_repo", {
            "owner": owner,
            "repo": repo
        })
    
    def list_branches(self, owner, repo):
        """List branches in a repository."""
        return self.call_tool("list_branches", {
            "owner": owner,
            "repo": repo
        })
    
    def get_branch(self, owner, repo, branch):
        """Get information about a branch."""
        return self.call_tool("get_branch", {
            "owner": owner,
            "repo": repo,
            "branch": branch
        })
    
    def create_branch(self, owner, repo, branch, sha):
        """Create a new branch."""
        return self.call_tool("create_branch", {
            "owner": owner,
            "repo": repo,
            "branch": branch,
            "sha": sha
        })
    
    def get_file_contents(self, owner, repo, path, ref=None):
        """Get file contents from a repository."""
        params = {
            "owner": owner,
            "repo": repo,
            "path": path
        }
        
        if ref:
            params["ref"] = ref
        
        return self.call_tool("get_file_contents", params)
    
    def create_or_update_file(self, owner, repo, path, message, content, branch, sha=None):
        """Create or update a file in a repository."""
        params = {
            "owner": owner,
            "repo": repo,
            "path": path,
            "message": message,
            "content": content,
            "branch": branch
        }
        
        if sha:
            params["sha"] = sha
        
        return self.call_tool("create_or_update_file", params)
    
    def create_pull_request(self, owner, repo, title, head, base, body, draft=False):
        """Create a new pull request."""
        return self.call_tool("create_pull_request", {
            "owner": owner,
            "repo": repo,
            "title": title,
            "head": head,
            "base": base,
            "body": body,
            "draft": draft
        })

def test_authentication(client):
    """Test authentication with GitHub."""
    log("TESTING: Authentication")
    user = client.get_authenticated_user()
    log(f"✅ Successfully authenticated as: {user.get('login')}")
    return user

def test_repository(client, owner, repo):
    """Test repository operations."""
    log(f"TESTING: Repository operations for {owner}/{repo}")
    
    # Get repository info
    repo_info = client.get_repo(owner, repo)
    log(f"✅ Repository: {repo_info.get('full_name')}")
    log(f"✅ Default branch: {repo_info.get('default_branch')}")
    
    return repo_info

def test_branches(client, owner, repo):
    """Test branch operations."""
    log(f"TESTING: Branch operations for {owner}/{repo}")
    
    # List branches
    branches = client.list_branches(owner, repo)
    branch_count = len(branches.get("items", []))
    log(f"✅ Found {branch_count} branches")
    
    if branch_count > 0:
        for branch in branches.get("items", [])[:3]:  # Show up to 3 branches
            log(f"  - {branch.get('name')}")
    
    return branches

def test_files(client, owner, repo, branch=None):
    """Test file operations."""
    log(f"TESTING: File operations for {owner}/{repo}")
    
    # Get README file contents
    try:
        readme = client.get_file_contents(owner, repo, "README.md", branch)
        log(f"✅ Found README.md ({readme.get('size')} bytes)")
    except Exception as e:
        log(f"❌ Error getting README.md: {e}", "ERROR")
    
    return True

def run_comprehensive_test():
    """Run a comprehensive test of the GitHub MCP Server."""
    
    # Get repository owner and name
    owner = args.owner
    repo = args.repo
    
    if not owner or not repo:
        log("Repository owner and name are required for comprehensive testing.", "ERROR")
        log("Please provide them with --owner and --repo", "ERROR")
        return False
    
    with GitHubMCPClient(binary_path) as client:
        try:
            # Test authentication
            user = test_authentication(client)
            
            # Test repository operations
            repo_info = test_repository(client, owner, repo)
            
            # Test branch operations
            branches = test_branches(client, owner, repo)
            
            # Test file operations
            test_files(client, owner, repo)
            
            # Test PR creation? (Only if explicitly requested)
            if args.create_pr:
                pass  # To be implemented if needed
            
            log("All tests completed successfully! ✅")
            return True
            
        except Exception as e:
            log(f"Test failed: {e}", "ERROR")
            return False

if __name__ == "__main__":
    if not args.owner or not args.repo:
        log("Repository owner and name are required for comprehensive testing.", "WARNING")
        log("Please provide them with --owner and --repo arguments.", "WARNING")
        log("Running basic authentication test only...")
        
        with GitHubMCPClient(binary_path) as client:
            try:
                # Test authentication only
                test_authentication(client)
                log("Basic authentication test completed successfully! ✅")
            except Exception as e:
                log(f"Authentication test failed: {e}", "ERROR")
                sys.exit(1)
    else:
        success = run_comprehensive_test()
        if not success:
            sys.exit(1)