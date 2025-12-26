package modal_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
)

func init() {
	system_testing.BuildSystem()
}

func skipIfNotConfigured(t *testing.T) {
	if !modal.Configured() {
		t.Skip("Modal client is not configured, skipping test")
	}
}

func TestSandboxConfig(t *testing.T) {
	t.Run("SandboxConfig with all fields", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-account-123")
		volumeName := "test-volume"
		volumeMountPath := "/mnt/workspace"
		workdir := "/mnt/workspace"
		timestamp := time.Now().Unix()

		imageConfig := &modal.ImageConfig{
			BaseImage:          "alpine:3.21",
			DockerfileCommands: []string{"RUN apk add --no-cache bash"},
		}

		s3Config := &modal.S3MountConfig{
			BucketName: "test-bucket",
			SecretName: "s3-secret",
			KeyPrefix:  "docs/test-account/",
			MountPath:  "/mnt/s3-bucket",
			ReadOnly:   true,
			Timestamp:  timestamp,
		}

		secrets := map[string]string{
			"API_KEY": "test-key",
		}

		envVars := map[string]string{
			"ENV_VAR": "value",
		}

		// Act
		config := &modal.SandboxConfig{
			AccountID:       accountID,
			Image:           imageConfig,
			VolumeName:      volumeName,
			VolumeMountPath: volumeMountPath,
			S3Config:        s3Config,
			Workdir:         workdir,
			Secrets:         secrets,
			EnvironmentVars: envVars,
		}

		// Assert
		if config.AccountID != accountID {
			t.Fatalf("expected AccountID %s, got %s", accountID, config.AccountID)
		}
		if config.Image.BaseImage != "alpine:3.21" {
			t.Fatalf("expected BaseImage alpine:3.21, got %s", config.Image.BaseImage)
		}
		if len(config.Image.DockerfileCommands) != 1 {
			t.Fatalf("expected 1 Dockerfile command, got %d", len(config.Image.DockerfileCommands))
		}
		if config.VolumeName != volumeName {
			t.Fatalf("expected VolumeName %s, got %s", volumeName, config.VolumeName)
		}
		if config.VolumeMountPath != volumeMountPath {
			t.Fatalf("expected VolumeMountPath %s, got %s", volumeMountPath, config.VolumeMountPath)
		}
		if config.S3Config.BucketName != "test-bucket" {
			t.Fatalf("expected S3 BucketName test-bucket, got %s", config.S3Config.BucketName)
		}
		if config.S3Config.Timestamp != timestamp {
			t.Fatalf("expected S3 Timestamp %d, got %d", timestamp, config.S3Config.Timestamp)
		}
		if config.Workdir != workdir {
			t.Fatalf("expected Workdir %s, got %s", workdir, config.Workdir)
		}
		if config.Secrets["API_KEY"] != "test-key" {
			t.Fatalf("expected Secrets API_KEY test-key, got %s", config.Secrets["API_KEY"])
		}
		if config.EnvironmentVars["ENV_VAR"] != "value" {
			t.Fatalf("expected EnvironmentVars ENV_VAR value, got %s", config.EnvironmentVars["ENV_VAR"])
		}
	})

	t.Run("SandboxConfig with minimal fields", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-account-456")

		imageConfig := &modal.ImageConfig{
			BaseImage: "alpine:3.21",
		}

		// Act
		config := &modal.SandboxConfig{
			AccountID:       accountID,
			Image:           imageConfig,
			VolumeName:      "minimal-volume",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}

		// Assert
		if config.AccountID != accountID {
			t.Fatalf("expected AccountID %s, got %s", accountID, config.AccountID)
		}
		if config.S3Config != nil {
			t.Fatal("expected S3Config to be nil")
		}
		if config.Secrets != nil {
			t.Fatal("expected Secrets to be nil")
		}
		if config.EnvironmentVars != nil {
			t.Fatal("expected EnvironmentVars to be nil")
		}
	})

	t.Run("ImageConfig with base image only", func(t *testing.T) {
		// Arrange & Act
		imageConfig := &modal.ImageConfig{
			BaseImage: "ubuntu:22.04",
		}

		// Assert
		if imageConfig.BaseImage != "ubuntu:22.04" {
			t.Fatalf("expected BaseImage ubuntu:22.04, got %s", imageConfig.BaseImage)
		}
		if len(imageConfig.DockerfileCommands) != 0 {
			t.Fatalf("expected 0 Dockerfile commands, got %d", len(imageConfig.DockerfileCommands))
		}
	})

	t.Run("ImageConfig with custom Dockerfile commands", func(t *testing.T) {
		// Arrange & Act
		commands := []string{
			"RUN apt-get update",
			"RUN apt-get install -y curl",
			"ENV PATH=/usr/local/bin:$PATH",
		}
		imageConfig := &modal.ImageConfig{
			BaseImage:          "ubuntu:22.04",
			DockerfileCommands: commands,
		}

		// Assert
		if len(imageConfig.DockerfileCommands) != 3 {
			t.Fatalf("expected 3 Dockerfile commands, got %d", len(imageConfig.DockerfileCommands))
		}
		if imageConfig.DockerfileCommands[0] != "RUN apt-get update" {
			t.Fatalf("expected first command to be 'RUN apt-get update', got %s", imageConfig.DockerfileCommands[0])
		}
	})

	t.Run("S3MountConfig with timestamp versioning", func(t *testing.T) {
		// Arrange
		timestamp := int64(1704067200)
		accountID := "test-account-789"
		keyPrefix := "docs/" + accountID + "/1704067200/"

		// Act
		s3Config := &modal.S3MountConfig{
			BucketName: "tb-prod-agent-docs",
			SecretName: "s3-bucket",
			KeyPrefix:  keyPrefix,
			MountPath:  "/mnt/s3-bucket",
			ReadOnly:   true,
			Timestamp:  timestamp,
		}

		// Assert
		if s3Config.BucketName != "tb-prod-agent-docs" {
			t.Fatalf("expected BucketName tb-prod-agent-docs, got %s", s3Config.BucketName)
		}
		if s3Config.SecretName != "s3-bucket" {
			t.Fatalf("expected SecretName s3-bucket, got %s", s3Config.SecretName)
		}
		if s3Config.KeyPrefix != keyPrefix {
			t.Fatalf("expected KeyPrefix %s, got %s", keyPrefix, s3Config.KeyPrefix)
		}
		if s3Config.MountPath != "/mnt/s3-bucket" {
			t.Fatalf("expected MountPath /mnt/s3-bucket, got %s", s3Config.MountPath)
		}
		if !s3Config.ReadOnly {
			t.Fatal("expected ReadOnly to be true")
		}
		if s3Config.Timestamp != timestamp {
			t.Fatalf("expected Timestamp %d, got %d", timestamp, s3Config.Timestamp)
		}
	})
}

