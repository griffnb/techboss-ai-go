#!/bin/bash

# Ralph Loop Watcher for GitHub Actions
# Monitors .claude/ralph-loop.local.md and writes updates to GitHub Step Summary
# Run this as a background process: ./ralph-watcher.sh &

set -euo pipefail

RALPH_STATE=".claude/ralph-loop.local.md"
LAST_ITERATION=0
INITIALIZED=false
POLL_INTERVAL="${RALPH_WATCHER_INTERVAL:-5}"

# Check if running in GitHub Actions
if [[ -z "${GITHUB_STEP_SUMMARY:-}" ]]; then
  echo "Warning: GITHUB_STEP_SUMMARY not set, running in dry-run mode" >&2
  GITHUB_STEP_SUMMARY="/dev/stdout"
fi

while true; do
  sleep "$POLL_INTERVAL"
  
  if [[ -f "$RALPH_STATE" ]]; then
    # Parse frontmatter
    FRONTMATTER=$(sed -n '/^---$/,/^---$/{ /^---$/d; p; }' "$RALPH_STATE" 2>/dev/null || echo "")
    ITERATION=$(echo "$FRONTMATTER" | grep '^iteration:' | sed 's/iteration: *//' || echo "0")
    MAX_ITERATIONS=$(echo "$FRONTMATTER" | grep '^max_iterations:' | sed 's/max_iterations: *//' || echo "0")
    COMPLETION_PROMISE=$(echo "$FRONTMATTER" | grep '^completion_promise:' | sed 's/completion_promise: *//' | sed 's/^"\(.*\)"$/\1/' || echo "")
    STARTED_AT=$(echo "$FRONTMATTER" | grep '^started_at:' | sed 's/started_at: *//' | sed 's/^"\(.*\)"$/\1/' || echo "")
    
    # Extract prompt (everything after second ---)
    PROMPT=$(awk '/^---$/{i++; next} i>=2' "$RALPH_STATE" 2>/dev/null | head -5 || echo "")
    
    # Initialize summary on first detection
    if [[ "$INITIALIZED" != "true" ]] && [[ -n "$ITERATION" ]]; then
      INITIALIZED=true
      MAX_DISPLAY="$MAX_ITERATIONS"
      [[ "$MAX_ITERATIONS" == "0" ]] && MAX_DISPLAY="unlimited"
      
      PROMISE_DISPLAY="_none_"
      [[ -n "$COMPLETION_PROMISE" ]] && [[ "$COMPLETION_PROMISE" != "null" ]] && PROMISE_DISPLAY="\`$COMPLETION_PROMISE\`"
      
      cat >> "$GITHUB_STEP_SUMMARY" <<INIT_EOF
## 🔄 Ralph Loop Initialized

| Setting | Value |
|---------|-------|
| **Started At** | $STARTED_AT |
| **Max Iterations** | $MAX_DISPLAY |
| **Completion Promise** | $PROMISE_DISPLAY |

### Prompt
\`\`\`
$PROMPT
\`\`\`

---

INIT_EOF
      echo "::notice title=Ralph Loop Started::Max iterations: $MAX_DISPLAY | Promise: ${COMPLETION_PROMISE:-none}"
    fi
    
    # Log iteration changes
    if [[ "$ITERATION" =~ ^[0-9]+$ ]] && [[ $ITERATION -gt $LAST_ITERATION ]]; then
      LAST_ITERATION=$ITERATION
      MAX_DISPLAY="∞"
      [[ "$MAX_ITERATIONS" =~ ^[0-9]+$ ]] && [[ $MAX_ITERATIONS -gt 0 ]] && MAX_DISPLAY="$MAX_ITERATIONS"
      
      cat >> "$GITHUB_STEP_SUMMARY" <<ITER_EOF
### 🔄 Iteration $ITERATION / $MAX_DISPLAY
- **Time**: $(date -u +%Y-%m-%dT%H:%M:%SZ)

ITER_EOF
      echo "::notice title=Ralph Iteration $ITERATION::Progressing..."
    fi
  else
    # State file removed = loop completed or stopped
    if [[ "$INITIALIZED" == "true" ]]; then
      cat >> "$GITHUB_STEP_SUMMARY" <<DONE_EOF
### ✅ Ralph Loop Completed
- **Ended At**: $(date -u +%Y-%m-%dT%H:%M:%SZ)
- **Final Iteration**: $LAST_ITERATION

DONE_EOF
      echo "::notice title=Ralph Loop Complete::Finished after $LAST_ITERATION iterations"
      exit 0
    fi
  fi
done
