#!/bin/bash
set -euo pipefail

# Test script for capture-session-learnings.sh
# Usage: ./test-capture-learnings.sh <session_id> <transcript_path>
#
# Example:
#   ./test-capture-learnings.sh test-session-123 /path/to/transcript.md

if [ $# -lt 2 ]; then
  echo "Usage: $0 <session_id> <transcript_path>"
  echo ""
  echo "Example:"
  echo "  $0 test-session-123 /path/to/transcript.md"
  exit 1
fi

session_id="$1"
transcript_path="$2"
cwd="${3:-$(pwd)}"

# Get the script directory
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
capture_script="$script_dir/capture-session-learnings.sh"

if [ ! -f "$capture_script" ]; then
  echo "Error: capture-session-learnings.sh not found at: $capture_script"
  exit 1
fi

if [ ! -f "$transcript_path" ]; then
  echo "Warning: Transcript file does not exist: $transcript_path"
  read -p "Continue anyway? (y/n) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
  fi
fi

# Export CLAUDE_PROJECT_DIR if not set
if [ -z "${CLAUDE_PROJECT_DIR:-}" ]; then
  export CLAUDE_PROJECT_DIR="$(cd "$script_dir/../../.." && pwd)"
  echo "Setting CLAUDE_PROJECT_DIR=$CLAUDE_PROJECT_DIR"
fi

# Build the JSON input that the hook expects
# Set use_prompt_mode to false for easier testing (allows interactive mode)
hook_input=$(jq -n \
  --arg session_id "$session_id" \
  --arg transcript_path "$transcript_path" \
  --arg cwd "$cwd" \
  --argjson use_prompt_mode false \
  '{
    session_id: $session_id,
    transcript_path: $transcript_path,
    cwd: $cwd,
    use_prompt_mode: $use_prompt_mode
  }')

echo "=== Test Capture Session Learnings ==="
echo "Session ID: $session_id"
echo "Transcript: $transcript_path"
echo "CWD: $cwd"
echo "Hook Script: $capture_script"
echo ""
echo "Sending JSON input:"
echo "$hook_input" | jq .
echo ""
echo "==================================="
echo ""

# Create a temporary file to capture output
output_file="/tmp/test-capture-output-$session_id.txt"

# Call the capture script with the JSON input, tee output to both console and file
echo "$hook_input" | "$capture_script" 2>&1 | tee "$output_file"

exit_code=${PIPESTATUS[1]}

echo ""
echo "==================================="
echo "Script exited with code: $exit_code"
echo ""

# Display the debug log which contains the Claude interaction
debug_log="/tmp/capture-session-debug-$session_id.log"
if [ -f "$debug_log" ]; then
  echo "==================================="
  echo "=== Claude Debug Log ==="
  echo "==================================="
  cat "$debug_log"
  echo ""
fi

# Display any learning output
learning_logs=$(ls /tmp/*session-learning-$session_id.log 2>/dev/null || true)
if [ -n "$learning_logs" ]; then
  echo "==================================="
  echo "=== Learning Capture Output ==="
  echo "==================================="
  for log in $learning_logs; do
    echo "--- $log ---"
    cat "$log"
    echo ""
  done
fi

echo "==================================="
echo "All logs:"
echo "  Output: $output_file"
echo "  Debug: /tmp/capture-session-debug-$session_id.log"
echo "  Learning: /tmp/*session-learning-$session_id.log"
echo "  Error: /tmp/capture-session-error-$session_id.log (if failed)"

exit $exit_code
