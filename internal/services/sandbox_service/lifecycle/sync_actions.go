package lifecycle

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service/state_files"
	"github.com/pkg/errors"
)

// ExecuteSyncActions executes the sync actions defined in a StateDiff.
// It downloads files from S3, deletes local files, and tracks statistics.
// Returns SyncStats with counts and any errors encountered.
//
// This function performs the actual file operations based on the state diff:
// - Downloads files from S3 mount to local volume (new or updated files)
// - Deletes local files that don't exist in S3 (removed files)
// - Tracks detailed statistics: files downloaded, deleted, skipped, bytes, duration
// - Collects non-fatal errors to allow partial sync success
//
// Requirements satisfied:
// - 3.4: Download files in S3 but not local
// - 3.6: Download files with different checksums
// - 3.7: Delete files in local but not S3
// - 3.10: Return detailed sync stats
func ExecuteSyncActions(
	ctx context.Context,
	client modal.APIClientInterface,
	sandboxInfo *modal.SandboxInfo,
	diff *state_files.StateDiff,
) (*state_files.SyncStats, error) {
	// Validate inputs
	if client == nil {
		return nil, errors.New("client cannot be nil")
	}
	if sandboxInfo == nil {
		return nil, errors.New("sandboxInfo cannot be nil")
	}
	if diff == nil {
		return nil, errors.New("diff cannot be nil")
	}
	if sandboxInfo.Config == nil {
		return nil, errors.New("sandboxInfo.Config cannot be nil")
	}
	if sandboxInfo.Config.S3Config == nil {
		return nil, errors.New("sandboxInfo.Config.S3Config is required")
	}

	startTime := time.Now()

	stats := &state_files.SyncStats{
		FilesDownloaded:  0,
		FilesDeleted:     0,
		FilesSkipped:     len(diff.FilesToSkip),
		BytesTransferred: 0,
		Duration:         0,
		Errors:           []error{},
	}

	// Download files from S3 mount to local volume
	for _, file := range diff.FilesToDownload {
		err := downloadFile(ctx, sandboxInfo, file)
		if err != nil {
			stats.Errors = append(stats.Errors, errors.Wrapf(err, "failed to download file %s", file.Path))
			continue
		}

		// Successfully downloaded
		stats.FilesDownloaded++
		stats.BytesTransferred += file.Size
	}

	// Delete local files not in S3
	for _, filePath := range diff.FilesToDelete {
		err := deleteFile(ctx, sandboxInfo, filePath)
		if err != nil {
			stats.Errors = append(stats.Errors, errors.Wrapf(err, "failed to delete file %s", filePath))
			continue
		}

		// Successfully deleted
		stats.FilesDeleted++
	}

	// Set final duration
	stats.Duration = time.Since(startTime)

	// Return error if we had any errors during sync
	if len(stats.Errors) > 0 {
		return stats, errors.Errorf("sync completed with %d errors", len(stats.Errors))
	}

	return stats, nil
}

// downloadFile downloads a single file from S3 mount to local volume.
// It creates parent directories if needed and preserves timestamps with cp -p flag.
func downloadFile(ctx context.Context, sandboxInfo *modal.SandboxInfo, file state_files.FileEntry) error {
	sourcePath := fmt.Sprintf("%s/%s", sandboxInfo.Config.S3Config.MountPath, file.Path)
	destPath := fmt.Sprintf("%s/%s", sandboxInfo.Config.VolumeMountPath, file.Path)

	// Create parent directory if needed
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
		return errors.Wrapf(err, "failed to create directory for %s", file.Path)
	}
	_, _ = mkdirProcess.Wait(ctx)

	// Copy file from S3 mount to local volume with preserved timestamps
	cpCmd := []string{
		"sh", "-c",
		fmt.Sprintf("cp -p %s %s 2>&1", sourcePath, destPath),
	}

	cpProcess, err := sandboxInfo.Sandbox.Exec(ctx, cpCmd, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to execute copy command for %s", file.Path)
	}

	// Read output for error details
	var output bytes.Buffer
	scanner := bufio.NewScanner(cpProcess.Stdout)
	for scanner.Scan() {
		output.Write(scanner.Bytes())
		output.WriteByte('\n')
	}

	exitCode, err := cpProcess.Wait(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to wait for copy command for %s", file.Path)
	}

	if exitCode != 0 {
		return errors.Errorf("copy failed for %s: exit code %d, output: %s",
			file.Path, exitCode, output.String())
	}

	return nil
}

// deleteFile deletes a single file from the local volume.
func deleteFile(ctx context.Context, sandboxInfo *modal.SandboxInfo, filePath string) error {
	localPath := fmt.Sprintf("%s/%s", sandboxInfo.Config.VolumeMountPath, filePath)

	// Delete file
	rmCmd := []string{
		"sh", "-c",
		fmt.Sprintf("rm -f %s 2>&1", localPath),
	}

	rmProcess, err := sandboxInfo.Sandbox.Exec(ctx, rmCmd, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to execute delete command for %s", filePath)
	}

	exitCode, err := rmProcess.Wait(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to wait for delete command for %s", filePath)
	}

	if exitCode != 0 {
		return errors.Errorf("delete failed for %s: exit code %d", filePath, exitCode)
	}

	return nil
}
