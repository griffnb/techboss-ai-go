#!/bin/bash
set -euo pipefail

# Stop hook script that captures learnings before session ends
# This script runs SYNCHRONOUSLY and blocks session ending until complete
# 
# Behavior differs based on environment:
# - CI/CD: Stage changes, commit, and add PR comment
# - Local: Create/update skill files but do NOT stage or commit

# Read hook input from stdin
input=$(cat)

# Extract session info
session_id=$(echo "$input" | jq -r '.session_id // "unknown"')
transcript_path=$(echo "$input" | jq -r '.transcript_path // ""')
cwd=$(echo "$input" | jq -r '.cwd // "."')

# Detect CI/CD environment
# Check common CI environment variables
is_ci="false"
if [ -n "${CI:-}" ] || \
   [ -n "${GITHUB_ACTIONS:-}" ] || \
   [ -n "${GITLAB_CI:-}" ] || \
   [ -n "${CIRCLECI:-}" ] || \
   [ -n "${JENKINS_URL:-}" ] || \
   [ -n "${TRAVIS:-}" ] || \
   [ -n "${BUILDKITE:-}" ] || \
   [ -n "${CODEBUILD_BUILD_ID:-}" ] || \
   [ -n "${TF_BUILD:-}" ]; then
  is_ci="true"
fi

# Extract GitHub-specific info for PR comments
github_repo="${GITHUB_REPOSITORY:-}"
github_pr_number="${GITHUB_PR_NUMBER:-}"
github_token="${GITHUB_TOKEN:-}"

# If in GitHub Actions, try to get PR number from event
if [ "$is_ci" = "true" ] && [ -n "${GITHUB_EVENT_PATH:-}" ] && [ -f "${GITHUB_EVENT_PATH}" ]; then
  github_pr_number=$(jq -r '.pull_request.number // .issue.number // empty' "$GITHUB_EVENT_PATH" 2>/dev/null || echo "")
fi

# Read configuration if it exists
config_file="$CLAUDE_PROJECT_DIR/.claude/session-learner.local.md"
quality_threshold="true"
dry_run="false"
skills_path=".claude/skills/"

if [ -f "$config_file" ]; then
  # Extract YAML frontmatter values
  quality_threshold=$(sed -n '/^---$/,/^---$/p' "$config_file" | grep "quality_threshold:" | awk '{print $2}' 2>/dev/null || echo "true")
  dry_run=$(sed -n '/^---$/,/^---$/p' "$config_file" | grep "dry_run:" | awk '{print $2}' 2>/dev/null || echo "false")
  skills_path=$(sed -n '/^---$/,/^---$/p' "$config_file" | grep "skills_path:" | awk '{print $2}' 2>/dev/null | tr -d '"' || echo ".claude/skills/")
fi

# Quick quality check - if session is trivial, skip immediately
if [ "$quality_threshold" = "true" ]; then
  # Check if transcript exists and has reasonable size
  if [ -n "$transcript_path" ] && [ -f "$transcript_path" ]; then
    # Count lines in transcript as rough proxy for session complexity
    line_count=$(wc -l < "$transcript_path" 2>/dev/null || echo "0")

    # Skip if transcript is too short (< 50 lines = very brief session)
    if [ "$line_count" -lt 50 ]; then
      echo "{
        \"decision\": \"approve\",
        \"reason\": \"Session too brief for learning capture (${line_count} lines)\",
        \"systemMessage\": \"Session ending. No learnings captured (session too brief).\"
      }" | jq -c .
      exit 0
    fi
  fi
fi

# If dry_run mode, skip actual capture but show what would happen
if [ "$dry_run" = "true" ]; then
  echo "{
    \"decision\": \"approve\",
    \"reason\": \"Dry-run mode enabled\",
    \"systemMessage\": \"Session ending. Dry-run mode: would have captured learnings to $skills_path\"
  }" | jq -c .
  exit 0
fi

# Build analysis prompt for the agent based on environment
if [ "$is_ci" = "true" ]; then
  git_instructions="
Git Workflow (CI/CD Mode):
1. After creating/updating skill files, stage them with: git add $skills_path
2. Commit the changes with message: 'chore: capture session learnings [skip ci]'
3. Push the changes to the current branch
"
  pr_instructions=""
  if [ -n "$github_repo" ] && [ -n "$github_pr_number" ] && [ -n "$github_token" ]; then
    pr_instructions="
PR Comment Instructions:
- Add a comment to PR #$github_pr_number in $github_repo
- Use the GitHub API: POST /repos/$github_repo/issues/$github_pr_number/comments
- Include a summary of learnings captured in the comment
- Format as markdown with '## 🧠 Session Learnings Captured' header
- Use curl with GITHUB_TOKEN for authentication
"
  fi
else
  git_instructions="
Git Workflow (Local Mode):
- DO NOT stage, commit, or push any changes
- Only create/update the skill files
- Output a summary of what files were created/modified so the user can review
"
  pr_instructions=""
fi

analysis_prompt="You are analyzing a completed development session to extract learnings.

Session Info:
- Session ID: $session_id
- Transcript File: $transcript_path
- Working Directory: $cwd
- Environment: $([ "$is_ci" = "true" ] && echo "CI/CD" || echo "Local Development")

CRITICAL FIRST STEPS:
1. Read the ENTIRE transcript file at: $transcript_path
   - Use Bash: wc -l $transcript_path to get total line count
   - Read sequentially in chunks of ~500 lines using: sed -n '1,500p' $transcript_path, sed -n '501,1000p' $transcript_path, etc.
   - Continue until you've read the whole file
   - Take notes on valuable patterns as you go, you can store them in a temporary file if needed in .claude/tmp/session-learning-$session_id.log
2. Use Bash to run 'git diff' and 'git status' in $cwd to see actual code changes
3. Correlate the transcript with the actual file changes

Analysis Guidelines:
- Read the ENTIRE transcript - don't skip sections
- Look for: error resolution, design decisions, non-obvious solutions, patterns discovered
- Note places where the agent struggled or had to try multiple approaches
- Skip capturing: trivial changes, simple CRUD, basic file reads without problem-solving

$git_instructions
$pr_instructions
Configuration:
- Skills path: $skills_path
- This is running in Stop hook - be efficient and complete quickly

IMPORTANT: You MUST read the ENTIRE transcript file sequentially. Do NOT skip sections.
IMPORTANT: If session has no valuable learnings, output 'No valuable learnings found' and exit.
"

# Run learning capture SYNCHRONOUSLY (blocking)
# Session won't end until this completes
capture_output=$(cd "$cwd" && claude -p "/learning-analyzer $analysis_prompt" 2>&1 | head -100)

# Log output for debugging
if [ -n "$capture_output" ]; then
  echo "$capture_output" > /tmp/session-learning-$session_id.log 2>&1 || true
fi

# Build appropriate system message based on environment
if [ "$is_ci" = "true" ]; then
  system_msg="Session ending. Learning capture completed in CI/CD mode - changes committed to branch."
else
  system_msg="Session ending. Learning capture completed - skill files updated locally (not staged/committed). Log: /tmp/session-learning-$session_id.log Transcript: $transcript_path INPUT: $input"
fi

# Always approve stopping after capture completes
echo "{
  \"decision\": \"approve\",
  \"reason\": \"Learning capture completed\",
  \"systemMessage\": \"$system_msg\"
}" | jq -c .

exit 0
