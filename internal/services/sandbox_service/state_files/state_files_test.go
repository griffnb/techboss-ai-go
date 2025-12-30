package state_files_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service/state_files"
)

func init() {
	system_testing.BuildSystem()
}

// skipIfNotConfigured skips the test if Modal is not configured
func skipIfNotConfigured(t *testing.T) {
	if !modal.Configured() {
		t.Skip("Modal client is not configured, skipping integration test")
	}
}

// createTestSandbox creates a simple sandbox for testing state file operations
func createTestSandbox(ctx context.Context, t *testing.T, accountID types.UUID) *modal.SandboxInfo {
	t.Helper()

	client := modal.Client()
	config := &modal.SandboxConfig{
		AccountID: accountID,
		Image: &modal.ImageConfig{
			BaseImage: "alpine:3.21",
			DockerfileCommands: []string{
				"RUN apk add --no-cache bash coreutils findutils",
			},
		},
		VolumeName:      "test-state-volume",
		VolumeMountPath: "/mnt/workspace",
		Workdir:         "/mnt/workspace",
	}

	sandboxInfo, err := client.CreateSandbox(ctx, config)
	assert.NoError(t, err)
	assert.NotEmpty(t, sandboxInfo)
	assert.Equal(t, modal.SandboxStatusRunning, sandboxInfo.Status)

	return sandboxInfo
}

// cleanupSandbox terminates a test sandbox
func cleanupSandbox(ctx context.Context, t *testing.T, sandboxInfo *modal.SandboxInfo) {
	t.Helper()

	if sandboxInfo != nil && sandboxInfo.Sandbox != nil {
		client := modal.Client()
		err := client.TerminateSandbox(ctx, sandboxInfo, false)
		if err != nil {
			t.Logf("Warning: failed to cleanup sandbox: %v", err)
		}
	}
}

// createTestFiles creates test files in the sandbox at the specified path
func createTestFiles(ctx context.Context, t *testing.T, sandboxInfo *modal.SandboxInfo, basePath string, fileCount int) {
	t.Helper()

	// Create directory if it doesn't exist
	cmd := []string{"sh", "-c", fmt.Sprintf("mkdir -p %s", basePath)}
	process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, nil)
	assert.NoError(t, err)
	exitCode, err := process.Wait(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, exitCode)

	// Create test files
	for i := 0; i < fileCount; i++ {
		content := fmt.Sprintf("Test content for file %d", i)
		filePath := fmt.Sprintf("%s/testfile%d.txt", basePath, i)
		cmd := []string{"sh", "-c", fmt.Sprintf("echo '%s' > %s", content, filePath)}
		process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, nil)
		assert.NoError(t, err)
		exitCode, err := process.Wait(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode)
	}
}

// Test_GenerateStateFile_emptydirectory tests generating state file from empty directory
func Test_GenerateStateFile_emptydirectory(t *testing.T) {
	skipIfNotConfigured(t)

	ctx := context.Background()
	accountID := types.UUID("test-state-empty-dir-123")
	sandboxInfo := createTestSandbox(ctx, t, accountID)
	defer cleanupSandbox(ctx, t, sandboxInfo)

	t.Run("Generate state from empty directory", func(t *testing.T) {
		// Arrange - workspace is already empty in new sandbox

		// Act
		stateFile, err := state_files.GenerateStateFile(ctx, sandboxInfo, "/mnt/workspace")

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, stateFile)
		assert.Equal(t, "1.0", stateFile.Version)
		assert.Equal(t, 0, len(stateFile.Files))
		assert.NotEqual(t, int64(0), stateFile.LastSyncedAt)
	})
}

