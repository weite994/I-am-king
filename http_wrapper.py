#!/usr/bin/env python3
"""
HTTP Wrapper for GitHub MCP Server.

This script creates a simple HTTP server that forwards requests to the GitHub MCP Server
using stdio transport and returns the responses.

Usage:
    python3 http_wrapper.py [--port PORT] [--verbose]
"""

import argparse
import json
import os
import subprocess
import sys
import threading
import time
from http.server import HTTPServer, BaseHTTPRequestHandler
from token_helper import get_github_token

# Process to communicate with
mcp_process = None
token = None
verbose = False

def log(message, level="INFO"):
    """Log a message with timestamp."""
    timestamp = time.strftime("%Y-%m-%d %H:%M:%S", time.localtime())
    print(f"[{timestamp}] {level}: {message}")

def debug(message):
    """Log a debug message if verbose is enabled."""
    if verbose:
        log(message, level="DEBUG")

class MCPRequestHandler(BaseHTTPRequestHandler):
    """HTTP Request Handler for GitHub MCP Server requests."""
    
    def do_GET(self):
        """Handle GET requests."""
        if self.path == "/health":
            self.send_response(200)
            self.send_header("Content-type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps({"status": "ok"}).encode())
        else:
            self.send_response(404)
            self.send_header("Content-type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps({"error": "Not found"}).encode())
    
    def do_POST(self):
        """Handle POST requests."""
        global mcp_process
        
        if self.path == "/sse":
            # Get request body length
            content_length = int(self.headers["Content-Length"])
            
            # Read request body
            request_body = self.rfile.read(content_length).decode()
            debug(f"Received request: {request_body}")
            
            try:
                # Parse request body
                request = json.loads(request_body)
                
                # Check if process is still running
                if mcp_process.poll() is not None:
                    log("MCP process exited unexpectedly. Restarting...", "WARNING")
                    start_mcp_server()
                
                # Add newline to request
                request_str = json.dumps(request) + "\n"
                
                # Send request to MCP process
                debug("Sending request to MCP process")
                mcp_process.stdin.write(request_str)
                mcp_process.stdin.flush()
                
                # Read response
                debug("Reading response from MCP process")
                response_str = mcp_process.stdout.readline()
                
                if not response_str:
                    log("No response received from MCP process", "ERROR")
                    self.send_response(500)
                    self.send_header("Content-type", "application/json")
                    self.end_headers()
                    self.wfile.write(json.dumps({"error": "No response from MCP server"}).encode())
                    return
                
                debug(f"Received response: {response_str}")
                
                # Parse response
                response = json.loads(response_str)
                
                # Send response as SSE
                self.send_response(200)
                self.send_header("Content-type", "text/event-stream")
                self.send_header("Cache-Control", "no-cache")
                self.send_header("Connection", "keep-alive")
                self.end_headers()
                
                # Send data event
                event = f"event: data\ndata: {json.dumps(response)}\n\n"
                self.wfile.write(event.encode())
                
            except json.JSONDecodeError as e:
                log(f"JSON decode error: {e}", "ERROR")
                self.send_response(400)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({"error": "Invalid JSON"}).encode())
            except Exception as e:
                log(f"Error processing request: {e}", "ERROR")
                self.send_response(500)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(json.dumps({"error": str(e)}).encode())
        else:
            self.send_response(404)
            self.send_header("Content-type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps({"error": "Not found"}).encode())
    
    def log_message(self, format, *args):
        """Override log_message to use our custom logger."""
        if verbose:
            log(f"{self.address_string()} - {format % args}", "HTTP")

def start_mcp_server():
    """Start the GitHub MCP Server process."""
    global mcp_process, token
    
    log("Starting GitHub MCP Server process")
    
    # Environment variables
    env = os.environ.copy()
    env["GITHUB_PERSONAL_ACCESS_TOKEN"] = token
    
    # Start process
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
        log("Failed to start GitHub MCP Server process", "ERROR")
        return False
    
    # Start a thread to read stderr
    def read_stderr():
        while mcp_process.poll() is None:
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

def main():
    """Main function."""
    global token, verbose
    
    parser = argparse.ArgumentParser(description="HTTP Wrapper for GitHub MCP Server")
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
        httpd = HTTPServer(server_address, MCPRequestHandler)
        
        log(f"Starting HTTP server on port {args.port}")
        log("Press Ctrl+C to stop")
        
        # Start HTTP server
        httpd.serve_forever()
        
    except KeyboardInterrupt:
        log("Keyboard interrupt received, shutting down")
    finally:
        # Stop servers
        log("Stopping HTTP server")
        httpd.shutdown()
        
        stop_mcp_server()

if __name__ == "__main__":
    main()