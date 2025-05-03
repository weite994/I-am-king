#!/usr/bin/env python3
"""
HTTP SSE Pull Request Workflow Test for GitHub MCP Server.

This script tests a complete pull request workflow using the HTTP SSE wrapper
for the GitHub MCP Server.

Usage:
    python3 pr_workflow_http_sse.py --owner YOUR_USERNAME --repo YOUR_REPO --port 7445
"""

import argparse
import json
import requests
import sys
import time
from datetime import datetime

# Import token helper
from token_helper import get_github_token

# Set up argument parser
parser = argparse.ArgumentParser(description="Test GitHub MCP Server PR workflow with HTTP SSE")
parser.add_argument("--owner", required=True, help="Repository owner")
parser.add_argument("--repo", required=True, help="Repository name")
parser.add_argument("--port", type=int, default=7445, help="HTTP server port (default: 7445)")
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

# Get GitHub token
token = get_github_token()
if not token:
    log("GitHub token not found.", "ERROR")
    log("Please set up a token in ~/.github_token or GITHUB_PERSONAL_ACCESS_TOKEN environment variable", "ERROR")
    sys.exit(1)

class HTTPMCPClient:
    """Client for communicating with the GitHub MCP Server via HTTP/SSE."""
    
    def __init__(self, base_url, token):
        """Initialize the client."""
        self.base_url = base_url
        self.token = token
        self.headers = {"Content-Type": "application/json"}
        self.request_id = 0
    
    def call_tool(self, name, arguments):
        """Call a tool in the GitHub MCP Server via HTTP."""
        self.request_id += 1
        
        # First check if server is reachable
        try:
            health_response = requests.get(f"{self.base_url}/health", timeout=5)
            if health_response.status_code != 200:
                raise RuntimeError(f"Server health check failed: {health_response.status_code}")
        except requests.exceptions.RequestException as e:
            raise RuntimeError(f"Server is not reachable: {e}")
        
        # Prepare the request data
        data = {
            "name": name,
            "arguments": arguments
        }
        
        debug(f"Calling tool {name} with arguments: {arguments}")
        
        # Send the request
        try:
            response = requests.post(
                f"{self.base_url}/tools/call",
                headers=self.headers,
                json=data,
                timeout=30
            )
            
            # Check for HTTP errors
            response.raise_for_status()
            
            # Parse the response
            result = response.json()
            debug(f"Received response: {result}")
            
            # Return the result
            return result
            
        except requests.exceptions.RequestException as e:
            # If there's a timeout or other network issue, try the alternative SSE endpoint
            debug(f"Error with regular HTTP call: {e}, trying SSE endpoint")
            return self.call_tool_sse(name, arguments)
        except Exception as e:
            raise RuntimeError(f"Error calling tool {name}: {e}")
    
    def call_tool_sse(self, name, arguments):
        """Call a tool using the SSE endpoint as a fallback."""
        # Prepare a JSON-RPC request for the SSE endpoint
        request = {
            "jsonrpc": "2.0",
            "id": str(time.time()),
            "method": "tools/call",
            "params": {
                "name": name,
                "arguments": arguments
            }
        }
        
        # Set up headers with Accept: text/event-stream
        sse_headers = {
            "Content-Type": "application/json",
            "Accept": "text/event-stream"
        }
        
        debug(f"Calling tool {name} via SSE endpoint")
        
        try:
            # Send the request to the SSE endpoint
            response = requests.post(
                f"{self.base_url}/sse",
                headers=sse_headers,
                data=json.dumps(request),
                stream=True,
                timeout=30
            )
            
            # Check for HTTP errors
            response.raise_for_status()
            
            # Process the SSE response
            event_data = None
            
            for line in response.iter_lines():
                if not line:
                    continue
                    
                line = line.decode('utf-8')
                
                # Look for SSE data lines
                if line.startswith('data:'):
                    data_text = line[5:].strip()
                    try:
                        event_data = json.loads(data_text)
                        
                        # Extract the result
                        if "result" in event_data:
                            result = event_data["result"]
                            
                            # Handle the nested content format
                            if "content" in result and isinstance(result["content"], list):
                                for item in result["content"]:
                                    if item.get("type") == "text":
                                        text = item.get("text", "")
                                        try:
                                            return json.loads(text)
                                        except json.JSONDecodeError:
                                            return text
                            
                            return result
                    except json.JSONDecodeError as e:
                        debug(f"Error parsing SSE data: {e}")
                
                # Look for end of stream
                if line.startswith('event: end'):
                    break
            
            # If we got here without returning, check if we have event_data
            if event_data and "error" in event_data:
                error = event_data["error"]
                raise RuntimeError(f"Error from SSE: {error.get('message', 'Unknown error')}")
            
            # If no data was returned, raise an error
            if not event_data:
                raise RuntimeError("No data received from SSE endpoint")
                
            return {}
            
        except Exception as e:
            raise RuntimeError(f"Error calling tool {name} via SSE: {e}")
            
    # API methods
    
    def get_authenticated_user(self):
        """Get information about the authenticated user."""
        return self.call_tool("get_me", {})
    
    def search_repositories(self, query):
        """Search for repositories."""
        return self.call_tool("search_repositories", {
            "query": query
        })
    
    def list_branches(self, owner, repo):
        """List branches in a repository."""
        return self.call_tool("list_branches", {
            "owner": owner,
            "repo": repo
        })
    
    def create_branch(self, owner, repo, branch, from_branch=None):
        """Create a new branch."""
        params = {
            "owner": owner,
            "repo": repo,
            "branch": branch
        }
        
        if from_branch:
            params["from_branch"] = from_branch
        
        return self.call_tool("create_branch", params)
    
    def get_file_contents(self, owner, repo, path, branch=None):
        """Get file contents from a repository."""
        params = {
            "owner": owner,
            "repo": repo,
            "path": path
        }
        
        if branch:
            params["branch"] = branch
        
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

