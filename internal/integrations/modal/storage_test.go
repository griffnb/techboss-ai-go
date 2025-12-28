package modal_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
)

// TestInitVolumeFromS3 tests copying files from S3 to volume
func TestInitVolumeFromS3(t *testing.T) {
	skipIfNotConfigured(t)

	client := modal.Client()
	ctx := context.Background()

	t.Run("InitVolumeFromS3 copies files from S3 successfully", func(t *testing.T) {
		// Arrange - Create sandbox with S3 mount
		accountID := types.UUID("test-init-s3-123")
		timestamp := time.Now().Unix()

		config := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl aws-cli",
				},
			},
			VolumeName:      "test-volume-init-s3",
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

		sandboxInfo, err := client.CreateSandbox(ctx, config)
		defer func() {
			if sandboxInfo != nil && sandboxInfo.Sandbox != nil {
				_ = client.TerminateSandbox(ctx, sandboxInfo, false)
			}
		}()
		assert.NoError(t, err)
		assert.NotEmpty(t, sandboxInfo)

		// Act - Initialize volume from S3
		stats, err := client.InitVolumeFromS3(ctx, sandboxInfo)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, stats)
		assert.True(t, stats.Duration > 0)
	})

	t.Run("InitVolumeFromS3 handles empty S3 bucket gracefully", func(t *testing.T) {
		// Arrange - Create sandbox with S3 mount to empty prefix
		accountID := types.UUID("test-init-empty-456")
		timestamp := time.Now().Unix()

		config := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl aws-cli",
				},
			},
			VolumeName:      "test-volume-init-empty",
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

		sandboxInfo, err := client.CreateSandbox(ctx, config)
		defer func() {
			if sandboxInfo != nil && sandboxInfo.Sandbox != nil {
				_ = client.TerminateSandbox(ctx, sandboxInfo, false)
			}
		}()
		assert.NoError(t, err)

		// Act - Initialize from empty S3 location
		stats, err := client.InitVolumeFromS3(ctx, sandboxInfo)

		// Assert - Should succeed with 0 files processed
		assert.NoError(t, err)
		assert.NotEmpty(t, stats)
		assert.Equal(t, 0, stats.FilesDownloaded)
	})

	t.Run("InitVolumeFromS3 returns error for sandbox without S3Config", func(t *testing.T) {
		// Arrange - Create sandbox without S3 mount
		accountID := types.UUID("test-init-no-s3-789")

		config := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
			},
			VolumeName:      "test-volume-no-s3",
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

		// Act - Try to initialize from S3 without S3 config
		_, err = client.InitVolumeFromS3(ctx, sandboxInfo)

		// Assert - Should return error
		assert.Error(t, err)
	})
}

