#!/bin/bash

# Script to create or remove symlinks for Go folders from go-core/ai/go/
# Usage: ./symlink.sh link <path_to_go_core>
#        ./symlink.sh unlink <path_to_go_core>

set -e

if [ $# -ne 2 ]; then
    echo "Usage: $0 <link|unlink> <path_to_go_core>"
    echo "Example: $0 link ../../griffnb/core/backend"
    echo "Example: $0 unlink ../../griffnb/core/backend"
    exit 1
fi

COMMAND="$1"
GO_CORE_PATH="$2"

# Validate command
if [ "$COMMAND" != "link" ] && [ "$COMMAND" != "unlink" ]; then
    echo "Error: Command must be 'link' or 'unlink'"
    exit 1
fi

# Check if the Go path exists
if [ ! -d "$GO_CORE_PATH" ]; then
    echo "Error: Go path '$GO_CORE_PATH' does not exist"
    exit 1
fi

# Convert to absolute path to avoid symlink resolution issues
GO_ABS_PATH=$(realpath "$GO_CORE_PATH")

# List of Go folders to symlink
GO_ITEMS=(".claude" "docs/CONTROLLERS.md" "docs/MODELS.md" "CLAUDE.md" "AGENTS.md" ".mcp.json" ".github/instructions" ".github/prompts" ".vscode" "scripts/code_gen.sh" "scripts/make_mcp.sh" "scripts/tools.json")

if [ "$COMMAND" = "link" ]; then
    echo "Creating Go symlinks using go-core path: $GO_CORE_PATH"
    echo "Go source path: $GO_CORE_PATH"
    echo "Go absolute path: $GO_ABS_PATH"
    
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

elif [ "$COMMAND" = "unlink" ]; then
    echo "Unlinking and copying actual files from: $GO_CORE_PATH"
    echo "Go source path: $GO_CORE_PATH"
    echo "Go absolute path: $GO_ABS_PATH"
    
    for item in "${GO_ITEMS[@]}"; do
        echo "Processing $item..."
        
        # Check if the item exists and is a symlink
        if [ -L "$item" ]; then
            echo "  Found symlink: $item"
            
            # Get the target of the symlink
            target=$(readlink "$item")
            
            # Check if the target exists
            if [ -e "$target" ]; then
                echo "  Removing symlink: $item"
                rm "$item"
                
                # Copy the actual file/directory content
                if [ -d "$target" ]; then
                    echo "  Copying directory: $target -> $item"
                    cp -r "$target" "$item"
                else
                    # Create parent directories if needed
                    parent_dir=$(dirname "$item")
                    if [ "$parent_dir" != "." ] && [ ! -d "$parent_dir" ]; then
                        echo "  Creating parent directory: $parent_dir"
                        mkdir -p "$parent_dir"
                    fi
                    
                    echo "  Copying file: $target -> $item"
                    cp "$target" "$item"
                fi
            else
                echo "  Warning: Symlink target does not exist, removing symlink only"
                rm "$item"
            fi
        elif [ -e "$item" ]; then
            echo "  Item exists but is not a symlink, skipping: $item"
        else
            echo "  Item does not exist, skipping: $item"
        fi
    done
    
    echo "Symlinks unlinked and replaced with actual files successfully!"
fi