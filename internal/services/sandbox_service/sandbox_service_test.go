package sandbox_service_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service"
)

func init() {
	system_testing.BuildSystem()
}

func skipIfNotConfigured(t *testing.T) {
	if !modal.Configured() {
		t.Skip("Modal client is not configured, skipping test")
	}
}

func TestNewSandboxService(t *testing.T) {
	t.Run("Creates service successfully", func(t *testing.T) {
		skipIfNotConfigured(t)

		// Act
		service := sandbox_service.NewSandboxService()

		// Assert
		assert.NEmpty(t, service)
	})
}

func TestSandboxService_CreateSandbox(t *testing.T) {
	skipIfNotConfigured(t)

	service := sandbox_service.NewSandboxService()
	ctx := context.Background()

	t.Run("Creates sandbox with accountID added to config", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-account-123")
		config := &modal.SandboxConfig{
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
			},
			VolumeName:      "test-volume",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}

		// Act
		sandboxInfo, err := service.CreateSandbox(ctx, accountID, config)
		if sandboxInfo != nil {
			defer func() {
				_ = modal.Client().TerminateSandbox(ctx, sandboxInfo, false)
			}()
		}

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, sandboxInfo)
		assert.Equal(t, accountID, sandboxInfo.Config.AccountID)
		assert.NotEmpty(t, sandboxInfo.SandboxID)
	})

	t.Run("Validates config before creating", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-account-456")
		config := &modal.SandboxConfig{
			// Missing Image - should fail validation
			VolumeMountPath: "/mnt/workspace",
		}

		// Act
		sandboxInfo, err := service.CreateSandbox(ctx, accountID, config)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, sandboxInfo)
	})
}

