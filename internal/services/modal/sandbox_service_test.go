package modal_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	modalservice "github.com/griffnb/techboss-ai-go/internal/services/modal"
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
		service := modalservice.NewSandboxService()

		// Assert
		assert.NEmpty(t, service)
	})
}

func TestSandboxService_CreateSandbox(t *testing.T) {
	skipIfNotConfigured(t)

	service := modalservice.NewSandboxService()
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

	service := modalservice.NewSandboxService()
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

	service := modalservice.NewSandboxService()
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
			Prompt:          "Say hello",
			OutputFormat:    "text",
			SkipPermissions: true,
		}

		// Create mock response writer
		responseWriter := httptest.NewRecorder()

		// Act
		err = service.ExecuteClaudeStream(ctx, sandboxInfo, claudeConfig, responseWriter)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, responseWriter.Body.String())
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
		err = service.ExecuteClaudeStream(ctx, sandboxInfo, claudeConfig, responseWriter)

		// Assert
		assert.Error(t, err)
	})
}

func TestSandboxService_InitFromS3(t *testing.T) {
	skipIfNotConfigured(t)

	service := modalservice.NewSandboxService()
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

	service := modalservice.NewSandboxService()
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
