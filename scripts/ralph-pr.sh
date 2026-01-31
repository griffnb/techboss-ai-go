#!/bin/bash
# Wrapper for github-agent-pr.sh with Claude defaults
# Customize TEMPLATE and NAME here, then run with same arguments as github-agent-pr.sh

export TEMPLATE='Start or continue implementing the ${FOLDER_NAME} tasks from ${TASK}.'
export NAME="claude"

# Run the main script with all passed arguments
exec "$(dirname "$0")/github-agent-pr.sh" "$@"