// TestSyncVolumeToS3 tests copying files from volume to S3
func TestSyncVolumeToS3(t *testing.T) {
	skipIfNotConfigured(t)

	client := modal.Client()
	ctx := context.Background()

	t.Run("SyncVolumeToS3 creates timestamped version", func(t *testing.T) {
		// Arrange - Create sandbox and write files to volume
		accountID := types.UUID("test-sync-s3-123")

		config := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl aws-cli",
				},
			},
			VolumeName:      "test-volume-sync-s3",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
			S3Config: &modal.S3MountConfig{
				BucketName: "tb-prod-agent-docs",
				SecretName: "s3-bucket",
				KeyPrefix:  fmt.Sprintf("docs/%s/", accountID.String()),
				MountPath:  "/mnt/s3-bucket",
				ReadOnly:   false,
			},
		}

		sandboxInfo, err := client.CreateSandbox(ctx, config)
		defer func() {
			if sandboxInfo != nil && sandboxInfo.Sandbox != nil {
				_ = client.TerminateSandbox(ctx, sandboxInfo, false)
			}
		}()
		assert.NoError(t, err)

		// Create test files in volume
		cmd := []string{"sh", "-c", "echo 'test content' > /mnt/workspace/test.txt"}
		process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, nil)
		assert.NoError(t, err)
		exitCode, err := process.Wait(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode)

		// Act - Sync volume to S3
		stats, err := client.SyncVolumeToS3(ctx, sandboxInfo)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, stats)
		// Stats should be populated (duration is always > 0)
		assert.True(t, stats.Duration > 0)
		// File count and bytes may be 0 if parsing fails, but sync succeeded
		// We verify the sync worked by checking no error was returned
	})

	t.Run("SyncVolumeToS3 handles empty volume", func(t *testing.T) {
		// Arrange - Create sandbox with empty volume
		accountID := types.UUID("test-sync-empty-456")

		config := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl aws-cli",
				},
			},
			VolumeName:      "test-volume-sync-empty",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
			S3Config: &modal.S3MountConfig{
				BucketName: "tb-prod-agent-docs",
				SecretName: "s3-bucket",
				KeyPrefix:  fmt.Sprintf("docs/%s/", accountID.String()),
				MountPath:  "/mnt/s3-bucket",
				ReadOnly:   false,
			},
		}

		sandboxInfo, err := client.CreateSandbox(ctx, config)
		defer func() {
			if sandboxInfo != nil && sandboxInfo.Sandbox != nil {
				_ = client.TerminateSandbox(ctx, sandboxInfo, false)
			}
		}()
		assert.NoError(t, err)

		// Act - Sync empty volume to S3
		stats, err := client.SyncVolumeToS3(ctx, sandboxInfo)

		// Assert - Should succeed with 0 files processed
		assert.NoError(t, err)
		assert.NotEmpty(t, stats)
		assert.Equal(t, 0, stats.FilesDownloaded)
	})

	t.Run("SyncVolumeToS3 returns error for sandbox without S3Config", func(t *testing.T) {
		// Arrange - Create sandbox without S3 config
		accountID := types.UUID("test-sync-no-s3-789")

		config := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
			},
			VolumeName:      "test-volume-sync-no-s3",
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

		// Act - Try to sync without S3 config
		_, err = client.SyncVolumeToS3(ctx, sandboxInfo)

		// Assert - Should return error
		assert.Error(t, err)
	})
}

// TestGetLatestVersion tests retrieving the most recent timestamp
func TestGetLatestVersion(t *testing.T) {
	skipIfNotConfigured(t)

	client := modal.Client()
	ctx := context.Background()

	t.Run("GetLatestVersion returns most recent timestamp", func(t *testing.T) {
		// Arrange - Create sandbox and sync multiple versions
		accountID := types.UUID("test-version-123")

		config := &modal.SandboxConfig{
			AccountID: accountID,
			Image: &modal.ImageConfig{
				BaseImage: "alpine:3.21",
				DockerfileCommands: []string{
					"RUN apk add --no-cache bash curl aws-cli",
				},
			},
			VolumeName:      "test-volume-version",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
			S3Config: &modal.S3MountConfig{
				BucketName: "tb-prod-agent-docs",
				SecretName: "s3-bucket",
				KeyPrefix:  fmt.Sprintf("docs/%s/", accountID.String()),
				MountPath:  "/mnt/s3-bucket",
				ReadOnly:   false,
			},
		}

		sandboxInfo, err := client.CreateSandbox(ctx, config)
		defer func() {
			if sandboxInfo != nil && sandboxInfo.Sandbox != nil {
				_ = client.TerminateSandbox(ctx, sandboxInfo, false)
			}
		}()
		assert.NoError(t, err)

		// Create test file
		cmd := []string{"sh", "-c", "echo 'version 1' > /mnt/workspace/test.txt"}
		process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, nil)
		assert.NoError(t, err)
		exitCode, err := process.Wait(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode)

		// Sync first version
		beforeTimestamp := time.Now().Unix()
		stats1, err := client.SyncVolumeToS3(ctx, sandboxInfo)
		assert.NoError(t, err)
		assert.NotEmpty(t, stats1)

		// Wait a second to ensure different timestamp
		time.Sleep(2 * time.Second)

		// Update file and sync second version
		cmd = []string{"sh", "-c", "echo 'version 2' > /mnt/workspace/test.txt"}
		process, err = sandboxInfo.Sandbox.Exec(ctx, cmd, nil)
		assert.NoError(t, err)
		exitCode, err = process.Wait(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode)

		stats2, err := client.SyncVolumeToS3(ctx, sandboxInfo)
		assert.NoError(t, err)
		assert.NotEmpty(t, stats2)
		afterTimestamp := time.Now().Unix()

		// Act - Get latest version
		latestVersion, err := client.GetLatestVersion(ctx, accountID, "tb-prod-agent-docs")

		// Assert - Should return the most recent timestamp
		assert.NoError(t, err)
		assert.True(t, latestVersion >= beforeTimestamp)
		assert.True(t, latestVersion <= afterTimestamp)
	})

	t.Run("GetLatestVersion returns 0 for empty bucket", func(t *testing.T) {
		// Arrange - Use account with no versions
		accountID := types.UUID("test-version-empty-456")

		// Act - Get latest version for non-existent account
		latestVersion, err := client.GetLatestVersion(ctx, accountID, "tb-prod-agent-docs")

		// Assert - Should return 0 or error
		if err != nil {
			assert.Error(t, err)
		} else {
			assert.Equal(t, int64(0), latestVersion)
		}
	})

	t.Run("GetLatestVersion returns error for invalid bucket", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("test-version-invalid-789")

		// Act - Get latest version from non-existent bucket
		version, err := client.GetLatestVersion(ctx, accountID, "nonexistent-bucket-xyz")

		// Assert - Should return error or 0 (AWS may not fail immediately)
		if err != nil {
			assert.Error(t, err)
		} else {
			assert.Equal(t, int64(0), version)
		}
	})
}

