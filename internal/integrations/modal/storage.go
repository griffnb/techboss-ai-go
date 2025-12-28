package modal

// https://github.com/awslabs/mountpoint-s3/blob/main/doc/SEMANTICS.md
// S3 mount uses mountpoint-s3 which provides POSIX-like semantics for S3 buckets.

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/griffnb/core/lib/types"
	"github.com/modal-labs/libmodal/modal-go"
	"github.com/pkg/errors"
)

// SyncStats contains statistics about S3 sync operations.
// It tracks the number of files downloaded, deleted, skipped, bytes transferred,
// operation duration, and any non-fatal errors encountered during the sync.
// This aligns with state-file-based sync per design phase 3.1.
type SyncStats struct {
	FilesDownloaded  int           // Number of files downloaded from S3
	FilesDeleted     int           // Number of local files deleted
	FilesSkipped     int           // Number of files unchanged (skipped)
	BytesTransferred int64         // Total bytes transferred
	Duration         time.Duration // Time taken
	Errors           []error       // Any non-fatal errors
}

// InitVolumeFromS3 copies files from S3 bucket to volume on sandbox startup.
// It uses the cp command to recursively copy files from the S3 mount to the volume.
// This is typically called after sandbox creation to restore previous work state.
// Returns SyncStats with files processed count and duration. Handles empty S3 gracefully.
func (c *APIClient) InitVolumeFromS3(ctx context.Context, sandboxInfo *SandboxInfo) (*SyncStats, error) {
	if sandboxInfo == nil || sandboxInfo.Sandbox == nil {
		return nil, errors.New("sandboxInfo or sandbox is nil")
	}
	if sandboxInfo.Config.S3Config == nil {
		return nil, errors.New("S3Config is required for InitVolumeFromS3")
	}

	startTime := time.Now()

	// Build command to copy files from S3 mount to volume as claudeuser
	// Run as claudeuser to ensure proper file ownership in workspace
	// Use cp with recursive and verbose options
	// Note: The || true ensures the command succeeds even if source is empty
	// The 2>&1 redirects stderr to stdout so we can capture all output
	cmd := []string{
		"runuser", "-u", ClaudeUserName, "--",
		"sh", "-c",
		fmt.Sprintf(
			"cp -rv %s/* %s/ 2>&1 || true",
			sandboxInfo.Config.S3Config.MountPath,
			sandboxInfo.Config.VolumeMountPath,
		),
	}

	// Execute command
	process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, &modal.SandboxExecParams{
		Workdir: "/",
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute InitVolumeFromS3 command")
	}

	// Read output
	output, err := io.ReadAll(process.Stdout)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read InitVolumeFromS3 output")
	}

	// Wait for process to complete
	exitCode, err := process.Wait(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to wait for InitVolumeFromS3 process")
	}

	// Check exit code (allow 0 for success, ignore errors since we use || true)
	// Note: cp returns non-zero when source is empty, which is expected behavior
	// We only fail if it's a real error (not "No such file" which indicates empty S3)
	if exitCode != 0 {
		// Log but don't fail - empty directories return non-zero
		outputStr := string(output)
		if !strings.Contains(outputStr, "No such file or directory") {
			return nil, errors.Errorf("InitVolumeFromS3 failed with exit code %d: %s", exitCode, outputStr)
		}
	}

	// Parse output to get file count
	filesProcessed := 0
	bytesTransferred := int64(0)

	outputStr := string(output)
	scanner := bufio.NewScanner(strings.NewReader(outputStr))
	for scanner.Scan() {
		line := scanner.Text()
		// Count copied files (cp -v outputs each file)
		if strings.HasPrefix(line, "'") || strings.Contains(line, " -> ") {
			filesProcessed++
		}
	}

	duration := time.Since(startTime)

	stats := &SyncStats{
		FilesDownloaded:  filesProcessed,
		FilesDeleted:     0,
		FilesSkipped:     0,
		BytesTransferred: bytesTransferred,
		Duration:         duration,
		Errors:           []error{},
	}

	return stats, nil
}

