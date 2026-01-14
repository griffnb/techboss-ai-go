#!/bin/bash
# Auto-commit and push script for CI/CD pipelines
# Used as a stop hook to ensure all changes are committed before session ends

set -e

# Get the directory where the script is located, then navigate to repo root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

cd "$REPO_ROOT"

echo "🔍 Checking for uncommitted changes in $REPO_ROOT..."

# Check if there are any changes (staged or unstaged)
if git diff --quiet && git diff --cached --quiet; then
    echo "✅ No changes to commit"
    exit 0
fi

# Get current branch name
BRANCH=$(git rev-parse --abbrev-ref HEAD)
echo "📌 Current branch: $BRANCH"

# Stage all changes
echo "📦 Staging all changes..."
git add -A

# Check if there's anything to commit after staging
if git diff --cached --quiet; then
    echo "✅ No changes to commit after staging"
    exit 0
fi

# Show what will be committed
echo "📋 Changes to be committed:"
git diff --cached --stat

# Create commit with timestamp
TIMESTAMP=$(date -u +"%Y-%m-%d %H:%M:%S UTC")
COMMIT_MSG="chore: auto-commit from CI/CD pipeline [${TIMESTAMP}]"

echo "💾 Committing with message: $COMMIT_MSG"
git commit -m "$COMMIT_MSG"

# Push to remote
echo "🚀 Pushing to origin/$BRANCH..."
git push origin "$BRANCH"

echo "✅ Successfully committed and pushed all changes"
