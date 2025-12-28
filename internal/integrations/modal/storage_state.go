package modal

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/griffnb/core/lib/log"
	modal "github.com/modal-labs/libmodal/modal-go"
	"github.com/pkg/errors"
)

// stateFileName is the name of the state file (duplicated from state_files to avoid import cycle)
const stateFileName = ".sandbox-state"

// stateFileVersion is the current state file schema version
const stateFileVersion = "1.0"

// StateFile represents the .sandbox-state file format used to track file synchronization
// state between local volumes and S3. This is duplicated from state_files package to avoid import cycle.
type StateFile struct {
	Version      string      `json:"version"`        // Schema version (e.g., "1.0")
	LastSyncedAt int64       `json:"last_synced_at"` // Unix timestamp of last sync
	Files        []FileEntry `json:"files"`          // Array of tracked files
}

// FileEntry represents a single file tracked in the state file.
// Duplicated from state_files package to avoid import cycle.
type FileEntry struct {
	Path       string `json:"path"`        // Relative path from volume root
	Checksum   string `json:"checksum"`    // MD5 hash (matches S3 ETag format)
	Size       int64  `json:"size"`        // File size in bytes
	ModifiedAt int64  `json:"modified_at"` // Unix timestamp of last modification
}

// StateDiff represents the differences between local and S3 state files.
// Duplicated from state_files package to avoid import cycle.
type StateDiff struct {
	FilesToDownload []FileEntry // Files to download from S3 (new or updated)
	FilesToDelete   []string    // Local file paths to delete (removed from S3)
	FilesToSkip     []FileEntry // Files that match (no action needed)
}

// compareStateFiles compares local and S3 state files to determine which sync actions are needed.
// This is duplicated from state_files package to avoid import cycle.
// It identifies files to download (new or updated), files to skip (unchanged), and files to delete (removed from S3).
func compareStateFiles(localState *StateFile, s3State *StateFile) *StateDiff {
	// Initialize result
	diff := &StateDiff{
		FilesToDownload: []FileEntry{},
		FilesToDelete:   []string{},
		FilesToSkip:     []FileEntry{},
	}

	// Create maps for O(1) lookups
	localFiles := make(map[string]FileEntry)
	s3Files := make(map[string]FileEntry)

	// Build local file map
	if localState != nil {
		for _, f := range localState.Files {
			localFiles[f.Path] = f
		}
	}

	// Build S3 file map
	if s3State != nil {
		for _, f := range s3State.Files {
			s3Files[f.Path] = f
		}
	}

	// Process S3 files: determine if download or skip
	for path, s3File := range s3Files {
		localFile, existsLocal := localFiles[path]

		if existsLocal {
			// File exists in both states - compare checksums
			if localFile.Checksum == s3File.Checksum {
				// Matching checksum -> skip
				diff.FilesToSkip = append(diff.FilesToSkip, s3File)
			} else {
				// Different checksum -> download
				diff.FilesToDownload = append(diff.FilesToDownload, s3File)
			}
		} else {
			// File only in S3 -> download
			diff.FilesToDownload = append(diff.FilesToDownload, s3File)
		}
	}

	// Process local files: identify files to delete
	for path := range localFiles {
		if _, existsS3 := s3Files[path]; !existsS3 {
			// File only in local -> delete
			diff.FilesToDelete = append(diff.FilesToDelete, path)
		}
	}

	return diff
}

// parseStateFile parses JSON bytes into a StateFile struct.
// Duplicated from state_files to avoid import cycle.
func parseStateFile(data []byte) (*StateFile, error) {
	if len(data) == 0 {
		return nil, errors.New("state file data is empty")
	}

	var stateFile StateFile
	err := json.Unmarshal(data, &stateFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal state file JSON")
	}

	// Validate version compatibility
	if stateFile.Version != stateFileVersion {
		return nil, errors.Errorf(
			"incompatible state file version: got %s, expected %s",
			stateFile.Version,
			stateFileVersion,
		)
	}

	return &stateFile, nil
}

