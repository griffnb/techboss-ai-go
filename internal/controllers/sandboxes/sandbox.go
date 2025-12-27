package sandboxes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service"
)

// CreateSandboxTemplateRequest holds request data for sandbox creation using templates.
// Frontend only specifies the provider type and agent, not configuration details.
type CreateSandboxTemplateRequest struct {
	Provider sandbox.Provider `json:"provider"` // Provider type (1=Claude Code)
	AgentID  types.UUID       `json:"agent_id"` // Agent ID (optional, for future use)
}

// SyncSandboxRequest holds request data for S3 sync operations.
// No parameters needed - syncs the sandbox's configured S3 bucket.
type SyncSandboxRequest struct {
	// Empty struct - no parameters needed
}

// createSandbox creates a new sandbox using a premade template based on provider/agent.
// It saves the sandbox to the database with ExternalID, Provider, AgentID, Status, and empty MetaData.
func createSandbox(_ http.ResponseWriter, req *http.Request) (*sandbox.Sandbox, int, error) {
	// Get authenticated user session
	userSession := request.GetReqSession(req)
	accountID := userSession.User.ID()

	// Parse request body
	data, err := request.GetJSONPostAs[*CreateSandboxTemplateRequest](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*sandbox.Sandbox](err)
	}

	// Get premade template for provider/agent
	template, err := sandbox_service.GetSandboxTemplate(data.Provider, data.AgentID)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*sandbox.Sandbox](err)
	}

	// Build config from template
	config := template.BuildSandboxConfig(accountID)

	// Create sandbox via service
	service := sandbox_service.NewSandboxService()
	sandboxInfo, err := service.CreateSandbox(req.Context(), accountID, config)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*sandbox.Sandbox](err)
	}

	// Initialize from S3 if template specifies
	if template.InitFromS3 && config.S3Config != nil {
		_, err := service.InitFromS3(req.Context(), sandboxInfo)
		if err != nil {
			log.ErrorContext(err, req.Context())
			log.Infof("Warning: failed to initialize from S3: %v", err)
		}
	}

	// Save to database for persistent storage
	// ExternalID stores the Modal sandbox ID (sb-xxx) for API operations
	// Status tracks sandbox state (active, terminated, etc.)
	// MetaData stores sync timestamps and statistics in JSONB format
	sandboxModel := sandbox.New()
	sandboxModel.AccountID.Set(accountID)
	sandboxModel.AgentID.Set(data.AgentID)
	sandboxModel.Provider.Set(data.Provider)
	sandboxModel.ExternalID.Set(sandboxInfo.SandboxID)
	sandboxModel.Status.Set(constants.STATUS_ACTIVE)
	sandboxModel.MetaData.Set(&sandbox.MetaData{})

	err = sandboxModel.Save(userSession.User)
	if err != nil {
		log.ErrorContext(err, req.Context())
		// Note: Sandbox was created in Modal but DB save failed
		// TODO: Consider adding cleanup logic here or async cleanup task
		return response.AdminBadRequestError[*sandbox.Sandbox](err)
	}

	log.Infof("Created sandbox %s (external_id: %s) for account %s",
		sandboxModel.ID(), sandboxInfo.SandboxID, accountID)

	return response.Success(sandboxModel)
}

// authDelete terminates a sandbox and soft-deletes the database record.
// The auth framework already handles ownership verification.
// If Modal termination fails, logs a warning but continues with soft delete.
func authDelete(_ http.ResponseWriter, req *http.Request) (*sandbox.Sandbox, int, error) {
	userSession := request.GetReqSession(req)
	accountID := userSession.User.ID()
	id := chi.URLParam(req, "id")

	log.Infof("authDelete called for sandbox ID: %s", id)

	// Query database for sandbox by ID with ownership verification
	// Auth framework ensures only the owner can access this sandbox
	sandboxModel, err := sandbox.Get(req.Context(), types.UUID(id))
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*sandbox.Sandbox](err)
	}

	// Reconstruct SandboxInfo for Modal termination
	sandboxInfo := sandbox_service.ReconstructSandboxInfo(sandboxModel, accountID)

	// Terminate sandbox via service with S3 sync
	service := sandbox_service.NewSandboxService()
	err = service.TerminateSandbox(req.Context(), sandboxInfo, true)
	if err != nil {
		log.ErrorContext(err, req.Context())
		// Log error but continue with soft delete
		log.Infof("Warning: failed to terminate Modal sandbox %s: %v", sandboxModel.ExternalID.Get(), err)
	}

	// Update database record: set status to deleted and mark as soft-deleted
	// This preserves the record for audit purposes while hiding it from queries
	sandboxModel.Status.Set(constants.STATUS_DELETED)
	sandboxModel.Deleted.Set(1)
	err = sandboxModel.Save(userSession.User)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*sandbox.Sandbox](err)
	}

	log.Infof("Terminated and deleted sandbox %s (external_id: %s)",
		sandboxModel.ID(), sandboxModel.ExternalID.Get())

	return response.Success(sandboxModel)
}
