#!/bin/bash
# Quick Agent PR Creator
# Usage: ./scripts/claude-pr.sh "task description" [base-branch]
#        ./scripts/claude-pr.sh "./agents/specs/task-name/tasks.md" [base-branch]
#
# Environment Variables:
#   NAME     - Agent name (default: "claude")
#   TEMPLATE - Custom prompt template for file-based tasks
#              Available variables: ${FOLDER_NAME}, ${TASK}
#              Default: "Start implementing the ${FOLDER_NAME} tasks from ${TASK}, delegating each task to sub-agents as documented."

set -e

# Agent name from env, default to "claude"
AGENT_NAME="${NAME:-claude}"
AGENT_NAME_LOWER=$(echo "$AGENT_NAME" | tr '[:upper:]' '[:lower:]')
AGENT_NAME_TITLE=$(echo "$AGENT_NAME" | sed 's/./\U&/')

TASK="$1"
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
BASE_BRANCH="${2:-$CURRENT_BRANCH}"

if [ -z "$TASK" ]; then
    echo "Usage: $0 \"task description\" [base-branch]"
    echo "Example: $0 \"Add user authentication\" main"
    echo "Example: $0 \".agents/specs/task-sdk-complete/tasks.md\" development"
    echo ""
    echo "Environment Variables:"
    echo "  NAME     - Agent name (default: claude)"
    echo "  TEMPLATE - Custom prompt template for file-based tasks"
    exit 1
fi

# Check if TASK is a file path
if [[ "$TASK" == *.md ]] || [[ -f "$TASK" ]]; then
    echo "📁 Detected file path task"
    
    # Extract the parent directory name (folder before the file)
    PARENT_DIR=$(dirname "$TASK")
    FOLDER_NAME=$(basename "$PARENT_DIR")
    
    # Create branch name from folder
    BRANCH_NAME="${AGENT_NAME_LOWER}/${FOLDER_NAME}"
    
    # Use custom template if provided, otherwise use default
    if [ -n "$TEMPLATE" ]; then
        # Replace placeholders in the template
        PROMPT_BODY="${TEMPLATE//\$\{FOLDER_NAME\}/$FOLDER_NAME}"
        PROMPT_BODY="${PROMPT_BODY//\$\{TASK\}/$TASK}"
    else
        PROMPT_BODY="Start implementing the ${FOLDER_NAME} tasks from ${TASK}, delegating each task to sub-agents as documented."
    fi
    
    # Create prompt message
    PR_TITLE="${AGENT_NAME_TITLE}: ${FOLDER_NAME}"
    PR_BODY="@${AGENT_NAME_LOWER} ${PROMPT_BODY}"
    COMMIT_MSG="${AGENT_NAME_TITLE} task: ${FOLDER_NAME} from ${TASK}"
    
    echo "📂 Task folder: ${FOLDER_NAME}"
    echo "📄 Task file: ${TASK}"
else
    echo "📝 Detected text task"
    
    
    # Create a branch name from the first 4 words of the task description
    FIRST_WORDS=$(echo "$TASK" | awk '{for(i=1;i<=4 && i<=NF;i++) printf "%s%s", $i, (i<4 && i<NF ? " " : "")}')
    BRANCH_NAME="${AGENT_NAME_LOWER}/$(echo "$FIRST_WORDS" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')"
    
    
    PR_TITLE="${AGENT_NAME_TITLE}: $TASK"
    PR_BODY="@${AGENT_NAME_LOWER} $TASK"
    COMMIT_MSG="${AGENT_NAME_TITLE} task: $TASK"
fi

echo "🌿 Creating branch: $BRANCH_NAME"
git checkout -b "$BRANCH_NAME" "$BASE_BRANCH"

# Check if there are uncommitted changes in .agents folder
if git status --porcelain .agents 2>/dev/null | grep -q .; then
    echo "📄 Committing changes in .agents folder..."
    git add .agents
    git commit -m "$COMMIT_MSG"
else
    echo "📄 Creating empty commit to enable PR..."
    git commit --allow-empty -m "$COMMIT_MSG"
fi

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

echo "💬 Adding task as review comment for Claude..."
PR_NUMBER=$(echo "$PR_URL" | grep -o '[0-9]*$')
gh pr review "$PR_NUMBER" --comment --body "$PR_BODY"

echo "🔙 Returning to original branch: $CURRENT_BRANCH"
git checkout "$CURRENT_BRANCH"

echo "✅ Done! ${AGENT_NAME_TITLE} will start working automatically."
echo "📍 PR: $PR_URL"