// SyncVolumeToS3 copies files from volume to S3 bucket with timestamp versioning.
// It generates a new timestamp-based path in S3 (docs/{account}/{timestamp}/) and uses
// AWS CLI sync to upload files. This preserves work history with immutable versions.
// Returns SyncStats with files processed, bytes transferred, and duration.
func (c *APIClient) SyncVolumeToS3(ctx context.Context, sandboxInfo *SandboxInfo) (*SyncStats, error) {
	if sandboxInfo == nil || sandboxInfo.Sandbox == nil {
		return nil, errors.New("sandboxInfo or sandbox is nil")
	}
	if sandboxInfo.Config.S3Config == nil {
		return nil, errors.New("S3Config is required for SyncVolumeToS3")
	}

	startTime := time.Now()

	// Generate timestamp for versioning
	// This creates immutable snapshots of workspace state over time
	// Format: s3://bucket/docs/{account_id}/{unix_timestamp}/
	// Each sync creates a new version, preserving history
	timestamp := time.Now().Unix()

	// Build S3 path with account and timestamp
	s3Path := fmt.Sprintf("s3://%s/docs/%s/%d/",
		sandboxInfo.Config.S3Config.BucketName,
		sandboxInfo.Config.AccountID,
		timestamp,
	)

	// Retrieve AWS credentials from Modal secrets
	secret, err := c.client.Secrets.FromName(ctx, sandboxInfo.Config.S3Config.SecretName, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get secret %s for S3 sync", sandboxInfo.Config.S3Config.SecretName)
	}

	// Build AWS CLI sync command running as claudeuser
	// AWS CLI runs directly as claudeuser with credentials passed via secrets
	// Note: Modal sandboxes have "no new privileges" flag set, so sudo cannot be used
	syncCmd := fmt.Sprintf("aws s3 sync %s %s --exact-timestamps 2>&1", sandboxInfo.Config.VolumeMountPath, s3Path)
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
		return nil, errors.Wrapf(err, "failed to execute SyncVolumeToS3 command")
	}

	// Read output
	output, err := io.ReadAll(process.Stdout)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read SyncVolumeToS3 output")
	}

	// Wait for process to complete
	exitCode, err := process.Wait(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to wait for SyncVolumeToS3 process")
	}

	// Check exit code
	if exitCode != 0 {
		return nil, errors.Errorf("SyncVolumeToS3 failed with exit code %d: %s", exitCode, string(output))
	}

	// Count files in volume to get stats
	// Use ls -laR to recursively list all files, grep to filter only regular files (not dirs)
	// This provides metrics for monitoring and billing purposes
	// Run as claudeuser to ensure proper permissions
	countCmd := []string{
		"runuser", "-u", ClaudeUserName, "--",
		"sh", "-c",
		fmt.Sprintf("ls -laR %s 2>/dev/null | grep '^-' | wc -l", sandboxInfo.Config.VolumeMountPath),
	}
	countProcess, err := sandboxInfo.Sandbox.Exec(ctx, countCmd, &modal.SandboxExecParams{
		Workdir: "/",
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to count files")
	}

	countOutput, err := io.ReadAll(countProcess.Stdout)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read count output")
	}

	countExitCode, err := countProcess.Wait(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to wait for count process")
	}

	filesProcessed := 0
	if countExitCode == 0 {
		countStr := strings.TrimSpace(string(countOutput))
		if countStr != "" {
			count, err := strconv.Atoi(countStr)
			if err == nil {
				filesProcessed = count
			}
		}
	}

	// Get total bytes using ls -l and awk
	bytesTransferred := int64(0)
	sizeCmd := []string{
		"sh",
		"-c",
		fmt.Sprintf("ls -laR %s 2>/dev/null | grep '^-' | awk '{sum+=$5} END {print sum}'", sandboxInfo.Config.VolumeMountPath),
	}
	sizeProcess, err := sandboxInfo.Sandbox.Exec(ctx, sizeCmd, &modal.SandboxExecParams{
		Workdir: "/",
	})
	if err == nil {
		sizeOutput, err := io.ReadAll(sizeProcess.Stdout)
		if err == nil {
			sizeExitCode, _ := sizeProcess.Wait(ctx)
			if sizeExitCode == 0 {
				sizeStr := strings.TrimSpace(string(sizeOutput))
				if sizeStr != "" && sizeStr != "0" {
					size, err := strconv.ParseInt(sizeStr, 10, 64)
					if err == nil {
						bytesTransferred = size
					}
				}
			}
		}
	}

	duration := time.Since(startTime)

	stats := &SyncStats{
		FilesDownloaded:  filesProcessed,
		FilesDeleted:     0,
		FilesSkipped:     0,
		BytesTransferred: bytesTransferred,
		Duration:         duration,
		Errors:           []error{},
	}

	return stats, nil
}

