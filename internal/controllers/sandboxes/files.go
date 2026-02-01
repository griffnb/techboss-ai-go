package sandboxes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service"
	"github.com/pkg/errors"
)

// parseFileListOptions extracts and validates query parameters from HTTP request.
// It applies default values for missing parameters and validates all inputs.
//
// Default values:
// - source: "volume"
// - page: 1
// - per_page: 100
// - recursive: true
// - path: "" (empty = workspace root)
func parseFileListOptions(req *http.Request) (*sandbox_service.FileListOptions, error) {
	query := req.URL.Query()

	// Parse source (default: "volume")
	source := query.Get("source")
	if source == "" {
		source = "volume"
	}
	if source != "volume" && source != "s3" {
		return nil, errors.New("source must be 'volume' or 's3'")
	}

	// Parse page (default: 1)
	page := 1
	if pageStr := query.Get("page"); pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err == nil {
			page = parsedPage
		}
	}
	if page < 1 {
		return nil, errors.New("page must be at least 1")
	}

	// Parse per_page (default: 100, max: 1000)
	perPage := 100
	if perPageStr := query.Get("per_page"); perPageStr != "" {
		parsedPerPage, err := strconv.Atoi(perPageStr)
		if err == nil {
			perPage = parsedPerPage
		}
	}
	if perPage < 1 || perPage > 1000 {
		return nil, errors.New("per_page must be between 1 and 1000")
	}

	// Parse recursive (default: true)
	recursive := true
	if recursiveStr := query.Get("recursive"); recursiveStr != "" {
		if recursiveStr == "false" || recursiveStr == "0" {
			recursive = false
		}
	}

	// Parse path (default: "")
	path := query.Get("path")

	return &sandbox_service.FileListOptions{
		Source:    source,
		Path:      path,
		Recursive: recursive,
		Page:      page,
		PerPage:   perPage,
	}, nil
}

// Lists files in a sandbox volume or S3 bucket with pagination
//
//	@Title			List Sandbox Files
//	@Summary		List files in sandbox
//	@Description	Lists files in a sandbox volume or S3 bucket with pagination
//	@Tags			Sandbox
//	@Tags			AdminOnly
//	@Accept			json
//	@Produce		json
//	@Param			id			path	string	true	"Sandbox ID"
//	@Param			source		query	string	false	"Source location (volume or s3)"	default(volume)	enums(volume, s3)
//	@Param			path		query	string	false	"Root path to list from (e.g., /workspace/src)"
//	@Param			recursive	query	bool	false	"Include subdirectories"	default(true)
//	@Param			page		query	int		false	"Page number"	default(1)	minimum(1)
//	@Param			per_page	query	int		false	"Items per page"	default(100)	minimum(1)	maximum(1000)
//	@Success		200	{object}	response.SuccessResponse{data=sandbox_service.FileListResponse}
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Router			/admin/sandbox/{id}/files [get]
func adminListFiles(_ http.ResponseWriter, req *http.Request) (*sandbox_service.FileListResponse, int, error) {
	// Extract sandbox ID from URL parameter
	id := chi.URLParam(req, "id")

	// Parse query parameters into FileListOptions
	opts, err := parseFileListOptions(req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*sandbox_service.FileListResponse](err)
	}

	// Load sandbox from database
	sandboxModel, err := sandbox.Get(req.Context(), types.UUID(id))
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*sandbox_service.FileListResponse](err)
	}

	// Reconstruct sandbox info for service layer
	sandboxInfo, err := sandbox_service.ReconstructSandboxInfo(req.Context(), sandboxModel, sandboxModel.AccountID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*sandbox_service.FileListResponse](err)
	}

	// Get user session for database updates
	userSession := request.GetReqSession(req)

	// Call service to list files (handles auto-restart and database updates internally)
	service := sandbox_service.NewSandboxService()
	files, err := service.ListFiles(req.Context(), sandboxInfo, sandboxModel, userSession.User, opts)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*sandbox_service.FileListResponse](err)
	}

	return response.Success(files)
}

