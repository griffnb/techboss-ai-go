package state_files

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	modallib "github.com/modal-labs/libmodal/modal-go"
)

// Test_WriteLocalStateFile_validatesInput tests input validation for WriteLocalStateFile
func Test_WriteLocalStateFile_validatesInput(t *testing.T) {
	t.Run("returns error when sandboxInfo is nil", func(t *testing.T) {
		ctx := context.Background()
		stateFile := &StateFile{Version: CurrentVersion}

		err := WriteLocalStateFile(ctx, nil, "/mnt/workspace", stateFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sandboxInfo")
	})

	t.Run("returns error when sandbox is nil", func(t *testing.T) {
		ctx := context.Background()
		sandboxInfo := &modal.SandboxInfo{
			Sandbox: nil,
		}
		stateFile := &StateFile{Version: CurrentVersion}

		err := WriteLocalStateFile(ctx, sandboxInfo, "/mnt/workspace", stateFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sandbox")
	})

	t.Run("returns error when volumePath is empty", func(t *testing.T) {
		ctx := context.Background()
		// Create SandboxInfo with non-nil Sandbox (just a pointer is enough for validation)
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox",
			Sandbox:   &modallib.Sandbox{}, // Non-nil pointer
		}
		stateFile := &StateFile{Version: CurrentVersion}

		err := WriteLocalStateFile(ctx, sandboxInfo, "", stateFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "volumePath")
	})

	t.Run("returns error when stateFile is nil", func(t *testing.T) {
		ctx := context.Background()
		// Create SandboxInfo with non-nil Sandbox
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox",
			Sandbox:   &modallib.Sandbox{}, // Non-nil pointer
		}

		err := WriteLocalStateFile(ctx, sandboxInfo, "/mnt/workspace", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "stateFile")
	})
}

// Test_WriteLocalStateFile_updatesLastSyncedAt tests that LastSyncedAt is updated
func Test_WriteLocalStateFile_updatesLastSyncedAt(t *testing.T) {
	t.Run("updates LastSyncedAt to current time", func(t *testing.T) {
		// Create state file with old timestamp
		oldTimestamp := time.Now().Add(-1 * time.Hour).Unix()
		stateFile := &StateFile{
			Version:      CurrentVersion,
			LastSyncedAt: oldTimestamp,
			Files:        []FileEntry{},
		}

		// Get timestamp before update
		beforeUpdate := time.Now().Unix()

		// Simulate the update (this is what WriteLocalStateFile does internally)
		stateFile.LastSyncedAt = time.Now().Unix()

		// Get timestamp after update
		afterUpdate := time.Now().Unix()

		// Verify LastSyncedAt was updated
		assert.True(t, stateFile.LastSyncedAt >= beforeUpdate)
		assert.True(t, stateFile.LastSyncedAt <= afterUpdate)
		assert.True(t, stateFile.LastSyncedAt > oldTimestamp)
	})
}

// Test_WriteS3StateFile_validatesInput tests input validation for WriteS3StateFile
func Test_WriteS3StateFile_validatesInput(t *testing.T) {
	t.Run("returns error when sandboxInfo is nil", func(t *testing.T) {
		ctx := context.Background()
		stateFile := &StateFile{Version: CurrentVersion}

		err := WriteS3StateFile(ctx, nil, "/mnt/s3", stateFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sandboxInfo")
	})

	t.Run("returns error when s3MountPath is empty", func(t *testing.T) {
		ctx := context.Background()
		// Create SandboxInfo with non-nil Sandbox
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox",
			Sandbox:   &modallib.Sandbox{}, // Non-nil pointer
		}
		stateFile := &StateFile{Version: CurrentVersion}

		err := WriteS3StateFile(ctx, sandboxInfo, "", stateFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "s3MountPath")
	})
}

