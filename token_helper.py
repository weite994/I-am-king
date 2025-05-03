#!/usr/bin/env python3
"""
Token Helper Module for GitHub MCP Server Tests.

This module provides functions to get and set GitHub tokens from a token file.
"""

import os
import sys
from pathlib import Path

# Default token file path
TOKEN_FILE = os.path.expanduser("~/.github_token")

def get_github_token():
    """
    Get GitHub token from token file or environment variable.
    
    Returns:
        str: GitHub token
    """
    # Try token file first
    if os.path.exists(TOKEN_FILE):
        with open(TOKEN_FILE, "r") as f:
            token = f.read().strip()
            if token:
                return token
    
    # Try environment variable as fallback
    token = os.environ.get("GITHUB_PERSONAL_ACCESS_TOKEN")
    if token:
        return token
    
    # No token found
    print("ERROR: GitHub token not found.")
    print(f"Please create a token file at {TOKEN_FILE} or set GITHUB_PERSONAL_ACCESS_TOKEN environment variable.")
    return None

def create_token_file(token):
    """
    Create or update the GitHub token file.
    
    Args:
        token (str): GitHub token
    """
    with open(TOKEN_FILE, "w") as f:
        f.write(token)
    os.chmod(TOKEN_FILE, 0o600)  # Set permissions to owner read/write only
    print(f"Token saved to {TOKEN_FILE}")

def ensure_token_exists():
    """
    Ensure a GitHub token exists, prompting the user if necessary.
    
    Returns:
        str: GitHub token
    """
    token = get_github_token()
    if not token:
        token = input("Enter your GitHub token: ").strip()
        if token:
            create_token_file(token)
        else:
            print("No token provided. Exiting.")
            sys.exit(1)
    return token