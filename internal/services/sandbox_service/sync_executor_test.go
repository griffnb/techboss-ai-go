package sandbox_service

import (
	"context"
	"testing"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
)

// Test_ExecuteSyncActions_nilInputs tests error handling for nil inputs.
// Requirement: Validate all inputs before processing
func Test_ExecuteSyncActions_nilInputs(t *testing.T) {
	t.Run("returns error for nil client", func(t *testing.T) {
		ctx := context.Background()
		sandboxInfo := &modal.SandboxInfo{}
		diff := &modal.StateDiff{}

		stats, err := ExecuteSyncActions(ctx, nil, sandboxInfo, diff)

		assert.Error(t, err)
		assert.True(t, stats == nil, "expected stats to be nil")
	})

	t.Run("returns error for nil sandboxInfo", func(t *testing.T) {
		ctx := context.Background()
		mockClient := modal.MockClient()
		diff := &modal.StateDiff{}

		stats, err := ExecuteSyncActions(ctx, mockClient, nil, diff)

		assert.Error(t, err)
		assert.True(t, stats == nil, "expected stats to be nil")
	})

	t.Run("returns error for nil diff", func(t *testing.T) {
		ctx := context.Background()
		mockClient := modal.MockClient()
		sandboxInfo := &modal.SandboxInfo{}

		stats, err := ExecuteSyncActions(ctx, mockClient, sandboxInfo, nil)

		assert.Error(t, err)
		assert.True(t, stats == nil, "expected stats to be nil")
	})

	t.Run("returns error for nil sandboxInfo.Config", func(t *testing.T) {
		ctx := context.Background()
		mockClient := modal.MockClient()
		sandboxInfo := &modal.SandboxInfo{
			Config: nil,
		}
		diff := &modal.StateDiff{}

		stats, err := ExecuteSyncActions(ctx, mockClient, sandboxInfo, diff)

		assert.Error(t, err)
		assert.True(t, stats == nil, "expected stats to be nil")
	})

	t.Run("returns error for nil S3Config", func(t *testing.T) {
		ctx := context.Background()
		mockClient := modal.MockClient()
		sandboxInfo := &modal.SandboxInfo{
			Config: &modal.SandboxConfig{
				VolumeMountPath: "/mnt/workspace",
				S3Config:        nil,
			},
		}
		diff := &modal.StateDiff{}

		stats, err := ExecuteSyncActions(ctx, mockClient, sandboxInfo, diff)

		assert.Error(t, err)
		assert.True(t, stats == nil, "expected stats to be nil")
	})
}

// Test_ExecuteSyncActions_emptyDiff tests that empty diff results in no actions.
// Requirement: Handle empty diffs gracefully
func Test_ExecuteSyncActions_emptyDiff(t *testing.T) {
	t.Run("empty diff returns zero stats", func(t *testing.T) {
		ctx := context.Background()
		mockClient := modal.MockClient()

		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox",
			Config: &modal.SandboxConfig{
				VolumeMountPath: "/mnt/workspace",
				S3Config: &modal.S3MountConfig{
					MountPath: "/mnt/s3",
				},
			},
		}

		diff := &modal.StateDiff{
			FilesToDownload: []modal.FileEntry{},
			FilesToDelete:   []string{},
			FilesToSkip:     []modal.FileEntry{},
		}

		stats, err := ExecuteSyncActions(ctx, mockClient, sandboxInfo, diff)

		assert.NoError(t, err)
		assert.True(t, stats != nil, "expected stats to not be nil")
		assert.Equal(t, 0, stats.FilesDownloaded)
		assert.Equal(t, 0, stats.FilesDeleted)
		assert.Equal(t, 0, stats.FilesSkipped)
		assert.Equal(t, int64(0), stats.BytesTransferred)
		assert.True(t, stats.Duration > 0, "expected non-zero duration")
		assert.Equal(t, 0, len(stats.Errors))
	})

	t.Run("diff with only skipped files", func(t *testing.T) {
		ctx := context.Background()
		mockClient := modal.MockClient()

		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox",
			Config: &modal.SandboxConfig{
				VolumeMountPath: "/mnt/workspace",
				S3Config: &modal.S3MountConfig{
					MountPath: "/mnt/s3",
				},
			},
		}

		diff := &modal.StateDiff{
			FilesToDownload: []modal.FileEntry{},
			FilesToDelete:   []string{},
			FilesToSkip: []modal.FileEntry{
				{Path: "file1.txt", Size: 100},
				{Path: "file2.txt", Size: 200},
				{Path: "file3.txt", Size: 300},
			},
		}

		stats, err := ExecuteSyncActions(ctx, mockClient, sandboxInfo, diff)

		assert.NoError(t, err)
		assert.True(t, stats != nil, "expected stats to not be nil")
		assert.Equal(t, 0, stats.FilesDownloaded)
		assert.Equal(t, 0, stats.FilesDeleted)
		assert.Equal(t, 3, stats.FilesSkipped)
		assert.Equal(t, int64(0), stats.BytesTransferred)
		assert.True(t, stats.Duration > 0, "expected non-zero duration")
	})
}

// NOTE: Tests for actual file operations (downloads, deletes) require integration tests
// with real sandbox instances. These will be covered in integration_test.go where we can
// create real Modal sandboxes and test the full sync workflow.
//
// The unit tests above cover:
// - Input validation
// - Empty diff handling
// - Statistics initialization
//
// Integration tests will cover:
// - Downloading files from S3 to volume
// - Deleting local files
// - Mixed actions (download + delete)
// - Error handling during file operations
// - Accurate byte counting and duration tracking
