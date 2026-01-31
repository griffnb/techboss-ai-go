package sandbox_service

import (
	"context"
	"fmt"

	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service/lifecycle"
)

// GitHubTemplateConfig contains configuration for GitHub-integrated sandboxes.
// This enables Claude to work directly with GitHub repositories, creating branches and PRs.
type GitHubTemplateConfig struct {
	InstallationID string // GitHub App installation ID
	Repository     string // Format: "owner/repo"
	SourceBranch   string // Branch to base work on
	TargetBranch   string // New branch to create for changes
	PRTargetBranch string // Base branch for the pull request
	PRTitle        string // Title of the pull request
	PRBody         string // Description of the pull request
	GitUserName    string // Git user name for commits (optional)
	GitUserEmail   string // Git user email for commits (optional)
}

// GetGitHubImage returns an image config with Git and GitHub CLI preinstalled.
// This image is optimized for GitHub operations with the gh CLI tool.
func GetGitHubImage() *modal.ImageConfig {
	return &modal.ImageConfig{
		BaseImage: "python:3.11-slim",
		DockerfileCommands: []string{
			"RUN apt-get update && apt-get install -y git curl",
			"RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg",
			"RUN echo \"deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main\" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null",
			"RUN apt-get update && apt-get install -y gh",
			"RUN apt-get clean && rm -rf /var/lib/apt/lists/*",
		},
	}
}

// GetGitHubTemplate returns a sandbox template configured for GitHub operations.
// The template includes lifecycle hooks that:
// - OnColdStart: Clone repo, create branch, create draft PR
// - OnTerminate: Commit and push any final changes
func GetGitHubTemplate(config *GitHubTemplateConfig) *SandboxTemplate {
	// Set default git user if not provided
	gitUserName := config.GitUserName
	if gitUserName == "" {
		gitUserName = "TechBoss AI"
	}
	gitUserEmail := config.GitUserEmail
	if gitUserEmail == "" {
		gitUserEmail = "ai@techboss.app"
	}

	return &SandboxTemplate{
		Type:         sandbox.TYPE_CLAUDE_CODE,
		ImageConfig:  GetGitHubImage(),
		VolumeName:   "",
		S3BucketName: "",
		S3KeyPrefix:  "",
		InitFromS3:   false,
		Hooks: &lifecycle.LifecycleHooks{
			OnColdStart: func(_ context.Context, hookData *lifecycle.HookData) error {
				// Build the git setup script
				script := fmt.Sprintf(`#!/bin/bash
set -euo pipefail

# Set up authentication
echo "$GITHUB_TOKEN" | gh auth login --with-token

# Configure Git
git config --global user.name "%s"
git config --global user.email "%s"

# Clone repository
gh repo clone "$REPOSITORY" /mnt/workspace
cd /mnt/workspace

# Checkout source branch
git checkout "$SOURCE_BRANCH"
git pull origin "$SOURCE_BRANCH"

# Create new branch
git checkout -b "$TARGET_BRANCH"

# Create draft PR
gh pr create \
  --title "$PR_TITLE" \
  --body "$PR_BODY" \
  --base "$PR_TARGET_BRANCH" \
  --head "$TARGET_BRANCH" \
  --draft

echo "GitHub setup complete"
`, gitUserName, gitUserEmail)

				// Execute the script in the sandbox
				// TODO: This needs to be implemented via Modal API
				// For now, return nil as this will be executed by Modal's lifecycle system
				_ = script
				_ = hookData
				return nil
			},
			OnMessage:      lifecycle.DefaultOnMessage,
			OnStreamFinish: lifecycle.DefaultOnStreamFinish,
			OnTerminate: func(_ context.Context, hookData *lifecycle.HookData) error {
				// Build the cleanup script
				script := `#!/bin/bash
set -euo pipefail

cd /mnt/workspace

# Check if there are uncommitted changes
if [[ -n $(git status --porcelain) ]]; then
  git add .
  git commit -m "Final changes from AI session"
fi

# Push changes
git push origin "$TARGET_BRANCH"

echo "GitHub cleanup complete"
`
				// Execute the script in the sandbox
				// TODO: This needs to be implemented via Modal API
				// For now, return nil as this will be executed by Modal's lifecycle system
				_ = script
				_ = hookData
				return nil
			},
		},
	}
}
