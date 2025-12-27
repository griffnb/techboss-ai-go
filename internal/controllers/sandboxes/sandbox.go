package sandboxes

import (
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service"
	"github.com/pkg/errors"
)

// sandboxCache provides in-memory storage for sandbox info by ID.
// This is a temporary solution for Phase 1 testing. Phase 2 will implement
// proper database persistence for production use.
// Thread-safe with sync.Map for concurrent access.
var sandboxCache = sync.Map{}

// CreateSandboxRequest holds request data for sandbox creation.
// It defines the Docker image, storage volumes, and S3 configuration for the sandbox.
// Supports either image_template for pre-configured images or custom image_base + dockerfile_commands.
type CreateSandboxRequest struct {
	ImageTemplate      string   `json:"image_template"`      // Pre-configured template (e.g., "claude")
	ImageBase          string   `json:"image_base"`          // Custom base image (required if no template)
	DockerfileCommands []string `json:"dockerfile_commands"` // Custom Dockerfile commands (optional)
	VolumeName         string   `json:"volume_name"`         // Volume name for persistent storage
	S3BucketName       string   `json:"s3_bucket_name"`      // S3 bucket name (optional)
	S3KeyPrefix        string   `json:"s3_key_prefix"`       // S3 key prefix (optional)
	InitFromS3         bool     `json:"init_from_s3"`        // Initialize volume from S3
}