// Lists files in a sandbox volume or S3 bucket with pagination for authenticated users
//
//	@Title			List Sandbox Files
//	@Public
//	@Summary		List files in sandbox
//	@Description	Lists files in a sandbox volume or S3 bucket with pagination
//	@Tags			Sandbox
//	@Accept			json
//	@Produce		json
//	@Param			id			path	string	true	"Sandbox ID"
//	@Param			source		query	string	false	"Source location (volume or s3)"	default(volume)	enums(volume, s3)
//	@Param			path		query	string	false	"Root path to list from (e.g., /workspace/src)"
//	@Param			recursive	query	bool	false	"Include subdirectories"	default(true)
//	@Param			page		query	int		false	"Page number"	default(1)	minimum(1)
//	@Param			per_page	query	int		false	"Items per page"	default(100)	minimum(1)	maximum(1000)
//	@Success		200	{object}	response.SuccessResponse{data=sandbox_service.FileListResponse}
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Router			/sandbox/{id}/files [get]
func authListFiles(_ http.ResponseWriter, req *http.Request) (*sandbox_service.FileListResponse, int, error) {
	// Extract sandbox ID from URL parameter
	id := chi.URLParam(req, "id")

	// Parse query parameters into FileListOptions
	opts, err := parseFileListOptions(req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*sandbox_service.FileListResponse]()
	}

	// Load sandbox from database
	sandboxModel, err := sandbox.Get(req.Context(), types.UUID(id))
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*sandbox_service.FileListResponse]()
	}

	// Reconstruct sandbox info for service layer
	sandboxInfo, err := sandbox_service.ReconstructSandboxInfo(req.Context(), sandboxModel, sandboxModel.AccountID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*sandbox_service.FileListResponse]()
	}

	// Get user session for database updates
	userSession := request.GetReqSession(req)

	// Call service to list files (handles auto-restart and database updates internally)
	service := sandbox_service.NewSandboxService()
	files, err := service.ListFiles(req.Context(), sandboxInfo, sandboxModel, userSession.User, opts)
	if err != nil {
		log.ErrorContext(errors.WithMessagef(err, "failed pulling for sandbox id:%s ExternalID: %s", id, sandboxInfo.SandboxID), req.Context())
		return response.PublicBadRequestError[*sandbox_service.FileListResponse]()
	}

	return response.Success(files)
}

// Retrieves and serves raw file content from a sandbox
//
//	@Title			Get File Content
//	@Summary		Get file content from sandbox
//	@Description	Retrieves and serves the raw content of a file from a sandbox (not wrapped in JSON)
//	@Tags			Sandbox
//	@Tags			AdminOnly
//	@Accept			json
//	@Produce		octet-stream
//	@Param			id			path	string	true	"Sandbox ID"
//	@Param			file_path	query	string	true	"Full path to the file (e.g., /workspace/test.txt)"
//	@Param			source		query	string	false	"Source location (volume or s3)"	default(volume)	enums(volume, s3)
//	@Param			download	query	bool	false	"Download as attachment (true) or display inline/stream (false)"	default(false)
//	@Success		200	{file}		binary	"File content"
//	@Failure		400	{string}	string	"Bad Request"
//	@Failure		404	{string}	string	"Not Found"
//	@Failure		500	{string}	string	"Internal Server Error"
//	@Router			/admin/sandbox/{id}/files/content [get]
func adminGetFileContent(w http.ResponseWriter, req *http.Request) {
	// Extract sandbox ID from URL parameter
	id := chi.URLParam(req, "id")

	// Parse query parameters
	source := req.URL.Query().Get("source")
	if source == "" {
		source = "volume"
	}

	filePath := req.URL.Query().Get("file_path")
	if filePath == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("file_path query parameter is required"))
		return
	}

	// Parse download parameter (default: false = inline/stream)
	download := req.URL.Query().Get("download") == "true"

	// Load sandbox from database
	sandboxModel, err := sandbox.Get(req.Context(), types.UUID(id))
	if err != nil {
		log.ErrorContext(err, req.Context())
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	// Reconstruct sandbox info for service layer
	sandboxInfo, err := sandbox_service.ReconstructSandboxInfo(req.Context(), sandboxModel, sandboxModel.AccountID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	// Get user session for database updates
	userSession := request.GetReqSession(req)

	// Call service to get file content (handles auto-restart and database updates internally)
	service := sandbox_service.NewSandboxService()
	fileContent, err := service.GetFileContent(req.Context(), sandboxInfo, sandboxModel, userSession.User, source, filePath)
	if err != nil {
		log.ErrorContext(err, req.Context())
		// Check if error message contains "file not found" to return 404
		errMsg := err.Error()
		if len(errMsg) >= 14 && errMsg[:14] == "file not found" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(err.Error()))
		} else {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
		}
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", fileContent.ContentType)
	if download {
		w.Header().Set("Content-Disposition", "attachment; filename=\""+fileContent.FileName+"\"")
	} else {
		w.Header().Set("Content-Disposition", "inline; filename=\""+fileContent.FileName+"\"")
	}
	w.Header().Set("Content-Length", strconv.FormatInt(fileContent.Size, 10))

	// Write content
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(fileContent.Content)
}

