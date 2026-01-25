// https://github.com/modal-labs/libmodal/blob/main/modal-go/examples/sandbox-cloud-bucket/main.go
package modal

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/core/lib/types"
	"github.com/modal-labs/libmodal/modal-go"
	"github.com/pkg/errors"
)

// SandboxConfig holds configuration for creating a Modal sandbox.
// It defines the Docker image, storage volumes, S3 mounts, working directory,
// secrets, and environment variables for the sandbox environment.
type SandboxConfig struct {
	AccountID       types.UUID        // Account scoping for the sandbox
	Image           *ImageConfig      // Docker image configuration
	DockerFilePath  string            // Path to Dockerfile (alternative to Image)
	VolumeName      string            // Volume name for persistent storage
	VolumeMountPath string            // Where to mount volume (e.g., "/mnt/workspace")
	S3Config        *S3MountConfig    // Optional S3 bucket mount
	Workdir         string            // Working directory for processes
	Secrets         map[string]string // Additional secrets to inject
	EnvironmentVars map[string]string // Custom environment variables
}

// ImageConfig defines the Docker image to use for a sandbox.
// It specifies a base registry image and optional Dockerfile commands to customize the image.
type ImageConfig struct {
	BaseImage          string   // Base registry image (e.g., "alpine:3.21")
	DockerfileCommands []string // Custom Dockerfile commands for image building
}

// S3MountConfig defines S3 bucket mount configuration with timestamp versioning.
// It enables mounting S3 buckets into sandboxes with read-only or read-write access.
// The KeyPrefix supports timestamp-based versioning for data isolation.
type S3MountConfig struct {
	BucketName string // S3 bucket name
	SecretName string // Modal secret for AWS credentials
	KeyPrefix  string // S3 key prefix (e.g., "docs/{account}/{timestamp}/")
	MountPath  string // Where to mount in container
	ReadOnly   bool   // Mount as read-only
	Timestamp  int64  // Unix timestamp for versioning
}

// SandboxInfo contains sandbox metadata and state.
// It provides access to the underlying Modal sandbox and tracks configuration and status.
type SandboxInfo struct {
	SandboxID string         // Modal sandbox ID
	Sandbox   *modal.Sandbox // Underlying Modal sandbox
	Config    *SandboxConfig // Original configuration
	CreatedAt time.Time      // Creation timestamp
	Status    SandboxStatus  // Current status
}

// SandboxStatus represents the current state of a sandbox.
type SandboxStatus string

const (
	// SandboxStatusRunning indicates the sandbox is currently active.
	SandboxStatusRunning SandboxStatus = "running"
	// SandboxStatusTerminated indicates the sandbox has been stopped.
	SandboxStatusTerminated SandboxStatus = "terminated"
	// SandboxStatusError indicates the sandbox encountered an error.
	SandboxStatusError SandboxStatus = "error"
)

// CreateSandbox creates a new Modal sandbox with the given configuration.
// It creates a Modal app (scoped by accountID), builds the Docker image,
// creates or reuses a volume, optionally mounts an S3 bucket, and starts the sandbox.
// Returns SandboxInfo containing the sandbox ID and handle for further operations.
func (c *APIClient) CreateSandbox(ctx context.Context, config *SandboxConfig) (*SandboxInfo, error) {
	if config == nil {
		return nil, errors.New("SandboxConfig cannot be nil")
	}
	if tools.Empty(config.AccountID) {
		return nil, errors.New("AccountID cannot be empty")
	}
	if config.Image == nil {
		return nil, errors.New("ImageConfig cannot be nil")
	}

	// Create or get Modal app scoped by accountID
	appName := fmt.Sprintf("app-%s", config.AccountID.String())
	app, err := c.client.Apps.FromName(ctx, appName, &modal.AppFromNameParams{
		CreateIfMissing: true,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get or create app for account %s", config.AccountID)
	}

	// Build image from ImageConfig
	image := c.client.Images.FromRegistry(config.Image.BaseImage, nil)
	if len(config.Image.DockerfileCommands) > 0 {
		image = image.DockerfileCommands(config.Image.DockerfileCommands, nil)
	}

	// Create or get volume
	volumeName := config.VolumeName
	if volumeName == "" {
		volumeName = fmt.Sprintf("volume-%s", config.AccountID.String())
	}
	volume, err := c.client.Volumes.FromName(ctx, volumeName, &modal.VolumeFromNameParams{
		CreateIfMissing: true,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create volume %s for account %s", volumeName, config.AccountID)
	}

	// Prepare sandbox creation params
	sandboxParams := &modal.SandboxCreateParams{
		Volumes: map[string]*modal.Volume{
			config.VolumeMountPath: volume,
		},
	}

	// Add workdir if specified
	if config.Workdir != "" {
		sandboxParams.Workdir = config.Workdir
	}

	// Add S3 mount if configured
	if config.S3Config != nil {
		// Retrieve S3 secret
		secret, err := c.client.Secrets.FromName(ctx, config.S3Config.SecretName, nil)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get secret %s for S3 mount", config.S3Config.SecretName)
		}

		// Create CloudBucketMount with timestamp key prefix
		cloudBucketMount, err := c.client.CloudBucketMounts.New(
			config.S3Config.BucketName,
			&modal.CloudBucketMountParams{
				Secret:    secret,
				KeyPrefix: &config.S3Config.KeyPrefix,
				ReadOnly:  config.S3Config.ReadOnly,
			},
		)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create cloud bucket mount for bucket %s", config.S3Config.BucketName)
		}

		sandboxParams.CloudBucketMounts = map[string]*modal.CloudBucketMount{
			config.S3Config.MountPath: cloudBucketMount,
		}
	}

	// Create sandbox
	sb, err := c.client.Sandboxes.Create(ctx, app, image, sandboxParams)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create sandbox for account %s", config.AccountID)
	}

	// Return SandboxInfo with metadata
	sandboxInfo := &SandboxInfo{
		SandboxID: sb.SandboxID,
		Sandbox:   sb,
		Config:    config,
		CreatedAt: time.Now(),
		Status:    SandboxStatusRunning,
	}

	return sandboxInfo, nil
}