func TestSandboxInfo(t *testing.T) {
	t.Run("SandboxInfo with all fields", func(t *testing.T) {
		// Arrange
		sandboxID := "sb_test123"
		createdAt := time.Now()
		config := &modal.SandboxConfig{
			AccountID:       types.UUID("test-account"),
			Image:           &modal.ImageConfig{BaseImage: "alpine:3.21"},
			VolumeName:      "test-volume",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}

		// Act
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: sandboxID,
			Sandbox:   nil, // Will be nil in tests without actual Modal sandbox
			Config:    config,
			CreatedAt: createdAt,
			Status:    modal.SandboxStatusRunning,
		}

		// Assert
		if sandboxInfo.SandboxID != sandboxID {
			t.Fatalf("expected SandboxID %s, got %s", sandboxID, sandboxInfo.SandboxID)
		}
		if sandboxInfo.Config.AccountID != types.UUID("test-account") {
			t.Fatalf("expected AccountID test-account, got %s", sandboxInfo.Config.AccountID)
		}
		if sandboxInfo.Status != modal.SandboxStatusRunning {
			t.Fatalf("expected Status %s, got %s", modal.SandboxStatusRunning, sandboxInfo.Status)
		}
		if !sandboxInfo.CreatedAt.Equal(createdAt) {
			t.Fatalf("expected CreatedAt %v, got %v", createdAt, sandboxInfo.CreatedAt)
		}
	})
}

