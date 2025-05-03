#!/usr/bin/env python3
"""
HTTP and SSE Wrapper for GitHub MCP Server.

This script creates an HTTP server that communicates with the GitHub MCP Server
using stdio transport and returns responses via HTTP or Server-Sent Events (SSE).
"""

import argparse
import http.server
import json
import os
import socketserver
import subprocess
import sys
import threading
import time
import uuid
from urllib.parse import parse_qs, urlparse
from token_helper import get_github_token

# Global process to communicate with
mcp_process = None
token = None
verbose = False

def log(message, level="INFO"):
    """Log a message with timestamp."""
    timestamp = time.strftime("%Y-%m-%d %H:%M:%S", time.localtime())
    print(f"[{timestamp}] {level}: {message}", flush=True)

def debug(message):
    """Log a debug message if verbose is enabled."""
    if verbose:
        log(message, level="DEBUG")

def start_mcp_server():
    """Start the GitHub MCP Server process."""
    global mcp_process, token
    
    log("Starting GitHub MCP Server process")
    
    # Environment variables
    env = os.environ.copy()
    env["GITHUB_PERSONAL_ACCESS_TOKEN"] = token
    
    # Start process
    try:
        mcp_process = subprocess.Popen(
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
        if mcp_process.poll() is not None:
            stderr_output = mcp_process.stderr.read()
            log(f"Failed to start GitHub MCP Server process: {stderr_output}", "ERROR")
            return False
            
        log(f"Process started with PID: {mcp_process.pid}")
    except Exception as e:
        log(f"Error starting MCP process: {e}", "ERROR")
        return False
    
    # Start a thread to read stderr
    def read_stderr():
        while mcp_process and mcp_process.poll() is None:
            line = mcp_process.stderr.readline()
            if line:
                log(f"MCP: {line.strip()}", "MCP")
    
    stderr_thread = threading.Thread(target=read_stderr)
    stderr_thread.daemon = True
    stderr_thread.start()
    
    log("GitHub MCP Server process started successfully")
    return True

def stop_mcp_server():
    """Stop the GitHub MCP Server process."""
    global mcp_process
    
    if mcp_process and mcp_process.poll() is None:
        log("Stopping GitHub MCP Server process")
        mcp_process.terminate()
        try:
            mcp_process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            log("Process did not terminate, killing", "WARNING")
            mcp_process.kill()
        log("GitHub MCP Server process stopped")

def call_mcp_tool(name, arguments):
    """Call a tool in the GitHub MCP Server."""
    global mcp_process
    
    if not mcp_process or mcp_process.poll() is not None:
        if not start_mcp_server():
            return {"error": "Failed to start MCP server"}
    
    # Create the request
    request = {
        "jsonrpc": "2.0",
        "id": str(time.time()),
        "method": "tools/call",
        "params": {
            "name": name,
            "arguments": arguments
        }
    }
    
    # Convert to JSON and add newline
    request_str = json.dumps(request) + "\n"
    debug(f"Sending request: {request_str}")
    
    try:
        # Send the request
        mcp_process.stdin.write(request_str)
        mcp_process.stdin.flush()
        
        # Read the response
        debug("Reading response...")
        response_str = mcp_process.stdout.readline()
        
        if not response_str:
            log("No response received", "ERROR")
            return {"error": "No response from MCP server"}
        
        debug(f"Received response: {response_str}")
        
        # Parse the response
        response = json.loads(response_str)
        
        # Check for errors
        if "error" in response:
            error = response["error"]
            error_message = error.get("message", "Unknown error")
            error_code = error.get("code", -1)
            return {"error": f"{error_message} (code {error_code})"}
        
        # Extract the result
        result = {}
        if "result" in response:
            result = response["result"]
            
            # Check if result contains content field
            if "content" in result and isinstance(result["content"], list):
                for item in result["content"]:
                    if item.get("type") == "text":
                        text = item.get("text", "")
                        
                        # Try to parse as JSON
                        try:
                            return json.loads(text)
                        except json.JSONDecodeError:
                            return {"text": text}
        
        return result
        
    except Exception as e:
        log(f"Error calling tool: {e}", "ERROR")
        return {"error": str(e)}

class MCPRequestHandler(http.server.SimpleHTTPRequestHandler):
    """HTTP Request Handler for GitHub MCP Server requests."""
    
    def log_message(self, format, *args):
        """Override log_message to use our custom logger."""
        if verbose:
            log(f"{self.address_string()} - {format % args}", "HTTP")
    
    def do_GET(self):
        """Handle GET requests."""
        # Parse URL and query parameters
        url = urlparse(self.path)
        path = url.path
        query = parse_qs(url.query)
        
        # Health check endpoint
        if path == "/health":
            self.send_response(200)
            self.send_header("Content-type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps({"status": "ok"}).encode())
            return
        
        # Get authenticated user
        if path == "/user":
            result = call_mcp_tool("get_me", {})
            self.send_response(200)
            self.send_header("Content-type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps(result).encode())
            return
        
        # Search repositories
        if path == "/search/repositories":
            q = query.get("q", ["language:go stars:>1000"])[0]
            result = call_mcp_tool("search_repositories", {"query": q})
            self.send_response(200)
            self.send_header("Content-type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps(result).encode())
            return
        
        # List tools
        if path == "/tools":
            # Create a request to list all available tools
            request = {
                "jsonrpc": "2.0",
                "id": str(time.time()),
                "method": "tools/list",
                "params": {}
            }
            
            # Convert to JSON and add newline
            request_str = json.dumps(request) + "\n"
            debug(f"Sending tools/list request: {request_str}")
            
            try:
                # Send the request
                mcp_process.stdin.write(request_str)
                mcp_process.stdin.flush()
                
                # Read the response
                response_str = mcp_process.stdout.readline()
                
                if not response_str:
                    self.send_response(500)
                    self.send_header("Content-type", "application/json")
                    self.end_headers()
                    self.wfile.write(json.dumps({"error": "No response from MCP server"}).encode())
                    return
                
                # Parse the response
                response = json.loads(response_str)
                
                # Extract tools
                result = {}
                if "result" in response:
                    result = response["result"]
                
                self.send_response(200)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps(result).encode())
                return
                
            except Exception as e:
                self.send_response(500)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({"error": str(e)}).encode())
                return
        
        # Default response for other paths
        self.send_response(404)
        self.send_header("Content-type", "application/json")
        self.end_headers()
        self.wfile.write(json.dumps({"error": "Not found"}).encode())
    
    def do_POST(self):
        """Handle POST requests."""
        # Parse URL
        url = urlparse(self.path)
        path = url.path
        
        # Read request body
        content_length = int(self.headers.get("Content-Length", 0))
        post_data = self.rfile.read(content_length).decode("utf-8")
        
        try:
            # Parse JSON data
            data = json.loads(post_data)
        except json.JSONDecodeError:
            self.send_response(400)
            self.send_header("Content-type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps({"error": "Invalid JSON"}).encode())
            return
        
        # SSE endpoint
        if path == "/sse":
            debug(f"SSE request received: {post_data}")
            
            # Extract request method and params
            if "jsonrpc" not in data or "method" not in data:
                self.send_response(400)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({"error": "Invalid JSON-RPC request"}).encode())
                return
            
            # Set up SSE response headers
            self.send_response(200)
            self.send_header("Content-type", "text/event-stream")
            self.send_header("Cache-Control", "no-cache")
            self.send_header("Connection", "keep-alive")
            self.send_header("X-Accel-Buffering", "no")  # For NGINX
            self.end_headers()
            
            # Process the request and stream the response
            self.handle_sse_request(data)
            return
        
        # Generic tool endpoint
        if path == "/tools/call":
            tool_name = data.get("name")
            tool_args = data.get("arguments", {})
            
            if not tool_name:
                self.send_response(400)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({"error": "Tool name is required"}).encode())
                return
            
            result = call_mcp_tool(tool_name, tool_args)
            self.send_response(200)
            self.send_header("Content-type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps(result).encode())
            return
        
        # Default response for other paths
        self.send_response(404)
        self.send_header("Content-type", "application/json")
        self.end_headers()
        self.wfile.write(json.dumps({"error": "Not found"}).encode())
        
    def handle_sse_request(self, request_data):
        """Handle an SSE request by streaming the response."""
        # Extract method and params
        method = request_data.get("method")
        params = request_data.get("params", {})
        request_id = request_data.get("id", str(uuid.uuid4()))
        
        debug(f"Processing SSE request - Method: {method}, ID: {request_id}")
        
        # Handle tools/list method
        if method == "tools/list":
            # Create the request
            request = {
                "jsonrpc": "2.0",
                "id": request_id,
                "method": "tools/list",
                "params": params
            }
            
            # Send the request and stream the response
            self.stream_sse_response(request)
            return
            
        # Handle tools/call method
        elif method == "tools/call":
            if "name" not in params:
                self.send_sse_error("Tool name is required", request_id)
                return
                
            # Create the request
            request = {
                "jsonrpc": "2.0",
                "id": request_id,
                "method": "tools/call",
                "params": params
            }
            
            # Send the request and stream the response
            self.stream_sse_response(request)
            return
            
        # Unsupported method
        else:
            self.send_sse_error(f"Unsupported method: {method}", request_id)
            return
    
    def stream_sse_response(self, request):
        """Stream a response as SSE events."""
        global mcp_process
        
        # Make sure the MCP process is running
        if not mcp_process or mcp_process.poll() is not None:
            if not start_mcp_server():
                self.send_sse_error("Failed to start MCP server", request.get("id", "unknown"))
                return
        
        try:
            # Convert to JSON and add newline
            request_str = json.dumps(request) + "\n"
            debug(f"Sending SSE request: {request_str}")
            
            # Send the request
            mcp_process.stdin.write(request_str)
            mcp_process.stdin.flush()
            
            # Read the response
            response_str = mcp_process.stdout.readline()
            
            if not response_str:
                self.send_sse_error("No response from MCP server", request.get("id", "unknown"))
                return
            
            # Parse the response
            debug(f"Received SSE response: {response_str}")
            response = json.loads(response_str)
            
            # Send the response as an SSE event
            event_data = json.dumps(response)
            self.wfile.write(f"event: data\n".encode())
            self.wfile.write(f"data: {event_data}\n\n".encode())
            self.wfile.flush()
            
            # End the stream
            self.wfile.write(f"event: end\n".encode())
            self.wfile.write(f"data: end\n\n".encode())
            self.wfile.flush()
            
        except Exception as e:
            error_message = str(e)
            log(f"Error streaming SSE response: {error_message}", "ERROR")
            self.send_sse_error(error_message, request.get("id", "unknown"))
    
    def send_sse_error(self, message, request_id):
        """Send an error as an SSE event."""
        error_data = json.dumps({
            "jsonrpc": "2.0",
            "id": request_id,
            "error": {
                "code": -32000,
                "message": message
            }
        })
        
        self.wfile.write(f"event: error\n".encode())
        self.wfile.write(f"data: {error_data}\n\n".encode())
        self.wfile.flush()

def main():
    """Main function."""
    global token, verbose
    
    parser = argparse.ArgumentParser(description="HTTP and SSE Wrapper for GitHub MCP Server")
    parser.add_argument("--port", type=int, default=7444, help="HTTP server port (default: 7444)")
    parser.add_argument("--verbose", action="store_true", help="Enable verbose output")
    args = parser.parse_args()
    
    verbose = args.verbose
    
    # Get GitHub token
    token = get_github_token()
    if not token:
        log("GitHub token not found", "ERROR")
        log("Please set up a token in ~/.github_token or GITHUB_PERSONAL_ACCESS_TOKEN environment variable")
        sys.exit(1)
    
    # Start MCP server
    if not start_mcp_server():
        sys.exit(1)
    
    try:
        # Create HTTP server
        server_address = ("", args.port)
        httpd = socketserver.ThreadingTCPServer(server_address, MCPRequestHandler)
        
        log(f"Starting HTTP and SSE server on port {args.port}")
        log("Available endpoints:")
        log("  GET /health - Health check")
        log("  GET /user - Get authenticated user")
        log("  GET /search/repositories?q=query - Search repositories")
        log("  GET /tools - List all available tools")
        log("  POST /tools/call - Call a specific tool")
        log("  POST /sse - Server-Sent Events endpoint for streaming responses")
        log("Press Ctrl+C to stop")
        
        # Start HTTP server
        httpd.serve_forever()
        
    except KeyboardInterrupt:
        log("Keyboard interrupt received, shutting down")
    except Exception as e:
        log(f"Error: {e}", "ERROR")
    finally:
        # Stop servers
        log("Stopping HTTP server")
        try:
            httpd.server_close()
        except:
            pass
        
        stop_mcp_server()

if __name__ == "__main__":
    main()