// TerminateSandbox terminates a sandbox and optionally syncs data to S3.
// If syncToS3 is true and S3Config is present, it syncs the volume to S3 before termination.
// The sandbox status is updated to SandboxStatusTerminated on success.
func (c *APIClient) TerminateSandbox(ctx context.Context, sandboxInfo *SandboxInfo, syncToS3 bool) error {
	if sandboxInfo == nil || sandboxInfo.Sandbox == nil {
		return errors.New("sandboxInfo or sandbox is nil")
	}

	// Sync to S3 before termination if requested
	if syncToS3 && sandboxInfo.Config.S3Config != nil {
		_, err := c.SyncVolumeToS3(ctx, sandboxInfo)
		if err != nil {
			// Log error but don't fail termination
			log.Printf("Warning: failed to sync volume to S3 before termination: %v", err)
		}
	}

	err := sandboxInfo.Sandbox.Terminate(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to terminate sandbox %s", sandboxInfo.SandboxID)
	}

	sandboxInfo.Status = SandboxStatusTerminated
	return nil
}

// GetSandboxStatus returns the current status of a sandbox by ID.
// This method currently requires SandboxInfo and returns an error directing
// callers to use GetSandboxStatusFromInfo instead.
func (c *APIClient) GetSandboxStatus(_ context.Context, sandboxID string) (SandboxStatus, error) {
	if sandboxID == "" {
		return "", errors.New("sandboxID cannot be empty")
	}

	// Note: This implementation has a limitation - we need the SandboxInfo to query status
	// For now, we'll use a simple approach based on the cached status in SandboxInfo
	// In a real implementation, you would either:
	// 1. Store SandboxInfo in a cache/database and retrieve it by sandboxID
	// 2. Use Modal API to lookup sandbox by ID (if such API exists)
	//
	// For Task 3, we'll use the approach where the caller must maintain SandboxInfo
	// and we'll query the sandbox's status via Poll()

	return "", errors.New("GetSandboxStatus requires SandboxInfo - use GetSandboxStatusFromInfo instead")
}

// GetSandboxStatusFromInfo returns the current status of a sandbox using SandboxInfo.
// It polls the sandbox to check if it's still running or has terminated.
// The SandboxInfo.Status field is updated with the current status.
func (c *APIClient) GetSandboxStatusFromInfo(ctx context.Context, sandboxInfo *SandboxInfo) (SandboxStatus, error) {
	if sandboxInfo == nil || sandboxInfo.Sandbox == nil {
		return "", errors.New("sandboxInfo or sandbox is nil")
	}

	// Use Poll to check if sandbox is still running
	exitCode, err := sandboxInfo.Sandbox.Poll(ctx)
	if err != nil {
		return SandboxStatusError, errors.Wrapf(err, "failed to poll sandbox %s", sandboxInfo.SandboxID)
	}

	// If Poll returns nil, sandbox is still running
	if exitCode == nil {
		sandboxInfo.Status = SandboxStatusRunning
		return SandboxStatusRunning, nil
	}

	// If Poll returns an exit code, sandbox has terminated
	sandboxInfo.Status = SandboxStatusTerminated
	return SandboxStatusTerminated, nil
}

// GetSandbox retrieves a sandbox reference by ID.
// This returns a sandbox handle that can be used for operations.
// Note: This does not validate if the sandbox actually exists.
func (c *APIClient) GetSandbox(ctx context.Context, sandboxID string) (*modal.Sandbox, error) {
	sandbox, err := c.client.Sandboxes.FromID(ctx, sandboxID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get sandbox %s", sandboxID)
	}
	return sandbox, nil
}
