// https://github.com/modal-labs/libmodal/blob/main/modal-go/examples/sandbox-cloud-bucket/main.go
package modal

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/core/lib/types"
	"github.com/modal-labs/libmodal/modal-go"
	"github.com/pkg/errors"
)

// SandboxConfig holds configuration for creating a sandbox.
type SandboxConfig struct {
	AccountID       types.UUID        // Account scoping for the sandbox
	Image           *ImageConfig      // Docker image configuration
	VolumeName      string            // Volume name for persistent storage
	VolumeMountPath string            // Where to mount volume (e.g., "/mnt/workspace")
	S3Config        *S3MountConfig    // Optional S3 bucket mount
	Workdir         string            // Working directory for processes
	Secrets         map[string]string // Additional secrets to inject
	EnvironmentVars map[string]string // Custom environment variables
}

// ImageConfig defines Docker image to use.
type ImageConfig struct {
	BaseImage          string   // Base registry image (e.g., "alpine:3.21")
	DockerfileCommands []string // Custom Dockerfile commands for image building
}

// S3MountConfig defines S3 bucket mount configuration with timestamp versioning.
type S3MountConfig struct {
	BucketName string // S3 bucket name
	SecretName string // Modal secret for AWS credentials
	KeyPrefix  string // S3 key prefix (e.g., "docs/{account}/{timestamp}/")
	MountPath  string // Where to mount in container
	ReadOnly   bool   // Mount as read-only
	Timestamp  int64  // Unix timestamp for versioning
}

// SandboxInfo contains sandbox metadata and state.
type SandboxInfo struct {
	SandboxID string         // Modal sandbox ID
	Sandbox   *modal.Sandbox // Underlying Modal sandbox
	Config    *SandboxConfig // Original configuration
	CreatedAt time.Time      // Creation timestamp
	Status    SandboxStatus  // Current status
}

// SandboxStatus represents sandbox state.
type SandboxStatus string

const (
	SandboxStatusRunning    SandboxStatus = "running"
	SandboxStatusTerminated SandboxStatus = "terminated"
	SandboxStatusError      SandboxStatus = "error"
)

