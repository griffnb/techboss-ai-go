#!/bin/bash

# Ralph GitHub Comment Updater
# Posts/updates a PR comment with Ralph loop status
# Requires: GITHUB_TOKEN, GITHUB_REPOSITORY, and either PR_NUMBER or ability to detect it

set -uo pipefail

# Arguments
ACTION="${1:-update}"  # start, iteration, complete, stopped, error
ITERATION="${2:-1}"
MAX_ITERATIONS="${3:-5}"
COMPLETION_PROMISE="${4:-}"
MESSAGE="${5:-}"

# Check for required environment variables
if [[ -z "${GITHUB_TOKEN:-}" ]]; then
  echo "GITHUB_TOKEN not set, skipping comment update" >&2
  exit 0
fi

if [[ -z "${GITHUB_REPOSITORY:-}" ]]; then
  echo "GITHUB_REPOSITORY not set, skipping comment update" >&2
  exit 0
fi

# Try to get PR number from various sources
PR_NUMBER="${PR_NUMBER:-}"
if [[ -z "$PR_NUMBER" ]] && [[ -n "${GITHUB_EVENT_PATH:-}" ]] && [[ -f "$GITHUB_EVENT_PATH" ]]; then
  PR_NUMBER=$(jq -r '.pull_request.number // .issue.number // empty' "$GITHUB_EVENT_PATH" 2>/dev/null) || true
fi

if [[ -z "$PR_NUMBER" ]]; then
  echo "Could not determine PR number, skipping comment update" >&2
  exit 0
fi

# Comment marker to identify our comment
COMMENT_MARKER="<!-- ralph-loop-status -->"

# GitHub API base
API_BASE="https://api.github.com/repos/${GITHUB_REPOSITORY}"

# Function to find existing comment
find_comment() {
  curl -s -H "Authorization: token $GITHUB_TOKEN" \
    -H "Accept: application/vnd.github.v3+json" \
    "${API_BASE}/issues/${PR_NUMBER}/comments" | \
    jq -r ".[] | select(.body | contains(\"$COMMENT_MARKER\")) | .id" | head -1
}

# Function to create/update comment
post_comment() {
  local body="$1"
  local comment_id
  comment_id=$(find_comment)
  
  # Escape body for JSON
  local json_body
  json_body=$(jq -n --arg body "$body" '{body: $body}')
  
  if [[ -n "$comment_id" ]]; then
    # Update existing comment
    curl -s -X PATCH \
      -H "Authorization: token $GITHUB_TOKEN" \
      -H "Accept: application/vnd.github.v3+json" \
      "${API_BASE}/issues/comments/${comment_id}" \
      -d "$json_body" > /dev/null
  else
    # Create new comment
    curl -s -X POST \
      -H "Authorization: token $GITHUB_TOKEN" \
      -H "Accept: application/vnd.github.v3+json" \
      "${API_BASE}/issues/${PR_NUMBER}/comments" \
      -d "$json_body" > /dev/null
  fi
}

# Format max iterations display
MAX_DISPLAY="$MAX_ITERATIONS"
[[ "$MAX_ITERATIONS" == "0" ]] && MAX_DISPLAY="∞"

# Format promise display
PROMISE_DISPLAY="None"
[[ -n "$COMPLETION_PROMISE" ]] && [[ "$COMPLETION_PROMISE" != "null" ]] && PROMISE_DISPLAY="\`$COMPLETION_PROMISE\`"

# Current timestamp
TIMESTAMP=$(date -u +"%Y-%m-%d %H:%M:%S UTC")

# Build status emoji and text based on action
case "$ACTION" in
  start)
    STATUS_EMOJI="🔄"
    STATUS_TEXT="Running"
    ;;
  iteration)
    STATUS_EMOJI="🔄"
    STATUS_TEXT="Running (Iteration $ITERATION)"
    ;;
  complete)
    STATUS_EMOJI="✅"
    STATUS_TEXT="Complete"
    ;;
  stopped)
    STATUS_EMOJI="🛑"
    STATUS_TEXT="Stopped (Max iterations)"
    ;;
  error)
    STATUS_EMOJI="⚠️"
    STATUS_TEXT="Error"
    ;;
  *)
    STATUS_EMOJI="🔄"
    STATUS_TEXT="Running"
    ;;
esac

# Build the comment body
COMMENT_BODY="${COMMENT_MARKER}
## ${STATUS_EMOJI} Ralph Loop Status

| Property | Value |
|----------|-------|
| **Status** | ${STATUS_TEXT} |
| **Current Iteration** | ${ITERATION} / ${MAX_DISPLAY} |
| **Completion Promise** | ${PROMISE_DISPLAY} |
| **Last Updated** | ${TIMESTAMP} |

"

# Add message if provided
if [[ -n "$MESSAGE" ]]; then
  COMMENT_BODY="${COMMENT_BODY}
### Latest Activity
${MESSAGE}

"
fi

# Add progress bar
if [[ "$MAX_ITERATIONS" != "0" ]] && [[ "$MAX_ITERATIONS" =~ ^[0-9]+$ ]]; then
  PROGRESS=$((ITERATION * 100 / MAX_ITERATIONS))
  FILLED=$((PROGRESS / 5))
  EMPTY=$((20 - FILLED))
  PROGRESS_BAR=$(printf '█%.0s' $(seq 1 $FILLED 2>/dev/null) || echo "")$(printf '░%.0s' $(seq 1 $EMPTY 2>/dev/null) || echo "")
  COMMENT_BODY="${COMMENT_BODY}
### Progress
\`${PROGRESS_BAR}\` ${PROGRESS}%

"
fi

# Add footer
COMMENT_BODY="${COMMENT_BODY}
---
<sub>🤖 This comment is automatically updated by Ralph Loop</sub>"

# Post the comment
post_comment "$COMMENT_BODY"

echo "Ralph status comment updated: $ACTION (iteration $ITERATION)"