// Test_GenerateStateFile_withfiles tests generating state file with various file counts
func Test_GenerateStateFile_withfiles(t *testing.T) {
	skipIfNotConfigured(t)

	ctx := context.Background()
	accountID := types.UUID("test-state-gen-files-456")
	sandboxInfo := createTestSandbox(ctx, t, accountID)
	defer cleanupSandbox(ctx, t, sandboxInfo)

	t.Run("Generate state with small file count", func(t *testing.T) {
		// Arrange - Create 5 test files
		createTestFiles(ctx, t, sandboxInfo, "/mnt/workspace", 5)

		// Act
		stateFile, err := state_files.GenerateStateFile(ctx, sandboxInfo, "/mnt/workspace")

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, stateFile)
		assert.Equal(t, "1.0", stateFile.Version)
		assert.Equal(t, 5, len(stateFile.Files))

		// Verify each file entry has required fields
		for _, fileEntry := range stateFile.Files {
			assert.NotEmpty(t, fileEntry.Path)
			assert.NotEmpty(t, fileEntry.Checksum)
			assert.NotEqual(t, int64(0), fileEntry.Size)
			assert.NotEqual(t, int64(0), fileEntry.ModifiedAt)
		}
	})

	t.Run("Generate state with larger file count", func(t *testing.T) {
		// Arrange - Create 50 additional test files
		createTestFiles(ctx, t, sandboxInfo, "/mnt/workspace/large", 50)

		// Act
		stateFile, err := state_files.GenerateStateFile(ctx, sandboxInfo, "/mnt/workspace")

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, stateFile)
		assert.Equal(t, 55, len(stateFile.Files)) // 5 from previous + 50 new
	})
}

// Test_WriteAndReadStateFile_localvolume tests the complete write and read cycle for local volume
func Test_WriteAndReadStateFile_localvolume(t *testing.T) {
	skipIfNotConfigured(t)

	ctx := context.Background()
	accountID := types.UUID("test-state-write-read-789")
	sandboxInfo := createTestSandbox(ctx, t, accountID)
	defer cleanupSandbox(ctx, t, sandboxInfo)

	t.Run("Write and read state file to local volume", func(t *testing.T) {
		// Arrange - Create test files and generate state
		createTestFiles(ctx, t, sandboxInfo, "/mnt/workspace", 10)
		originalState, err := state_files.GenerateStateFile(ctx, sandboxInfo, "/mnt/workspace")
		assert.NoError(t, err)

		// Act - Write state file
		err = state_files.WriteLocalStateFile(ctx, sandboxInfo, "/mnt/workspace", originalState)
		assert.NoError(t, err)

		// Read state file back
		readState, err := state_files.ReadLocalStateFile(ctx, sandboxInfo, "/mnt/workspace")

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, readState)
		assert.Equal(t, originalState.Version, readState.Version)
		assert.Equal(t, len(originalState.Files), len(readState.Files))
		assert.NotEqual(t, int64(0), readState.LastSyncedAt)

		// Verify file entries match
		assert.Equal(t, len(originalState.Files), len(readState.Files))
	})

	t.Run("Read non-existent state file returns nil", func(t *testing.T) {
		// Arrange - Use a path where no state file exists
		testPath := "/mnt/workspace/nonexistent"
		cmd := []string{"sh", "-c", fmt.Sprintf("mkdir -p %s", testPath)}
		process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, nil)
		assert.NoError(t, err)
		_, err = process.Wait(ctx)
		assert.NoError(t, err)

		// Act
		stateFile, err := state_files.ReadLocalStateFile(ctx, sandboxInfo, testPath)

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, stateFile) // Should return nil for missing file
	})
}

