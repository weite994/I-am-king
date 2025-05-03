#!/bin/bash
# Simple script to test GitHub MCP Server PR workflow with fixed tool names

# Function to print messages with timestamp
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Make sure scripts are executable
chmod +x token_helper.py pr_workflow_fixed.py

# Make sure the GitHub MCP Server binary is executable
chmod +x ./github-mcp-server

# Simple help function
show_help() {
    echo "Usage: $0 --owner OWNER --repo REPO"
    echo ""
    echo "This script tests the GitHub MCP Server by creating a pull request."
    echo "It uses the correct tool names discovered from tools/list."
    echo ""
    echo "Arguments:"
    echo "  --owner     GitHub repository owner (username)"
    echo "  --repo      GitHub repository name"
    echo ""
    echo "Example:"
    echo "  $0 --owner octocat --repo hello-world"
    exit 1
}

# Parse arguments
OWNER=""
REPO=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --owner)
            OWNER="$2"
            shift 2
            ;;
        --repo)
            REPO="$2"
            shift 2
            ;;
        --help|-h)
            show_help
            ;;
        *)
            echo "Unknown option: $1"
            show_help
            ;;
    esac
done

# Check if owner and repo are provided
if [ -z "$OWNER" ] || [ -z "$REPO" ]; then
    log "ERROR: Owner and repo are required."
    show_help
fi

# Run the PR workflow test
log "Running PR workflow test with fixed tool names..."
python3 pr_workflow_fixed.py --owner "$OWNER" --repo "$REPO" --verbose
if [ $? -eq 0 ]; then
    log "🎉 PR workflow test completed successfully!"
else
    log "❌ PR workflow test failed."
    exit 1
fi