// readLocalStateFile reads the .sandbox-state file from the local volume in the sandbox.
// This is a modal-specific wrapper around state file reading that executes commands in the sandbox.
// Returns nil without error if the file doesn't exist (treat as empty state).
func readLocalStateFile(
	ctx context.Context,
	sandboxInfo *SandboxInfo,
	volumePath string,
) (*StateFile, error) {
	if sandboxInfo == nil || sandboxInfo.Sandbox == nil {
		return nil, errors.New("sandboxInfo or sandbox cannot be nil")
	}
	if volumePath == "" {
		return nil, errors.New("volumePath cannot be empty")
	}

	// Build file path
	filePath := fmt.Sprintf("%s/%s", volumePath, stateFileName)

	// Execute cat command to read file
	cmd := []string{
		"sh", "-c",
		fmt.Sprintf("test -f %s && cat %s || echo 'FILE_NOT_FOUND'", filePath, filePath),
	}

	process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute cat command for state file at %s", filePath)
	}

	// Read stdout
	var output bytes.Buffer
	scanner := bufio.NewScanner(process.Stdout)
	for scanner.Scan() {
		output.Write(scanner.Bytes())
		output.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrapf(err, "failed to read process output")
	}

	// Wait for command to complete
	exitCode, err := process.Wait(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to wait for cat command")
	}

	if exitCode != 0 {
		return nil, errors.Errorf("cat command failed with exit code %d", exitCode)
	}

	// Check if file doesn't exist
	outputStr := strings.TrimSpace(output.String())
	if outputStr == "FILE_NOT_FOUND" {
		return nil, nil
	}

	// Parse the state file
	stateFile, err := parseStateFile([]byte(outputStr))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse state file from %s", filePath)
	}

	return stateFile, nil
}

// readS3StateFile reads the .sandbox-state file from the S3 mount path in the sandbox.
// Returns nil without error if the file doesn't exist.
func readS3StateFile(
	ctx context.Context,
	sandboxInfo *SandboxInfo,
	s3MountPath string,
) (*StateFile, error) {
	if sandboxInfo == nil || sandboxInfo.Sandbox == nil {
		return nil, errors.New("sandboxInfo or sandbox cannot be nil")
	}
	if s3MountPath == "" {
		return nil, errors.New("s3MountPath cannot be empty")
	}

	// Build file path for S3 mount
	filePath := fmt.Sprintf("%s/%s", s3MountPath, stateFileName)

	// Execute cat command
	cmd := []string{
		"sh", "-c",
		fmt.Sprintf("test -f %s && cat %s || echo 'FILE_NOT_FOUND'", filePath, filePath),
	}

	process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute cat command for S3 state file at %s", filePath)
	}

	// Read stdout
	var output bytes.Buffer
	scanner := bufio.NewScanner(process.Stdout)
	for scanner.Scan() {
		output.Write(scanner.Bytes())
		output.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrapf(err, "failed to read S3 process output")
	}

	// Wait for command to complete
	exitCode, err := process.Wait(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to wait for S3 cat command")
	}

	if exitCode != 0 {
		return nil, errors.Errorf("S3 cat command failed with exit code %d", exitCode)
	}

	// Check if file doesn't exist
	outputStr := strings.TrimSpace(output.String())
	if outputStr == "FILE_NOT_FOUND" {
		return nil, nil
	}

	// Parse the state file
	stateFile, err := parseStateFile([]byte(outputStr))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse S3 state file from %s", filePath)
	}

	return stateFile, nil
}