// Test_CompareStateFiles_syncscenarios tests various sync scenarios using state file comparison
func Test_CompareStateFiles_syncscenarios(t *testing.T) {
	skipIfNotConfigured(t)

	ctx := context.Background()
	accountID := types.UUID("test-state-compare-321")
	sandboxInfo := createTestSandbox(ctx, t, accountID)
	defer cleanupSandbox(ctx, t, sandboxInfo)

	t.Run("Compare identical states - all files skipped", func(t *testing.T) {
		// Arrange - Create files and generate state
		createTestFiles(ctx, t, sandboxInfo, "/mnt/workspace", 5)
		state, err := state_files.GenerateStateFile(ctx, sandboxInfo, "/mnt/workspace")
		assert.NoError(t, err)

		// Act - Compare identical states
		diff := state_files.CompareStateFiles(state, state)

		// Assert
		assert.NotEmpty(t, diff)
		assert.Equal(t, 0, len(diff.FilesToDownload))
		assert.Equal(t, 0, len(diff.FilesToDelete))
		assert.Equal(t, 5, len(diff.FilesToSkip))
	})

	t.Run("Compare empty local with S3 state - all files to download", func(t *testing.T) {
		// Arrange - Create S3 state with files, empty local state
		createTestFiles(ctx, t, sandboxInfo, "/mnt/workspace/s3sim", 3)
		s3State, err := state_files.GenerateStateFile(ctx, sandboxInfo, "/mnt/workspace/s3sim")
		assert.NoError(t, err)

		emptyLocal := &state_files.StateFile{
			Version:      "1.0",
			LastSyncedAt: time.Now().Unix(),
			Files:        []state_files.FileEntry{},
		}

		// Act
		diff := state_files.CompareStateFiles(emptyLocal, s3State)

		// Assert
		assert.NotEmpty(t, diff)
		assert.Equal(t, 3, len(diff.FilesToDownload))
		assert.Equal(t, 0, len(diff.FilesToDelete))
		assert.Equal(t, 0, len(diff.FilesToSkip))
	})

	t.Run("Compare local with empty S3 - all files to delete", func(t *testing.T) {
		// Arrange
		createTestFiles(ctx, t, sandboxInfo, "/mnt/workspace/local", 4)
		localState, err := state_files.GenerateStateFile(ctx, sandboxInfo, "/mnt/workspace/local")
		assert.NoError(t, err)

		emptyS3 := &state_files.StateFile{
			Version:      "1.0",
			LastSyncedAt: time.Now().Unix(),
			Files:        []state_files.FileEntry{},
		}

		// Act
		diff := state_files.CompareStateFiles(localState, emptyS3)

		// Assert
		assert.NotEmpty(t, diff)
		assert.Equal(t, 0, len(diff.FilesToDownload))
		assert.Equal(t, 4, len(diff.FilesToDelete))
		assert.Equal(t, 0, len(diff.FilesToSkip))
	})
}

// Test_CompareStateFiles_modifiedfiles tests detecting modified files by checksum
func Test_CompareStateFiles_modifiedfiles(t *testing.T) {
	skipIfNotConfigured(t)

	ctx := context.Background()
	accountID := types.UUID("test-state-modified-654")
	sandboxInfo := createTestSandbox(ctx, t, accountID)
	defer cleanupSandbox(ctx, t, sandboxInfo)

	t.Run("Detect modified file by checksum difference", func(t *testing.T) {
		// Arrange - Create file and get initial state
		createTestFiles(ctx, t, sandboxInfo, "/mnt/workspace", 1)
		initialState, err := state_files.GenerateStateFile(ctx, sandboxInfo, "/mnt/workspace")
		assert.NoError(t, err)
		assert.Equal(t, 1, len(initialState.Files))

		// Modify the file to change its checksum
		cmd := []string{"sh", "-c", "echo 'Modified content' > /mnt/workspace/testfile0.txt"}
		process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, nil)
		assert.NoError(t, err)
		exitCode, err := process.Wait(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode)

		// Generate new state after modification
		modifiedState, err := state_files.GenerateStateFile(ctx, sandboxInfo, "/mnt/workspace")
		assert.NoError(t, err)

		// Act - Compare states
		diff := state_files.CompareStateFiles(initialState, modifiedState)

		// Assert - File should be flagged for download due to checksum change
		assert.NotEmpty(t, diff)
		assert.Equal(t, 1, len(diff.FilesToDownload))
		assert.Equal(t, 0, len(diff.FilesToDelete))
		assert.Equal(t, 0, len(diff.FilesToSkip))
	})
}

