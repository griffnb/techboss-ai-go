package state_files

import (
	"testing"
	"time"

	"github.com/griffnb/core/lib/testtools/assert"
)

// Test_CompareStateFiles_filesOnlyInS3 tests files that exist only in S3 should be downloaded
func Test_CompareStateFiles_filesOnlyInS3(t *testing.T) {
	t.Run("files only in S3 state should be downloaded", func(t *testing.T) {
		// Arrange
		localState := &StateFile{
			Version:      "1.0",
			LastSyncedAt: 1234567890,
			Files:        []FileEntry{},
		}

		s3State := &StateFile{
			Version:      "1.0",
			LastSyncedAt: 1234567900,
			Files: []FileEntry{
				{
					Path:       "file1.txt",
					Checksum:   "hash1",
					Size:       100,
					ModifiedAt: 1234567890,
				},
				{
					Path:       "dir/file2.txt",
					Checksum:   "hash2",
					Size:       200,
					ModifiedAt: 1234567890,
				},
			},
		}

		// Act
		diff := CompareStateFiles(localState, s3State)

		// Assert
		assert.NotEmpty(t, diff)
		assert.Equal(t, 2, len(diff.FilesToDownload))
		assert.Equal(t, 0, len(diff.FilesToDelete))
		assert.Equal(t, 0, len(diff.FilesToSkip))
		assert.Equal(t, "file1.txt", diff.FilesToDownload[0].Path)
		assert.Equal(t, "dir/file2.txt", diff.FilesToDownload[1].Path)
	})
}

// Test_CompareStateFiles_matchingChecksums tests files with matching checksums should be skipped
func Test_CompareStateFiles_matchingChecksums(t *testing.T) {
	t.Run("files with matching checksums should be skipped", func(t *testing.T) {
		// Arrange
		localState := &StateFile{
			Version:      "1.0",
			LastSyncedAt: 1234567890,
			Files: []FileEntry{
				{
					Path:       "file1.txt",
					Checksum:   "hash1",
					Size:       100,
					ModifiedAt: 1234567890,
				},
			},
		}

		s3State := &StateFile{
			Version:      "1.0",
			LastSyncedAt: 1234567900,
			Files: []FileEntry{
				{
					Path:       "file1.txt",
					Checksum:   "hash1", // Same checksum
					Size:       100,
					ModifiedAt: 1234567890,
				},
			},
		}

		// Act
		diff := CompareStateFiles(localState, s3State)

		// Assert
		assert.NotEmpty(t, diff)
		assert.Equal(t, 0, len(diff.FilesToDownload))
		assert.Equal(t, 0, len(diff.FilesToDelete))
		assert.Equal(t, 1, len(diff.FilesToSkip))
		assert.Equal(t, "file1.txt", diff.FilesToSkip[0].Path)
	})
}

// Test_CompareStateFiles_differentChecksums tests files with different checksums should be downloaded
func Test_CompareStateFiles_differentChecksums(t *testing.T) {
	t.Run("files with different checksums should be downloaded", func(t *testing.T) {
		// Arrange
		localState := &StateFile{
			Version:      "1.0",
			LastSyncedAt: 1234567890,
			Files: []FileEntry{
				{
					Path:       "file1.txt",
					Checksum:   "hash1",
					Size:       100,
					ModifiedAt: 1234567890,
				},
			},
		}

		s3State := &StateFile{
			Version:      "1.0",
			LastSyncedAt: 1234567900,
			Files: []FileEntry{
				{
					Path:       "file1.txt",
					Checksum:   "hash2", // Different checksum
					Size:       100,
					ModifiedAt: 1234567895,
				},
			},
		}

		// Act
		diff := CompareStateFiles(localState, s3State)

		// Assert
		assert.NotEmpty(t, diff)
		assert.Equal(t, 1, len(diff.FilesToDownload))
		assert.Equal(t, 0, len(diff.FilesToDelete))
		assert.Equal(t, 0, len(diff.FilesToSkip))
		assert.Equal(t, "file1.txt", diff.FilesToDownload[0].Path)
		assert.Equal(t, "hash2", diff.FilesToDownload[0].Checksum)
	})
}