func TestSandboxStatus(t *testing.T) {
	t.Run("SandboxStatus constants", func(t *testing.T) {
		// Assert constants are defined correctly
		if modal.SandboxStatusRunning != "running" {
			t.Fatalf("expected SandboxStatusRunning to be 'running', got %s", modal.SandboxStatusRunning)
		}
		if modal.SandboxStatusTerminated != "terminated" {
			t.Fatalf("expected SandboxStatusTerminated to be 'terminated', got %s", modal.SandboxStatusTerminated)
		}
		if modal.SandboxStatusError != "error" {
			t.Fatalf("expected SandboxStatusError to be 'error', got %s", modal.SandboxStatusError)
		}
	})

	t.Run("SandboxStatus type usage", func(t *testing.T) {
		// Arrange & Act
		var status modal.SandboxStatus
		status = modal.SandboxStatusRunning

		// Assert
		if status != modal.SandboxStatusRunning {
			t.Fatalf("expected status to be running, got %s", status)
		}

		// Change status
		status = modal.SandboxStatusTerminated
		if status != modal.SandboxStatusTerminated {
			t.Fatalf("expected status to be terminated, got %s", status)
		}
	})
}

// TestCreateSandbox tests sandbox creation with various configurations
func TestCreateSandbox(t *testing.T) {
	skipIfNotConfigured(t)

	client := modal.Client()
	ctx := context.Background()

	t.Run("Basic sandbox with volume only", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-create-basic-123")
		config := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
			},
			VolumeName:      "test-volume",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}

		// Act
		sandboxInfo, err := client.CreateSandbox(ctx, config)
		defer func() {
			if sandboxInfo != nil && sandboxInfo.Sandbox != nil {
				_ = client.TerminateSandbox(ctx, sandboxInfo, false)
			}
		}()

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, sandboxInfo.SandboxID)
		assert.Equal(t, modal.SandboxStatusRunning, sandboxInfo.Status)
		assert.NotEmpty(t, sandboxInfo.Sandbox)
		assert.Equal(t, accountID, sandboxInfo.Config.AccountID)

		err = sandboxInfo.Sandbox.Terminate(context.Background())
		assert.NoError(t, err)
	})

	t.Run("Sandbox with custom Docker commands", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-create-custom-456")
		config := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep aws-cli",
				},
			},
			VolumeName:      "test-volume-custom",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}

		// Act
		sandboxInfo, err := client.CreateSandbox(ctx, config)
		defer func() {
			if sandboxInfo != nil && sandboxInfo.Sandbox != nil {
				_ = client.TerminateSandbox(ctx, sandboxInfo, false)
			}
		}()

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, sandboxInfo.SandboxID)
		assert.Equal(t, modal.SandboxStatusRunning, sandboxInfo.Status)
		assert.Equal(t, 1, len(sandboxInfo.Config.Image.DockerfileCommands))

		err = sandboxInfo.Sandbox.Terminate(context.Background())
		assert.NoError(t, err)
	})

	t.Run("Sandbox with S3 bucket mount", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-create-s3-789")
		timestamp := time.Now().Unix()
		config := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
			},
			VolumeName:      "test-volume-s3",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
			S3Config: &modal.S3MountConfig{
				BucketName: "tb-prod-agent-docs",
				SecretName: "s3-bucket",
				KeyPrefix:  fmt.Sprintf("docs/%s/%d/", accountID.String(), timestamp),
				MountPath:  "/mnt/s3-bucket",
				ReadOnly:   true,
				Timestamp:  timestamp,
			},
		}

		// Act
		sandboxInfo, err := client.CreateSandbox(ctx, config)
		defer func() {
			if sandboxInfo != nil && sandboxInfo.Sandbox != nil {
				_ = client.TerminateSandbox(ctx, sandboxInfo, false)
			}
		}()

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, sandboxInfo.SandboxID)
		assert.Equal(t, modal.SandboxStatusRunning, sandboxInfo.Status)
		assert.NotEmpty(t, sandboxInfo.Config.S3Config)
		assert.Equal(t, timestamp, sandboxInfo.Config.S3Config.Timestamp)

		err = sandboxInfo.Sandbox.Terminate(context.Background())
		assert.NoError(t, err)
	})

	t.Run("Error handling for invalid configs", func(t *testing.T) {
		// Test case 1: Missing AccountID
		t.Run("Missing AccountID", func(t *testing.T) {
			config := &modal.SandboxConfig{
				Image: &modal.ImageConfig{
					BaseImage: "alpine:3.21",
				},
				VolumeMountPath: "/mnt/workspace",
			}

			_, err := client.CreateSandbox(ctx, config)
			assert.Error(t, err)
		})

		// Test case 2: Missing Image config
		t.Run("Missing Image config", func(t *testing.T) {
			config := &modal.SandboxConfig{
				AccountID:       types.UUID("test-error-456"),
				VolumeMountPath: "/mnt/workspace",
			}

			_, err := client.CreateSandbox(ctx, config)
			assert.Error(t, err)
		})

		// Test case 3: Invalid secret name for S3
		t.Run("Invalid S3 secret name", func(t *testing.T) {
			config := &modal.SandboxConfig{
				AccountID: types.UUID("test-error-789"),
				Image: &modal.ImageConfig{
					BaseImage: "alpine:3.21",
				},
				VolumeMountPath: "/mnt/workspace",
				S3Config: &modal.S3MountConfig{
					BucketName: "tb-prod-agent-docs",
					SecretName: "nonexistent-secret-xyz",
					KeyPrefix:  "docs/test/",
					MountPath:  "/mnt/s3-bucket",
				},
			}

			_, err := client.CreateSandbox(ctx, config)
			assert.Error(t, err)
		})
	})
}