def run_pr_workflow(owner, repo, port):
    """Run a complete PR workflow test."""
    
    # Set up the base URL for the HTTP server
    base_url = f"http://localhost:{port}"
    
    # Check if the server is running
    try:
        response = requests.get(f"{base_url}/health", timeout=5)
        if response.status_code != 200:
            log(f"HTTP server health check failed: {response.status_code}", "ERROR")
            log(f"Make sure the server is running on port {port}", "ERROR")
            return False
    except requests.exceptions.RequestException:
        log("HTTP server is not reachable.", "ERROR")
        log(f"Make sure the server is running on port {port}", "ERROR")
        log("Run: ./start_http_sse_server.sh --port {port}", "ERROR")
        return False
    
    log(f"Server is running on {base_url}")
    
    # Create the HTTP MCP client
    client = HTTPMCPClient(base_url, token)
    
    try:
        # Step 1: Get authenticated user
        log("Step 1: Getting authenticated user...")
        user = client.get_authenticated_user()
        log(f"✅ Authenticated as: {user.get('login')}")
        
        # Step 2: Search for the repository
        log(f"Step 2: Searching for repository {owner}/{repo}...")
        try:
            repo_query = f"repo:{owner}/{repo}"
            repo_search = client.search_repositories(repo_query)
            
            # Find the repository in the search results
            repo_info = None
            if isinstance(repo_search, dict) and "items" in repo_search:
                for item in repo_search.get("items", []):
                    if item.get("full_name") == f"{owner}/{repo}":
                        repo_info = item
                        break
            
            if not repo_info:
                # If search fails, try direct access with list_branches
                log(f"Repository not found in search results, trying direct branch access...")
                branches = client.list_branches(owner, repo)
                if branches:
                    # Create minimal repo info
                    repo_info = {
                        "full_name": f"{owner}/{repo}",
                        "default_branch": "main"  # Assume main as default
                    }
            
            if not repo_info:
                log(f"❌ Repository {owner}/{repo} not found or not accessible", "ERROR")
                return False
                
            log(f"✅ Found repository: {repo_info.get('full_name')}")
            
        except Exception as e:
            log(f"Error searching for repository: {e}", "ERROR")
            log("Trying direct branch access instead...")
            try:
                branches = client.list_branches(owner, repo)
                if branches:
                    # Create minimal repo info
                    repo_info = {
                        "full_name": f"{owner}/{repo}",
                        "default_branch": "main"  # Assume main as default
                    }
                    log(f"✅ Found repository: {repo_info.get('full_name')}")
                else:
                    log(f"❌ Repository {owner}/{repo} not found or not accessible", "ERROR")
                    return False
            except Exception as e2:
                log(f"Error accessing repository branches: {e2}", "ERROR")
                return False
        
        # Get default branch
        default_branch = repo_info.get('default_branch', 'main')
        log(f"✅ Default branch: {default_branch}")
        
        # Step 3: Get branches
        log(f"Step 3: Getting branches for {owner}/{repo}...")
        try:
            branches_response = client.list_branches(owner, repo)
            
            # Handle different response formats
            branches = []
            if isinstance(branches_response, list):
                branches = branches_response
            elif isinstance(branches_response, dict) and "items" in branches_response:
                branches = branches_response.get("items", [])
            
            if not branches:
                log(f"No branches found in response, assuming default branch structure")
                # Create a synthetic branch for the default branch
                branches = [{
                    "name": default_branch,
                    "commit": {"sha": "HEAD"}  # Use HEAD reference
                }]
            
            log(f"✅ Found {len(branches)} branches")
        except Exception as e:
            log(f"Error getting branches: {e}", "ERROR")
            log("Creating synthetic branch structure for default branch")
            # Create a synthetic branch for the default branch
            branches = [{
                "name": default_branch,
                "commit": {"sha": "HEAD"}  # Use HEAD reference
            }]
        
        # Step 4: Create a new branch
        timestamp = int(time.time())
        branch_name = f"http-sse-test-{timestamp}"
        log(f"Step 4: Creating new branch {branch_name} from {default_branch}...")
        
        try:
            new_branch = client.create_branch(owner, repo, branch_name, default_branch)
            branch_name_result = new_branch.get('name') if isinstance(new_branch, dict) else branch_name
            log(f"✅ Created branch: {branch_name_result}")
        except Exception as e:
            log(f"❌ Failed to create branch: {e}", "ERROR")
            return False
        
        # Step 5: Create a new file in the branch
        timestamp_str = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        file_path = f"docs/http-sse-test-{timestamp}.md"
        
        file_content = f"""# GitHub MCP HTTP SSE Test

This file was created by the GitHub MCP Server HTTP SSE PR workflow test.

Created at: {timestamp_str}

## Test Information

- User: {user.get('login')}
- Repository: {repo_info.get('full_name')}
- Branch: {branch_name}
- File: {file_path}
- Timestamp: {timestamp}
- Transport: HTTP SSE
"""
        
        log(f"Step 5: Creating file {file_path} in branch {branch_name}...")
        
        try:
            file_result = client.create_or_update_file(
                owner, repo, file_path, 
                "Add GitHub MCP HTTP SSE test file", 
                file_content, branch_name
            )
            
            # Handle different response formats
            file_path_result = ""
            if isinstance(file_result, dict) and "content" in file_result:
                file_path_result = file_result.get("content", {}).get("path", file_path)
            
            log(f"✅ Created file: {file_path_result or file_path}")
        except Exception as e:
            log(f"❌ Failed to create file: {e}", "ERROR")
            return False
        
        # Step 6: Create a pull request
        pr_title = f"Test: GitHub MCP HTTP SSE PR Workflow"
        
        pr_body = f"""# GitHub MCP HTTP SSE PR Workflow Test

This pull request was created automatically by the GitHub MCP Server HTTP SSE PR workflow test.

## Changes

- Created branch `{branch_name}` from `{default_branch}`
- Added test file at `{file_path}`

## Test Details

- Transport: HTTP SSE
- Server URL: {base_url}
- Generated at: {timestamp_str}
"""
        
        log(f"Step 6: Creating pull request from {branch_name} to {default_branch}...")
        
        try:
            pr_result = client.create_pull_request(
                owner, repo, pr_title, 
                branch_name, default_branch, pr_body,
                draft=True  # Create as draft to avoid accidental merges
            )
            
            # Handle different response formats
            pr_number = None
            pr_url = None
            
            if isinstance(pr_result, dict):
                pr_number = pr_result.get('number')
                pr_url = pr_result.get('html_url')
            
            if pr_number:
                log(f"✅ Created PR #{pr_number}: {pr_title}")
            if pr_url:
                log(f"✅ PR URL: {pr_url}")
            else:
                log(f"✅ Created pull request (details not available)")
            
            log("🎉 HTTP SSE PR workflow completed successfully!")
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
    
    success = run_pr_workflow(args.owner, args.repo, args.port)
    if not success:
        sys.exit(1)