#!/bin/bash
# Script to run GitHub MCP Server tests using a token file

# Path to token file
TOKEN_FILE="$HOME/.github_token"

# Function to print messages with timestamp
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Create token file if it doesn't exist
if [ ! -f "$TOKEN_FILE" ]; then
    read -p "Enter your GitHub token: " github_token
    echo "$github_token" > "$TOKEN_FILE"
    chmod 600 "$TOKEN_FILE"
    log "Token saved to $TOKEN_FILE"
fi

# Read token from file
GITHUB_TOKEN=$(cat "$TOKEN_FILE")
if [ -z "$GITHUB_TOKEN" ]; then
    log "ERROR: Token file is empty."
    exit 1
fi

# Export token as environment variable for the tests
export GITHUB_PERSONAL_ACCESS_TOKEN="$GITHUB_TOKEN"

# Check if GitHub MCP Server binary exists
if [ ! -f "./github-mcp-server" ]; then
    log "ERROR: GitHub MCP Server binary not found."
    log "Please build it first or make sure it exists in the current directory."
    exit 1
fi

# Make sure the binary is executable
chmod +x ./github-mcp-server

# Function to run a test
run_test() {
    log "Running test: $1"
    python3 $1 "${@:2}"
    if [ $? -eq 0 ]; then
        log "✅ Test passed: $1"
    else
        log "❌ Test failed: $1"
        exit 1
    fi
}

# Parse arguments
OWNER=""
REPO=""
TEST=""

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
        --test)
            TEST="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Check if a specific test was requested
if [ -n "$TEST" ]; then
    case $TEST in
        simple)
            run_test simple_test.py
            ;;
        comprehensive)
            if [ -z "$OWNER" ] || [ -z "$REPO" ]; then
                log "ERROR: Owner and repo are required for comprehensive test."
                exit 1
            fi
            run_test comprehensive_test.py --owner "$OWNER" --repo "$REPO" --verbose
            ;;
        pr)
            if [ -z "$OWNER" ] || [ -z "$REPO" ]; then
                log "ERROR: Owner and repo are required for PR workflow test."
                exit 1
            fi
            run_test pr_workflow_test.py --owner "$OWNER" --repo "$REPO" --verbose
            ;;
        *)
            log "ERROR: Unknown test: $TEST"
            log "Available tests: simple, comprehensive, pr"
            exit 1
            ;;
    esac
    exit 0
fi

# If no specific test was requested, run the simple test
if [ -z "$OWNER" ] || [ -z "$REPO" ]; then
    log "No owner/repo provided. Running simple authentication test only."
    run_test simple_test.py
else
    # Run all tests
    log "Running all tests with owner=$OWNER, repo=$REPO"
    
    # Run simple authentication test
    run_test simple_test.py
    
    # Run comprehensive test
    run_test comprehensive_test.py --owner "$OWNER" --repo "$REPO" --verbose
    
    # Ask if user wants to run PR workflow test
    read -p "Do you want to run the PR workflow test? (y/n): " answer
    if [[ "$answer" == "y" ]]; then
        run_test pr_workflow_test.py --owner "$OWNER" --repo "$REPO" --verbose
    fi
    
    log "🎉 All tests completed successfully!"
fi