// writeLocalStateFile writes the .sandbox-state file to the local volume atomically.
func writeLocalStateFile(
	ctx context.Context,
	sandboxInfo *SandboxInfo,
	volumePath string,
	stateFile *StateFile,
) error {
	if sandboxInfo == nil || sandboxInfo.Sandbox == nil {
		return errors.New("sandboxInfo or sandbox cannot be nil")
	}
	if volumePath == "" {
		return errors.New("volumePath cannot be empty")
	}
	if stateFile == nil {
		return errors.New("stateFile cannot be nil")
	}

	// Update LastSyncedAt to current time
	stateFile.LastSyncedAt = time.Now().Unix()

	// Marshal to JSON
	data, err := json.MarshalIndent(stateFile, "", "  ")
	if err != nil {
		return errors.Wrapf(err, "failed to marshal state file to JSON")
	}

	// Build file paths
	finalPath := fmt.Sprintf("%s/%s", volumePath, stateFileName)
	tempPath := fmt.Sprintf("%s/%s.tmp", volumePath, stateFileName)

	// Escape single quotes in JSON for shell command
	jsonContent := strings.ReplaceAll(string(data), "'", "'\\''")

	// Build atomic write command
	cmd := []string{
		"sh", "-c",
		fmt.Sprintf("printf '%%s' '%s' > %s && mv %s %s", jsonContent, tempPath, tempPath, finalPath),
	}

	// Execute command
	process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to execute write command for state file at %s", finalPath)
	}

	// Wait for command to complete
	exitCode, err := process.Wait(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to wait for write command")
	}

	if exitCode != 0 {
		// Read error output
		var stderr bytes.Buffer
		scanner := bufio.NewScanner(process.Stderr)
		for scanner.Scan() {
			stderr.Write(scanner.Bytes())
			stderr.WriteByte('\n')
		}
		return errors.Errorf("write command failed with exit code %d: %s", exitCode, stderr.String())
	}

	return nil
}

// writeS3StateFile writes the .sandbox-state file to the S3 mount path atomically.
func writeS3StateFile(
	ctx context.Context,
	sandboxInfo *SandboxInfo,
	s3MountPath string,
	stateFile *StateFile,
) error {
	if sandboxInfo == nil || sandboxInfo.Sandbox == nil {
		return errors.New("sandboxInfo or sandbox cannot be nil")
	}
	if s3MountPath == "" {
		return errors.New("s3MountPath cannot be empty")
	}
	if stateFile == nil {
		return errors.New("stateFile cannot be nil")
	}

	// Update LastSyncedAt to current time
	stateFile.LastSyncedAt = time.Now().Unix()

	// Marshal to JSON
	data, err := json.MarshalIndent(stateFile, "", "  ")
	if err != nil {
		return errors.Wrapf(err, "failed to marshal state file to JSON for S3")
	}

	// Build file paths for S3
	finalPath := fmt.Sprintf("%s/%s", s3MountPath, stateFileName)
	tempPath := fmt.Sprintf("%s/%s.tmp", s3MountPath, stateFileName)

	// Escape single quotes in JSON for shell command
	jsonContent := strings.ReplaceAll(string(data), "'", "'\\''")

	// Build atomic write command
	cmd := []string{
		"sh", "-c",
		fmt.Sprintf("printf '%%s' '%s' > %s && mv %s %s", jsonContent, tempPath, tempPath, finalPath),
	}

	// Execute command
	process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to execute write command for S3 state file at %s", finalPath)
	}

	// Wait for command to complete
	exitCode, err := process.Wait(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to wait for S3 write command")
	}

	if exitCode != 0 {
		// Read error output
		var stderr bytes.Buffer
		scanner := bufio.NewScanner(process.Stderr)
		for scanner.Scan() {
			stderr.Write(scanner.Bytes())
			stderr.WriteByte('\n')
		}
		return errors.Errorf("S3 write command failed with exit code %d: %s", exitCode, stderr.String())
	}

	return nil
}

