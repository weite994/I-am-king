#!/usr/bin/env python3
"""
Pull Request Workflow Test for GitHub MCP Server.

This script tests a complete pull request workflow:
1. Authentication
2. Getting repository info
3. Creating a new branch
4. Creating a new file
5. Creating a pull request

Usage:
    export GITHUB_PERSONAL_ACCESS_TOKEN=your_token_here
    python3 pr_workflow_test.py --owner YOUR_USERNAME --repo YOUR_REPO
"""

import os
import sys
import json
import subprocess
import time
import argparse
from datetime import datetime

# Set up argument parser
parser = argparse.ArgumentParser(description="Test GitHub MCP Server PR workflow")
parser.add_argument("--owner", required=True, help="Repository owner")
parser.add_argument("--repo", required=True, help="Repository name")
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

def run_pr_workflow(owner, repo):
    """Run a complete PR workflow test."""
    
    with GitHubMCPClient(binary_path) as client:
        try:
            # Step 1: Get authenticated user
            log("Step 1: Getting authenticated user...")
            user = client.get_authenticated_user()
            log(f"✅ Authenticated as: {user.get('login')}")
            
            # Step 2: Get repository information
            log(f"Step 2: Getting repository information for {owner}/{repo}...")
            repo_info = client.get_repo(owner, repo)
            log(f"✅ Repository: {repo_info.get('full_name')}")
            
            # Get default branch
            default_branch = repo_info.get('default_branch', 'main')
            log(f"✅ Default branch: {default_branch}")
            
            # Step 3: Get default branch information
            log(f"Step 3: Getting information for branch: {default_branch}...")
            branch_info = client.get_branch(owner, repo, default_branch)
            base_sha = branch_info.get('commit', {}).get('sha')
            
            if not base_sha:
                log(f"❌ Failed to get SHA for branch {default_branch}", "ERROR")
                return False
                
            log(f"✅ Base SHA: {base_sha}")
            
            # Step 4: Create a new branch
            timestamp = int(time.time())
            branch_name = f"mcp-test-{timestamp}"
            log(f"Step 4: Creating new branch {branch_name}...")
            
            try:
                new_branch = client.create_branch(owner, repo, branch_name, base_sha)
                log(f"✅ Created branch: {new_branch.get('name')}")
            except Exception as e:
                log(f"❌ Failed to create branch: {e}", "ERROR")
                return False
            
            # Step 5: Create a new file in the branch
            timestamp_str = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            file_path = f"docs/mcp-test-{timestamp}.md"
            
            file_content = f"""# GitHub MCP Test

This file was created by the GitHub MCP Server PR workflow test.

Created at: {timestamp_str}

## Test Information

- User: {user.get('login')}
- Repository: {repo_info.get('full_name')}
- Branch: {branch_name}
- File: {file_path}
- Timestamp: {timestamp}
"""
            
            log(f"Step 5: Creating file {file_path} in branch {branch_name}...")
            
            try:
                file_result = client.create_or_update_file(
                    owner, repo, file_path, 
                    "Add GitHub MCP test file", 
                    file_content, branch_name
                )
                log(f"✅ Created file: {file_result.get('content', {}).get('path')}")
            except Exception as e:
                log(f"❌ Failed to create file: {e}", "ERROR")
                return False
            
            # Step 6: Create a pull request
            pr_title = f"Test: GitHub MCP PR Workflow"
            
            pr_body = f"""# GitHub MCP PR Workflow Test

This pull request was created automatically by the GitHub MCP Server PR workflow test.

## Changes

- Created branch `{branch_name}` from `{default_branch}`
- Added test file at `{file_path}`

## Timestamp

Generated at: {timestamp_str}
"""
            
            log(f"Step 6: Creating pull request from {branch_name} to {default_branch}...")
            
            try:
                pr_result = client.create_pull_request(
                    owner, repo, pr_title, 
                    branch_name, default_branch, pr_body,
                    draft=True  # Create as draft to avoid accidental merges
                )
                
                pr_number = pr_result.get('number')
                pr_url = pr_result.get('html_url')
                
                log(f"✅ Created PR #{pr_number}: {pr_title}")
                log(f"✅ PR URL: {pr_url}")
                
                log("🎉 PR workflow completed successfully!")
                return True
                
            except Exception as e:
                log(f"❌ Failed to create pull request: {e}", "ERROR")
                return False
                
        except Exception as e:
            log(f"❌ Error during workflow: {e}", "ERROR")
            return False

if __name__ == "__main__":
    if not args.owner or not args.repo:
        log("Repository owner and name are required.", "ERROR")
        log("Please provide them with --owner and --repo arguments.", "ERROR")
        sys.exit(1)
    
    success = run_pr_workflow(args.owner, args.repo)
    if not success:
        sys.exit(1)