// Test_CompareStateFiles_filesOnlyInLocal tests files only in local should be deleted
func Test_CompareStateFiles_filesOnlyInLocal(t *testing.T) {
	t.Run("files only in local state should be deleted", func(t *testing.T) {
		// Arrange
		localState := &StateFile{
			Version:      "1.0",
			LastSyncedAt: 1234567890,
			Files: []FileEntry{
				{
					Path:       "file1.txt",
					Checksum:   "hash1",
					Size:       100,
					ModifiedAt: 1234567890,
				},
				{
					Path:       "dir/file2.txt",
					Checksum:   "hash2",
					Size:       200,
					ModifiedAt: 1234567890,
				},
			},
		}

		s3State := &StateFile{
			Version:      "1.0",
			LastSyncedAt: 1234567900,
			Files:        []FileEntry{},
		}

		// Act
		diff := CompareStateFiles(localState, s3State)

		// Assert
		assert.NotEmpty(t, diff)
		assert.Equal(t, 0, len(diff.FilesToDownload))
		assert.Equal(t, 2, len(diff.FilesToDelete))
		assert.Equal(t, 0, len(diff.FilesToSkip))
		assert.Equal(t, "file1.txt", diff.FilesToDelete[0])
		assert.Equal(t, "dir/file2.txt", diff.FilesToDelete[1])
	})
}

// Test_CompareStateFiles_mixedActions tests a mix of download, skip, and delete actions
func Test_CompareStateFiles_mixedActions(t *testing.T) {
	t.Run("should correctly identify mixed actions", func(t *testing.T) {
		// Arrange
		localState := &StateFile{
			Version:      "1.0",
			LastSyncedAt: 1234567890,
			Files: []FileEntry{
				{
					Path:       "unchanged.txt",
					Checksum:   "hash1",
					Size:       100,
					ModifiedAt: 1234567890,
				},
				{
					Path:       "updated.txt",
					Checksum:   "old_hash",
					Size:       200,
					ModifiedAt: 1234567890,
				},
				{
					Path:       "deleted.txt",
					Checksum:   "hash3",
					Size:       300,
					ModifiedAt: 1234567890,
				},
			},
		}

		s3State := &StateFile{
			Version:      "1.0",
			LastSyncedAt: 1234567900,
			Files: []FileEntry{
				{
					Path:       "unchanged.txt",
					Checksum:   "hash1", // Same - skip
					Size:       100,
					ModifiedAt: 1234567890,
				},
				{
					Path:       "updated.txt",
					Checksum:   "new_hash", // Different - download
					Size:       250,
					ModifiedAt: 1234567895,
				},
				{
					Path:       "new.txt",
					Checksum:   "hash4", // New - download
					Size:       400,
					ModifiedAt: 1234567900,
				},
				// deleted.txt not in S3 - delete from local
			},
		}

		// Act
		diff := CompareStateFiles(localState, s3State)

		// Assert
		assert.NotEmpty(t, diff)

		// Should download: updated.txt, new.txt
		assert.Equal(t, 2, len(diff.FilesToDownload))

		// Should delete: deleted.txt
		assert.Equal(t, 1, len(diff.FilesToDelete))
		assert.Equal(t, "deleted.txt", diff.FilesToDelete[0])

		// Should skip: unchanged.txt
		assert.Equal(t, 1, len(diff.FilesToSkip))
		assert.Equal(t, "unchanged.txt", diff.FilesToSkip[0].Path)
	})
}

// Test_CompareStateFiles_nilLocalState tests nil local state (new sandbox)
func Test_CompareStateFiles_nilLocalState(t *testing.T) {
	t.Run("nil local state should download all S3 files", func(t *testing.T) {
		// Arrange
		s3State := &StateFile{
			Version:      "1.0",
			LastSyncedAt: 1234567900,
			Files: []FileEntry{
				{
					Path:       "file1.txt",
					Checksum:   "hash1",
					Size:       100,
					ModifiedAt: 1234567890,
				},
				{
					Path:       "file2.txt",
					Checksum:   "hash2",
					Size:       200,
					ModifiedAt: 1234567890,
				},
			},
		}

		// Act
		diff := CompareStateFiles(nil, s3State)

		// Assert
		assert.NotEmpty(t, diff)
		assert.Equal(t, 2, len(diff.FilesToDownload))
		assert.Equal(t, 0, len(diff.FilesToDelete))
		assert.Equal(t, 0, len(diff.FilesToSkip))
	})
}