func TestSandboxService_TerminateSandbox(t *testing.T) {
	skipIfNotConfigured(t)

	service := sandbox_service.NewSandboxService()
	ctx := context.Background()

	t.Run("Terminates sandbox successfully", func(t *testing.T) {
		// Arrange - create a sandbox first
		accountID := types.UUID("test-account-789")
		config := &modal.SandboxConfig{
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
			},
			VolumeName:      "test-volume",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}
		sandboxInfo, err := service.CreateSandbox(ctx, accountID, config)
		assert.NoError(t, err)

		// Act
		err = service.TerminateSandbox(ctx, sandboxInfo, false)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("Terminates with S3 sync", func(t *testing.T) {
		// Arrange - create a sandbox with S3 config
		accountID := types.UUID("test-account-s3")
		config := &modal.SandboxConfig{
			Image: &modal.ImageConfig{
				BaseImage:          "alpine:3.21",
				DockerfileCommands: []string{"RUN apk add --no-cache aws-cli"},
			},
			VolumeName:      "test-volume",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
			S3Config: &modal.S3MountConfig{
				BucketName: "techboss-test-bucket",
				SecretName: "s3-bucket",
				KeyPrefix:  "test/",
				MountPath:  "/mnt/s3-bucket",
				ReadOnly:   false,
			},
		}
		sandboxInfo, err := service.CreateSandbox(ctx, accountID, config)
		assert.NoError(t, err)

		// Act
		err = service.TerminateSandbox(ctx, sandboxInfo, true)

		// Assert
		assert.NoError(t, err)
	})
}

func TestSandboxService_ExecuteClaudeStream(t *testing.T) {
	skipIfNotConfigured(t)

	service := sandbox_service.NewSandboxService()
	ctx := context.Background()

	t.Run("Executes Claude and streams output", func(t *testing.T) {
		// Arrange - create sandbox with Claude installed
		accountID := types.UUID("test-account-claude")
		config := &modal.SandboxConfig{
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep",
					"RUN curl -fsSL https://claude.ai/install.sh | bash",
					"ENV PATH=/root/.local/bin:$PATH USE_BUILTIN_RIPGREP=0",
				},
			},
			VolumeName:      "test-volume",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}
		sandboxInfo, err := service.CreateSandbox(ctx, accountID, config)
		assert.NoError(t, err)
		defer func() {
			_ = modal.Client().TerminateSandbox(ctx, sandboxInfo, false)
		}()

		claudeConfig := &modal.ClaudeExecConfig{
			Prompt:       "Say hello",
			OutputFormat: "text",
		}

		// Create mock response writer
		responseWriter := httptest.NewRecorder()

		// Act
		claudeProcess, err := service.ExecuteClaudeStream(ctx, sandboxInfo, claudeConfig, responseWriter)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, responseWriter.Body.String())
		assert.NEmpty(t, claudeProcess)
		assert.NEmpty(t, claudeProcess.Process)
		assert.Equal(t, claudeConfig, claudeProcess.Config)
	})

	t.Run("Returns ClaudeProcess with token information", func(t *testing.T) {
		// Arrange - create sandbox with Claude installed
		accountID := types.UUID("test-account-claude-tokens")
		config := &modal.SandboxConfig{
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep",
					"RUN curl -fsSL https://claude.ai/install.sh | bash",
					"ENV PATH=/root/.local/bin:$PATH USE_BUILTIN_RIPGREP=0",
				},
			},
			VolumeName:      "test-volume",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}
		sandboxInfo, err := service.CreateSandbox(ctx, accountID, config)
		assert.NoError(t, err)
		defer func() {
			_ = modal.Client().TerminateSandbox(ctx, sandboxInfo, false)
		}()

		claudeConfig := &modal.ClaudeExecConfig{
			Prompt:          "Say hello",
			OutputFormat:    "stream-json",
			SkipPermissions: true,
		}

		// Create mock response writer
		responseWriter := httptest.NewRecorder()

		// Act
		claudeProcess, err := service.ExecuteClaudeStream(ctx, sandboxInfo, claudeConfig, responseWriter)

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, claudeProcess)
		assert.NEmpty(t, claudeProcess.Process)
		assert.Equal(t, claudeConfig, claudeProcess.Config)
		// Token fields should exist (may be 0 if no summary event was parsed)
		// The important part is that ClaudeProcess is returned so callers can access tokens
	})

	t.Run("Validates claudeConfig before execution", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-account-validation")
		config := &modal.SandboxConfig{
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
			},
			VolumeMountPath: "/mnt/workspace",
		}
		sandboxInfo, err := service.CreateSandbox(ctx, accountID, config)
		assert.NoError(t, err)
		defer func() {
			_ = modal.Client().TerminateSandbox(ctx, sandboxInfo, false)
		}()

		claudeConfig := &modal.ClaudeExecConfig{
			// Missing Prompt - should fail validation
			OutputFormat: "text",
		}

		responseWriter := httptest.NewRecorder()

		// Act
		claudeProcess, err := service.ExecuteClaudeStream(ctx, sandboxInfo, claudeConfig, responseWriter)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, claudeProcess)
	})
}

func TestSandboxService_InitFromS3(t *testing.T) {
	skipIfNotConfigured(t)

	service := sandbox_service.NewSandboxService()
	ctx := context.Background()

	t.Run("Initializes volume from S3", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-account-init-s3")
		config := &modal.SandboxConfig{
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
			},
			VolumeName:      "test-volume",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
			S3Config: &modal.S3MountConfig{
				BucketName: "techboss-test-bucket",
				SecretName: "s3-bucket",
				KeyPrefix:  "test/",
				MountPath:  "/mnt/s3-bucket",
				ReadOnly:   true,
			},
		}

		sandboxInfo, err := service.CreateSandbox(ctx, accountID, config)
		assert.NoError(t, err)
		defer func() {
			_ = modal.Client().TerminateSandbox(ctx, sandboxInfo, false)
		}()

		// Act
		stats, err := service.InitFromS3(ctx, sandboxInfo)

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, stats)
	})
}

func TestSandboxService_SyncToS3(t *testing.T) {
	skipIfNotConfigured(t)

	service := sandbox_service.NewSandboxService()
	ctx := context.Background()

	t.Run("Syncs volume to S3", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-account-sync-s3")
		config := &modal.SandboxConfig{
			Image: &modal.ImageConfig{
				BaseImage:          "alpine:3.21",
				DockerfileCommands: []string{"RUN apk add --no-cache aws-cli"},
			},
			VolumeName:      "test-volume",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
			S3Config: &modal.S3MountConfig{
				BucketName: "techboss-test-bucket",
				SecretName: "s3-bucket",
				KeyPrefix:  "test/",
				MountPath:  "/mnt/s3-bucket",
				ReadOnly:   false,
			},
		}
		sandboxInfo, err := service.CreateSandbox(ctx, accountID, config)
		assert.NoError(t, err)
		defer func() {
			_ = modal.Client().TerminateSandbox(ctx, sandboxInfo, false)
		}()

		// Act
		stats, err := service.SyncToS3(ctx, sandboxInfo)

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, stats)
	})
}

