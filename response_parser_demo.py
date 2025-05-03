#!/usr/bin/env python3
"""
Response Parser Demo for GitHub MCP Server Binary Implementation.

This script demonstrates the complex response parsing required for
the GitHub MCP Server Binary API responses.

IMPORTANT NOTE: This demo specifically addresses the response format of
the GitHub MCP Server Binary implementation. Other implementations might
have different response formats.

Usage:
    python3 response_parser_demo.py
"""

import json
import sys

# Import token helper
from token_helper import ensure_token_exists

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
    print("\n=== RESPONSE PARSING DEMO ===\n")
    
    print(f"Original response: {json.dumps(response, indent=2)}")
    
    if "error" in response:
        error = response["error"]
        error_message = error.get("message", "Unknown error")
        error_code = error.get("code", -1)
        print(f"\nERROR detected: {error_message} (code {error_code})")
        return None
    
    if "result" not in response:
        print("\nNo 'result' field found in response")
        return {}
    
    result = response["result"]
    print(f"\nExtracted 'result': {json.dumps(result, indent=2)}")
    
    # Check if result contains content field (the new format)
    if "content" in result and isinstance(result["content"], list):
        print("\nDetected 'content' array in result")
        
        for i, item in enumerate(result["content"]):
            print(f"\nProcessing content[{i}]:")
            
            if item.get("type") == "text":
                text = item.get("text", "")
                print(f"Found text content: {text[:100]}...")
                
                # Try to parse the text as JSON
                try:
                    json_data = json.loads(text)
                    print(f"Successfully parsed as JSON: {type(json_data).__name__}")
                    print(f"Parsed data: {json.dumps(json_data, indent=2)[:200]}...")
                    return json_data
                except json.JSONDecodeError as e:
                    print(f"Not valid JSON: {e}")
                    # If it's not valid JSON, return the text as is
                    return text
            else:
                print(f"Content has type '{item.get('type')}', not 'text'")
    else:
        print("\nNo 'content' array found or not an array, returning result directly")
    
    # If no content field or parsing failed, return the result as is
    return result

def main():
    # Example 1: Simple response with direct result
    simple_response = {
        "jsonrpc": "2.0",
        "id": "1",
        "result": {
            "login": "octocat",
            "id": 1,
            "name": "The Octocat"
        }
    }
    
    # Example 2: Complex response with nested content
    complex_response = {
        "jsonrpc": "2.0",
        "id": "2",
        "result": {
            "content": [
                {
                    "type": "text",
                    "text": "{\"items\": [{\"name\": \"main\", \"commit\": {\"sha\": \"abcdef1234567890\"}}]}"
                }
            ]
        }
    }
    
    # Example 3: Error response
    error_response = {
        "jsonrpc": "2.0",
        "id": "3",
        "error": {
            "code": 401,
            "message": "Bad credentials"
        }
    }
    
    # Demonstrate parsing each response type
    print("\n=== EXAMPLE 1: SIMPLE RESPONSE ===")
    simple_result = parse_response(simple_response)
    print(f"\nFinal parsed result (type: {type(simple_result).__name__}):")
    print(json.dumps(simple_result, indent=2))
    
    print("\n=== EXAMPLE 2: COMPLEX NESTED RESPONSE ===")
    complex_result = parse_response(complex_response)
    print(f"\nFinal parsed result (type: {type(complex_result).__name__}):")
    print(json.dumps(complex_result, indent=2))
    
    print("\n=== EXAMPLE 3: ERROR RESPONSE ===")
    error_result = parse_response(error_response)
    print(f"\nFinal parsed result: {error_result}")
    
    # Example for handling different response structures for different tools
    print("\n=== RESPONSE TYPE VARIATIONS ===\n")
    
    # Example for list_branches (direct array)
    branches_list_response = {
        "jsonrpc": "2.0",
        "id": "4",
        "result": {
            "content": [
                {
                    "type": "text",
                    "text": "[{\"name\": \"main\", \"commit\": {\"sha\": \"abc123\"}}, {\"name\": \"dev\", \"commit\": {\"sha\": \"def456\"}}]"
                }
            ]
        }
    }
    
    # Example for search_repositories (dict with items array)
    search_response = {
        "jsonrpc": "2.0",
        "id": "5",
        "result": {
            "content": [
                {
                    "type": "text",
                    "text": "{\"total_count\": 2, \"items\": [{\"name\": \"repo1\", \"full_name\": \"user/repo1\"}, {\"name\": \"repo2\", \"full_name\": \"user/repo2\"}]}"
                }
            ]
        }
    }
    
    print("Parsing branches (list response):")
    branches_result = parse_response(branches_list_response)
    
    # Handle different response formats for branches
    branches = []
    if isinstance(branches_result, list):
        branches = branches_result
        print("Direct list response detected")
    elif isinstance(branches_result, dict) and "items" in branches_result:
        branches = branches_result.get("items", [])
        print("Dictionary with items array detected")
    
    print(f"Extracted {len(branches)} branches: {json.dumps(branches, indent=2)}")
    
    print("\nParsing repository search (dict with items):")
    search_result = parse_response(search_response)
    
    # Handle different response formats for search
    repos = []
    if isinstance(search_result, list):
        repos = search_result
        print("Direct list response detected")
    elif isinstance(search_result, dict) and "items" in search_result:
        repos = search_result.get("items", [])
        print("Dictionary with items array detected")
    
    print(f"Extracted {len(repos)} repositories: {json.dumps(repos, indent=2)}")

if __name__ == "__main__":
    main()