package sandbox_test

import (
	"testing"
	"time"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
)

func init() {
	system_testing.BuildSystem()
}

func Test_UpdateLastSync_setsTimestampAndStats(t *testing.T) {
	t.Run("sets timestamp and stats on first sync", func(t *testing.T) {
		// Arrange
		metadata := &sandbox.MetaData{}
		filesProcessed := 10
		bytesTransferred := int64(1024)
		durationMs := int64(500)

		// Act
		beforeTime := time.Now().Unix()
		metadata.UpdateLastSync(filesProcessed, bytesTransferred, durationMs)
		afterTime := time.Now().Unix()

		// Assert
		assert.NEmpty(t, metadata.LastS3Sync)
		assert.NEmpty(t, metadata.SyncStats)

		// Verify timestamp is within reasonable range
		assert.True(t, *metadata.LastS3Sync >= beforeTime)
		assert.True(t, *metadata.LastS3Sync <= afterTime)

		// Verify stats
		assert.Equal(t, filesProcessed, metadata.SyncStats.FilesProcessed)
		assert.Equal(t, bytesTransferred, metadata.SyncStats.BytesTransferred)
		assert.Equal(t, durationMs, metadata.SyncStats.DurationMs)
	})

	t.Run("updates timestamp and stats on subsequent sync", func(t *testing.T) {
		// Arrange
		metadata := &sandbox.MetaData{}

		// First sync
		metadata.UpdateLastSync(5, 512, 250)
		firstTimestamp := *metadata.LastS3Sync

		// Wait a bit to ensure different timestamp
		time.Sleep(1 * time.Second)

		// Act - Second sync
		metadata.UpdateLastSync(20, 2048, 1000)

		// Assert
		assert.NEmpty(t, metadata.LastS3Sync)
		assert.NEmpty(t, metadata.SyncStats)

		// Verify timestamp was updated
		assert.True(t, *metadata.LastS3Sync > firstTimestamp)

		// Verify stats were updated
		assert.Equal(t, 20, metadata.SyncStats.FilesProcessed)
		assert.Equal(t, int64(2048), metadata.SyncStats.BytesTransferred)
		assert.Equal(t, int64(1000), metadata.SyncStats.DurationMs)
	})

	t.Run("handles zero values", func(t *testing.T) {
		// Arrange
		metadata := &sandbox.MetaData{}

		// Act
		metadata.UpdateLastSync(0, 0, 0)

		// Assert
		// Verify pointers are not nil
		if metadata.LastS3Sync == nil {
			t.Fatal("LastS3Sync should not be nil")
		}
		if metadata.SyncStats == nil {
			t.Fatal("SyncStats should not be nil")
		}

		// Verify zero values are properly stored
		assert.Equal(t, 0, metadata.SyncStats.FilesProcessed)
		assert.Equal(t, int64(0), metadata.SyncStats.BytesTransferred)
		assert.Equal(t, int64(0), metadata.SyncStats.DurationMs)
	})
}