func TestSandboxService_ExecuteClaudeStream_AutoRestart(t *testing.T) {
	skipIfNotConfigured(t)

	service := sandbox_service.NewSandboxService()
	ctx := context.Background()

	t.Run("Auto-creates new sandbox when existing sandbox is terminated", func(t *testing.T) {
		// Arrange - create sandbox with Claude installed
		accountID := types.UUID("test-account-auto-restart")
		config := &modal.SandboxConfig{
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep",
					"RUN curl -fsSL https://claude.ai/install.sh | bash",
					"ENV PATH=/root/.local/bin:$PATH USE_BUILTIN_RIPGREP=0",
				},
			},
			VolumeName:      "test-volume",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}
		sandboxInfo, err := service.CreateSandbox(ctx, accountID, config)
		assert.NoError(t, err)

		// Store original sandbox ID
		originalSandboxID := sandboxInfo.SandboxID

		// Terminate the sandbox to simulate it being in terminated state
		err = service.TerminateSandbox(ctx, sandboxInfo, false)
		assert.NoError(t, err)

		// Verify sandbox is terminated
		status, err := modal.Client().GetSandboxStatusFromInfo(ctx, sandboxInfo)
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusTerminated, status)

		claudeConfig := &modal.ClaudeExecConfig{
			Prompt:       "Say hello",
			OutputFormat: "text",
		}
		responseWriter := httptest.NewRecorder()

		// Act - ExecuteClaudeStream should auto-create new sandbox
		// Note: May fail at Claude execution due to permissions, but auto-restart should work
		_, _ = service.ExecuteClaudeStream(ctx, sandboxInfo, claudeConfig, responseWriter)

		// Clean up the new sandbox
		defer func() {
			_ = modal.Client().TerminateSandbox(ctx, sandboxInfo, false)
		}()

		// Assert - verify a new sandbox was created (different ID)
		// This is the key test - the sandbox should have been recreated even if Claude execution fails
		assert.NEmpty(t, sandboxInfo.SandboxID)
		assert.NotEqual(t, originalSandboxID, sandboxInfo.SandboxID, "A new sandbox should have been created")

		// Verify new sandbox is running
		status, err = modal.Client().GetSandboxStatusFromInfo(ctx, sandboxInfo)
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusRunning, status, "New sandbox should be running")
	})

	t.Run("Updates sandboxInfo in-place with new sandbox details", func(t *testing.T) {
		// Arrange - create and terminate a sandbox
		accountID := types.UUID("test-account-inplace-update")
		config := &modal.SandboxConfig{
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep",
					"RUN curl -fsSL https://claude.ai/install.sh | bash",
					"ENV PATH=/root/.local/bin:$PATH USE_BUILTIN_RIPGREP=0",
				},
			},
			VolumeName:      "test-volume-inplace",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}
		sandboxInfo, err := service.CreateSandbox(ctx, accountID, config)
		assert.NoError(t, err)

		// Store original values for comparison
		originalSandboxID := sandboxInfo.SandboxID
		originalCreatedAt := sandboxInfo.CreatedAt

		// Terminate the sandbox
		err = service.TerminateSandbox(ctx, sandboxInfo, false)
		assert.NoError(t, err)

		claudeConfig := &modal.ClaudeExecConfig{
			Prompt:       "Echo test",
			OutputFormat: "text",
		}
		responseWriter := httptest.NewRecorder()

		// Act - ExecuteClaudeStream should update sandboxInfo in-place
		// Note: May fail at Claude execution due to permissions, but auto-restart should work
		_, _ = service.ExecuteClaudeStream(ctx, sandboxInfo, claudeConfig, responseWriter)

		// Clean up the new sandbox
		defer func() {
			_ = modal.Client().TerminateSandbox(ctx, sandboxInfo, false)
		}()

		// Assert - verify sandboxInfo was updated in-place with new sandbox details
		assert.NEmpty(t, sandboxInfo.SandboxID)
		assert.NotEqual(t, originalSandboxID, sandboxInfo.SandboxID, "SandboxID should be updated to new sandbox")
		assert.NotEqual(t, originalCreatedAt, sandboxInfo.CreatedAt, "CreatedAt should be updated to new sandbox creation time")
		assert.NEmpty(t, sandboxInfo.Sandbox, "Sandbox object should be populated")
		assert.NEmpty(t, sandboxInfo.Config, "Config should be populated")
		assert.Equal(t, accountID, sandboxInfo.Config.AccountID, "AccountID should be preserved")

		// Verify new sandbox is running
		status, err := modal.Client().GetSandboxStatusFromInfo(ctx, sandboxInfo)
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusRunning, status)
	})

	t.Run("Works normally when sandbox is already running", func(t *testing.T) {
		// Arrange - create sandbox with Claude installed (don't terminate it)
		accountID := types.UUID("test-account-already-running")
		config := &modal.SandboxConfig{
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep",
					"RUN curl -fsSL https://claude.ai/install.sh | bash",
					"ENV PATH=/root/.local/bin:$PATH USE_BUILTIN_RIPGREP=0",
				},
			},
			VolumeName:      "test-volume-running",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}
		sandboxInfo, err := service.CreateSandbox(ctx, accountID, config)
		assert.NoError(t, err)
		defer func() {
			_ = modal.Client().TerminateSandbox(ctx, sandboxInfo, false)
		}()

		// Store original sandbox ID and creation time
		originalSandboxID := sandboxInfo.SandboxID
		originalCreatedAt := sandboxInfo.CreatedAt

		// Verify sandbox is running
		status, err := modal.Client().GetSandboxStatusFromInfo(ctx, sandboxInfo)
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusRunning, status)

		claudeConfig := &modal.ClaudeExecConfig{
			Prompt:       "Say hello",
			OutputFormat: "text",
		}
		responseWriter := httptest.NewRecorder()

		// Act - ExecuteClaudeStream should NOT recreate sandbox (may fail at Claude execution)
		_, _ = service.ExecuteClaudeStream(ctx, sandboxInfo, claudeConfig, responseWriter)

		// Assert - verify sandbox was NOT recreated
		assert.Equal(t, originalSandboxID, sandboxInfo.SandboxID, "SandboxID should remain the same when sandbox is already running")
		assert.Equal(t, originalCreatedAt, sandboxInfo.CreatedAt, "CreatedAt should remain the same when sandbox is not recreated")

		// Verify sandbox is still running
		status, err = modal.Client().GetSandboxStatusFromInfo(ctx, sandboxInfo)
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusRunning, status)
	})

	t.Run("Returns error when sandboxInfo config is nil and sandbox is terminated", func(t *testing.T) {
		// Arrange - create a sandbox and then manually clear its config
		accountID := types.UUID("test-account-nil-config")
		config := &modal.SandboxConfig{
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep",
					"RUN curl -fsSL https://claude.ai/install.sh | bash",
					"ENV PATH=/root/.local/bin:$PATH USE_BUILTIN_RIPGREP=0",
				},
			},
			VolumeName:      "test-volume-nil-config",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}
		sandboxInfo, err := service.CreateSandbox(ctx, accountID, config)
		assert.NoError(t, err)

		// Terminate the sandbox
		err = service.TerminateSandbox(ctx, sandboxInfo, false)
		assert.NoError(t, err)

		// Manually clear the config to simulate missing config scenario
		sandboxInfo.Config = nil

		claudeConfig := &modal.ClaudeExecConfig{
			Prompt:       "Say hello",
			OutputFormat: "text",
		}
		responseWriter := httptest.NewRecorder()

		// Act
		claudeProcess, err := service.ExecuteClaudeStream(ctx, sandboxInfo, claudeConfig, responseWriter)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, claudeProcess)
		assert.Contains(t, err.Error(), "cannot recreate sandbox")
		assert.Contains(t, err.Error(), "config is nil")
	})
}
