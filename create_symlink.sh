#!/bin/bash

# Script to create symlinks for Go folders from go-core/ai/go/
# Usage: ./symlink_go <path_to_go_core>

set -e

if [ $# -ne 1 ]; then
    echo "Usage: $0 <path_to_go_core>"
    echo "Example: $0 ../../griffnb/core/backend"
    exit 1
fi

GO_CORE_PATH="$1"

# Check if the Go path exists
if [ ! -d "$GO_CORE_PATH" ]; then
    echo "Error: Go path '$GO_CORE_PATH' does not exist"
    exit 1
fi

# Convert to absolute path to avoid symlink resolution issues
GO_ABS_PATH=$(realpath "$GO_CORE_PATH")

echo "Creating Go symlinks using go-core path: $GO_CORE_PATH"
echo "Go source path: $GO_CORE_PATH"
echo "Go absolute path: $GO_ABS_PATH"

# List of Go folders to symlink
GO_ITEMS=("docs/CONTROLLERS.md" "docs/MODELS.md" "CLAUDE.md" "AGENTS.md" ".mcp.json" ".github/instructions" ".github/prompts" ".vscode" "scripts/code_gen.sh" "scripts/make_mcp.sh" "scripts/tools.json")

for item in "${GO_ITEMS[@]}"; do
    echo "Processing $item..."
    
    # If the item exists, remove it (whether it's a file, directory, or symlink)
    if [ -e "$item" ] || [ -L "$item" ]; then
        echo "  Removing existing $item"
        rm -rf "$item"
    fi
    
    # Create parent directories if needed
    parent_dir=$(dirname "$item")
    if [ "$parent_dir" != "." ] && [ ! -d "$parent_dir" ]; then
        echo "  Creating parent directory: $parent_dir"
        mkdir -p "$parent_dir"
    fi
    
    # Create the symlink
    if [ -e "${GO_ABS_PATH}/${item}" ]; then
        echo "  Creating symlink: $item -> ${GO_ABS_PATH}/${item}"
        ln -s "${GO_ABS_PATH}/${item}" "$item"
    else
        echo "  Warning: Source ${GO_ABS_PATH}/${item} does not exist, skipping"
    fi
done

echo "Go symlinks created successfully!"