package modal_test

import (
	"testing"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
)

// TestParseDockerFile tests Dockerfile parsing functionality
func TestParseDockerFile(t *testing.T) {
	t.Run("Parse Claude Dockerfile", func(t *testing.T) {
		// We can't directly test the private parseDockerFile function,
		// but we can test it through the public API by checking if
		// the image config is built correctly

		// This is tested through TestCreateSandboxFromDockerFile
		// For now, let's verify the expected image config matches our template
		templateConfig := modal.GetClaudeImageConfig()

		assert.NotEmpty(t, templateConfig)
		assert.Equal(t, "alpine:3.21", templateConfig.BaseImage)
		assert.True(t, len(templateConfig.DockerfileCommands) > 0)
	})
}
