package sandboxes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service"
)

// Updates sandbox metadata (TODO: Implement)
//
//	@Title			Update Sandbox
//	@Public
//	@Summary		Update sandbox
//	@Description	Updates sandbox metadata
//	@Tags			Sandbox
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string			true	"Sandbox ID"
//	@Param			data	body	sandbox.Sandbox	true	"Sandbox Data"
//	@Success		200	{object}	response.SuccessResponse{data=sandbox.Sandbox}
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Router			/sandbox/{id} [put]
func authUpdate(_ http.ResponseWriter, req *http.Request) (*sandbox.Sandbox, int, error) {
	user := request.GetReqSession(req).User

	id := chi.URLParam(req, "id")
	sandboxObj, err := sandbox.GetRestricted(req.Context(), types.UUID(id), user)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*sandbox.Sandbox]()

	}

	return response.Success(sandboxObj)
}

// SyncSandboxResponse holds response data for S3 sync operations.
// Updated per design phase 3.1 to include detailed sync metrics.
type SyncSandboxResponse struct {
	SandboxID        string `json:"sandbox_id"`
	FilesDownloaded  int    `json:"files_downloaded"`
	FilesDeleted     int    `json:"files_deleted"`
	FilesSkipped     int    `json:"files_skipped"`
	BytesTransferred int64  `json:"bytes_transferred"`
	DurationMs       int64  `json:"duration_ms"`
}

// Syncs the sandbox volume to S3 without terminating
//
//	@Title			Sync Sandbox
//	@Public
//	@Summary		Sync sandbox to S3
//	@Description	Syncs the sandbox volume to S3 without terminating for manual backups/snapshots
//	@Tags			Sandbox
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"Sandbox ID"
//	@Success		200	{object}	response.SuccessResponse{data=SyncSandboxResponse}
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Router			/sandbox/{id}/sync [post]
func syncSandbox(_ http.ResponseWriter, req *http.Request) (*SyncSandboxResponse, int, error) {
	userSession := request.GetReqSession(req)
	accountID := userSession.User.ID()
	id := chi.URLParam(req, "id")

	log.Infof("syncSandbox called for sandbox ID: %s", id)

	// Query database for sandbox by ID with ownership verification
	sandboxModel, err := sandbox.Get(req.Context(), types.UUID(id))
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*SyncSandboxResponse](err)
	}

	// Reconstruct SandboxInfo for Modal sync
	sandboxInfo, err := sandbox_service.ReconstructSandboxInfo(req.Context(), sandboxModel, accountID)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*SyncSandboxResponse](err)
	}

	// Sync to S3 via service layer
	service := sandbox_service.NewSandboxService()
	stats, err := service.SyncToS3(req.Context(), sandboxInfo)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*SyncSandboxResponse](err)
	}

	log.Infof("Synced sandbox %s to S3: %d downloaded, %d deleted, %d skipped, %d bytes, %dms",
		id, stats.FilesDownloaded, stats.FilesDeleted, stats.FilesSkipped, stats.BytesTransferred, stats.Duration.Milliseconds())

	// Update database metadata with sync statistics and timestamp
	metadata, err := sandboxModel.MetaData.Get()
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*SyncSandboxResponse](err)
	}
	if metadata == nil {
		metadata = &sandbox.MetaData{}
	}
	metadata.UpdateLastSync(
		stats.FilesDownloaded,
		stats.FilesDeleted,
		stats.FilesSkipped,
		stats.BytesTransferred,
		stats.Duration.Milliseconds(),
	)
	sandboxModel.MetaData.Set(metadata)

	err = sandboxModel.Save(userSession.User)
	if err != nil {
		log.ErrorContext(err, req.Context())
		log.Infof("Warning: failed to update metadata after sync: %v", err)
	}

	// Return response with sync stats
	resp := &SyncSandboxResponse{
		SandboxID:        sandboxModel.ID().String(),
		FilesDownloaded:  stats.FilesDownloaded,
		FilesDeleted:     stats.FilesDeleted,
		FilesSkipped:     stats.FilesSkipped,
		BytesTransferred: stats.BytesTransferred,
		DurationMs:       stats.Duration.Milliseconds(),
	}

	return response.Success(resp)
}

// Creates a new sandbox using a premade template based on provider/agent for authenticated users
//
//	@Title			Create Sandbox
//	@Public
//	@Summary		Create sandbox
//	@Description	Creates a new sandbox using a premade template based on provider/agent
//	@Tags			Sandbox
//	@Accept			json
//	@Produce		json
//	@Param			data	body	CreateSandboxTemplateRequest	true	"Sandbox creation request"
//	@Success		200	{object}	response.SuccessResponse{data=sandbox.Sandbox}
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Router			/sandbox/ [post]
func authCreateSandbox(_ http.ResponseWriter, req *http.Request) (*sandbox.Sandbox, int, error) {
	// Get authenticated user session
	usr := helpers.GetLoadedUser(req)
	// Parse request body
	data, err := request.GetJSONPostAs[*CreateSandboxTemplateRequest](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*sandbox.Sandbox](err)
	}

	// Get premade template for provider/agent
	template, err := sandbox_service.GetSandboxTemplate(data.Type, data.AgentID)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*sandbox.Sandbox](err)
	}

	// Build config from template
	config := template.BuildSandboxConfig(usr.ID())

	// Create sandbox via service
	service := sandbox_service.NewSandboxService()
	sandboxInfo, err := service.CreateSandbox(req.Context(), usr.ID(), config)
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
	sandboxModel.AccountID.Set(usr.ID())
	if data.AgentID != "" {
		sandboxModel.AgentID.Set(data.AgentID)
	}
	sandboxModel.Type.Set(data.Type)
	sandboxModel.ExternalID.Set(sandboxInfo.SandboxID)
	sandboxModel.Status.Set(constants.STATUS_ACTIVE)
	sandboxModel.OrganizationID.Set(usr.OrganizationID.Get())
	sandboxModel.MetaData.Set(&sandbox.MetaData{})

	err = sandboxModel.Save(usr)
	if err != nil {
		log.ErrorContext(err, req.Context())
		// Note: Sandbox was created in Modal but DB save failed
		// TODO: Consider adding cleanup logic here or async cleanup task
		return response.AdminBadRequestError[*sandbox.Sandbox](err)
	}

	return response.Success(sandboxModel)
}