// Retrieves and serves raw file content from a sandbox for authenticated users
//
//	@Title			Get File Content
//	@Public
//	@Summary		Get file content from sandbox
//	@Description	Retrieves and serves the raw content of a file from a sandbox (not wrapped in JSON)
//	@Tags			Sandbox
//	@Accept			json
//	@Produce		octet-stream
//	@Param			id			path	string	true	"Sandbox ID"
//	@Param			file_path	query	string	true	"Full path to the file (e.g., /workspace/test.txt)"
//	@Param			source		query	string	false	"Source location (volume or s3)"	default(volume)	enums(volume, s3)
//	@Param			download	query	bool	false	"Download as attachment (true) or display inline/stream (false)"	default(false)
//	@Success		200	{file}		binary	"File content"
//	@Failure		400	{string}	string	"Bad Request"
//	@Failure		404	{string}	string	"Not Found"
//	@Failure		500	{string}	string	"Internal Server Error"
//	@Router			/sandbox/{id}/files/content [get]
func authGetFileContent(w http.ResponseWriter, req *http.Request) {
	// Extract sandbox ID from URL parameter
	id := chi.URLParam(req, "id")

	// Parse query parameters
	source := req.URL.Query().Get("source")
	if source == "" {
		source = "volume"
	}

	filePath := req.URL.Query().Get("file_path")
	if filePath == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("file_path query parameter is required"))
		return
	}

	// Parse download parameter (default: false = inline/stream)
	download := req.URL.Query().Get("download") == "true"

	// Load sandbox from database
	sandboxModel, err := sandbox.Get(req.Context(), types.UUID(id))
	if err != nil {
		log.ErrorContext(err, req.Context())
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	// Reconstruct sandbox info for service layer
	sandboxInfo, err := sandbox_service.ReconstructSandboxInfo(req.Context(), sandboxModel, sandboxModel.AccountID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	// Get user session for database updates
	userSession := request.GetReqSession(req)

	// Call service to get file content (handles auto-restart and database updates internally)
	service := sandbox_service.NewSandboxService()
	fileContent, err := service.GetFileContent(req.Context(), sandboxInfo, sandboxModel, userSession.User, source, filePath)
	if err != nil {
		log.ErrorContext(err, req.Context())
		// Check if error message contains "file not found" to return 404
		errMsg := err.Error()
		if len(errMsg) >= 14 && errMsg[:14] == "file not found" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(err.Error()))
		} else {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
		}
		return
	}
	if download {
		w.Header().Set("Content-Disposition", "attachment; filename=\""+fileContent.FileName+"\"")
	} else {
		w.Header().Set("Content-Disposition", "inline; filename=\""+fileContent.FileName+"\"")
	}
	// Set response headers
	w.Header().Set("Content-Type", fileContent.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(fileContent.Size, 10))

	// Write content
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(fileContent.Content)
}

// Retrieves a hierarchical tree structure of files in a sandbox
//
//	@Title			Get File Tree
//	@Summary		Get file tree structure
//	@Description	Retrieves a hierarchical tree structure of files in a sandbox
//	@Tags			Sandbox
//	@Tags			AdminOnly
//	@Accept			json
//	@Produce		json
//	@Param			id			path	string	true	"Sandbox ID"
//	@Param			source		query	string	false	"Source location (volume or s3)"	default(volume)	enums(volume, s3)
//	@Param			path		query	string	false	"Root path to list from"	default()
//	@Param			recursive	query	bool	false	"Include subdirectories"	default(true)
//	@Success		200	{object}	response.SuccessResponse{data=sandbox_service.FileTreeNode}
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Router			/admin/sandbox/{id}/files/tree [get]
func adminGetFileTree(_ http.ResponseWriter, req *http.Request) (*sandbox_service.FileTreeNode, int, error) {
	// Extract sandbox ID from URL parameter
	id := chi.URLParam(req, "id")

	// Parse query parameters into FileListOptions
	opts, err := parseFileListOptions(req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*sandbox_service.FileTreeNode](err)
	}

	// Load sandbox from database
	sandboxModel, err := sandbox.Get(req.Context(), types.UUID(id))
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*sandbox_service.FileTreeNode](err)
	}

	// Reconstruct sandbox info for service layer
	sandboxInfo, err := sandbox_service.ReconstructSandboxInfo(req.Context(), sandboxModel, sandboxModel.AccountID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*sandbox_service.FileTreeNode](err)
	}

	// Get user session for database updates
	userSession := request.GetReqSession(req)

	// Call service to list files first (handles auto-restart and database updates internally)
	service := sandbox_service.NewSandboxService()
	fileList, err := service.ListFiles(req.Context(), sandboxInfo, sandboxModel, userSession.User, opts)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*sandbox_service.FileTreeNode](err)
	}

	// Determine root path for tree building - use actual mount paths
	// The files returned from ListFiles have mount paths (/mnt/workspace, /mnt/s3-bucket)
	// so we need to match those paths when building the tree
	rootPath := sandbox_service.VOLUME_MOUNT_PATH // /mnt/workspace
	if opts.Source == "s3" {
		rootPath = sandbox_service.S3_MOUNT_PATH // /mnt/s3-bucket
	}
	if opts.Path != "" && opts.Path != "/" {
		// Convert user-facing path to mount path
		if strings.HasPrefix(opts.Path, "/workspace") {
			rootPath = strings.Replace(opts.Path, "/workspace", sandbox_service.VOLUME_MOUNT_PATH, 1)
		} else if strings.HasPrefix(opts.Path, "/s3-bucket") {
			rootPath = strings.Replace(opts.Path, "/s3-bucket", sandbox_service.S3_MOUNT_PATH, 1)
		}
	}

	// Build tree structure from flat file list using mount paths
	tree, err := service.BuildFileTree(fileList.Files, rootPath)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*sandbox_service.FileTreeNode](err)
	}

	// Convert mount paths back to user-facing paths for API response
	sandbox_service.ConvertTreePathsToUserFacing(tree)

	return response.Success(tree)
}

