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
				{
					Path:       "file2.txt",
					Checksum:   "hash2",
					Size:       200,
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
				{
					Path:       "file2.txt",
					Checksum:   "hash2", // Same checksum
					Size:       200,
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
		assert.Equal(t, 2, len(diff.FilesToSkip))
		assert.Equal(t, "file1.txt", diff.FilesToSkip[0].Path)
		assert.Equal(t, "file2.txt", diff.FilesToSkip[1].Path)
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
					Checksum:   "hash1_old",
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
					Checksum:   "hash1_new", // Different checksum
					Size:       150,
					ModifiedAt: 1234567900,
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
		assert.Equal(t, "hash1_new", diff.FilesToDownload[0].Checksum)
	})
}

// Test_CompareStateFiles_filesOnlyLocal tests files only in local state should be deleted
func Test_CompareStateFiles_filesOnlyLocal(t *testing.T) {
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
					Path:       "old_file.txt",
					Checksum:   "hash_old",
					Size:       50,
					ModifiedAt: 1234567800,
				},
			},
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
				// old_file.txt is not in S3
			},
		}

		// Act
		diff := CompareStateFiles(localState, s3State)

		// Assert
		assert.NotEmpty(t, diff)
		assert.Equal(t, 0, len(diff.FilesToDownload))
		assert.Equal(t, 1, len(diff.FilesToDelete))
		assert.Equal(t, 1, len(diff.FilesToSkip))
		assert.Equal(t, "old_file.txt", diff.FilesToDelete[0])
	})
}

