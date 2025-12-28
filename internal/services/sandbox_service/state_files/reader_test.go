package state_files

import (
	"testing"

	"github.com/griffnb/core/lib/testtools/assert"
)

// Test_ParseStateFile_validJSON tests parsing valid state file JSON
func Test_ParseStateFile_validJSON(t *testing.T) {
	t.Run("parses valid JSON successfully", func(t *testing.T) {
		validJSON := []byte(`{
			"version": "1.0",
			"last_synced_at": 1234567890,
			"files": [
				{
					"path": "test.txt",
					"checksum": "abc123",
					"size": 1024,
					"modified_at": 1234567890
				}
			]
		}`)

		// Call function
		stateFile, err := ParseStateFile(validJSON)

		// Assert expectations
		assert.NoError(t, err)
		assert.NotEmpty(t, stateFile)
		assert.Equal(t, "1.0", stateFile.Version)
		assert.Equal(t, int64(1234567890), stateFile.LastSyncedAt)
		assert.Equal(t, 1, len(stateFile.Files))
		assert.Equal(t, "test.txt", stateFile.Files[0].Path)
		assert.Equal(t, "abc123", stateFile.Files[0].Checksum)
		assert.Equal(t, int64(1024), stateFile.Files[0].Size)
		assert.Equal(t, int64(1234567890), stateFile.Files[0].ModifiedAt)
	})
}

// Test_ParseStateFile_invalidJSON tests that invalid JSON returns an error
func Test_ParseStateFile_invalidJSON(t *testing.T) {
	t.Run("returns error for invalid JSON", func(t *testing.T) {
		invalidJSON := []byte(`{invalid json}`)

		// Call function
		stateFile, err := ParseStateFile(invalidJSON)

		// Assert error
		assert.Error(t, err)
		assert.Empty(t, stateFile)
	})
}

// Test_ParseStateFile_incompatibleVersion tests that future versions return an error
func Test_ParseStateFile_incompatibleVersion(t *testing.T) {
	t.Run("returns error for incompatible version", func(t *testing.T) {
		futureVersionJSON := []byte(`{
			"version": "99.0",
			"last_synced_at": 1234567890,
			"files": []
		}`)

		// Call function
		stateFile, err := ParseStateFile(futureVersionJSON)

		// Assert error for incompatible version
		assert.Error(t, err)
		assert.Empty(t, stateFile)
	})
}

// Test_ParseStateFile_emptyFiles tests parsing state file with empty file list
func Test_ParseStateFile_emptyFiles(t *testing.T) {
	t.Run("parses state file with empty file list", func(t *testing.T) {
		emptyFilesJSON := []byte(`{
			"version": "1.0",
			"last_synced_at": 1234567890,
			"files": []
		}`)

		// Call function
		stateFile, err := ParseStateFile(emptyFilesJSON)

		// Assert expectations
		assert.NoError(t, err)
		assert.NotEmpty(t, stateFile)
		assert.Equal(t, "1.0", stateFile.Version)
		assert.Equal(t, 0, len(stateFile.Files))
	})
}

// Test_ParseStateFile_emptyData tests that empty data returns an error
func Test_ParseStateFile_emptyData(t *testing.T) {
	t.Run("returns error for empty data", func(t *testing.T) {
		emptyJSON := []byte(``)

		// Call function
		stateFile, err := ParseStateFile(emptyJSON)

		// Assert error
		assert.Error(t, err)
		assert.Empty(t, stateFile)
	})
}

// Test_ParseStateFile_multipleFiles tests parsing state file with multiple files
func Test_ParseStateFile_multipleFiles(t *testing.T) {
	t.Run("parses state file with multiple files", func(t *testing.T) {
		multipleFilesJSON := []byte(`{
			"version": "1.0",
			"last_synced_at": 1234567890,
			"files": [
				{
					"path": "file1.txt",
					"checksum": "hash1",
					"size": 100,
					"modified_at": 1234567890
				},
				{
					"path": "dir/file2.txt",
					"checksum": "hash2",
					"size": 200,
					"modified_at": 1234567900
				},
				{
					"path": "dir/subdir/file3.txt",
					"checksum": "hash3",
					"size": 300,
					"modified_at": 1234567910
				}
			]
		}`)

		// Call function
		stateFile, err := ParseStateFile(multipleFilesJSON)

		// Assert expectations
		assert.NoError(t, err)
		assert.NotEmpty(t, stateFile)
		assert.Equal(t, "1.0", stateFile.Version)
		assert.Equal(t, 3, len(stateFile.Files))
		assert.Equal(t, "file1.txt", stateFile.Files[0].Path)
		assert.Equal(t, "dir/file2.txt", stateFile.Files[1].Path)
		assert.Equal(t, "dir/subdir/file3.txt", stateFile.Files[2].Path)
	})
}
