package sandboxes

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/log"
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

// adminListFiles lists files in a sandbox volume or S3 bucket with pagination.
// It extracts the sandbox ID from URL, parses query parameters, loads the sandbox,
// and calls the sandbox service to retrieve files.
//
// Query parameters:
// - source: "volume" or "s3" (default: "volume")
// - path: Root path to list from (e.g., "/workspace/src")
// - recursive: Include subdirectories (default: true)
// - page: Page number (default: 1)
// - per_page: Items per page (default: 100, max: 1000)
//
// Returns FileListResponse with files array, pagination metadata, and HTTP status.
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

	// Call service to list files
	service := sandbox_service.NewSandboxService()
	files, err := service.ListFiles(req.Context(), sandboxInfo, opts)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*sandbox_service.FileListResponse](err)
	}

	return response.Success(files)
}

// authListFiles lists files in a sandbox volume or S3 bucket with pagination for authenticated users.
// The auth framework verifies ownership automatically, ensuring only the sandbox owner can access.
//
// Query parameters:
// - source: "volume" or "s3" (default: "volume")
// - path: Root path to list from (e.g., "/workspace/src")
// - recursive: Include subdirectories (default: true)
// - page: Page number (default: 1)
// - per_page: Items per page (default: 100, max: 1000)
//
// Returns FileListResponse with files array, pagination metadata, and HTTP status.
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

	// Call service to list files
	service := sandbox_service.NewSandboxService()
	files, err := service.ListFiles(req.Context(), sandboxInfo, opts)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*sandbox_service.FileListResponse]()
	}

	return response.Success(files)
}

// adminGetFileContent retrieves and serves the raw content of a file from a sandbox.
// This endpoint returns the actual file content directly, not wrapped in JSON.
// It sets appropriate Content-Type, Content-Disposition, and Content-Length headers.
//
// Query parameters:
// - file_path: Required. Full path to the file (e.g., "/workspace/test.txt")
// - source: "volume" or "s3" (default: "volume")
//
// Response headers:
// - Content-Type: MIME type based on file extension
// - Content-Disposition: attachment; filename="<filename>"
// - Content-Length: Size of the file in bytes
//
// Returns:
// - 200 OK: File content successfully retrieved
// - 400 Bad Request: Missing file_path parameter or invalid sandbox ID
// - 404 Not Found: File does not exist at the specified path
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

	// Call service to get file content
	service := sandbox_service.NewSandboxService()
	fileContent, err := service.GetFileContent(req.Context(), sandboxInfo, source, filePath)
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
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileContent.FileName+"\"")
	w.Header().Set("Content-Length", strconv.FormatInt(fileContent.Size, 10))

	// Write content
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(fileContent.Content)
}

// authGetFileContent retrieves and serves the raw content of a file from a sandbox for authenticated users.
// The auth framework verifies ownership automatically, ensuring only the sandbox owner can access.
// This endpoint returns the actual file content directly, not wrapped in JSON.
// It sets appropriate Content-Type, Content-Disposition, and Content-Length headers.
//
// Query parameters:
// - file_path: Required. Full path to the file (e.g., "/workspace/test.txt")
// - source: "volume" or "s3" (default: "volume")
//
// Response headers:
// - Content-Type: MIME type based on file extension
// - Content-Disposition: attachment; filename="<filename>"
// - Content-Length: Size of the file in bytes
//
// Returns:
// - 200 OK: File content successfully retrieved
// - 400 Bad Request: Missing file_path parameter or invalid sandbox ID
// - 404 Not Found: File does not exist at the specified path
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

	// Call service to get file content
	service := sandbox_service.NewSandboxService()
	fileContent, err := service.GetFileContent(req.Context(), sandboxInfo, source, filePath)
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
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileContent.FileName+"\"")
	w.Header().Set("Content-Length", strconv.FormatInt(fileContent.Size, 10))

	// Write content
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(fileContent.Content)
}

// adminGetFileTree retrieves a hierarchical tree structure of files in a sandbox.
// It extracts the sandbox ID from URL, parses query parameters (source, path),
// loads the sandbox, lists files, and builds a tree structure.
//
// Query parameters:
// - source: "volume" or "s3" (default: "volume")
// - path: Root path to list from (default: "" = workspace root)
// - recursive: Include subdirectories (default: true)
//
// Returns FileTreeNode with nested children structure and HTTP status.
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

	// Call service to list files first
	service := sandbox_service.NewSandboxService()
	fileList, err := service.ListFiles(req.Context(), sandboxInfo, opts)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*sandbox_service.FileTreeNode](err)
	}

	// Determine root path based on source
	rootPath := "/workspace"
	if opts.Source == "s3" {
		rootPath = "/s3-bucket"
	}
	if opts.Path != "" && opts.Path != "/" {
		rootPath = opts.Path
	}

	// Build tree structure from flat file list
	tree, err := service.BuildFileTree(fileList.Files, rootPath)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*sandbox_service.FileTreeNode](err)
	}

	return response.Success(tree)
}

// authGetFileTree retrieves a hierarchical tree structure of files in a sandbox for authenticated users.
// The auth framework verifies ownership automatically, ensuring only the sandbox owner can access.
//
// Query parameters:
// - source: "volume" or "s3" (default: "volume")
// - path: Root path to list from (default: "" = workspace root)
// - recursive: Include subdirectories (default: true)
//
// Returns FileTreeNode with nested children structure and HTTP status.
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

	// Call service to list files first
	service := sandbox_service.NewSandboxService()
	fileList, err := service.ListFiles(req.Context(), sandboxInfo, opts)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*sandbox_service.FileTreeNode]()
	}

	// Determine root path based on source
	rootPath := "/workspace"
	if opts.Source == "s3" {
		rootPath = "/s3-bucket"
	}
	if opts.Path != "" && opts.Path != "/" {
		rootPath = opts.Path
	}

	// Build tree structure from flat file list
	tree, err := service.BuildFileTree(fileList.Files, rootPath)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*sandbox_service.FileTreeNode]()
	}

	return response.Success(tree)
}
