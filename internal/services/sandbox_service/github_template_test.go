package sandbox_service

import (
	"strings"
	"testing"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
)

func Test_GetGitHubImage(t *testing.T) {
	t.Run("returns valid image config", func(t *testing.T) {
		config := GetGitHubImage()
		assert.NEmpty(t, config)
		assert.Equal(t, "python:3.11-slim", config.BaseImage)
		assert.True(t, len(config.DockerfileCommands) > 0)

		hasGit := false
		hasGH := false
		for _, cmd := range config.DockerfileCommands {
			if strings.Contains(cmd, "git") {
				hasGit = true
			}
			if strings.Contains(cmd, "gh") {
				hasGH = true
			}
		}
		assert.True(t, hasGit)
		assert.True(t, hasGH)
	})
}

func Test_GetGitHubTemplate(t *testing.T) {
	t.Run("returns valid template", func(t *testing.T) {
		config := &GitHubTemplateConfig{
			InstallationID: "12345",
			Repository:     "owner/repo",
			SourceBranch:   "main",
			TargetBranch:   "feature/test",
			PRTargetBranch: "main",
			PRTitle:        "Test PR",
			PRBody:         "Test body",
		}

		template := GetGitHubTemplate(config)

		assert.NEmpty(t, template)
		assert.Equal(t, sandbox.PROVIDER_CLAUDE_CODE, template.Provider)
		assert.NEmpty(t, template.ImageConfig)
		assert.NEmpty(t, template.Hooks)
		assert.NEmpty(t, template.Hooks.OnColdStart)
		assert.NEmpty(t, template.Hooks.OnTerminate)
	})
}
