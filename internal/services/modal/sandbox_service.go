// Package modal provides service layer for Modal sandbox operations.
// It acts as an intermediary between HTTP controllers and the Modal integration layer,
// providing validation, business logic, and higher-level operations.
package modal

import (
	"context"
	"net/http"

	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/pkg/errors"
)

// SandboxService handles business logic for sandbox operations.
// It provides a clean interface for controllers and wraps the Modal integration layer.
type SandboxService struct {
	client *modal.APIClient
}

// NewSandboxService creates a new sandbox service using the singleton Modal client.
// The service provides validation and business logic on top of the Modal integration.
func NewSandboxService() *SandboxService {
	return &SandboxService{
		client: modal.Client(),
	}
}

// CreateSandbox creates a new sandbox with the given configuration.
// It adds the accountID to the config and validates required fields before
// calling the integration layer. This ensures sandboxes are properly scoped
// to accounts for multi-tenancy.
//
// TODO: Add business logic checks (quotas, permissions)
// TODO: Store sandbox metadata in database for retrieval
func (s *SandboxService) CreateSandbox(
	ctx context.Context,
	accountID types.UUID,
	config *modal.SandboxConfig,
) (*modal.SandboxInfo, error) {
	// Validate inputs
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}
	if config.Image == nil {
		return nil, errors.New("image config is required")
	}

	// Add account ID to config for multi-tenant scoping
	config.AccountID = accountID

	// Create sandbox via integration layer
	sandboxInfo, err := s.client.CreateSandbox(ctx, config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create sandbox for account %s", accountID)
	}

	// TODO: Store sandboxInfo in database/cache for later retrieval
	// TODO: Track sandbox creation for billing/usage metrics

	return sandboxInfo, nil
}

// TerminateSandbox terminates a sandbox and optionally syncs data to S3.
// If syncToS3 is true and S3Config is present, volume data is synced before termination.
//
// TODO: Update database record to mark sandbox as terminated
// TODO: Emit metrics for sandbox lifetime
func (s *SandboxService) TerminateSandbox(
	ctx context.Context,
	sandboxInfo *modal.SandboxInfo,
	syncToS3 bool,
) error {
	if sandboxInfo == nil {
		return errors.New("sandboxInfo cannot be nil")
	}

	// Terminate via integration layer
	err := s.client.TerminateSandbox(ctx, sandboxInfo, syncToS3)
	if err != nil {
		return errors.Wrapf(err, "failed to terminate sandbox %s", sandboxInfo.SandboxID)
	}

	// TODO: Update database to mark as terminated
	// TODO: Emit termination metrics

	return nil
}

// ExecuteClaudeStream executes Claude and streams output to HTTP response.
// It combines ExecClaude and StreamClaudeOutput into a single operation for convenience.
// This method validates the config and handles the complete streaming lifecycle.
func (s *SandboxService) ExecuteClaudeStream(
	ctx context.Context,
	sandboxInfo *modal.SandboxInfo,
	config *modal.ClaudeExecConfig,
	responseWriter http.ResponseWriter,
) error {
	// Validate inputs
	if sandboxInfo == nil {
		return errors.New("sandboxInfo cannot be nil")
	}
	if config == nil {
		return errors.New("config cannot be nil")
	}
	if config.Prompt == "" {
		return errors.New("prompt is required")
	}
	if responseWriter == nil {
		return errors.New("responseWriter cannot be nil")
	}

	// Execute Claude via integration layer
	claudeProcess, err := s.client.ExecClaude(ctx, sandboxInfo, config)
	if err != nil {
		return errors.Wrapf(err, "failed to execute Claude in sandbox %s", sandboxInfo.SandboxID)
	}

	// Stream output to HTTP response
	err = s.client.StreamClaudeOutput(ctx, claudeProcess, responseWriter)
	if err != nil {
		return errors.Wrapf(err, "failed to stream Claude output for sandbox %s", sandboxInfo.SandboxID)
	}

	// TODO: Log Claude execution for audit trail
	// TODO: Track Claude usage for billing

	return nil
}

// InitFromS3 initializes volume from S3 bucket.
// It copies files from the S3 mount to the volume, typically used on sandbox startup
// to restore previous work state.
func (s *SandboxService) InitFromS3(
	ctx context.Context,
	sandboxInfo *modal.SandboxInfo,
) (*modal.SyncStats, error) {
	if sandboxInfo == nil {
		return nil, errors.New("sandboxInfo cannot be nil")
	}
	if sandboxInfo.Config.S3Config == nil {
		return nil, errors.New("S3Config is required for InitFromS3")
	}

	// Initialize from S3 via integration layer
	stats, err := s.client.InitVolumeFromS3(ctx, sandboxInfo)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to initialize volume from S3 for sandbox %s", sandboxInfo.SandboxID)
	}

	// TODO: Log sync stats for monitoring
	// TODO: Track data transfer for billing

	return stats, nil
}

// SyncToS3 syncs volume to S3 bucket with timestamp versioning.
// It uploads the current volume state to S3, creating a new timestamped version
// for historical tracking and rollback capabilities.
func (s *SandboxService) SyncToS3(
	ctx context.Context,
	sandboxInfo *modal.SandboxInfo,
) (*modal.SyncStats, error) {
	if sandboxInfo == nil {
		return nil, errors.New("sandboxInfo cannot be nil")
	}
	if sandboxInfo.Config.S3Config == nil {
		return nil, errors.New("S3Config is required for SyncToS3")
	}

	// Sync to S3 via integration layer
	stats, err := s.client.SyncVolumeToS3(ctx, sandboxInfo)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to sync volume to S3 for sandbox %s", sandboxInfo.SandboxID)
	}

	// TODO: Log sync stats for monitoring
	// TODO: Track data transfer for billing

	return stats, nil
}