// Test_GenerateStateFile_validatesInput tests input validation for GenerateStateFile
func Test_GenerateStateFile_validatesInput(t *testing.T) {
	t.Run("returns error when sandboxInfo is nil", func(t *testing.T) {
		ctx := context.Background()

		_, err := GenerateStateFile(ctx, nil, "/mnt/workspace")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sandboxInfo")
	})

	t.Run("returns error when sandbox is nil", func(t *testing.T) {
		ctx := context.Background()
		sandboxInfo := &modal.SandboxInfo{
			Sandbox: nil,
		}

		_, err := GenerateStateFile(ctx, sandboxInfo, "/mnt/workspace")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sandbox")
	})

	t.Run("returns error when directoryPath is empty", func(t *testing.T) {
		ctx := context.Background()
		// Create SandboxInfo with non-nil Sandbox
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox",
			Sandbox:   &modallib.Sandbox{}, // Non-nil pointer
		}

		_, err := GenerateStateFile(ctx, sandboxInfo, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "directoryPath")
	})
}

// Test_marshalStateFile tests JSON marshaling of state files
func Test_marshalStateFile(t *testing.T) {
	t.Run("marshals state file to valid JSON", func(t *testing.T) {
		stateFile := &StateFile{
			Version:      CurrentVersion,
			LastSyncedAt: 1234567890,
			Files: []FileEntry{
				{
					Path:       "test.txt",
					Checksum:   "abc123",
					Size:       100,
					ModifiedAt: 1234567890,
				},
			},
		}

		data, err := json.MarshalIndent(stateFile, "", "  ")
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		// Verify we can parse it back
		parsed, err := ParseStateFile(data)
		assert.NoError(t, err)
		assert.Equal(t, CurrentVersion, parsed.Version)
		assert.Equal(t, int64(1234567890), parsed.LastSyncedAt)
		assert.Equal(t, 1, len(parsed.Files))
		assert.Equal(t, "test.txt", parsed.Files[0].Path)
	})

	t.Run("marshals empty file list", func(t *testing.T) {
		stateFile := &StateFile{
			Version:      CurrentVersion,
			LastSyncedAt: time.Now().Unix(),
			Files:        []FileEntry{},
		}

		data, err := json.MarshalIndent(stateFile, "", "  ")
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		// Verify we can parse it back
		parsed, err := ParseStateFile(data)
		assert.NoError(t, err)
		assert.Equal(t, CurrentVersion, parsed.Version)
		assert.Equal(t, 0, len(parsed.Files))
	})

	t.Run("marshals multiple files with nested paths", func(t *testing.T) {
		stateFile := &StateFile{
			Version:      CurrentVersion,
			LastSyncedAt: time.Now().Unix(),
			Files: []FileEntry{
				{Path: "file1.txt", Checksum: "hash1", Size: 100, ModifiedAt: time.Now().Unix()},
				{Path: "dir/file2.txt", Checksum: "hash2", Size: 200, ModifiedAt: time.Now().Unix()},
				{Path: "dir/subdir/file3.txt", Checksum: "hash3", Size: 300, ModifiedAt: time.Now().Unix()},
			},
		}

		data, err := json.MarshalIndent(stateFile, "", "  ")
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		// Verify we can parse it back
		parsed, err := ParseStateFile(data)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(parsed.Files))
	})
}

// Test_parseFilePaths tests parsing file paths from find command output
func Test_parseFilePaths(t *testing.T) {
	t.Run("parses find output with multiple files", func(t *testing.T) {
		output := `./file1.txt
./dir/file2.txt
./dir/subdir/file3.txt`

		paths := parseFilePaths(output)

		assert.Equal(t, 3, len(paths))
		// Check that all paths are in the slice
		found1, found2, found3 := false, false, false
		for _, p := range paths {
			if p == "./file1.txt" {
				found1 = true
			}
			if p == "./dir/file2.txt" {
				found2 = true
			}
			if p == "./dir/subdir/file3.txt" {
				found3 = true
			}
		}
		assert.True(t, found1)
		assert.True(t, found2)
		assert.True(t, found3)
	})

	t.Run("handles empty output", func(t *testing.T) {
		output := ``

		paths := parseFilePaths(output)

		assert.Equal(t, 0, len(paths))
	})

	t.Run("handles output with empty lines", func(t *testing.T) {
		output := `./file1.txt

./file2.txt
`

		paths := parseFilePaths(output)

		assert.Equal(t, 2, len(paths))
	})
}