// TestTerminateSandbox tests sandbox termination
func TestTerminateSandbox(t *testing.T) {
	skipIfNotConfigured(t)

	client := modal.Client()
	ctx := context.Background()

	t.Run("Terminate sandbox successfully", func(t *testing.T) {
		// Arrange - Create a sandbox first
		accountID := types.UUID("test-terminate-123")
		config := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
			},
			VolumeName:      "test-volume-terminate",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}

		sandboxInfo, err := client.CreateSandbox(ctx, config)
		assert.NoError(t, err)
		assert.NotEmpty(t, sandboxInfo)
		assert.Equal(t, modal.SandboxStatusRunning, sandboxInfo.Status)

		// Act - Terminate the sandbox
		err = client.TerminateSandbox(ctx, sandboxInfo, false)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusTerminated, sandboxInfo.Status)
	})

	t.Run("Terminate with syncToS3 parameter accepted", func(t *testing.T) {
		// Arrange - Create a sandbox
		accountID := types.UUID("test-terminate-sync-456")
		config := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
			},
			VolumeName:      "test-volume-terminate-sync",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}

		sandboxInfo, err := client.CreateSandbox(ctx, config)
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusRunning, sandboxInfo.Status)

		// Act - Terminate with syncToS3 = true (placeholder for Task 4)
		err = client.TerminateSandbox(ctx, sandboxInfo, true)

		// Assert - Should accept the parameter without error (sync logic not yet implemented)
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusTerminated, sandboxInfo.Status)
	})

	t.Run("Terminate already terminated sandbox", func(t *testing.T) {
		// Arrange - Create and terminate a sandbox
		accountID := types.UUID("test-terminate-twice-789")
		config := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
			},
			VolumeName:      "test-volume-terminate-twice",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}

		sandboxInfo, err := client.CreateSandbox(ctx, config)
		assert.NoError(t, err)

		// First termination
		err = client.TerminateSandbox(ctx, sandboxInfo, false)
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusTerminated, sandboxInfo.Status)

		// Act - Try to terminate again
		err = client.TerminateSandbox(ctx, sandboxInfo, false)

		// Assert - Modal SDK handles this gracefully without error
		// The sandbox is already terminated, so this is idempotent
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusTerminated, sandboxInfo.Status)
	})

	t.Run("Terminate with nil sandboxInfo", func(t *testing.T) {
		// Act
		err := client.TerminateSandbox(ctx, nil, false)

		// Assert
		assert.Error(t, err)
	})

	t.Run("Terminate with nil sandbox object", func(t *testing.T) {
		// Arrange
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "fake-id",
			Sandbox:   nil, // Nil sandbox
			Config:    nil,
			CreatedAt: time.Now(),
			Status:    modal.SandboxStatusRunning,
		}

		// Act
		err := client.TerminateSandbox(ctx, sandboxInfo, false)

		// Assert
		assert.Error(t, err)
	})
}