func (this *APIClient) BuildSandbox(ctx context.Context, accountID types.UUID) (*modal.Sandbox, error) {
	app, err := this.client.Apps.FromName(ctx, "app-"+accountID.String(), &modal.AppFromNameParams{CreateIfMissing: true})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get or create App for account %s", accountID)
	}

	image := this.client.Images.FromRegistry("alpine:3.21", nil).DockerfileCommands([]string{
		"RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep",
		"RUN curl -fsSL https://claude.ai/install.sh | bash",
		"ENV PATH=/root/.local/bin:$PATH USE_BUILTIN_RIPGREP=0",
	}, nil)

	// standard volume
	volume, err := this.client.Volumes.FromName(ctx, fmt.Sprintf("volume-%s", accountID), &modal.VolumeFromNameParams{
		CreateIfMissing: true,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create Volume for account %s", accountID)
	}

	bucketSecret, err := this.client.Secrets.FromName(ctx, "s3-bucket", nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get Secret for account %s", accountID)
	}

	// S3 bucket mount
	keyPrefix := fmt.Sprintf("docs/%s/", accountID)
	cloudBucketMount, err := this.client.CloudBucketMounts.New("tb-prod-agent-docs", &modal.CloudBucketMountParams{
		Secret:    bucketSecret,
		KeyPrefix: &keyPrefix,
		ReadOnly:  true,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create Cloud Bucket Mount for account %s", accountID)
	}

	sb, err := this.client.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{
		Volumes: map[string]*modal.Volume{
			"/mnt/volume": volume,
		},
		CloudBucketMounts: map[string]*modal.CloudBucketMount{
			"/mnt/s3-bucket": cloudBucketMount,
		},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create writer Sandbox for account %s", accountID)
	}

	return sb, nil
}

// CreateSandbox creates a new Modal sandbox with the given configuration.
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
func (c *APIClient) TerminateSandbox(ctx context.Context, sandboxInfo *SandboxInfo, syncToS3 bool) error {
	if sandboxInfo == nil || sandboxInfo.Sandbox == nil {
		return errors.New("sandboxInfo or sandbox is nil")
	}

	// TODO: Implement syncToS3 logic when storage.go is implemented
	if syncToS3 {
		log.Printf("Warning: syncToS3 requested but not yet implemented")
	}

	err := sandboxInfo.Sandbox.Terminate(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to terminate sandbox %s", sandboxInfo.SandboxID)
	}

	sandboxInfo.Status = SandboxStatusTerminated
	return nil
}

func (this *APIClient) ExecClaude(
	ctx context.Context,
	sb *modal.Sandbox,
	prompt string,
) (*modal.ContainerProcess, error) {
	secrets, err := this.client.Secrets.FromMap(ctx, map[string]string{
		"ANTHROPIC_API_KEY":       "sk-xxxx",
		"AWS_BEDROCK_API_KEY":     "ABSKxxxx",
		"CLAUDE_CODE_USE_BEDROCK": "1",
		"AWS_REGION":              "us-east-1", // or your preferred region
	}, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get secrets for Sandbox %s", sb.SandboxID)
	}

	cmd := []string{"claude", "-c", "-p", prompt, "--dangerously-skip-permissions", "--output-format", "stream-json", "--verbose"}

	claude, err := sb.Exec(ctx, cmd, &modal.SandboxExecParams{
		PTY:     true, // Adding a PTY is important, since Claude requires it!
		Secrets: []*modal.Secret{secrets},
		Workdir: "/repo",
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute command in Sandbox %s", sb.SandboxID)
	}

	return claude, nil
}

func sandbox() {
	ctx := context.Background()
	mc, err := modal.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	app, err := mc.Apps.FromName(ctx, "libmodal-example", &modal.AppFromNameParams{CreateIfMissing: true})
	if err != nil {
		log.Fatalf("Failed to get or create App: %v", err)
	}

	image := mc.Images.FromRegistry("alpine:3.21", nil).DockerfileCommands([]string{
		"RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep",
		"RUN curl -fsSL https://claude.ai/install.sh | bash",
		"ENV PATH=/root/.local/bin:$PATH USE_BUILTIN_RIPGREP=0",
	}, nil)

	// standard volume
	volume, err := mc.Volumes.FromName(ctx, "libmodal-example-volume", &modal.VolumeFromNameParams{
		CreateIfMissing: true,
	})
	if err != nil {
		log.Fatalf("Failed to create Volume: %v", err)
	}

	secret, err := mc.Secrets.FromName(ctx, "libmodal-aws-bucket-secret", nil)
	if err != nil {
		log.Fatalf("Failed to get Secret: %v", err)
	}

	// S3 bucket mount
	keyPrefix := "data/"
	cloudBucketMount, err := mc.CloudBucketMounts.New("my-s3-bucket", &modal.CloudBucketMountParams{
		Secret:    secret,
		KeyPrefix: &keyPrefix,
		ReadOnly:  true,
	})
	if err != nil {
		log.Fatalf("Failed to create Cloud Bucket Mount: %v", err)
	}

	sb, err := mc.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{
		Command: []string{
			"sh",
			"-c",
			"echo 'Hello from writer Sandbox!' > /mnt/volume/message.txt",
		},
		Volumes: map[string]*modal.Volume{
			"/mnt/volume": volume,
		},
		CloudBucketMounts: map[string]*modal.CloudBucketMount{
			"/mnt/s3-bucket": cloudBucketMount,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create writer Sandbox: %v", err)
	}

	fmt.Printf("Writer Sandbox: %s\n", sb.SandboxID)
	defer func() {
		if err := sb.Terminate(context.Background()); err != nil {
			log.Fatalf("Failed to terminate Sandbox %s: %v", sb.SandboxID, err)
		}
	}()

	claudeCmd := []string{
		"claude",
		"-p",
		"Summarize what this repository is about. Don't modify any code or files.",
	}

	fmt.Println("\nRunning command:", claudeCmd)

	claudeSecret, err := mc.Secrets.FromName(ctx, "libmodal-anthropic-secret", &modal.SecretFromNameParams{
		RequiredKeys: []string{"ANTHROPIC_API_KEY"},
	})
	if err != nil {
		log.Fatalf("Failed to get secret: %v", err)
	}

	claude, err := sb.Exec(ctx, claudeCmd, &modal.SandboxExecParams{
		PTY:     true, // Adding a PTY is important, since Claude requires it!
		Secrets: []*modal.Secret{claudeSecret},
		Workdir: "/repo",
	})
	if err != nil {
		log.Fatalf("Failed to execute claude command: %v", err)
	}

	scanner := bufio.NewScanner(claude.Stdout)
	for scanner.Scan() {
		//line := scanner.Text()

		// Write the line to the response
		//_, err := fmt.Fprintf(responseWriter, "%s\n", line)
		//if err != nil {
		//	return errors.Wrap(err, "failed to write streaming response")
		//}
		//
		//// Flush the response
		//flusher.Flush()
		//
		//// Check if the stream is done
		//if line == "data: [DONE]" {
		//	break
		//}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading claude output: %v", err)
	}

	_, err = claude.Wait(ctx)
	if err != nil {
		log.Fatalf("Claude command failed: %v", err)
	}

	credentials, err := sb.CreateConnectToken(ctx, &modal.SandboxCreateConnectTokenParams{UserMetadata: "user_id=xxxx"})
	if err != nil {
		log.Fatalf("Failed to create connect token: %v", err)
	}
	fmt.Printf("Writer connect token: %s\n", credentials.Token)
	exitCode, err := sb.Wait(ctx)
	if err != nil {
		log.Fatalf("Failed to wait for writer Sandbox: %v", err)
	}
	fmt.Printf("Writer finished with exit code: %d\n", exitCode)
}
