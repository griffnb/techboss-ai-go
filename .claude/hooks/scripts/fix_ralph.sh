#!/bin/bash
set -euo pipefail

# Find the ralph-loop.md file in the file system
RALPH_FILE=$(find ~/.claude -name "ralph-loop.md" -type f 2>/dev/null | head -1)

if [ -z "$RALPH_FILE" ]; then
    echo "Error: ralph-loop.md file not found"
    exit 1
fi

echo "Found ralph-loop.md at: $RALPH_FILE"

# Create a temporary file
TEMP_FILE=$(mktemp)

# Use sed to replace the problematic code block
# Replace the ```! block with the Bash() function call
sed -e '/^```!$/,/^```$/ {
    /^```!$/c\
Bash("${CLAUDE_PLUGIN_ROOT}/scripts/setup-ralph-loop.sh" $ARGUMENTS)
    /^"${CLAUDE_PLUGIN_ROOT}\/scripts\/setup-ralph-loop\.sh" \$ARGUMENTS$/d
    /^```$/d
}' "$RALPH_FILE" > "$TEMP_FILE"

# Replace the original file
mv "$TEMP_FILE" "$RALPH_FILE"

echo "Successfully updated ralph-loop.md"