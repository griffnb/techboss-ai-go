#!/bin/bash

# Ralph Loop Watcher for GitHub Actions
# Monitors .claude/ralph-events.log and outputs ::notice annotations
# Run this as a background process

set -uo pipefail

RALPH_EVENTS=".claude/ralph-events.log"
POLL_INTERVAL="${RALPH_WATCHER_INTERVAL:-3}"
LAST_LINE=0

echo "Ralph watcher started, monitoring $RALPH_EVENTS every ${POLL_INTERVAL}s"

# Create the events file if it doesn't exist
touch "$RALPH_EVENTS"

while true; do
  sleep "$POLL_INTERVAL"
  
  if [[ -f "$RALPH_EVENTS" ]]; then
    CURRENT_LINES=$(wc -l < "$RALPH_EVENTS")
    
    if [[ $CURRENT_LINES -gt $LAST_LINE ]]; then
      # Output new lines as workflow annotations
      tail -n +$((LAST_LINE + 1)) "$RALPH_EVENTS" | while read -r line; do
        if [[ -n "$line" ]]; then
          echo "$line"
        fi
      done
      LAST_LINE=$CURRENT_LINES
    fi
  fi
  
  # Check if ralph loop is still active
  if [[ ! -f ".claude/ralph-loop.local.md" ]] && [[ $LAST_LINE -gt 0 ]]; then
    echo "::notice title=Ralph Loop::Loop completed or stopped"
    exit 0
  fi
done
