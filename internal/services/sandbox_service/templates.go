package sandbox_service

import (
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
	"github.com/pkg/errors"
)

// SandboxTemplate defines a premade sandbox configuration.
// Templates enable the frontend to create sandboxes without specifying configuration details.
type SandboxTemplate struct {
	Provider     sandbox.Provider
	ImageConfig  *modal.ImageConfig
	VolumeName   string
	S3BucketName string
	S3KeyPrefix  string
	InitFromS3   bool
}

// GetSandboxTemplate returns a premade template based on provider and agent.
// This enables the frontend to create sandboxes without configuration details.
func GetSandboxTemplate(provider sandbox.Provider, agentID types.UUID) (*SandboxTemplate, error) {
	switch provider {
	case sandbox.PROVIDER_CLAUDE_CODE:
		return getClaudeCodeTemplate(agentID), nil
	default:
		return nil, errors.Errorf("unsupported provider: %d", provider)
	}
}

// getClaudeCodeTemplate returns the Claude Code sandbox template
func getClaudeCodeTemplate(_ types.UUID) *SandboxTemplate {
	return &SandboxTemplate{
		Provider:     sandbox.PROVIDER_CLAUDE_CODE,
		ImageConfig:  modal.GetImageConfigFromTemplate("claude"),
		VolumeName:   "",
		S3BucketName: "",
		S3KeyPrefix:  "",
		InitFromS3:   false,
	}
}

// BuildSandboxConfig creates a modal.SandboxConfig from a template
func (t *SandboxTemplate) BuildSandboxConfig(accountID types.UUID) *modal.SandboxConfig {
	config := &modal.SandboxConfig{
		AccountID:       accountID,
		Image:           t.ImageConfig,
		VolumeName:      t.VolumeName,
		VolumeMountPath: "/mnt/workspace",
		Workdir:         "/mnt/workspace",
	}

	// Add S3 config if specified
	if t.S3BucketName != "" {
		config.S3Config = &modal.S3MountConfig{
			BucketName: t.S3BucketName,
			SecretName: "s3-bucket",
			KeyPrefix:  t.S3KeyPrefix,
			MountPath:  "/mnt/s3-bucket",
			ReadOnly:   true,
		}
	}

	return config
}
