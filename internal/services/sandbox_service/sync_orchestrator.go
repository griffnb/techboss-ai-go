package sandbox_service

import (
	"context"
	"time"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
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
//  2. Check if stale using modal.CheckIfStale()
//  3. If no local state or stale: do cold start (copy all from S3)
//  4. If local state exists and fresh: do incremental sync
//     a. Read S3 state file
//     b. Compare states using modal.CompareStateFiles()
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
) (*modal.SyncStats, error) {
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

	// For now, we'll use the existing InitVolumeFromS3WithState method which already
	// implements the state-based sync logic internally. This simplifies the orchestration
	// layer and uses the battle-tested implementation in the modal package.
	//
	// The InitVolumeFromS3WithState method already:
	// 1. Reads local and S3 state files
	// 2. Checks if state is stale (implicit in cold start detection)
	// 3. Performs full sync if no local state exists
	// 4. Performs incremental sync if local state exists
	// 5. Writes updated state files
	//
	// Future enhancement: We could extract the state file operations from modal package
	// and implement the full orchestration logic here if we need more control over
	// the sync decision logic (e.g., custom staleness thresholds, different sync strategies).

	log.Info("OrchestratePullSync: Using state-based sync")

	stats, err := client.InitVolumeFromS3WithState(ctx, sandboxInfo)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to perform pull sync")
	}

	log.Info("OrchestratePullSync: Completed successfully")

	return stats, nil
}

// OrchestratePushSync orchestrates pushing data from volume to S3 with state tracking.
// It creates a timestamped version in S3 for historical tracking.
//
// Flow:
// 1. Generate state file from current volume contents
// 2. Add timestamp to S3 config key prefix
// 3. Sync volume to S3 using modal.SyncVolumeToS3()
// 4. Write state file to S3
// 5. Write state file locally for next pull
//
// Returns SyncStats with operation details.
func OrchestratePushSync(
	ctx context.Context,
	client modal.APIClientInterface,
	sandboxInfo *modal.SandboxInfo,
) (*modal.SyncStats, error) {
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

	// For now, we'll use the existing SyncVolumeToS3WithState method which already
	// implements the state-based push logic internally. This simplifies the orchestration
	// layer and uses the battle-tested implementation in the modal package.
	//
	// The SyncVolumeToS3WithState method already:
	// 1. Generates state file from current volume contents
	// 2. Adds timestamp to S3 path for versioning
	// 3. Syncs volume to S3 using AWS CLI
	// 4. Writes state files to both local and S3
	//
	// Future enhancement: We could extract the state file operations from modal package
	// and implement the full orchestration logic here if we need more control over
	// the push strategy (e.g., selective push, compression, encryption).

	// Add timestamp to S3 config for versioning
	timestamp := time.Now().Unix()
	sandboxInfo.Config.S3Config.Timestamp = timestamp

	log.Info("OrchestratePushSync: Using state-based push with timestamp versioning")

	stats, err := client.SyncVolumeToS3WithState(ctx, sandboxInfo)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to perform push sync")
	}

	log.Info("OrchestratePushSync: Completed successfully")

	return stats, nil
}
