package sandbox_service

import (
	"time"

	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
)

// ReconstructSandboxInfo creates a modal.SandboxInfo from database model fields.
// This is needed for operations that interact with the Modal API after retrieving
// a sandbox from the database. The Sandbox handle is not reconstructed (set to nil).
func ReconstructSandboxInfo(model *sandbox.Sandbox, accountID types.UUID) *modal.SandboxInfo {
	// Get template to reconstruct config
	template, err := GetSandboxTemplate(
		model.Provider.Get(),
		model.AgentID.Get(),
	)

	var config *modal.SandboxConfig
	if err != nil || template == nil {
		// Fallback to basic config if template not found
		config = &modal.SandboxConfig{
			AccountID:       accountID,
			Image:           modal.GetImageConfigFromTemplate("claude"),
			VolumeName:      "",
			VolumeMountPath: "/mnt/workspace",
			Workdir:         "/mnt/workspace",
		}
	} else {
		config = template.BuildSandboxConfig(accountID)
	}

	// Map database status to Modal status
	var modalStatus modal.SandboxStatus
	if model.Deleted.Get() == 1 || model.Status.Get() != constants.STATUS_ACTIVE {
		modalStatus = modal.SandboxStatusTerminated
	} else {
		modalStatus = modal.SandboxStatusRunning
	}

	createdAt := model.CreatedAt.Get()
	if createdAt == nil {
		createdAt = new(time.Time)
	}

	return &modal.SandboxInfo{
		SandboxID: model.ExternalID.Get(),
		Config:    config,
		CreatedAt: *createdAt,
		Status:    modalStatus,
		Sandbox:   nil, // Not reconstructed from DB
	}
}
