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

// WriteLocalStateFile writes the .sandbox-state file to the local volume atomically.
// It updates the LastSyncedAt timestamp before writing and uses a temporary file
// followed by an atomic rename to prevent corruption if the process is interrupted.
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

	// Marshal to JSON with indentation for readability
	data, err := json.MarshalIndent(stateFile, "", "  ")
	if err != nil {
		return errors.Wrapf(err, "failed to marshal state file to JSON")
	}

	// Build file paths
	finalPath := fmt.Sprintf("%s/%s", volumePath, StateFileName)
	tempPath := fmt.Sprintf("%s/%s.tmp", volumePath, StateFileName)

	// Escape single quotes in JSON for shell command
	jsonContent := strings.ReplaceAll(string(data), "'", "'\\''")

	// Build atomic write command: write to temp file, then rename
	// Using printf instead of echo to handle special characters properly
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
// Similar to WriteLocalStateFile but writes to S3 mount location.
// S3 mounts support atomic renames, ensuring consistency.
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

	// Marshal to JSON with indentation for readability
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
// It executes a find command to list all files, calculates MD5 checksums for each file,
// retrieves file sizes and modification times, and builds the StateFile struct.
// The .sandbox-state file itself is excluded from the generated state (requirement 7.6).
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
	// Using -type f to get only files, not directories
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
		// Format: size|mtime|md5sum
		metaCmd := []string{
			"sh", "-c",
			fmt.Sprintf("stat -c '%%s|%%Y' %s && md5sum %s | cut -d' ' -f1", fullPath, fullPath),
		}

		metaProcess, err := sandboxInfo.Sandbox.Exec(ctx, metaCmd, nil)
		if err != nil {
			// Log warning but continue with other files
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
			// Skip files that can't be read
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
		Version:      CurrentVersion,
		LastSyncedAt: time.Now().Unix(),
		Files:        fileEntries,
	}

	return stateFile, nil
}

// parseFilePaths parses the output of find command into a slice of file paths.
// It filters out empty lines and trims whitespace.
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