// Test_CheckIfStale_thresholds tests staleness checking with various thresholds
func Test_CheckIfStale_thresholds(t *testing.T) {
	t.Run("Nil state file is stale", func(t *testing.T) {
		// Act
		isStale := state_files.CheckIfStale(nil, 3600)

		// Assert
		assert.Equal(t, true, isStale)
	})

	t.Run("State with zero LastSyncedAt is stale", func(t *testing.T) {
		// Arrange
		state := &state_files.StateFile{
			Version:      "1.0",
			LastSyncedAt: 0,
			Files:        []state_files.FileEntry{},
		}

		// Act
		isStale := state_files.CheckIfStale(state, 3600)

		// Assert
		assert.Equal(t, true, isStale)
	})

	t.Run("Recent state is not stale", func(t *testing.T) {
		// Arrange - State synced just now
		state := &state_files.StateFile{
			Version:      "1.0",
			LastSyncedAt: time.Now().Unix(),
			Files:        []state_files.FileEntry{},
		}

		// Act - Check with 1 hour threshold
		isStale := state_files.CheckIfStale(state, 3600)

		// Assert
		assert.Equal(t, false, isStale)
	})

	t.Run("Old state is stale", func(t *testing.T) {
		// Arrange - State synced 2 hours ago
		twoHoursAgo := time.Now().Add(-2 * time.Hour).Unix()
		state := &state_files.StateFile{
			Version:      "1.0",
			LastSyncedAt: twoHoursAgo,
			Files:        []state_files.FileEntry{},
		}

		// Act - Check with 1 hour threshold
		isStale := state_files.CheckIfStale(state, 3600)

		// Assert
		assert.Equal(t, true, isStale)
	})

	t.Run("State exactly at threshold is not stale", func(t *testing.T) {
		// Arrange - State synced exactly 1 hour ago
		oneHourAgo := time.Now().Add(-1 * time.Hour).Unix()
		state := &state_files.StateFile{
			Version:      "1.0",
			LastSyncedAt: oneHourAgo,
			Files:        []state_files.FileEntry{},
		}

		// Act - Check with 1 hour threshold
		isStale := state_files.CheckIfStale(state, 3600)

		// Assert - Should not be stale (age must EXCEED threshold)
		assert.Equal(t, false, isStale)
	})
}

// Test_CompleteStateCycle_integration tests the full cycle: generate → write → read → compare
func Test_CompleteStateCycle_integration(t *testing.T) {
	skipIfNotConfigured(t)

	ctx := context.Background()
	accountID := types.UUID("test-state-cycle-999")
	sandboxInfo := createTestSandbox(ctx, t, accountID)
	defer cleanupSandbox(ctx, t, sandboxInfo)

	t.Run("Complete state file lifecycle", func(t *testing.T) {
		// Step 1: Generate initial state with files
		createTestFiles(ctx, t, sandboxInfo, "/mnt/workspace", 7)
		initialState, err := state_files.GenerateStateFile(ctx, sandboxInfo, "/mnt/workspace")
		assert.NoError(t, err)
		assert.Equal(t, 7, len(initialState.Files))

		// Step 2: Write state file
		err = state_files.WriteLocalStateFile(ctx, sandboxInfo, "/mnt/workspace", initialState)
		assert.NoError(t, err)

		// Step 3: Read state file back
		readState, err := state_files.ReadLocalStateFile(ctx, sandboxInfo, "/mnt/workspace")
		assert.NoError(t, err)
		assert.NotEmpty(t, readState)
		assert.Equal(t, len(initialState.Files), len(readState.Files))

		// Step 4: Compare states - should be identical (all skipped)
		diff := state_files.CompareStateFiles(readState, initialState)
		assert.NotEmpty(t, diff)
		assert.Equal(t, 0, len(diff.FilesToDownload))
		assert.Equal(t, 0, len(diff.FilesToDelete))
		assert.Equal(t, 7, len(diff.FilesToSkip))

		// Step 5: Modify filesystem - add 3 files, remove 2 files
		createTestFiles(ctx, t, sandboxInfo, "/mnt/workspace/new", 3)

		// Remove 2 existing files
		cmd := []string{"sh", "-c", "rm -f /mnt/workspace/testfile0.txt /mnt/workspace/testfile1.txt"}
		process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, nil)
		assert.NoError(t, err)
		exitCode, err := process.Wait(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, exitCode)

		// Step 6: Generate new state
		newState, err := state_files.GenerateStateFile(ctx, sandboxInfo, "/mnt/workspace")
		assert.NoError(t, err)
		assert.Equal(t, 8, len(newState.Files)) // 7 - 2 + 3 = 8 files

		// Step 7: Compare old state (as if S3) with new state (as if local)
		// From new state perspective: old state is S3
		diff = state_files.CompareStateFiles(newState, readState)
		assert.NotEmpty(t, diff)
		// Should want to download the 2 deleted files from S3
		assert.Equal(t, 2, len(diff.FilesToDownload))
		// Should want to delete the 3 new local files not in S3
		assert.Equal(t, 3, len(diff.FilesToDelete))
		// Should skip the 5 unchanged files
		assert.Equal(t, 5, len(diff.FilesToSkip))
	})
}