// CreateSandboxResponse holds response data for sandbox creation.
// It returns the sandbox ID, status, and creation timestamp.
type CreateSandboxResponse struct {
	SandboxID string    `json:"sandbox_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// createSandbox creates a new sandbox for the authenticated user.
// It builds the sandbox configuration from the request, creates the sandbox via the service layer,
// and optionally initializes the volume from S3 if requested.
//
// TODO: Store sandboxInfo in database/cache for later retrieval (Phase 2)
func createSandbox(_ http.ResponseWriter, req *http.Request) (*CreateSandboxResponse, int, error) {
	// Get authenticated user session
	userSession := request.GetReqSession(req)
	accountID := userSession.User.ID()

	// Parse request body
	data, err := request.GetJSONPostAs[*CreateSandboxRequest](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*CreateSandboxResponse](err)
	}

	// Validate and build image config (either from template or custom)
	var imageConfig *modal.ImageConfig

	if !tools.Empty(data.ImageTemplate) {
		// Use pre-configured template
		imageConfig = modal.GetImageConfigFromTemplate(data.ImageTemplate)
		if imageConfig == nil {
			err := errors.Errorf("unknown image_template: %s", data.ImageTemplate)
			log.ErrorContext(err, req.Context())
			return response.AdminBadRequestError[*CreateSandboxResponse](err)
		}
	} else if !tools.Empty(data.ImageBase) {
		// Use custom image configuration
		imageConfig = &modal.ImageConfig{
			BaseImage:          data.ImageBase,
			DockerfileCommands: data.DockerfileCommands,
		}
	} else {
		// Neither template nor custom image provided
		err := errors.New("either image_template or image_base is required")
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*CreateSandboxResponse](err)
	}

	// Build sandbox config
	config := &modal.SandboxConfig{
		AccountID:       accountID,
		Image:           imageConfig,
		VolumeName:      data.VolumeName,
		VolumeMountPath: "/mnt/workspace",
		Workdir:         "/mnt/workspace",
	}

	// Add S3 config if provided
	if !tools.Empty(data.S3BucketName) {
		config.S3Config = &modal.S3MountConfig{
			BucketName: data.S3BucketName,
			SecretName: "s3-bucket", // Default secret name
			KeyPrefix:  data.S3KeyPrefix,
			MountPath:  "/mnt/s3-bucket",
			ReadOnly:   true,
		}
	}

	// Create sandbox via service
	service := sandbox_service.NewSandboxService()
	sandboxInfo, err := service.CreateSandbox(req.Context(), accountID, config)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*CreateSandboxResponse](err)
	}

	// Initialize from S3 if requested
	if data.InitFromS3 && config.S3Config != nil {
		_, err := service.InitFromS3(req.Context(), sandboxInfo)
		if err != nil {
			log.ErrorContext(err, req.Context())
			// Continue even if init fails - non-fatal
			log.Infof("Warning: failed to initialize from S3: %v", err)
		}
	}

	// TODO (Phase 2): Store sandboxInfo in database for persistence across restarts
	// TEMPORARY: Store in memory for Phase 1 testing
	// This cache is session-scoped and will be lost on server restart
	// For production, implement proper database persistence with a modal_sandboxes table
	sandboxCache.Store(sandboxInfo.SandboxID, sandboxInfo)
	log.Infof("Stored sandbox %s in memory cache", sandboxInfo.SandboxID)

	// Return response
	resp := &CreateSandboxResponse{
		SandboxID: sandboxInfo.SandboxID,
		Status:    string(sandboxInfo.Status),
		CreatedAt: sandboxInfo.CreatedAt,
	}

	return response.Success(resp)
}

// getSandbox retrieves sandbox status by ID.
// Currently uses in-memory cache for Phase 1 testing. Phase 2 will add database persistence.
//
// TODO: Retrieve sandboxInfo from database (Phase 2)
func getSandbox(_ http.ResponseWriter, req *http.Request) (*CreateSandboxResponse, int, error) {
	sandboxID := chi.URLParam(req, "sandboxID")

	log.Infof("getSandbox called with sandboxID: %s", sandboxID)

	// Retrieve from in-memory cache (temporary solution)
	// TODO (Phase 2): Query database instead of memory cache
	value, ok := sandboxCache.Load(sandboxID)
	if !ok {
		err := errors.Errorf("sandbox not found: %s", sandboxID)
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*CreateSandboxResponse](err)
	}

	sandboxInfo := value.(*modal.SandboxInfo)

	// Return response
	resp := &CreateSandboxResponse{
		SandboxID: sandboxInfo.SandboxID,
		Status:    string(sandboxInfo.Status),
		CreatedAt: sandboxInfo.CreatedAt,
	}

	return response.Success(resp)
}

// deleteSandbox terminates a sandbox by ID.
// Currently uses in-memory cache for Phase 1 testing. Phase 2 will add database persistence.
//
// TODO: Retrieve sandboxInfo from database (Phase 2)
func deleteSandbox(_ http.ResponseWriter, req *http.Request) (*CreateSandboxResponse, int, error) {
	sandboxID := chi.URLParam(req, "sandboxID")

	log.Infof("deleteSandbox called with sandboxID: %s", sandboxID)

	// Retrieve from in-memory cache (temporary solution)
	// TODO (Phase 2): Query database instead of memory cache
	value, ok := sandboxCache.Load(sandboxID)
	if !ok {
		err := errors.Errorf("sandbox not found: %s", sandboxID)
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*CreateSandboxResponse](err)
	}

	sandboxInfo := value.(*modal.SandboxInfo)

	// Terminate sandbox with S3 sync
	// The true parameter triggers volume sync to S3 before termination
	// This preserves the final workspace state
	service := sandbox_service.NewSandboxService()
	err := service.TerminateSandbox(req.Context(), sandboxInfo, true)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*CreateSandboxResponse](err)
	}

	// Remove from cache
	sandboxCache.Delete(sandboxID)
	log.Infof("Terminated and removed sandbox %s from cache", sandboxID)

	// Return response
	resp := &CreateSandboxResponse{
		SandboxID: sandboxInfo.SandboxID,
		Status:    "terminated",
		CreatedAt: sandboxInfo.CreatedAt,
	}

	return response.Success(resp)
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
//
// TODO: Retrieve sandboxInfo from database (Phase 2)
func syncSandbox(_ http.ResponseWriter, req *http.Request) (*SyncSandboxResponse, int, error) {
	sandboxID := chi.URLParam(req, "sandboxID")

	log.Infof("syncSandbox called with sandboxID: %s", sandboxID)

	// Retrieve from in-memory cache (temporary solution)
	// TODO (Phase 2): Query database instead of memory cache
	value, ok := sandboxCache.Load(sandboxID)
	if !ok {
		err := errors.Errorf("sandbox not found: %s", sandboxID)
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*SyncSandboxResponse](err)
	}

	sandboxInfo := value.(*modal.SandboxInfo)

	// Sync to S3 via service layer
	service := sandbox_service.NewSandboxService()
	stats, err := service.SyncToS3(req.Context(), sandboxInfo)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*SyncSandboxResponse](err)
	}

	log.Infof("Synced sandbox %s to S3: %d files, %d bytes, %dms",
		sandboxID, stats.FilesProcessed, stats.BytesTransferred, stats.Duration.Milliseconds())

	// Return response with sync stats
	resp := &SyncSandboxResponse{
		SandboxID:        sandboxInfo.SandboxID,
		FilesProcessed:   stats.FilesProcessed,
		BytesTransferred: stats.BytesTransferred,
		DurationMs:       stats.Duration.Milliseconds(),
	}

	return response.Success(resp)
}
