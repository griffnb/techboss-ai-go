package sandbox_service

import (
	"testing"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
)

// Test_ReconstructSandboxInfo_withValidProvider tests reconstruction with PROVIDER_CLAUDE_CODE
func Test_ReconstructSandboxInfo_withValidProvider(t *testing.T) {
	t.Run("reconstructs SandboxInfo with Claude Code provider", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-account-123")
		externalID := "sb-test-123"
		agentID := types.UUID("test-agent-456")

		model := sandbox.New()
		model.AccountID.Set(accountID)
		model.ExternalID.Set(externalID)
		model.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		model.AgentID.Set(agentID)
		model.Status.Set(constants.STATUS_ACTIVE)

		// Act
		sandboxInfo := ReconstructSandboxInfo(model, accountID)

		// Assert
		assert.NEmpty(t, sandboxInfo)
		assert.Equal(t, externalID, sandboxInfo.SandboxID)
		assert.Equal(t, modal.SandboxStatusRunning, sandboxInfo.Status)
		assert.NEmpty(t, sandboxInfo.Config)
		assert.Equal(t, accountID, sandboxInfo.Config.AccountID)
		assert.NEmpty(t, sandboxInfo.Config.Image)
		assert.Equal(t, "alpine:3.21", sandboxInfo.Config.Image.BaseImage)
		assert.Equal(t, "/mnt/workspace", sandboxInfo.Config.VolumeMountPath)
		assert.Equal(t, "/mnt/workspace", sandboxInfo.Config.Workdir)
		assert.Empty(t, sandboxInfo.Sandbox) // Sandbox handle is nil when reconstructed from DB
	})
}

// Test_ReconstructSandboxInfo_withDeletedSandbox tests status mapping for deleted sandbox
func Test_ReconstructSandboxInfo_withDeletedSandbox(t *testing.T) {
	t.Run("maps deleted sandbox to Terminated status", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-account-789")
		externalID := "sb-deleted-test"

		model := sandbox.New()
		model.AccountID.Set(accountID)
		model.ExternalID.Set(externalID)
		model.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		model.Status.Set(constants.STATUS_ACTIVE)
		model.Deleted.Set(1) // Soft deleted

		// Act
		sandboxInfo := ReconstructSandboxInfo(model, accountID)

		// Assert
		assert.NEmpty(t, sandboxInfo)
		assert.Equal(t, externalID, sandboxInfo.SandboxID)
		assert.Equal(t, modal.SandboxStatusTerminated, sandboxInfo.Status)
	})

	t.Run("maps disabled sandbox to Terminated status", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-account-999")
		externalID := "sb-disabled-test"

		model := sandbox.New()
		model.AccountID.Set(accountID)
		model.ExternalID.Set(externalID)
		model.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		model.Status.Set(constants.STATUS_DISABLED)

		// Act
		sandboxInfo := ReconstructSandboxInfo(model, accountID)

		// Assert
		assert.NEmpty(t, sandboxInfo)
		assert.Equal(t, modal.SandboxStatusTerminated, sandboxInfo.Status)
	})
}

// Test_ReconstructSandboxInfo_withActiveSandbox tests status mapping for active sandbox
func Test_ReconstructSandboxInfo_withActiveSandbox(t *testing.T) {
	t.Run("maps active sandbox to Running status", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-account-active")
		externalID := "sb-active-test"

		model := sandbox.New()
		model.AccountID.Set(accountID)
		model.ExternalID.Set(externalID)
		model.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		model.Status.Set(constants.STATUS_ACTIVE)
		model.Deleted.Set(0)

		// Act
		sandboxInfo := ReconstructSandboxInfo(model, accountID)

		// Assert
		assert.NEmpty(t, sandboxInfo)
		assert.Equal(t, modal.SandboxStatusRunning, sandboxInfo.Status)
	})
}

// Test_ReconstructSandboxInfo_withFallbackConfig tests fallback when template not found
func Test_ReconstructSandboxInfo_withFallbackConfig(t *testing.T) {
	t.Run("uses fallback config for unsupported provider", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-account-fallback")
		externalID := "sb-fallback-test"

		model := sandbox.New()
		model.AccountID.Set(accountID)
		model.ExternalID.Set(externalID)
		model.Provider.Set(sandbox.Provider(999)) // Unsupported provider
		model.Status.Set(constants.STATUS_ACTIVE)

		// Act
		sandboxInfo := ReconstructSandboxInfo(model, accountID)

		// Assert
		assert.NEmpty(t, sandboxInfo)
		assert.Equal(t, externalID, sandboxInfo.SandboxID)
		assert.Equal(t, modal.SandboxStatusRunning, sandboxInfo.Status)
		assert.NEmpty(t, sandboxInfo.Config)
		assert.Equal(t, accountID, sandboxInfo.Config.AccountID)
		// Should use fallback claude template
		assert.NEmpty(t, sandboxInfo.Config.Image)
		assert.Equal(t, "alpine:3.21", sandboxInfo.Config.Image.BaseImage)
	})
}
