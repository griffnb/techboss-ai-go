#!/bin/bash
# Quick Ralph PR Creator
# Usage: ./scripts/ralph-pr.sh "task description" [base-branch]
#        ./scripts/ralph-pr.sh "./agents/specs/task-name/tasks.md" [base-branch]

set -e

TASK="$1"
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
BASE_BRANCH="${2:-$CURRENT_BRANCH}"

if [ -z "$TASK" ]; then
    echo "Usage: $0 \"task description\" [base-branch]"
    echo "Example: $0 \"Add user authentication\" main"
    echo "Example: $0 \".agents/specs/task-sdk-complete/tasks.md\" development"
    exit 1
fi

# Check if TASK is a file path
if [[ "$TASK" == *.md ]] || [[ -f "$TASK" ]]; then
    echo "📁 Detected file path task"
    
    # Extract the parent directory name (folder before the file)
    PARENT_DIR=$(dirname "$TASK")
    FOLDER_NAME=$(basename "$PARENT_DIR")
    
    # Create branch name from folder
    BRANCH_NAME="ralph/${FOLDER_NAME}"
    
    # Create prompt message
    PR_TITLE="Ralph: ${FOLDER_NAME}"
    PR_BODY="@ralph Start implementing the ${FOLDER_NAME} tasks from ${TASK}."
    COMMIT_MSG="Ralph task: ${FOLDER_NAME} from ${TASK}"
    
    echo "📂 Task folder: ${FOLDER_NAME}"
    echo "📄 Task file: ${TASK}"
else
    echo "📝 Detected text task"
    
    # Create a branch name from the task description
    BRANCH_NAME="ralph/$(echo "$TASK" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | head -c 50)"
    
    PR_TITLE="Ralph: $TASK"
    PR_BODY="@ralph $TASK"
    COMMIT_MSG="Ralph task: $TASK"
fi

echo "🌿 Creating branch: $BRANCH_NAME"
git checkout -b "$BRANCH_NAME" "$BASE_BRANCH"

echo "📄 Creating empty commit to enable PR..."
git commit --allow-empty -m "$COMMIT_MSG"

echo "🚀 Pushing branch..."
git push -u origin "$BRANCH_NAME"

echo "🔧 Creating draft PR..."
PR_URL=$(gh pr create \
  --draft \
  --base "$BASE_BRANCH" \
  --title "$PR_TITLE" \
  --body "Task specifications and requirements." 2>&1)

if [ $? -ne 0 ]; then
  echo "❌ Failed to create PR"
  exit 1
fi

echo "💬 Adding task as review comment for Ralph..."
PR_NUMBER=$(echo "$PR_URL" | grep -o '[0-9]*$')
gh pr review "$PR_NUMBER" --comment --body "$PR_BODY"

echo "✅ Done! Ralph will start working automatically."
echo "📍 PR: $PR_URL"
