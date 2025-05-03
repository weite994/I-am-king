#!/bin/bash
# Start HTTP and SSE wrapper for GitHub MCP Server

# Default port
PORT=7444

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --port)
      PORT="$2"
      shift 2
      ;;
    --verbose)
      VERBOSE="--verbose"
      shift
      ;;
    *)
      echo "Unknown argument: $1"
      echo "Usage: $0 [--port PORT] [--verbose]"
      exit 1
      ;;
  esac
done

# Activate virtual environment if it exists
if [[ -d venv/bin ]]; then
  source venv/bin/activate
  echo "Activated virtual environment."
fi

# Install required packages if missing
pip show sseclient-py >/dev/null 2>&1 || pip install sseclient-py

# Get GitHub token
if [[ -f ~/.github_token ]]; then
  export GITHUB_PERSONAL_ACCESS_TOKEN=$(cat ~/.github_token)
elif [[ -z "${GITHUB_PERSONAL_ACCESS_TOKEN}" ]]; then
  echo "ERROR: GitHub token not found."
  echo "Please create a token file at ~/.github_token or set GITHUB_PERSONAL_ACCESS_TOKEN environment variable."
  exit 1
fi

echo "Starting HTTP and SSE wrapper for GitHub MCP Server on port ${PORT}..."
echo "Press Ctrl+C to stop the server."

# Save PID to file for easy termination
echo "$$" > http_mcp.pid

# Redirect logs
LOG_FILE="http_mcp.log"
echo "Logging to ${LOG_FILE}"
echo "Starting HTTP SSE server at $(date)" > ${LOG_FILE}

# Run the server
python http_mcp_wrapper.py --port ${PORT} ${VERBOSE:+--verbose} | tee -a ${LOG_FILE}

# Clean up
rm -f http_mcp.pid
echo "Server stopped."

# Deactivate virtual environment if it was activated
if [[ -d venv/bin ]]; then
  deactivate
fi