// TestGetSandboxStatus tests retrieving sandbox status
func TestGetSandboxStatus(t *testing.T) {
	skipIfNotConfigured(t)

	client := modal.Client()
	ctx := context.Background()

	t.Run("Get status of running sandbox", func(t *testing.T) {
		// Arrange - Create a sandbox
		accountID := types.UUID("test-status-running-123")
		config := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
			},
			VolumeName:      "test-volume-status",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}

		sandboxInfo, err := client.CreateSandbox(ctx, config)
		defer func() {
			if sandboxInfo != nil && sandboxInfo.Sandbox != nil {
				_ = client.TerminateSandbox(ctx, sandboxInfo, false)
			}
		}()
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusRunning, sandboxInfo.Status)

		// Act - Get status using SandboxInfo
		status, err := client.GetSandboxStatusFromInfo(ctx, sandboxInfo)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusRunning, status)
	})

	t.Run("Get status after termination", func(t *testing.T) {
		// Arrange - Create and terminate a sandbox
		accountID := types.UUID("test-status-terminated-456")
		config := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
			},
			VolumeName:      "test-volume-status-terminated",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}

		sandboxInfo, err := client.CreateSandbox(ctx, config)
		assert.NoError(t, err)

		// Terminate the sandbox
		err = client.TerminateSandbox(ctx, sandboxInfo, false)
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusTerminated, sandboxInfo.Status)

		// Act - Get status after termination using SandboxInfo
		status, err := client.GetSandboxStatusFromInfo(ctx, sandboxInfo)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, modal.SandboxStatusTerminated, status)
	})

	t.Run("Get status with nil sandboxInfo", func(t *testing.T) {
		// Act
		status, err := client.GetSandboxStatusFromInfo(ctx, nil)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, modal.SandboxStatus(""), status)
	})

	t.Run("Get status with nil sandbox object", func(t *testing.T) {
		// Arrange
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "fake-id",
			Sandbox:   nil,
			Config:    nil,
			CreatedAt: time.Now(),
			Status:    modal.SandboxStatusRunning,
		}

		// Act
		status, err := client.GetSandboxStatusFromInfo(ctx, sandboxInfo)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, modal.SandboxStatus(""), status)
	})

	t.Run("GetSandboxStatus with ID only returns error", func(t *testing.T) {
		// Act - Test the ID-based method (which is not fully implemented)
		status, err := client.GetSandboxStatus(ctx, "some-id")

		// Assert - Should return error indicating SandboxInfo is needed
		assert.Error(t, err)
		assert.Equal(t, modal.SandboxStatus(""), status)
	})

	t.Run("GetSandboxStatus with empty ID returns error", func(t *testing.T) {
		// Act
		status, err := client.GetSandboxStatus(ctx, "")

		// Assert
		assert.Error(t, err)
		assert.Equal(t, modal.SandboxStatus(""), status)
	})
}
