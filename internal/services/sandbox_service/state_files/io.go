package state_files

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/pkg/errors"
)

// ReadLocalStateFile reads the .sandbox-state file from the local volume in the sandbox.
// Returns nil without error if the file doesn't exist (treat as empty state).
func ReadLocalStateFile(
	ctx context.Context,
	sandboxInfo *modal.SandboxInfo,
	volumePath string,
) (*StateFile, error) {
	if sandboxInfo == nil || sandboxInfo.Sandbox == nil {
		return nil, errors.New("sandboxInfo or sandbox cannot be nil")
	}
	if volumePath == "" {
		return nil, errors.New("volumePath cannot be empty")
	}

	// Build file path
	filePath := fmt.Sprintf("%s/%s", volumePath, StateFileName)

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

// ReadS3StateFile reads the .sandbox-state file from the S3 mount path in the sandbox.
// Returns nil without error if the file doesn't exist.
func ReadS3StateFile(
	ctx context.Context,
	sandboxInfo *modal.SandboxInfo,
	s3MountPath string,
) (*StateFile, error) {
	if sandboxInfo == nil || sandboxInfo.Sandbox == nil {
		return nil, errors.New("sandboxInfo or sandbox cannot be nil")
	}
	if s3MountPath == "" {
		return nil, errors.New("s3MountPath cannot be empty")
	}

	// Build file path for S3 mount
	filePath := fmt.Sprintf("%s/%s", s3MountPath, StateFileName)

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

// WriteLocalStateFile writes the .sandbox-state file to the local volume atomically.
func WriteLocalStateFile(
	ctx context.Context,
	sandboxInfo *modal.SandboxInfo,
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
	finalPath := fmt.Sprintf("%s/%s", volumePath, StateFileName)
	tempPath := fmt.Sprintf("%s/%s.tmp", volumePath, StateFileName)

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

// WriteS3StateFile writes the .sandbox-state file to the S3 mount path atomically.
func WriteS3StateFile(
	ctx context.Context,
	sandboxInfo *modal.SandboxInfo,
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
	finalPath := fmt.Sprintf("%s/%s", s3MountPath, StateFileName)
	tempPath := fmt.Sprintf("%s/%s.tmp", s3MountPath, StateFileName)

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

// GenerateStateFile scans a directory and generates a StateFile with checksums and metadata.
func GenerateStateFile(
	ctx context.Context,
	sandboxInfo *modal.SandboxInfo,
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
		fmt.Sprintf("cd %s && find . -type f ! -name '%s' -print", directoryPath, StateFileName),
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
		Version:      StateFileVersion,
		LastSyncedAt: time.Now().Unix(),
		Files:        fileEntries,
	}

	return stateFile, nil
}

// parseStateFile parses JSON bytes into a StateFile struct.
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
	if stateFile.Version != StateFileVersion {
		return nil, errors.Errorf(
			"incompatible state file version: got %s, expected %s",
			stateFile.Version,
			StateFileVersion,
		)
	}

	return &stateFile, nil
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