// generateStateFile scans a directory and generates a StateFile with checksums and metadata.
func generateStateFile(
	ctx context.Context,
	sandboxInfo *SandboxInfo,
	directoryPath string,
) (*StateFile, error) {
	if sandboxInfo == nil || sandboxInfo.Sandbox == nil {
		return nil, errors.New("sandboxInfo or sandbox cannot be nil")
	}
	if directoryPath == "" {
		return nil, errors.New("directoryPath cannot be empty")
	}

	// Build find command to list all files (excluding .sandbox-state)
	cmd := []string{
		"sh", "-c",
		fmt.Sprintf("cd %s && find . -type f ! -name '%s' -print", directoryPath, stateFileName),
	}

	// Execute find command
	process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute find command in %s", directoryPath)
	}

	// Read file list from stdout
	var output bytes.Buffer
	scanner := bufio.NewScanner(process.Stdout)
	for scanner.Scan() {
		output.Write(scanner.Bytes())
		output.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrapf(err, "failed to read find command output")
	}

	// Wait for command to complete
	exitCode, err := process.Wait(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to wait for find command")
	}

	if exitCode != 0 {
		return nil, errors.Errorf("find command failed with exit code %d", exitCode)
	}

	// Parse file list
	filePaths := parseFilePaths(output.String())

	// Generate file entries with checksums and metadata
	fileEntries := []FileEntry{}
	for _, relPath := range filePaths {
		// Remove leading "./" if present
		cleanPath := strings.TrimPrefix(relPath, "./")

		// Build full path for stat and checksum commands
		fullPath := fmt.Sprintf("%s/%s", directoryPath, cleanPath)

		// Get file metadata and checksum in one command for efficiency
		metaCmd := []string{
			"sh", "-c",
			fmt.Sprintf("stat -c '%%s|%%Y' %s && md5sum %s | cut -d' ' -f1", fullPath, fullPath),
		}

		metaProcess, err := sandboxInfo.Sandbox.Exec(ctx, metaCmd, nil)
		if err != nil {
			continue
		}

		// Read metadata output
		var metaOutput bytes.Buffer
		metaScanner := bufio.NewScanner(metaProcess.Stdout)
		for metaScanner.Scan() {
			metaOutput.Write(metaScanner.Bytes())
			metaOutput.WriteByte('\n')
		}

		exitCode, err := metaProcess.Wait(ctx)
		if err != nil || exitCode != 0 {
			continue
		}

		// Parse metadata output
		lines := strings.Split(strings.TrimSpace(metaOutput.String()), "\n")
		if len(lines) < 2 {
			continue
		}

		// Parse size and mtime from first line
		parts := strings.Split(lines[0], "|")
		if len(parts) < 2 {
			continue
		}

		size, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			continue
		}

		modifiedAt, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}

		// MD5 checksum is on second line
		checksum := strings.TrimSpace(lines[1])

		// Create file entry
		fileEntries = append(fileEntries, FileEntry{
			Path:       cleanPath,
			Checksum:   checksum,
			Size:       size,
			ModifiedAt: modifiedAt,
		})
	}

	// Build and return StateFile
	stateFile := &StateFile{
		Version:      stateFileVersion,
		LastSyncedAt: time.Now().Unix(),
		Files:        fileEntries,
	}

	return stateFile, nil
}

// parseFilePaths parses the output of find command into a slice of file paths.
func parseFilePaths(output string) []string {
	paths := []string{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		paths = append(paths, line)
	}

	return paths
}