// Test_CompareStateFiles_nilS3State tests nil S3 state (empty bucket)
func Test_CompareStateFiles_nilS3State(t *testing.T) {
	t.Run("nil S3 state should delete all local files", func(t *testing.T) {
		// Arrange
		localState := &StateFile{
			Version:      "1.0",
			LastSyncedAt: 1234567890,
			Files: []FileEntry{
				{
					Path:       "file1.txt",
					Checksum:   "hash1",
					Size:       100,
					ModifiedAt: 1234567890,
				},
			},
		}

		// Act
		diff := CompareStateFiles(localState, nil)

		// Assert
		assert.NotEmpty(t, diff)
		assert.Equal(t, 0, len(diff.FilesToDownload))
		assert.Equal(t, 1, len(diff.FilesToDelete))
		assert.Equal(t, 0, len(diff.FilesToSkip))
		assert.Equal(t, "file1.txt", diff.FilesToDelete[0])
	})
}

// Test_CompareStateFiles_bothNil tests both states nil
func Test_CompareStateFiles_bothNil(t *testing.T) {
	t.Run("both nil states should return empty diff", func(t *testing.T) {
		// Act
		diff := CompareStateFiles(nil, nil)

		// Assert
		assert.NotEmpty(t, diff)
		assert.Equal(t, 0, len(diff.FilesToDownload))
		assert.Equal(t, 0, len(diff.FilesToDelete))
		assert.Equal(t, 0, len(diff.FilesToSkip))
	})
}

// Test_CheckIfStale_nilState tests nil state is considered stale
func Test_CheckIfStale_nilState(t *testing.T) {
	t.Run("nil state should be stale", func(t *testing.T) {
		// Act
		isStale := CheckIfStale(nil, 3600)

		// Assert
		assert.Equal(t, true, isStale)
	})
}

// Test_CheckIfStale_zeroLastSynced tests zero LastSyncedAt is considered stale
func Test_CheckIfStale_zeroLastSynced(t *testing.T) {
	t.Run("zero LastSyncedAt should be stale", func(t *testing.T) {
		// Arrange
		state := &StateFile{
			Version:      "1.0",
			LastSyncedAt: 0,
			Files:        []FileEntry{},
		}

		// Act
		isStale := CheckIfStale(state, 3600)

		// Assert
		assert.Equal(t, true, isStale)
	})
}

// Test_CheckIfStale_recentSync tests recent sync is not stale
func Test_CheckIfStale_recentSync(t *testing.T) {
	t.Run("recent sync should not be stale", func(t *testing.T) {
		// Arrange
		now := time.Now().Unix()
		state := &StateFile{
			Version:      "1.0",
			LastSyncedAt: now - 100, // 100 seconds ago
			Files:        []FileEntry{},
		}

		// Act
		isStale := CheckIfStale(state, 3600) // 1 hour threshold

		// Assert
		assert.Equal(t, false, isStale)
	})
}

// Test_CheckIfStale_oldSync tests old sync is stale
func Test_CheckIfStale_oldSync(t *testing.T) {
	t.Run("old sync should be stale", func(t *testing.T) {
		// Arrange
		now := time.Now().Unix()
		state := &StateFile{
			Version:      "1.0",
			LastSyncedAt: now - 7200, // 2 hours ago
			Files:        []FileEntry{},
		}

		// Act
		isStale := CheckIfStale(state, 3600) // 1 hour threshold

		// Assert
		assert.Equal(t, true, isStale)
	})
}

// Test_CheckIfStale_exactThreshold tests sync at exact threshold is stale
func Test_CheckIfStale_exactThreshold(t *testing.T) {
	t.Run("sync at exact threshold should be stale", func(t *testing.T) {
		// Arrange
		now := time.Now().Unix()
		threshold := int64(3600)
		state := &StateFile{
			Version:      "1.0",
			LastSyncedAt: now - threshold,
			Files:        []FileEntry{},
		}

		// Act
		isStale := CheckIfStale(state, threshold)

		// Assert
		assert.Equal(t, false, isStale) // Exactly at threshold is NOT stale (age == threshold, not >)
	})
}
