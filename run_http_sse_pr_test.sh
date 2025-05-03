#!/bin/bash
# Run PR workflow test using HTTP SSE transport

# Default values
PORT=7445
OWNER=""
REPO=""
VERBOSE=""

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --port)
      PORT="$2"
      shift 2
      ;;
    --owner)
      OWNER="$2"
      shift 2
      ;;
    --repo)
      REPO="$2"
      shift 2
      ;;
    --verbose)
      VERBOSE="--verbose"
      shift
      ;;
    *)
      echo "Unknown argument: $1"
      echo "Usage: $0 [--port PORT] [--owner OWNER] [--repo REPO] [--verbose]"
      exit 1
      ;;
  esac
done

# Check if owner and repo are provided
if [[ -z "$OWNER" || -z "$REPO" ]]; then
  echo "ERROR: Owner and repo are required."
  echo "Usage: $0 --owner OWNER --repo REPO [--port PORT] [--verbose]"
  exit 1
fi

# Activate virtual environment if it exists
if [[ -d venv/bin ]]; then
  source venv/bin/activate
  echo "Activated virtual environment."
fi

# Check if HTTP SSE server is running
if ! curl -s http://localhost:${PORT}/health > /dev/null; then
  echo "ERROR: HTTP SSE server is not running on port ${PORT}."
  echo "Please start the server with:"
  echo "./start_http_sse_server.sh --port ${PORT}"
  exit 1
fi

echo "HTTP SSE server is running on port ${PORT}. Starting PR workflow test..."

# Run the PR workflow test
python pr_workflow_http_sse.py --owner "${OWNER}" --repo "${REPO}" --port "${PORT}" ${VERBOSE}

# Store the exit code
EXIT_CODE=$?

# Deactivate virtual environment if it was activated
if [[ -d venv/bin ]]; then
  deactivate
fi

exit ${EXIT_CODE}