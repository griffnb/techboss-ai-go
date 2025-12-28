package state_files

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/pkg/errors"
)

const (
	// CurrentVersion is the current state file schema version
	CurrentVersion = "1.0"
	// StateFileName is the name of the state file
	StateFileName = ".sandbox-state"
)

// ReadLocalStateFile reads the .sandbox-state file from the local volume in the sandbox.
// It executes a cat command to read the file contents and parses the JSON.
// Returns nil without error if the file doesn't exist (treat as empty state).
// Returns an error if the file exists but is corrupted or cannot be parsed.
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
	// Use test command first to check if file exists
	cmd := []string{
		"sh", "-c",
		fmt.Sprintf("test -f %s && cat %s || echo 'FILE_NOT_FOUND'", filePath, filePath),
	}

	process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute cat command for state file at %s", filePath)
	}

	// Read stdout using scanner
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
	stateFile, err := ParseStateFile([]byte(outputStr))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse state file from %s", filePath)
	}

	return stateFile, nil
}

// ReadS3StateFile reads the .sandbox-state file from the S3 mount path in the sandbox.
// Similar to ReadLocalStateFile but reads from S3 mount location.
// Returns nil without error if the file doesn't exist.
// Returns an error if the file exists but is corrupted.
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

	// Execute cat command (same logic as local)
	cmd := []string{
		"sh", "-c",
		fmt.Sprintf("test -f %s && cat %s || echo 'FILE_NOT_FOUND'", filePath, filePath),
	}

	process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute cat command for S3 state file at %s", filePath)
	}

	// Read stdout using scanner
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
	stateFile, err := ParseStateFile([]byte(outputStr))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse S3 state file from %s", filePath)
	}

	return stateFile, nil
}

// ParseStateFile parses JSON bytes into a StateFile struct.
// It validates the version for compatibility and returns an error if the version
// is not supported or if the JSON is malformed.
func ParseStateFile(data []byte) (*StateFile, error) {
	if len(data) == 0 {
		return nil, errors.New("state file data is empty")
	}

	var stateFile StateFile
	err := json.Unmarshal(data, &stateFile)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal state file JSON")
	}

	// Validate version compatibility
	// Currently we only support version 1.0
	// Future versions (e.g., 2.0) would be incompatible
	if stateFile.Version != CurrentVersion {
		return nil, errors.Errorf(
			"incompatible state file version: got %s, expected %s",
			stateFile.Version,
			CurrentVersion,
		)
	}

	return &stateFile, nil
}