// InitVolumeFromS3WithState performs intelligent incremental sync from S3 to local volume using state files.
// It reads both local and S3 state files, compares them to determine which files need to be
// downloaded, updated, or deleted, then executes only the necessary sync actions.
//
// This implements state-file-based synchronization per design phase 3.2 and satisfies:
// - Requirement 3.1-3.12: State file comparison and incremental sync
// - Requirement 2.1-2.5: Cold start synchronization with state tracking
// - Requirement 4.4-4.5: Full sync for new sandboxes, incremental for existing
//
// The function performs the following steps:
// 1. Read local .sandbox-state (treat as empty if missing)
// 2. Read S3 .sandbox-state (generate if missing)
// 3. Compare states to determine sync actions (download, delete, skip)
// 4. Execute sync actions efficiently
// 5. Update local .sandbox-state with current state
//
// Returns SyncStats with files downloaded, deleted, skipped, bytes transferred, and duration.
func (c *APIClient) InitVolumeFromS3WithState(
	ctx context.Context,
	sandboxInfo *SandboxInfo,
) (*SyncStats, error) {
	if sandboxInfo == nil || sandboxInfo.Sandbox == nil {
		return nil, errors.New("sandboxInfo or sandbox is nil")
	}
	if sandboxInfo.Config.S3Config == nil {
		return nil, errors.New("S3Config is required for InitVolumeFromS3WithState")
	}

	startTime := time.Now()

	// Step 1: Read local .sandbox-state (nil if doesn't exist - new sandbox)
	localState, err := readLocalStateFile(ctx, sandboxInfo, sandboxInfo.Config.VolumeMountPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read local state file")
	}

	// Log cold start decision (Requirement 10.6 - OnColdStart vs incremental sync)
	if localState == nil {
		log.Info(fmt.Sprintf("[S3 Sync] OnColdStart decision: FULL SYNC (no local state file) sandbox=%s",
			sandboxInfo.SandboxID))
	} else {
		log.Info(fmt.Sprintf("[S3 Sync] OnColdStart decision: INCREMENTAL SYNC (local state exists, last_sync=%d) sandbox=%s",
			localState.LastSyncedAt,
			sandboxInfo.SandboxID))
	}

	// Step 2: Read S3 .sandbox-state (or generate if missing)
	s3State, err := readS3StateFile(ctx, sandboxInfo, sandboxInfo.Config.S3Config.MountPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read S3 state file")
	}

	// If S3 state doesn't exist, generate it from S3 directory scan
	if s3State == nil {
		log.Info(fmt.Sprintf("[S3 Sync] S3 state file not found, generating from directory scan sandbox=%s",
			sandboxInfo.SandboxID))
		s3State, err = generateStateFile(ctx, sandboxInfo, sandboxInfo.Config.S3Config.MountPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to generate S3 state file")
		}
	}

	// Step 3: Compare states to determine sync actions
	diff := compareStateFiles(localState, s3State)

	// Step 4: Execute sync actions (download, delete)
	stats, err := c.executeSyncActions(ctx, sandboxInfo, diff)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute sync actions")
	}

	// Step 5: Update local .sandbox-state with S3 state (we're now in sync)
	err = writeLocalStateFile(ctx, sandboxInfo, sandboxInfo.Config.VolumeMountPath, s3State)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to write local state file")
	}

	// Set final duration
	stats.Duration = time.Since(startTime)

	// Log sync statistics (Requirement 10.6)
	log.Info(fmt.Sprintf("[S3 Sync] InitVolumeFromS3WithState completed: downloaded=%d deleted=%d skipped=%d bytes=%d duration=%v sandbox=%s",
		stats.FilesDownloaded,
		stats.FilesDeleted,
		stats.FilesSkipped,
		stats.BytesTransferred,
		stats.Duration,
		sandboxInfo.SandboxID))

	return stats, nil
}

// SyncVolumeToS3WithState syncs the local volume to S3 with timestamp versioning and state file updates.
// It creates a new timestamped version in S3 (immutable snapshots), then updates state files in both
// local volume and S3 to track the sync.
//
// This implements state-file-based upload per design phase 3.2 and satisfies:
// - Requirement 5.2: Sync with new timestamp version
// - Requirement 5.3: Update state files in local and S3
// - Requirement 3.11: Preserve exact timestamps during sync
//
// The function performs the following steps:
// 1. Generate current state from local volume
// 2. Use AWS CLI sync to upload to timestamped S3 path
// 3. Parse AWS CLI output for stats
// 4. Write .sandbox-state to both local volume and S3
//
// Returns SyncStats with files uploaded, bytes transferred, and duration.
func (c *APIClient) SyncVolumeToS3WithState(
	ctx context.Context,
	sandboxInfo *SandboxInfo,
) (*SyncStats, error) {
	if sandboxInfo == nil || sandboxInfo.Sandbox == nil {
		return nil, errors.New("sandboxInfo or sandbox is nil")
	}
	if sandboxInfo.Config.S3Config == nil {
		return nil, errors.New("S3Config is required for SyncVolumeToS3WithState")
	}

	startTime := time.Now()

	// Step 1: Generate current state from local volume
	localState, err := generateStateFile(ctx, sandboxInfo, sandboxInfo.Config.VolumeMountPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate local state file")
	}

	// Step 2: Use AWS CLI sync to upload to timestamped S3 path
	timestamp := time.Now().Unix()
	s3Path := fmt.Sprintf("s3://%s/docs/%s/%d/",
		sandboxInfo.Config.S3Config.BucketName,
		sandboxInfo.Config.AccountID,
		timestamp,
	)

	stats, err := c.executeAWSSync(ctx, sandboxInfo, sandboxInfo.Config.VolumeMountPath, s3Path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute AWS sync")
	}

	// Step 3: Write .sandbox-state to both local and S3
	// Note: For S3, we write to the timestamped path that was just created
	err = writeLocalStateFile(ctx, sandboxInfo, sandboxInfo.Config.VolumeMountPath, localState)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to write local state file")
	}

	// Write to S3 mount path as well (this will be in the timestamped location)
	err = writeS3StateFile(ctx, sandboxInfo, sandboxInfo.Config.S3Config.MountPath, localState)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to write S3 state file")
	}

	// Set final duration
	stats.Duration = time.Since(startTime)

	// Log sync statistics (Requirement 10.6)
	log.Info(fmt.Sprintf("[S3 Sync] SyncVolumeToS3WithState completed: uploaded=%d bytes=%d duration=%v timestamp=%d sandbox=%s",
		stats.FilesDownloaded, // Using FilesDownloaded for upload count
		stats.BytesTransferred,
		stats.Duration,
		timestamp,
		sandboxInfo.SandboxID))

	return stats, nil
}