// GetLatestVersion retrieves the most recent timestamp version for an account.
// It creates a temporary sandbox to list S3 prefixes and finds the highest timestamp.
// Returns 0 if no versions exist. This is used to restore the most recent work state.
func (c *APIClient) GetLatestVersion(ctx context.Context, accountID types.UUID, bucketName string) (int64, error) {
	if accountID == "" {
		return 0, errors.New("accountID cannot be empty")
	}
	if bucketName == "" {
		return 0, errors.New("bucketName cannot be empty")
	}

	// Create a temporary sandbox to run AWS CLI commands
	// We need AWS credentials to list S3 buckets
	config := &SandboxConfig{
		AccountID: accountID,
		Image: &ImageConfig{
			BaseImage: "alpine:3.21",
			DockerfileCommands: []string{
				"RUN apk add --no-cache bash curl aws-cli",
			},
		},
		VolumeName:      fmt.Sprintf("temp-version-%s", accountID),
		VolumeMountPath: "/mnt/workspace",
		Workdir:         "/mnt/workspace",
		S3Config: &S3MountConfig{
			BucketName: bucketName,
			SecretName: "s3-bucket",
			KeyPrefix:  fmt.Sprintf("docs/%s/", accountID),
			MountPath:  "/mnt/s3-bucket",
			ReadOnly:   true,
		},
	}

	sandboxInfo, err := c.CreateSandbox(ctx, config)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to create temporary sandbox for GetLatestVersion")
	}
	defer func() {
		_ = c.TerminateSandbox(ctx, sandboxInfo, false)
	}()

	// Retrieve AWS credentials
	secret, err := c.client.Secrets.FromName(ctx, "s3-bucket", nil)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to get S3 secret for GetLatestVersion")
	}

	// Build AWS CLI command to list versions (timestamps)
	cmd := []string{
		"sh", "-c",
		fmt.Sprintf(
			"aws s3 ls s3://%s/docs/%s/ | grep PRE | awk '{print $2}' | sed 's/\\///g' | sort -n | tail -1",
			bucketName,
			accountID,
		),
	}

	// Execute with AWS credentials
	process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, &modal.SandboxExecParams{
		Secrets: []*modal.Secret{secret},
		Workdir: "/",
	})
	if err != nil {
		return 0, errors.Wrapf(err, "failed to execute GetLatestVersion command")
	}

	// Read output
	output, err := io.ReadAll(process.Stdout)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to read GetLatestVersion output")
	}

	// Wait for process to complete
	exitCode, err := process.Wait(ctx)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to wait for GetLatestVersion process")
	}

	// Check exit code
	if exitCode != 0 {
		// Empty result is not an error, just return 0
		outputStr := strings.TrimSpace(string(output))
		if outputStr == "" {
			return 0, nil
		}
		return 0, errors.Errorf("GetLatestVersion failed with exit code %d: %s", exitCode, string(output))
	}

	// Parse timestamp from output
	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		// No versions found
		return 0, nil
	}

	// Extract timestamp using regex (digits only)
	timestampRegex := regexp.MustCompile(`\d+`)
	timestampStr := timestampRegex.FindString(outputStr)
	if timestampStr == "" {
		return 0, errors.Errorf("failed to parse timestamp from output: %s", outputStr)
	}

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to convert timestamp string %s to int64", timestampStr)
	}

	return timestamp, nil
}
