package sandboxes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service"
)

// TODO: Implement authUpdate to allow updating sandbox metadata.
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
type SyncSandboxResponse struct {
	SandboxID        string `json:"sandbox_id"`
	FilesProcessed   int    `json:"files_processed"`
	BytesTransferred int64  `json:"bytes_transferred"`
	DurationMs       int64  `json:"duration_ms"`
}

// syncSandbox syncs the sandbox volume to S3 without terminating.
// This allows manual backups/snapshots of the current workspace state.
// Updates metadata with sync statistics (files processed, bytes transferred, duration).
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
	sandboxInfo := reconstructSandboxInfo(sandboxModel, accountID)

	// Sync to S3 via service layer
	service := sandbox_service.NewSandboxService()
	stats, err := service.SyncToS3(req.Context(), sandboxInfo)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*SyncSandboxResponse](err)
	}

	log.Infof("Synced sandbox %s to S3: %d files, %d bytes, %dms",
		id, stats.FilesProcessed, stats.BytesTransferred, stats.Duration.Milliseconds())

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
		stats.FilesProcessed,
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
		FilesProcessed:   stats.FilesProcessed,
		BytesTransferred: stats.BytesTransferred,
		DurationMs:       stats.Duration.Milliseconds(),
	}

	return response.Success(resp)
}