// executeSyncActions performs the actual file operations based on the state diff.
// It downloads files from S3 to local, deletes local files not in S3, and tracks stats.
//
// This satisfies:
// - Requirement 3.4: Download files in S3 but not local
// - Requirement 3.6: Download files with different checksums
// - Requirement 3.7: Delete files in local but not S3
// - Requirement 3.10: Return detailed sync stats
func (c *APIClient) executeSyncActions(
	ctx context.Context,
	sandboxInfo *SandboxInfo,
	diff *StateDiff,
) (*SyncStats, error) {
	stats := &SyncStats{
		FilesDownloaded:  0,
		FilesDeleted:     0,
		FilesSkipped:     len(diff.FilesToSkip),
		BytesTransferred: 0,
		Duration:         0,
		Errors:           []error{},
	}

	// Build commands to download files from S3 mount to local volume
	// We use cp with preserve flags to maintain timestamps and permissions
	for _, file := range diff.FilesToDownload {
		sourcePath := fmt.Sprintf("%s/%s", sandboxInfo.Config.S3Config.MountPath, file.Path)
		destPath := fmt.Sprintf("%s/%s", sandboxInfo.Config.VolumeMountPath, file.Path)

		// Create parent directory if needed
		// Extract directory from file path
		dirPath := destPath
		if lastSlash := strings.LastIndex(dirPath, "/"); lastSlash > 0 {
			dirPath = dirPath[:lastSlash]
		}

		// Create directory
		mkdirCmd := []string{
			"sh", "-c",
			fmt.Sprintf("mkdir -p %s", dirPath),
		}
		mkdirProcess, err := sandboxInfo.Sandbox.Exec(ctx, mkdirCmd, nil)
		if err != nil {
			stats.Errors = append(stats.Errors, errors.Wrapf(err, "failed to create directory for %s", file.Path))
			continue
		}
		_, _ = mkdirProcess.Wait(ctx)

		// Copy file from S3 mount to local volume with preserved timestamps
		cpCmd := []string{
			"sh", "-c",
			fmt.Sprintf("cp -p %s %s 2>&1", sourcePath, destPath),
		}

		cpProcess, err := sandboxInfo.Sandbox.Exec(ctx, cpCmd, nil)
		if err != nil {
			stats.Errors = append(stats.Errors, errors.Wrapf(err, "failed to copy file %s", file.Path))
			continue
		}

		// Read output
		var output bytes.Buffer
		scanner := bufio.NewScanner(cpProcess.Stdout)
		for scanner.Scan() {
			output.Write(scanner.Bytes())
			output.WriteByte('\n')
		}

		exitCode, err := cpProcess.Wait(ctx)
		if err != nil || exitCode != 0 {
			errMsg := fmt.Sprintf("copy failed for %s: exit code %d, output: %s", file.Path, exitCode, output.String())
			stats.Errors = append(stats.Errors, errors.New(errMsg))
			continue
		}

		// Successfully downloaded
		stats.FilesDownloaded++
		stats.BytesTransferred += file.Size
	}

	// Build commands to delete local files not in S3
	for _, filePath := range diff.FilesToDelete {
		localPath := fmt.Sprintf("%s/%s", sandboxInfo.Config.VolumeMountPath, filePath)

		// Delete file
		rmCmd := []string{
			"sh", "-c",
			fmt.Sprintf("rm -f %s 2>&1", localPath),
		}

		rmProcess, err := sandboxInfo.Sandbox.Exec(ctx, rmCmd, nil)
		if err != nil {
			stats.Errors = append(stats.Errors, errors.Wrapf(err, "failed to delete file %s", filePath))
			continue
		}

		exitCode, err := rmProcess.Wait(ctx)
		if err != nil || exitCode != 0 {
			stats.Errors = append(stats.Errors, errors.Errorf("delete failed for %s: exit code %d", filePath, exitCode))
			continue
		}

		// Successfully deleted
		stats.FilesDeleted++
	}

	// Return error if we had any errors during sync
	if len(stats.Errors) > 0 {
		return stats, errors.Errorf("sync completed with %d errors", len(stats.Errors))
	}

	return stats, nil
}