// Test_ConcurrentStateReads_threadsafety tests concurrent reads to verify thread safety
func Test_ConcurrentStateReads_threadsafety(t *testing.T) {
	skipIfNotConfigured(t)

	ctx := context.Background()
	accountID := types.UUID("test-state-concurrent-777")
	sandboxInfo := createTestSandbox(ctx, t, accountID)
	defer cleanupSandbox(ctx, t, sandboxInfo)

	t.Run("Concurrent state file reads", func(t *testing.T) {
		// Arrange - Create files and write state
		createTestFiles(ctx, t, sandboxInfo, "/mnt/workspace", 10)
		state, err := state_files.GenerateStateFile(ctx, sandboxInfo, "/mnt/workspace")
		assert.NoError(t, err)
		err = state_files.WriteLocalStateFile(ctx, sandboxInfo, "/mnt/workspace", state)
		assert.NoError(t, err)

		// Act - Perform multiple concurrent reads
		const numReads = 5
		errChan := make(chan error, numReads)
		stateChan := make(chan *state_files.StateFile, numReads)

		for i := 0; i < numReads; i++ {
			go func() {
				readState, err := state_files.ReadLocalStateFile(ctx, sandboxInfo, "/mnt/workspace")
				errChan <- err
				stateChan <- readState
			}()
		}

		// Assert - All reads should succeed with consistent results
		for i := 0; i < numReads; i++ {
			err := <-errChan
			assert.NoError(t, err)

			readState := <-stateChan
			assert.NotEmpty(t, readState)
			assert.Equal(t, 10, len(readState.Files))
			assert.Equal(t, "1.0", readState.Version)
		}
	})
}

// Test_LargeFileCount_performance tests performance with large number of files
func Test_LargeFileCount_performance(t *testing.T) {
	skipIfNotConfigured(t)

	ctx := context.Background()
	accountID := types.UUID("test-state-large-888")
	sandboxInfo := createTestSandbox(ctx, t, accountID)
	defer cleanupSandbox(ctx, t, sandboxInfo)

	t.Run("Generate state with 100 files", func(t *testing.T) {
		// Arrange
		createTestFiles(ctx, t, sandboxInfo, "/mnt/workspace/batch1", 50)
		createTestFiles(ctx, t, sandboxInfo, "/mnt/workspace/batch2", 50)

		// Act
		startTime := time.Now()
		state, err := state_files.GenerateStateFile(ctx, sandboxInfo, "/mnt/workspace")
		duration := time.Since(startTime)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, state)
		assert.Equal(t, 100, len(state.Files))

		// Log performance info
		t.Logf("Generated state for 100 files in %v", duration)
	})

	t.Run("Write and read state with 100 files", func(t *testing.T) {
		// Arrange
		state, err := state_files.GenerateStateFile(ctx, sandboxInfo, "/mnt/workspace")
		assert.NoError(t, err)

		// Act - Write
		startWrite := time.Now()
		err = state_files.WriteLocalStateFile(ctx, sandboxInfo, "/mnt/workspace", state)
		writeTime := time.Since(startWrite)
		assert.NoError(t, err)

		// Act - Read
		startRead := time.Now()
		readState, err := state_files.ReadLocalStateFile(ctx, sandboxInfo, "/mnt/workspace")
		readTime := time.Since(startRead)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, readState)
		assert.Equal(t, 100, len(readState.Files))

		// Log performance info
		t.Logf("Write time for 100 files: %v", writeTime)
		t.Logf("Read time for 100 files: %v", readTime)
	})
}
