package sandbox_service

import (
	"testing"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
)

// Test_GetSandboxTemplate_validProviders tests that GetSandboxTemplate returns a valid template for PROVIDER_CLAUDE_CODE
func Test_GetSandboxTemplate_validProviders(t *testing.T) {
	t.Run("returns Claude Code template for PROVIDER_CLAUDE_CODE", func(t *testing.T) {
		// Arrange
		provider := sandbox.PROVIDER_CLAUDE_CODE
		agentID := types.UUID("test-agent-123")

		// Act
		template, err := GetSandboxTemplate(provider, agentID)

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, template)
		assert.Equal(t, sandbox.PROVIDER_CLAUDE_CODE, template.Provider)
		assert.NEmpty(t, template.ImageConfig)
		assert.Equal(t, "alpine:3.21", template.ImageConfig.BaseImage)
	})
}

// Test_GetSandboxTemplate_unsupportedProvider tests that GetSandboxTemplate returns an error for unsupported providers
func Test_GetSandboxTemplate_unsupportedProvider(t *testing.T) {
	t.Run("returns error for unsupported provider", func(t *testing.T) {
		// Arrange
		provider := sandbox.Provider(999) // Invalid provider
		agentID := types.UUID("test-agent-123")

		// Act
		template, err := GetSandboxTemplate(provider, agentID)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, template)
	})
}

// Test_BuildSandboxConfig_createsValidConfig tests that BuildSandboxConfig creates a proper modal.SandboxConfig
func Test_BuildSandboxConfig_createsValidConfig(t *testing.T) {
	t.Run("creates config with proper defaults", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-account-456")
		agentID := types.UUID("test-agent-123")
		template, err := GetSandboxTemplate(sandbox.PROVIDER_CLAUDE_CODE, agentID)
		assert.NoError(t, err)

		// Act
		config := template.BuildSandboxConfig(accountID)

		// Assert
		assert.NEmpty(t, config)
		assert.Equal(t, accountID, config.AccountID)
		assert.NEmpty(t, config.Image)
		assert.Equal(t, "alpine:3.21", config.Image.BaseImage)
		assert.Equal(t, "/mnt/workspace", config.VolumeMountPath)
		assert.Equal(t, "/mnt/workspace", config.Workdir)
	})

	t.Run("does not include S3Config when S3BucketName is empty", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-account-789")
		agentID := types.UUID("test-agent-123")
		template, err := GetSandboxTemplate(sandbox.PROVIDER_CLAUDE_CODE, agentID)
		assert.NoError(t, err)

		// Act
		config := template.BuildSandboxConfig(accountID)

		// Assert
		assert.Empty(t, config.S3Config)
	})

	t.Run("includes S3Config when S3BucketName is not empty", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-account-999")
		template := &SandboxTemplate{
			Provider:     sandbox.PROVIDER_CLAUDE_CODE,
			ImageConfig:  nil, // Will be set by GetSandboxTemplate
			VolumeName:   "",
			S3BucketName: "test-bucket",
			S3KeyPrefix:  "test-prefix/",
			InitFromS3:   true,
		}

		// Act
		config := template.BuildSandboxConfig(accountID)

		// Assert
		assert.NEmpty(t, config.S3Config)
		assert.Equal(t, "test-bucket", config.S3Config.BucketName)
		assert.Equal(t, "s3-bucket", config.S3Config.SecretName)
		assert.Equal(t, "test-prefix/", config.S3Config.KeyPrefix)
		assert.Equal(t, "/mnt/s3-bucket", config.S3Config.MountPath)
		assert.Equal(t, true, config.S3Config.ReadOnly)
	})
}
