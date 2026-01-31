package sandbox_service

import (
	"context"
	"testing"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service/lifecycle"
)

// Test_GetSandboxTemplate_validProviders tests that GetSandboxTemplate returns a valid template for TYPE_CLAUDE_CODE
func Test_GetSandboxTemplate_validProviders(t *testing.T) {
	t.Run("returns Claude Code template for TYPE_CLAUDE_CODE", func(t *testing.T) {
		// Arrange
		provider := sandbox.TYPE_CLAUDE_CODE
		agentID := types.UUID("test-agent-123")

		// Act
		template, err := GetSandboxTemplate(provider, agentID)

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, template)
		assert.Equal(t, sandbox.TYPE_CLAUDE_CODE, template.Type)
		assert.NEmpty(t, template.ImageConfig)
		assert.Equal(t, "alpine:3.21", template.ImageConfig.BaseImage)
	})
}

// Test_GetSandboxTemplate_unsupportedProvider tests that GetSandboxTemplate returns an error for unsupported providers
func Test_GetSandboxTemplate_unsupportedProvider(t *testing.T) {
	t.Run("returns error for unsupported provider", func(t *testing.T) {
		// Arrange
		sandboxType := sandbox.Type(999) // Invalid type
		agentID := types.UUID("test-agent-123")

		// Act
		template, err := GetSandboxTemplate(sandboxType, agentID)
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
		template, err := GetSandboxTemplate(sandbox.TYPE_CLAUDE_CODE, agentID)
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
		template, err := GetSandboxTemplate(sandbox.TYPE_CLAUDE_CODE, agentID)
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
			Type:         sandbox.TYPE_CLAUDE_CODE,
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

// Test_GetSandboxTemplate_hooksRegistered tests that the template includes lifecycle hooks
func Test_GetSandboxTemplate_hooksRegistered(t *testing.T) {
	t.Run("Claude Code template has all lifecycle hooks registered", func(t *testing.T) {
		// Arrange
		provider := sandbox.TYPE_CLAUDE_CODE
		agentID := types.UUID("test-agent-123")

		// Act
		template, err := GetSandboxTemplate(provider, agentID)

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, template)
		assert.NEmpty(t, template.Hooks, "Hooks should be initialized")
		assert.NEmpty(t, template.Hooks.OnColdStart, "OnColdStart hook should be registered")
		assert.NEmpty(t, template.Hooks.OnMessage, "OnMessage hook should be registered")
		assert.NEmpty(t, template.Hooks.OnStreamFinish, "OnStreamFinish hook should be registered")
		assert.NEmpty(t, template.Hooks.OnTerminate, "OnTerminate hook should be registered")
	})

	t.Run("hooks are default implementations", func(t *testing.T) {
		// Arrange
		provider := sandbox.TYPE_CLAUDE_CODE
		agentID := types.UUID("test-agent-456")

		// Act
		template, err := GetSandboxTemplate(provider, agentID)

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, template)
		assert.NEmpty(t, template.Hooks)

		// Verify hooks are the expected default implementations by comparing function pointers
		// Note: We can't directly compare function pointers in Go, so we verify they're not nil
		// The actual behavior is tested in lifecycle/defaults_test.go
		assert.NEmpty(t, template.Hooks.OnColdStart)
		assert.NEmpty(t, template.Hooks.OnMessage)
		assert.NEmpty(t, template.Hooks.OnStreamFinish)
		assert.NEmpty(t, template.Hooks.OnTerminate)
	})

	t.Run("template without hooks field returns nil hooks", func(t *testing.T) {
		// Arrange - manually construct template with nil hooks
		template := &SandboxTemplate{
			Type:         sandbox.TYPE_CLAUDE_CODE,
			ImageConfig:  nil,
			VolumeName:   "",
			S3BucketName: "",
			S3KeyPrefix:  "",
			InitFromS3:   false,
			Hooks:        nil, // Explicitly nil
		}

		// Assert
		assert.Empty(t, template.Hooks, "Hooks should be nil when not initialized")
	})

	t.Run("can customize hooks after template creation", func(t *testing.T) {
		// Arrange
		provider := sandbox.TYPE_CLAUDE_CODE
		agentID := types.UUID("test-agent-789")
		template, err := GetSandboxTemplate(provider, agentID)
		assert.NoError(t, err)

		// Act - Override with custom hook
		customHookCalled := false
		template.Hooks.OnColdStart = func(_ context.Context, _ *lifecycle.HookData) error {
			customHookCalled = true
			return nil
		}

		// Execute the hook
		err = template.Hooks.OnColdStart(nil, &lifecycle.HookData{})

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, true, customHookCalled, "Custom hook should have been executed")
	})
}
