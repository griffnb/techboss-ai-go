package state_files

import (
	"context"

	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/pkg/errors"
)

// GetLatestVersion discovers the latest timestamped version in S3 for an account.
// It queries S3 to find the highest timestamp folder under docs/{accountID}/
//
// This is used to determine which S3 state to sync from when initializing a new sandbox.
// Returns the latest timestamp (Unix seconds) or 0 if no versions exist.
//
// Implementation approach:
// Since modal.APIClient already has GetLatestVersion(), we create a thin wrapper
// that delegates to it. This provides a service-layer entry point while reusing
// the existing battle-tested implementation.
//
// Flow:
// 1. Validate inputs
// 2. Delegate to modal client's GetLatestVersion()
// 3. Return timestamp
//
// Note: The modal implementation creates a temporary sandbox to query S3,
// which is necessary since we need AWS CLI access to list S3 prefixes.
func GetLatestVersion(
	ctx context.Context,
	client modal.APIClientInterface,
	accountID types.UUID,
	bucketName string,
) (int64, error) {
	// Validate inputs
	if client == nil {
		return 0, errors.New("client cannot be nil")
	}

	if accountID == "" {
		return 0, errors.New("accountID cannot be empty")
	}

	if bucketName == "" {
		return 0, errors.New("bucketName cannot be empty")
	}

	// Delegate to modal client
	timestamp, err := client.GetLatestVersion(ctx, accountID, bucketName)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to get latest version for account %s from bucket %s", accountID, bucketName)
	}

	return timestamp, nil
}