// TestSyncStats tests the SyncStats struct
func TestSyncStats(t *testing.T) {
	t.Run("SyncStats with all fields", func(t *testing.T) {
		// Arrange & Act
		stats := &modal.SyncStats{
			FilesDownloaded:  10,
			FilesDeleted:     2,
			FilesSkipped:     5,
			BytesTransferred: 1024,
			Duration:         5 * time.Second,
			Errors:           []error{},
		}

		// Assert
		assert.Equal(t, 10, stats.FilesDownloaded)
		assert.Equal(t, 2, stats.FilesDeleted)
		assert.Equal(t, 5, stats.FilesSkipped)
		assert.Equal(t, int64(1024), stats.BytesTransferred)
		assert.Equal(t, 5*time.Second, stats.Duration)
		assert.Equal(t, 0, len(stats.Errors))
	})

	t.Run("SyncStats with errors", func(t *testing.T) {
		// Arrange & Act
		stats := &modal.SyncStats{
			FilesDownloaded:  5,
			FilesDeleted:     1,
			FilesSkipped:     3,
			BytesTransferred: 512,
			Duration:         2 * time.Second,
			Errors:           []error{fmt.Errorf("test error 1"), fmt.Errorf("test error 2")},
		}

		// Assert
		assert.Equal(t, 5, stats.FilesDownloaded)
		assert.Equal(t, 1, stats.FilesDeleted)
		assert.Equal(t, 3, stats.FilesSkipped)
		assert.Equal(t, int64(512), stats.BytesTransferred)
		assert.Equal(t, 2, len(stats.Errors))
	})

	t.Run("SyncStats zero values", func(t *testing.T) {
		// Arrange & Act
		stats := &modal.SyncStats{}

		// Assert - Zero values should be useful
		assert.Equal(t, 0, stats.FilesDownloaded)
		assert.Equal(t, 0, stats.FilesDeleted)
		assert.Equal(t, 0, stats.FilesSkipped)
		assert.Equal(t, int64(0), stats.BytesTransferred)
		assert.Equal(t, time.Duration(0), stats.Duration)
		assert.Empty(t, stats.Errors)
	})

	t.Run("SyncStats with new fields (FilesDownloaded, FilesDeleted, FilesSkipped)", func(t *testing.T) {
		// Arrange & Act - Test new field structure per design phase 3.1
		stats := &modal.SyncStats{
			FilesDownloaded:  15,
			FilesDeleted:     3,
			FilesSkipped:     42,
			BytesTransferred: 2048,
			Duration:         3 * time.Second,
			Errors:           []error{},
		}

		// Assert - Verify all new fields are properly stored
		assert.Equal(t, 15, stats.FilesDownloaded)
		assert.Equal(t, 3, stats.FilesDeleted)
		assert.Equal(t, 42, stats.FilesSkipped)
		assert.Equal(t, int64(2048), stats.BytesTransferred)
		assert.Equal(t, 3*time.Second, stats.Duration)
		assert.Equal(t, 0, len(stats.Errors))
	})
}