// Test_CompareStateFiles_emptyStates tests comparing empty states returns empty diff
func Test_CompareStateFiles_emptyStates(t *testing.T) {
	t.Run("both states empty returns empty diff", func(t *testing.T) {
		// Arrange
		localState := &StateFile{
			Version:      "1.0",
			LastSyncedAt: 1234567890,
			Files:        []FileEntry{},
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
		assert.Equal(t, 0, len(diff.FilesToDelete))
		assert.Equal(t, 0, len(diff.FilesToSkip))
	})
}

// Test_CompareStateFiles_nilLocalState tests nil local state treated as empty
func Test_CompareStateFiles_nilLocalState(t *testing.T) {
	t.Run("nil local state treated as empty", func(t *testing.T) {
		// Arrange
		var localState *StateFile

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
	})
}

// Test_CompareStateFiles_nilS3State tests nil S3 state treated as empty
func Test_CompareStateFiles_nilS3State(t *testing.T) {
	t.Run("nil S3 state treated as empty", func(t *testing.T) {
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

		var s3State *StateFile

		// Act
		diff := CompareStateFiles(localState, s3State)

		// Assert
		assert.NotEmpty(t, diff)
		assert.Equal(t, 0, len(diff.FilesToDownload))
		assert.Equal(t, 1, len(diff.FilesToDelete))
		assert.Equal(t, 0, len(diff.FilesToSkip))
		assert.Equal(t, "file1.txt", diff.FilesToDelete[0])
	})
}

// Test_CompareStateFiles_mixedScenario tests complex scenario with all actions
func Test_CompareStateFiles_mixedScenario(t *testing.T) {
	t.Run("mixed scenario with download, skip, and delete", func(t *testing.T) {
		// Arrange
		localState := &StateFile{
			Version:      "1.0",
			LastSyncedAt: 1234567890,
			Files: []FileEntry{
				{
					Path:       "unchanged.txt",
					Checksum:   "hash_same",
					Size:       100,
					ModifiedAt: 1234567890,
				},
				{
					Path:       "updated.txt",
					Checksum:   "hash_old",
					Size:       200,
					ModifiedAt: 1234567890,
				},
				{
					Path:       "deleted.txt",
					Checksum:   "hash_delete",
					Size:       50,
					ModifiedAt: 1234567800,
				},
			},
		}

		s3State := &StateFile{
			Version:      "1.0",
			LastSyncedAt: 1234567900,
			Files: []FileEntry{
				{
					Path:       "unchanged.txt",
					Checksum:   "hash_same", // Same
					Size:       100,
					ModifiedAt: 1234567890,
				},
				{
					Path:       "updated.txt",
					Checksum:   "hash_new", // Different
					Size:       250,
					ModifiedAt: 1234567900,
				},
				{
					Path:       "new.txt", // New file
					Checksum:   "hash_new_file",
					Size:       300,
					ModifiedAt: 1234567900,
				},
				// deleted.txt is not in S3
			},
		}

		// Act
		diff := CompareStateFiles(localState, s3State)

		// Assert
		assert.NotEmpty(t, diff)
		assert.Equal(t, 2, len(diff.FilesToDownload)) // updated.txt and new.txt
		assert.Equal(t, 1, len(diff.FilesToDelete))   // deleted.txt
		assert.Equal(t, 1, len(diff.FilesToSkip))     // unchanged.txt

		// Verify download list contains updated and new files
		downloadPath0 := diff.FilesToDownload[0].Path
		downloadPath1 := diff.FilesToDownload[1].Path
		hasUpdated := downloadPath0 == "updated.txt" || downloadPath1 == "updated.txt"
		hasNew := downloadPath0 == "new.txt" || downloadPath1 == "new.txt"
		assert.True(t, hasUpdated)
		assert.True(t, hasNew)

		// Verify delete list
		assert.Equal(t, "deleted.txt", diff.FilesToDelete[0])

		// Verify skip list
		assert.Equal(t, "unchanged.txt", diff.FilesToSkip[0].Path)
	})
}

// Test_CheckIfStale_nilState tests nil state is considered stale
func Test_CheckIfStale_nilState(t *testing.T) {
	t.Run("nil state is considered stale", func(t *testing.T) {
		// Arrange
		var stateFile *StateFile
		thresholdSeconds := int64(3600) // 1 hour

		// Act
		isStale := CheckIfStale(stateFile, thresholdSeconds)

		// Assert
		assert.True(t, isStale)
	})
}

// Test_CheckIfStale_zeroLastSyncedAt tests zero LastSyncedAt is considered stale
func Test_CheckIfStale_zeroLastSyncedAt(t *testing.T) {
	t.Run("zero LastSyncedAt is considered stale", func(t *testing.T) {
		// Arrange
		stateFile := &StateFile{
			Version:      "1.0",
			LastSyncedAt: 0, // Never synced
			Files:        []FileEntry{},
		}
		thresholdSeconds := int64(3600)

		// Act
		isStale := CheckIfStale(stateFile, thresholdSeconds)

		// Assert
		assert.True(t, isStale)
	})
}

// Test_CheckIfStale_freshState tests fresh state is not stale
func Test_CheckIfStale_freshState(t *testing.T) {
	t.Run("fresh state is not stale", func(t *testing.T) {
		// Arrange
		// Use a timestamp that is recent relative to current time
		// Create state synced 30 minutes ago (1800 seconds)
		recentTime := time.Now().Unix() - 1800
		stateFile := &StateFile{
			Version:      "1.0",
			LastSyncedAt: recentTime, // 30 minutes ago from now
			Files:        []FileEntry{},
		}
		thresholdSeconds := int64(3600) // 1 hour

		// Act
		isStale := CheckIfStale(stateFile, thresholdSeconds)

		// Assert
		// State synced 30 minutes ago with 1 hour threshold should not be stale
		assert.True(t, !isStale)
	})
}

// Test_CheckIfStale_oldState tests old state is stale
func Test_CheckIfStale_oldState(t *testing.T) {
	t.Run("old state is stale", func(t *testing.T) {
		// Arrange
		oldTime := int64(1234567890)
		stateFile := &StateFile{
			Version:      "1.0",
			LastSyncedAt: oldTime, // Very old timestamp
			Files:        []FileEntry{},
		}
		thresholdSeconds := int64(3600) // 1 hour

		// Act
		isStale := CheckIfStale(stateFile, thresholdSeconds)

		// Assert
		assert.True(t, isStale)
	})
}

// Test_CheckIfStale_exactThreshold tests state at exact threshold boundary
func Test_CheckIfStale_exactThreshold(t *testing.T) {
	t.Run("state at exact threshold is stale", func(t *testing.T) {
		// Arrange
		thresholdSeconds := int64(3600)
		// Create state synced exactly threshold seconds ago
		exactThresholdTime := time.Now().Unix() - thresholdSeconds - 1 // Add 1 second to ensure > threshold
		stateFile := &StateFile{
			Version:      "1.0",
			LastSyncedAt: exactThresholdTime,
			Files:        []FileEntry{},
		}

		// Act
		isStale := CheckIfStale(stateFile, thresholdSeconds)

		// Assert
		// At threshold boundary, should be stale (age > threshold)
		assert.True(t, isStale)
	})
}