// executeAWSSync runs AWS CLI sync command to upload files to S3.
// It uses AWS CLI with --exact-timestamps flag to preserve modification times.
//
// This satisfies:
// - Requirement 3.11: Use --exact-timestamps flag
// - Requirement 5.2: Sync to timestamped S3 path
func (c *APIClient) executeAWSSync(
	ctx context.Context,
	sandboxInfo *SandboxInfo,
	sourcePath string,
	s3Path string,
) (*SyncStats, error) {
	if sandboxInfo == nil || sandboxInfo.Sandbox == nil {
		return nil, errors.New("sandboxInfo or sandbox is nil")
	}

	startTime := time.Now()

	// Retrieve AWS credentials from Modal secrets
	secret, err := c.client.Secrets.FromName(ctx, sandboxInfo.Config.S3Config.SecretName, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get secret %s for AWS sync", sandboxInfo.Config.S3Config.SecretName)
	}

	// Build AWS CLI sync command with exact timestamps
	syncCmd := fmt.Sprintf("aws s3 sync %s %s --exact-timestamps 2>&1", sourcePath, s3Path)
	cmd := []string{
		"runuser", "-u", ClaudeUserName, "--",
		"sh", "-c", syncCmd,
	}

	// Execute with AWS credentials
	process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, &modal.SandboxExecParams{
		Secrets: []*modal.Secret{secret},
		Workdir: "/",
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute AWS sync command")
	}

	// Read output
	var output bytes.Buffer
	scanner := bufio.NewScanner(process.Stdout)
	for scanner.Scan() {
		output.Write(scanner.Bytes())
		output.WriteByte('\n')
	}

	// Wait for process to complete
	exitCode, err := process.Wait(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to wait for AWS sync process")
	}

	// Check exit code
	if exitCode != 0 {
		return nil, errors.Errorf("AWS sync failed with exit code %d: %s", exitCode, output.String())
	}

	// Parse output to estimate stats
	// AWS CLI sync outputs lines like "upload: file.txt to s3://..."
	// We count upload lines to estimate files uploaded
	filesUploaded := 0
	outputStr := output.String()
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "upload:") || strings.Contains(line, "copy:") {
			filesUploaded++
		}
	}

	duration := time.Since(startTime)

	stats := &SyncStats{
		FilesDownloaded:  filesUploaded, // Using FilesDownloaded to track uploaded count
		FilesDeleted:     0,
		FilesSkipped:     0,
		BytesTransferred: 0, // AWS CLI doesn't provide bytes in output easily
		Duration:         duration,
		Errors:           []error{},
	}

	return stats, nil
}