// Retrieves a hierarchical tree structure of files in a sandbox for authenticated users
//
//	@Title			Get File Tree
//	@Public
//	@Summary		Get file tree structure
//	@Description	Retrieves a hierarchical tree structure of files in a sandbox
//	@Tags			Sandbox
//	@Accept			json
//	@Produce		json
//	@Param			id			path	string	true	"Sandbox ID"
//	@Param			source		query	string	false	"Source location (volume or s3)"	default(volume)	enums(volume, s3)
//	@Param			path		query	string	false	"Root path to list from"	default()
//	@Param			recursive	query	bool	false	"Include subdirectories"	default(true)
//	@Success		200	{object}	response.SuccessResponse{data=sandbox_service.FileTreeNode}
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Router			/sandbox/{id}/files/tree [get]
func authGetFileTree(_ http.ResponseWriter, req *http.Request) (*sandbox_service.FileTreeNode, int, error) {
	// Extract sandbox ID from URL parameter
	id := chi.URLParam(req, "id")

	// Parse query parameters into FileListOptions
	opts, err := parseFileListOptions(req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*sandbox_service.FileTreeNode]()
	}

	// Load sandbox from database
	sandboxModel, err := sandbox.Get(req.Context(), types.UUID(id))
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*sandbox_service.FileTreeNode]()
	}

	// Reconstruct sandbox info for service layer
	sandboxInfo, err := sandbox_service.ReconstructSandboxInfo(req.Context(), sandboxModel, sandboxModel.AccountID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*sandbox_service.FileTreeNode]()
	}

	// Get user session for database updates
	userSession := request.GetReqSession(req)

	// Call service to list files first (handles auto-restart and database updates internally)
	service := sandbox_service.NewSandboxService()
	fileList, err := service.ListFiles(req.Context(), sandboxInfo, sandboxModel, userSession.User, opts)
	if err != nil {
		log.ErrorContext(errors.WithMessagef(err, "failed pulling for sandbox id:%s ExternalID: %s", id, sandboxInfo.SandboxID), req.Context())
		return response.PublicBadRequestError[*sandbox_service.FileTreeNode]()
	}

	// Determine root path for tree building - use actual mount paths
	// The files returned from ListFiles have mount paths (/mnt/workspace, /mnt/s3-bucket)
	// so we need to match those paths when building the tree
	rootPath := sandbox_service.VOLUME_MOUNT_PATH // /mnt/workspace
	if opts.Source == "s3" {
		rootPath = sandbox_service.S3_MOUNT_PATH // /mnt/s3-bucket
	}
	if opts.Path != "" && opts.Path != "/" {
		// Convert user-facing path to mount path
		if strings.HasPrefix(opts.Path, "/workspace") {
			rootPath = strings.Replace(opts.Path, "/workspace", sandbox_service.VOLUME_MOUNT_PATH, 1)
		} else if strings.HasPrefix(opts.Path, "/s3-bucket") {
			rootPath = strings.Replace(opts.Path, "/s3-bucket", sandbox_service.S3_MOUNT_PATH, 1)
		}
	}

	// Build tree structure from flat file list using mount paths
	tree, err := service.BuildFileTree(fileList.Files, rootPath)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*sandbox_service.FileTreeNode]()
	}

	// Convert mount paths back to user-facing paths for API response
	sandbox_service.ConvertTreePathsToUserFacing(tree)

	return response.Success(tree)
}
