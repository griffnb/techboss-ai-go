#!/bin/bash

# Copy all Claude session files to ./.sessions for learning analysis

mkdir -p ./.sessions
SESSION_DIR="$HOME/.claude/projects"

if [ -d "$SESSION_DIR" ]; then
  # Copy all .jsonl session files (searches recursively in subfolders)
  COUNT=0
  find "$SESSION_DIR" -name "*.jsonl" -type f | while read -r SESSION_FILE; do
    BASENAME=$(basename "$SESSION_FILE")
    # If file with same name exists, add a counter
    DEST="./.sessions/$BASENAME"
    if [ -f "$DEST" ]; then
      COUNTER=1
      while [ -f "./.sessions/${BASENAME%.jsonl}_${COUNTER}.jsonl" ]; do
        COUNTER=$((COUNTER + 1))
      done
      DEST="./.sessions/${BASENAME%.jsonl}_${COUNTER}.jsonl"
    fi
    cp "$SESSION_FILE" "$DEST"
    echo "✅ Copied: $SESSION_FILE -> $DEST"
    COUNT=$((COUNT + 1))
  done
  echo "📁 Total session files copied: $(find ./.sessions -name '*.jsonl' | wc -l)"
else
  echo "⚠️ No Claude session directory found"
fi

ls -la ./.sessions/
