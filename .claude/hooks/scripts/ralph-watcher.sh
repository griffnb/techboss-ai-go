#!/bin/bash

# Ralph Loop Watcher for GitHub Actions
# Monitors .claude/ralph-loop.local.md and writes updates to GitHub Step Summary
# Run this as a background process: nohup ./ralph-watcher.sh &

# Don't use set -e as we want the loop to continue even if commands fail
set -uo pipefail

RALPH_STATE=".claude/ralph-loop.local.md"
LAST_ITERATION=0
INITIALIZED="false"
POLL_INTERVAL="${RALPH_WATCHER_INTERVAL:-5}"

# Check if running in GitHub Actions
if [[ -z "${GITHUB_STEP_SUMMARY:-}" ]]; then
  echo "Warning: GITHUB_STEP_SUMMARY not set, running in dry-run mode" >&2
  GITHUB_STEP_SUMMARY="/dev/stdout"
fi

# Helper function to write to summary with explicit sync
write_summary() {
  printf '%s\n' "$1" >> "$GITHUB_STEP_SUMMARY"
  sync 2>/dev/null || true
}

echo "Ralph watcher started, monitoring $RALPH_STATE every ${POLL_INTERVAL}s"
echo "Writing to GITHUB_STEP_SUMMARY: $GITHUB_STEP_SUMMARY"

while true; do
  sleep "$POLL_INTERVAL"
  
  if [[ -f "$RALPH_STATE" ]]; then
    # Parse frontmatter - use || true to prevent exit on grep no-match
    FRONTMATTER=$(sed -n '/^---$/,/^---$/{ /^---$/d; p; }' "$RALPH_STATE" 2>/dev/null) || true
    ITERATION=$(echo "$FRONTMATTER" | grep '^iteration:' | sed 's/iteration: *//' 2>/dev/null) || true
    MAX_ITERATIONS=$(echo "$FRONTMATTER" | grep '^max_iterations:' | sed 's/max_iterations: *//' 2>/dev/null) || true
    COMPLETION_PROMISE=$(echo "$FRONTMATTER" | grep '^completion_promise:' | sed 's/completion_promise: *//' | sed 's/^"\(.*\)"$/\1/' 2>/dev/null) || true
    STARTED_AT=$(echo "$FRONTMATTER" | grep '^started_at:' | sed 's/started_at: *//' | sed 's/^"\(.*\)"$/\1/' 2>/dev/null) || true
    
    # Default empty values
    ITERATION="${ITERATION:-0}"
    MAX_ITERATIONS="${MAX_ITERATIONS:-0}"
    
    # Extract prompt (everything after second ---)
    PROMPT=$(awk '/^---$/{i++; next} i>=2' "$RALPH_STATE" 2>/dev/null | head -5) || true
    
    # Initialize summary on first detection
    if [[ "$INITIALIZED" == "false" ]] && [[ -n "$ITERATION" ]] && [[ "$ITERATION" != "0" ]]; then
      INITIALIZED="true"
      MAX_DISPLAY="$MAX_ITERATIONS"
      [[ "$MAX_ITERATIONS" == "0" ]] && MAX_DISPLAY="unlimited"
      
      PROMISE_DISPLAY="_none_"
      [[ -n "$COMPLETION_PROMISE" ]] && [[ "$COMPLETION_PROMISE" != "null" ]] && PROMISE_DISPLAY="\`$COMPLETION_PROMISE\`"
      
      echo "Initializing Ralph Loop summary..."
      {
        echo "## 🔄 Ralph Loop Initialized"
        echo ""
        echo "| Setting | Value |"
        echo "|---------|-------|"
        echo "| **Started At** | $STARTED_AT |"
        echo "| **Max Iterations** | $MAX_DISPLAY |"
        echo "| **Completion Promise** | $PROMISE_DISPLAY |"
        echo ""
        echo "### Prompt"
        echo "\`\`\`"
        echo "$PROMPT"
        echo "\`\`\`"
        echo ""
        echo "---"
        echo ""
      } >> "$GITHUB_STEP_SUMMARY"
      sync 2>/dev/null || true
      
      # Debug: show what we wrote
      echo "Wrote to summary file. Current contents:"
      cat "$GITHUB_STEP_SUMMARY" 2>/dev/null | head -20 || echo "(could not read)"
      
      echo "::notice title=Ralph Loop Started::Max iterations: $MAX_DISPLAY | Promise: ${COMPLETION_PROMISE:-none}"
    fi
    
    # Log iteration changes
    if [[ "$ITERATION" =~ ^[0-9]+$ ]] && [[ $ITERATION -gt $LAST_ITERATION ]]; then
      LAST_ITERATION=$ITERATION
      MAX_DISPLAY="∞"
      [[ "$MAX_ITERATIONS" =~ ^[0-9]+$ ]] && [[ $MAX_ITERATIONS -gt 0 ]] && MAX_DISPLAY="$MAX_ITERATIONS"
      
      echo "Logging iteration $ITERATION..."
      {
        echo "### 🔄 Iteration $ITERATION / $MAX_DISPLAY"
        echo "- **Time**: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
        echo ""
      } >> "$GITHUB_STEP_SUMMARY"
      sync 2>/dev/null || true
      
      echo "::notice title=Ralph Iteration $ITERATION::Progressing..."
    fi
  else
    # State file removed = loop completed or stopped
    if [[ "$INITIALIZED" == "true" ]]; then
      echo "Ralph loop completed, writing final summary..."
      {
        echo "### ✅ Ralph Loop Completed"
        echo "- **Ended At**: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
        echo "- **Final Iteration**: $LAST_ITERATION"
        echo ""
      } >> "$GITHUB_STEP_SUMMARY"
      sync 2>/dev/null || true
      
      echo "::notice title=Ralph Loop Complete::Finished after $LAST_ITERATION iterations"
      exit 0
    fi
  fi
done
