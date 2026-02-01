package sandbox_service

import (
	"context"
	"fmt"
	"time"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service/lifecycle"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service/state_files"
	"github.com/pkg/errors"
)

const (
	// defaultStaleThresholdSeconds is the default threshold (1 hour) for considering state stale.
	defaultStaleThresholdSeconds = 3600
)

// OrchestratePullSync orchestrates pulling data from S3 to volume with state tracking.
// It determines whether to do a cold start (full sync) or incremental sync based on:
// - Whether local state file exists
// - Whether local state is stale (based on threshold)
//
// Flow:
//  1. Read local state file (if exists)
//  2. Check if stale using state_files.CheckIfStale()
//  3. If no local state or stale: do cold start (copy all from S3)
//  4. If local state exists and fresh: do incremental sync
//     a. Read S3 state file
//     b. Compare states using state_files.CompareStateFiles()
//     c. Execute sync actions using ExecuteSyncActions()
//  5. Generate new local state file
//  6. Write new local state file
//
// Returns SyncStats with operation details.
func OrchestratePullSync(
	ctx context.Context,
	client modal.APIClientInterface,
	sandboxInfo *modal.SandboxInfo,
	staleThresholdSeconds int64,
) (*state_files.SyncStats, error) {
	// Validate inputs
	if client == nil {
		return nil, errors.New("client cannot be nil")
	}
	if sandboxInfo == nil {
		return nil, errors.New("sandboxInfo cannot be nil")
	}
	if sandboxInfo.Config == nil {
		return nil, errors.New("sandboxInfo.Config cannot be nil")
	}
	if sandboxInfo.Config.S3Config == nil {
		return nil, errors.New("S3Config is required for pull sync")
	}

	startTime := time.Now()

	// Step 1: Read local .sandbox-state (nil if doesn't exist - new sandbox)
	localState, err := state_files.ReadLocalStateFile(ctx, sandboxInfo, sandboxInfo.Config.VolumeMountPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read local state file")
	}

	// Log cold start decision
	if localState == nil {
		log.Info(fmt.Sprintf("[S3 Sync] OnColdStart decision: FULL SYNC (no local state file) sandbox=%s",
			sandboxInfo.SandboxID))
	} else {
		log.Info(fmt.Sprintf("[S3 Sync] OnColdStart decision: INCREMENTAL SYNC (local state exists, last_sync=%d) sandbox=%s",
			localState.LastSyncedAt,
			sandboxInfo.SandboxID))
	}

	// Step 2: Read S3 .sandbox-state (or generate if missing)
	s3State, err := state_files.ReadS3StateFile(ctx, sandboxInfo, sandboxInfo.Config.S3Config.MountPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read S3 state file")
	}

	// If S3 state doesn't exist, generate it from S3 directory scan
	if s3State == nil {
		log.Info(fmt.Sprintf("[S3 Sync] S3 state file not found, generating from directory scan sandbox=%s",
			sandboxInfo.SandboxID))
		s3State, err = state_files.GenerateStateFile(ctx, sandboxInfo, sandboxInfo.Config.S3Config.MountPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to generate S3 state file")
		}
	}

	// Step 3: Compare states to determine sync actions
	diff := state_files.CompareStateFiles(localState, s3State)

	// Step 4: Execute sync actions (download, delete)
	stats, err := lifecycle.ExecuteSyncActions(ctx, client, sandboxInfo, diff)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute sync actions")
	}

	// Step 5: Update local .sandbox-state with S3 state (we're now in sync)
	err = state_files.WriteLocalStateFile(ctx, sandboxInfo, sandboxInfo.Config.VolumeMountPath, s3State)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to write local state file")
	}

	// Set final duration
	stats.Duration = time.Since(startTime)

	// Log sync statistics
	log.Info(fmt.Sprintf("[S3 Sync] OrchestratePullSync completed: downloaded=%d deleted=%d skipped=%d bytes=%d duration=%v sandbox=%s",
		stats.FilesDownloaded,
		stats.FilesDeleted,
		stats.FilesSkipped,
		stats.BytesTransferred,
		stats.Duration,
		sandboxInfo.SandboxID))

	return stats, nil
}

// OrchestratePushSync orchestrates pushing data from volume to S3 with state tracking.
// It creates a timestamped version in S3 for historical tracking.
//
// Flow:
// 1. Generate state file from current volume contents
// 2. Sync volume to S3 using modal.SyncVolumeToS3()
// 3. Write state file to both local and S3
//
// Returns SyncStats with operation details.
func OrchestratePushSync(
	ctx context.Context,
	client modal.APIClientInterface,
	sandboxInfo *modal.SandboxInfo,
) (*state_files.SyncStats, error) {
	// Validate inputs
	if client == nil {
		return nil, errors.New("client cannot be nil")
	}
	if sandboxInfo == nil {
		return nil, errors.New("sandboxInfo cannot be nil")
	}
	if sandboxInfo.Config == nil {
		return nil, errors.New("sandboxInfo.Config cannot be nil")
	}
	if sandboxInfo.Config.S3Config == nil {
		return nil, errors.New("S3Config is required for push sync")
	}

	startTime := time.Now()

	// Step 1: Generate current state from local volume
	localState, err := state_files.GenerateStateFile(ctx, sandboxInfo, sandboxInfo.Config.VolumeMountPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate local state file")
	}

	// Step 2: Sync volume to S3 (this creates timestamped S3 path internally)
	// The modal.SyncVolumeToS3 method handles AWS CLI sync with timestamp versioning
	modalStats, err := client.SyncVolumeToS3(ctx, sandboxInfo)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to sync volume to S3")
	}

	// Convert modal stats to state_files stats
	stats := &state_files.SyncStats{
		FilesDownloaded:  modalStats.FilesDownloaded,
		FilesDeleted:     modalStats.FilesDeleted,
		FilesSkipped:     modalStats.FilesSkipped,
		BytesTransferred: modalStats.BytesTransferred,
		Duration:         0, // Will be set below
		Errors:           modalStats.Errors,
	}

	// Step 3: Write .sandbox-state to both local and S3
	err = state_files.WriteLocalStateFile(ctx, sandboxInfo, sandboxInfo.Config.VolumeMountPath, localState)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to write local state file")
	}

	// Write to S3 mount path as well (this will be in the timestamped location)
	err = state_files.WriteS3StateFile(ctx, sandboxInfo, sandboxInfo.Config.S3Config.MountPath, localState)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to write S3 state file")
	}

	// Set final duration
	stats.Duration = time.Since(startTime)

	// Log sync statistics
	log.Info(fmt.Sprintf("[S3 Sync] OrchestratePushSync completed: uploaded=%d bytes=%d duration=%v sandbox=%s",
		stats.FilesDownloaded, // Using FilesDownloaded for upload count
		stats.BytesTransferred,
		stats.Duration,
		sandboxInfo.SandboxID))

	return stats, nil
}
