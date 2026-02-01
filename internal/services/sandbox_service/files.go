package sandbox_service

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/model/coremodel"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
	"github.com/pkg/errors"
)

const (
	VOLUME_MOUNT_PATH   = "/mnt/workspace"
	USER_WORKSPACE_PATH = "/workspace"
	S3_MOUNT_PATH       = "/mnt/s3-bucket"
	S3_USER_PATH        = "/s3-bucket"
)

// FileListOptions contains options for listing files in a sandbox.
// It supports filtering by source (volume/s3), pagination, and path filtering.
type FileListOptions struct {
	Source    string // "volume" or "s3"
	Path      string // Root path to list from (e.g., "/workspace/src")
	Recursive bool   // Include subdirectories
	Page      int    // Page number (1-indexed)
	PerPage   int    // Items per page (1-1000)
}

// FileInfo contains metadata about a single file or directory.
// This is returned for each file in a file listing response.
type FileInfo struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	ModifiedAt  time.Time `json:"modified_at"`
	IsDirectory bool      `json:"is_directory"`
	Checksum    string    `json:"checksum,omitempty"`
}

// FileListResponse contains a paginated list of files with metadata.
// This is the response format for file listing endpoints.
type FileListResponse struct {
	Files      []FileInfo `json:"files"`
	TotalCount int        `json:"total_count"`
	Page       int        `json:"page"`
	PerPage    int        `json:"per_page"`
	TotalPages int        `json:"total_pages"`
}

// FileContent represents the content of a file retrieved from a sandbox.
// Used when returning file content via the file content endpoint.
type FileContent struct {
	Content     []byte `json:"content"`
	ContentType string `json:"content_type"`
	FileName    string `json:"file_name"`
	Size        int64  `json:"size"`
}

// FileTreeNode represents a node in a hierarchical file tree structure.
// Used for tree-based file listing views where files and directories are
// organized in a parent-child relationship. Each node contains its own
// metadata and can have child nodes (subdirectories or files).
type FileTreeNode struct {
	Name        string          `json:"name"`
	Path        string          `json:"path"`
	IsDirectory bool            `json:"is_directory"`
	Size        int64           `json:"size,omitempty"`
	ModifiedAt  time.Time       `json:"modified_at,omitempty"`
	Children    []*FileTreeNode `json:"children,omitempty"`
}

// validatePath validates that a file path is safe and within allowed directories.
// It prevents directory traversal attacks and ensures paths are within
// /workspace or /s3-bucket directories.
// Returns an error if the path is invalid.
func validatePath(path string) error {
	// Empty path is treated as workspace root
	if path == "" {
		return nil
	}

	// Prevent directory traversal
	if strings.Contains(path, "..") {
		return errors.New("invalid path: directory traversal not allowed")
	}

	// Ensure path is within allowed directories
	if !strings.HasPrefix(path, VOLUME_MOUNT_PATH) && !strings.HasPrefix(path, S3_MOUNT_PATH) {
		return errors.Errorf("invalid path: must be within %s or %s", VOLUME_MOUNT_PATH, S3_MOUNT_PATH)
	}

	return nil
}

func cleanPath(path string) string {
	// Replace user-facing paths with mount paths
	if strings.HasPrefix(path, USER_WORKSPACE_PATH) {
		return strings.Replace(path, USER_WORKSPACE_PATH, VOLUME_MOUNT_PATH, 1)
	} else if strings.HasPrefix(path, S3_USER_PATH) {
		return strings.Replace(path, S3_USER_PATH, S3_MOUNT_PATH, 1)
	}
	return path
}

func isRootPath(path string) bool {
	return path == VOLUME_MOUNT_PATH || path == S3_MOUNT_PATH
}

// buildListFilesCommand constructs a find command for listing files based on options.
// It generates the appropriate find command based on source (volume/s3), path, and recursive flag.
// The command outputs metadata in the format: path|size|timestamp|type
//
// For volume source: uses /mnt/workspace as base path
// For s3 source: uses /mnt/s3-bucket as base path
// For non-recursive: includes -maxdepth 1 flag
// Always includes -type f -o -type d to find both files and directories
// Uses BusyBox-compatible commands (find + stat) instead of GNU find's -printf
//
// Path mapping: User-facing paths (/workspace, /s3-bucket) are converted to actual
// mount paths (/mnt/workspace, /mnt/s3-bucket) for sandbox execution.
func (s *SandboxService) buildListFilesCommand(opts *FileListOptions) string {
	// Determine base path based on source
	var basePath string
	if opts.Source == "s3" {
		basePath = S3_MOUNT_PATH // /mnt/s3-bucket
	} else {
		basePath = VOLUME_MOUNT_PATH // /mnt/workspace
	}

	opts.Path = cleanPath(opts.Path)

	// Build target path by converting user-facing paths to mount paths
	targetPath := basePath
	if opts.Path != "" && opts.Path != "/" {
		// Convert user-facing paths to actual mount paths
		// /workspace/test -> /mnt/workspace/test
		// /s3-bucket/test -> /mnt/s3-bucket/test
		if !isRootPath(opts.Path) {
			// If path doesn't start with expected prefix, append to basePath
			// This handles relative paths (though they should be rejected by validation)
			targetPath = basePath + "/" + strings.TrimPrefix(opts.Path, "/")
		}
	}

	// Build command string using BusyBox-compatible find + stat
	// We use find to get paths, then stat to get metadata for each file
	// Use -L flag to follow symbolic links (workspace paths are often symlinks)
	cmd := "find -L " + targetPath

	// Add maxdepth for non-recursive
	if !opts.Recursive {
		cmd += " -maxdepth 1"
	}

	// Add type filters (files or directories)
	cmd += " \\( -type f -o -type d \\)"

	// Use stat to output metadata in format: path|size|timestamp|type
	// BusyBox stat uses different format strings than GNU stat
	// Format: %n = name, %s = size, %Y = mtime (epoch), %F = file type
	cmd += " -exec stat -c '%n|%s|%Y|%F' {} \\;"

	return cmd
}

// parseFileMetadata parses stat command output into FileInfo structs.
// The input format is "path|size|timestamp|type" where:
// - path is the full file path
// - size is the file size in bytes (int64)
// - timestamp is Unix epoch time (integer)
// - type is the file type string from stat -c '%F' (e.g., "regular file", "directory")
//
// The excludePath parameter filters out the specified directory from results.
// This is typically used to exclude the root directory itself when listing its contents.
//
// Returns a slice of FileInfo structs and an error if parsing fails.
// Empty lines in the output are skipped.
func parseFileMetadata(output string, excludePath string) ([]FileInfo, error) {
	var files []FileInfo

	lines := strings.Split(output, "\n")
	for i, line := range lines {
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Split by delimiter
		parts := strings.Split(line, "|")
		if len(parts) != 4 {
			return nil, errors.Errorf("malformed metadata on line %d: expected 4 fields, got %d", i+1, len(parts))
		}

		path := parts[0]

		// Skip the root directory itself (we only want its contents)
		if path == excludePath {
			continue
		}

		sizeStr := parts[1]
		timestampStr := parts[2]
		typeStr := parts[3]

		// Parse size
		size, err := strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid size on line %d", i+1)
		}

		// Parse timestamp as integer (Unix epoch seconds)
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid timestamp on line %d", i+1)
		}
		modifiedAt := time.Unix(timestamp, 0)

		// Determine if directory based on stat %F output
		// BusyBox stat outputs: "regular file", "directory", "symbolic link", etc.
		isDirectory := strings.Contains(typeStr, "directory")

		// Extract name from path using the last component
		name := path
		if idx := strings.LastIndex(path, "/"); idx != -1 {
			name = path[idx+1:]
		}

		files = append(files, FileInfo{
			Name:        name,
			Path:        path,
			Size:        size,
			ModifiedAt:  modifiedAt,
			IsDirectory: isDirectory,
			Checksum:    "", // Empty for now
		})
	}

	return files, nil
}

// paginateFiles slices a file array based on page and per_page parameters.
// It calculates TotalCount, TotalPages, and returns the appropriate slice of files.
// If page <= 0, it is treated as page 1.
// If the page is out of range, it returns an empty Files slice.
// For empty file lists, TotalPages is 0.
func paginateFiles(files []FileInfo, opts *FileListOptions) *FileListResponse {
	totalCount := len(files)

	// Handle page <= 0 as page 1
	page := opts.Page
	if page <= 0 {
		page = 1
	}

	perPage := opts.PerPage

	// Calculate total pages (ceiling division)
	var totalPages int
	if totalCount == 0 {
		totalPages = 0
	} else {
		totalPages = (totalCount + perPage - 1) / perPage
	}

	// Calculate slice bounds
	start := (page - 1) * perPage
	end := start + perPage

	// Handle out of range
	var paginatedFiles []FileInfo
	if start >= totalCount {
		paginatedFiles = []FileInfo{}
	} else {
		if end > totalCount {
			end = totalCount
		}
		paginatedFiles = files[start:end]
	}

	return &FileListResponse{
		Files:      paginatedFiles,
		TotalCount: totalCount,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}
}

// parseFileListOptions extracts and validates query parameters from HTTP request.
// It applies default values for missing parameters and validates all inputs.
// Returns FileListOptions and an error if validation fails.
//
// Default values:
// - source: "volume"
// - page: 1
// - per_page: 100
// - recursive: true
// - path: "" (empty = workspace root)
func parseFileListOptions(req *http.Request) (*FileListOptions, error) {
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

	// Validate path
	if err := validatePath(path); err != nil {
		return nil, err
	}

	return &FileListOptions{
		Source:    source,
		Path:      path,
		Recursive: recursive,
		Page:      page,
		PerPage:   perPage,
	}, nil
}

// ListFiles retrieves a paginated list of files from the sandbox volume or S3 bucket.
// It executes a find command in the sandbox, parses the output, and applies pagination.
// The method validates inputs, builds a find command with metadata output, executes it
// via the Modal sandbox, parses the results, and returns a paginated response.
//
// If sandboxModel and user are provided, automatically handles database updates if sandbox auto-restarts.
//
// Returns an error if:
// - sandboxInfo or opts are nil
// - source is not "volume" or "s3"
// - path validation fails
// - command execution fails
// - output parsing fails
func (s *SandboxService) ListFiles(
	ctx context.Context,
	sandboxInfo *modal.SandboxInfo,
	sandboxModel *sandbox.Sandbox,
	user coremodel.Model,
	opts *FileListOptions,
) (*FileListResponse, error) {
	// Validate inputs
	if sandboxInfo == nil {
		return nil, errors.New("sandboxInfo cannot be nil")
	}
	if opts == nil {
		return nil, errors.New("opts cannot be nil")
	}

	// Validate source
	if opts.Source != "volume" && opts.Source != "s3" {
		return nil, errors.Errorf("invalid source: must be 'volume' or 's3', got '%s'", opts.Source)
	}

	// Validate path
	if err := validatePath(opts.Path); err != nil {
		return nil, errors.Wrap(err, "path validation failed")
	}

	// Check if sandbox is still running, auto-restart if needed
	// Automatically updates database if sandboxModel and user provided
	shouldContinue, err := s.ensureSandboxRunning(ctx, sandboxInfo, sandboxModel, user, "file listing")
	if err != nil {
		return nil, err
	}
	if !shouldContinue {
		// Sandbox is nil (test environment) - return empty results
		return &FileListResponse{
			Files:      []FileInfo{},
			TotalCount: 0,
			Page:       opts.Page,
			PerPage:    opts.PerPage,
			TotalPages: 0,
		}, nil
	}

	// Build the find command with metadata formatting
	// This outputs: path|size|timestamp|type
	cmdStr := s.buildListFilesCommand(opts)

	// Debug logging
	log.Infof("ListFiles command: %s (source=%s, path=%s, recursive=%v)", cmdStr, opts.Source, opts.Path, opts.Recursive)

	// DEBUG: First verify files exist with a simple ls
	testProcess, _ := sandboxInfo.Sandbox.Exec(ctx, []string{"sh", "-c", "ls -la /mnt/workspace"}, nil)
	if testProcess != nil {
		var testOut bytes.Buffer
		testScanner := bufio.NewScanner(testProcess.Stdout)
		for testScanner.Scan() {
			testOut.Write(testScanner.Bytes())
			testOut.WriteByte('\n')
		}
		testProcess.Wait(ctx)
		log.Infof("DEBUG ls -la /mnt/workspace:\n%s", testOut.String())
	}

	// DEBUG: Test find with -L flag
	testProcess2, _ := sandboxInfo.Sandbox.Exec(ctx, []string{"sh", "-c", "find -L /mnt/workspace \\( -type f -o -type d \\)"}, nil)
	if testProcess2 != nil {
		var testOut2 bytes.Buffer
		testScanner2 := bufio.NewScanner(testProcess2.Stdout)
		for testScanner2.Scan() {
			testOut2.Write(testScanner2.Bytes())
			testOut2.WriteByte('\n')
		}
		testProcess2.Wait(ctx)
		log.Infof("DEBUG find -L /mnt/workspace:\n%s", testOut2.String())
	}

	// Determine target path for filtering (convert user path to mount path)
	var targetPath string
	if opts.Source == "s3" {
		targetPath = S3_MOUNT_PATH
	} else {
		targetPath = VOLUME_MOUNT_PATH
	}
	if opts.Path != "" && opts.Path != "/" {
		if strings.HasPrefix(opts.Path, "/workspace") {
			targetPath = strings.Replace(opts.Path, "/workspace", VOLUME_MOUNT_PATH, 1)
		} else if strings.HasPrefix(opts.Path, "/s3-bucket") {
			targetPath = strings.Replace(opts.Path, "/s3-bucket", S3_MOUNT_PATH, 1)
		}
	}

	process, err := sandboxInfo.Sandbox.Exec(ctx, []string{"sh", "-c", cmdStr}, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute list files command")
	}

	// Read stdout
	var output bytes.Buffer
	scanner := bufio.NewScanner(process.Stdout)
	for scanner.Scan() {
		output.Write(scanner.Bytes())
		output.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to read command output")
	}

	// Read stderr for error details
	var stderr bytes.Buffer
	stderrScanner := bufio.NewScanner(process.Stderr)
	for stderrScanner.Scan() {
		stderr.Write(stderrScanner.Bytes())
		stderr.WriteByte('\n')
	}

	// Wait for command to complete
	exitCode, err := process.Wait(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wait for command completion")
	}

	if exitCode != 0 {
		return nil, errors.Errorf("find command failed with exit code %d. Command: %s. Stderr: %s", exitCode, cmdStr, stderr.String())
	}

	// Debug logging
	outputStr := output.String()
	if len(outputStr) > 500 {
		log.Infof("ListFiles raw output (first 500 chars): %s", outputStr[:500])
	} else {
		log.Infof("ListFiles raw output: %s", outputStr)
	}
	log.Infof("ListFiles targetPath for filtering: %s", targetPath)

	// Parse output into FileInfo structs
	// Filter out the target directory itself from results
	files, err := parseFileMetadata(output.String(), targetPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse file metadata")
	}

	log.Infof("ListFiles parsed %d files after filtering", len(files))

	// Apply pagination
	response := paginateFiles(files, opts)

	return response, nil
}

// detectMimeType returns the MIME type based on file extension.
// Used to set proper Content-Type headers when serving file content.
// Returns "application/octet-stream" for unknown file types.
func detectMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	mimeTypes := map[string]string{
		".txt":  "text/plain",
		".json": "application/json",
		".xml":  "application/xml",
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".py":   "text/x-python",
		".go":   "text/x-go",
		".java": "text/x-java",
		".c":    "text/x-c",
		".cpp":  "text/x-c++",
		".md":   "text/markdown",
		".yaml": "application/x-yaml",
		".yml":  "application/x-yaml",
		".sh":   "application/x-sh",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".pdf":  "application/pdf",
		".zip":  "application/zip",
		".tar":  "application/x-tar",
		".gz":   "application/gzip",
	}

	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}

// buildReadFileCommand constructs a command to read file contents.
// Uses cat for normal files, head -c for size-limited reads.
// For volume source: uses /workspace as base path
// For s3 source: uses /s3-bucket as base path
// If maxSize > 0: uses head -c to limit bytes read
// If maxSize = 0: uses cat to read entire file
func (s *SandboxService) buildReadFileCommand(source, filePath string, maxSize int64) string {
	// Determine base path
	var basePath string
	if source == "s3" {
		basePath = S3_MOUNT_PATH
	} else {
		basePath = VOLUME_MOUNT_PATH
	}

	// Build full path
	fullPath := basePath + filePath

	// Use head -c for size-limited reads, cat otherwise
	if maxSize > 0 {
		return fmt.Sprintf("head -c %d %s", maxSize, fullPath)
	}
	return fmt.Sprintf("cat %s", fullPath)
}

// GetFileContent retrieves the content of a specific file from the sandbox.
// It checks if the file exists, determines its size, and reads the content.
// For files larger than 100MB, it limits the read to the first 100MB.
// Returns FileContent with the file's content, MIME type, filename, and size.
//
// If sandboxModel and user are provided, automatically handles database updates if sandbox auto-restarts.
//
// Returns an error if:
// - sandboxInfo is nil
// - source is not "volume" or "s3"
// - path validation fails
// - file does not exist
// - command execution fails
func (s *SandboxService) GetFileContent(
	ctx context.Context,
	sandboxInfo *modal.SandboxInfo,
	sandboxModel *sandbox.Sandbox,
	user coremodel.Model,
	source string,
	filePath string,
) (*FileContent, error) {
	// Validate inputs
	if sandboxInfo == nil {
		return nil, errors.New("sandboxInfo cannot be nil")
	}

	if source != "volume" && source != "s3" {
		return nil, errors.Errorf("invalid source: must be 'volume' or 's3', got '%s'", source)
	}

	// Prevent directory traversal in the file path
	if strings.Contains(filePath, "..") {
		return nil, errors.Wrap(errors.New("invalid path: directory traversal not allowed"), "path validation failed")
	}

	filePath = strings.TrimPrefix(filePath, "/workspace")

	// Determine base path
	var basePath string
	if source == "s3" {
		basePath = S3_MOUNT_PATH
	} else {
		basePath = VOLUME_MOUNT_PATH
	}
	fullPath := basePath + filePath

	// Validate the full path is within allowed directories
	if err := validatePath(fullPath); err != nil {
		return nil, errors.Wrap(err, "path validation failed")
	}

	// Check if sandbox is still running, auto-restart if needed
	// Automatically updates database if sandboxModel and user provided
	shouldContinue, err := s.ensureSandboxRunning(ctx, sandboxInfo, sandboxModel, user, "file content retrieval")
	if err != nil {
		return nil, err
	}
	if !shouldContinue {
		// Sandbox is nil (test environment) - return error
		return nil, errors.New("file not found: sandbox not connected")
	}

	// Check file exists and get size using stat
	// stat -c '%s' outputs file size in bytes
	// 2>/dev/null suppresses error output if file doesn't exist
	statCmd := fmt.Sprintf("stat -c '%%s' %s 2>/dev/null || echo 'FILE_NOT_FOUND'", fullPath)

	statProcess, err := sandboxInfo.Sandbox.Exec(ctx, []string{"sh", "-c", statCmd}, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to stat file")
	}

	// Read stdout from stat command
	var statOutput bytes.Buffer
	scanner := bufio.NewScanner(statProcess.Stdout)
	for scanner.Scan() {
		statOutput.Write(scanner.Bytes())
	}
	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to read stat output")
	}

	// Wait for stat command to complete
	exitCode, err := statProcess.Wait(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wait for stat completion")
	}

	if exitCode != 0 {
		return nil, errors.Errorf("stat command failed with exit code %d", exitCode)
	}

	// Parse stat output
	statStr := strings.TrimSpace(statOutput.String())
	if statStr == "FILE_NOT_FOUND" || statStr == "" {
		return nil, errors.Errorf("file not found: %s", filePath)
	}

	size, err := strconv.ParseInt(statStr, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse file size")
	}

	// Determine max read size (100MB limit)
	const MaxFileContentSize = 100 * 1024 * 1024 // 100MB
	var maxSize int64
	if size > MaxFileContentSize {
		maxSize = MaxFileContentSize
	}

	// Build and execute read command
	readCmd := s.buildReadFileCommand(source, filePath, maxSize)
	readProcess, err := sandboxInfo.Sandbox.Exec(ctx, []string{"sh", "-c", readCmd}, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read file content")
	}

	// Read file content from stdout
	var contentBuffer bytes.Buffer
	contentScanner := bufio.NewScanner(readProcess.Stdout)
	// Set large buffer for binary content
	buf := make([]byte, 0, 64*1024)
	contentScanner.Buffer(buf, int(MaxFileContentSize))
	for contentScanner.Scan() {
		contentBuffer.Write(contentScanner.Bytes())
		contentBuffer.WriteByte('\n')
	}
	if err := contentScanner.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to read file content")
	}

	// Wait for read command to complete
	exitCode, err = readProcess.Wait(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wait for read completion")
	}

	if exitCode != 0 {
		return nil, errors.Errorf("read command failed with exit code %d", exitCode)
	}

	// Create FileContent
	content := contentBuffer.Bytes()
	fileName := filepath.Base(filePath)
	contentType := detectMimeType(fileName)

	return &FileContent{
		Content:     content,
		ContentType: contentType,
		FileName:    fileName,
		Size:        int64(len(content)),
	}, nil
}

// BuildFileTree converts a flat list of FileInfo into a hierarchical tree structure.
// It creates a FileTreeNode root and organizes all files/directories as children
// based on their path hierarchy. The algorithm sorts files by path to ensure
// parents are processed before children, then uses a map for O(n) lookups.
//
// Parameters:
// - files: flat list of FileInfo structs from ListFiles
// - rootPath: the root directory path (e.g., "/workspace" or "/s3-bucket")
//
// Returns:
// - *FileTreeNode: root node with nested children structure
// - error: if tree building fails
//
// For empty file lists, returns a valid root node with no children.
// The tree structure uses pointer slices for Children to maintain references
// and allow proper nesting through multiple levels.
func (s *SandboxService) BuildFileTree(files []FileInfo, rootPath string) (*FileTreeNode, error) {
	// Extract root name from rootPath (e.g., "/workspace" -> "workspace")
	rootPath = cleanPath(rootPath)

	rootName := rootPath
	if idx := strings.LastIndex(rootPath, "/"); idx != -1 {
		rootName = rootPath[idx+1:]
	}

	// Create root node
	root := &FileTreeNode{
		Name:        rootName,
		Path:        rootPath,
		IsDirectory: true,
		Children:    []*FileTreeNode{},
	}

	// Handle empty file list
	if len(files) == 0 {
		return root, nil
	}

	// Create a map for O(1) lookups: path -> node
	nodeMap := make(map[string]*FileTreeNode)
	nodeMap[rootPath] = root

	// Sort files by path to ensure parents come before children
	// This is critical for building the tree correctly
	sortedFiles := make([]FileInfo, len(files))
	copy(sortedFiles, files)
	// Simple bubble sort by path length first, then alphabetically
	for i := 0; i < len(sortedFiles); i++ {
		for j := i + 1; j < len(sortedFiles); j++ {
			// Sort by path depth (number of slashes) first, then alphabetically
			iDepth := strings.Count(sortedFiles[i].Path, "/")
			jDepth := strings.Count(sortedFiles[j].Path, "/")
			if iDepth > jDepth || (iDepth == jDepth && sortedFiles[i].Path > sortedFiles[j].Path) {
				sortedFiles[i], sortedFiles[j] = sortedFiles[j], sortedFiles[i]
			}
		}
	}

	// Build tree by iterating through sorted files
	for _, file := range sortedFiles {
		// Skip if this is the root path itself (already created)
		if file.Path == rootPath {
			// Update root metadata if it's in the file list
			root.Size = file.Size
			root.ModifiedAt = file.ModifiedAt
			continue
		}

		// Create node for this file
		node := &FileTreeNode{
			Name:        file.Name,
			Path:        file.Path,
			IsDirectory: file.IsDirectory,
			Size:        file.Size,
			ModifiedAt:  file.ModifiedAt,
			Children:    []*FileTreeNode{},
		}

		// Add to node map
		nodeMap[file.Path] = node

		// Find parent directory path
		parentPath := file.Path
		if idx := strings.LastIndex(parentPath, "/"); idx != -1 {
			parentPath = parentPath[:idx]
		}

		// Find parent node in map
		parentNode, exists := nodeMap[parentPath]
		if !exists {
			// Parent doesn't exist yet - this shouldn't happen with sorted files
			// Create parent as a directory node
			parentName := parentPath
			if idx := strings.LastIndex(parentPath, "/"); idx != -1 {
				parentName = parentPath[idx+1:]
			}
			parentNode = &FileTreeNode{
				Name:        parentName,
				Path:        parentPath,
				IsDirectory: true,
				Children:    []*FileTreeNode{},
			}
			nodeMap[parentPath] = parentNode

			// Also need to attach this parent to ITS parent
			grandParentPath := parentPath
			if idx := strings.LastIndex(grandParentPath, "/"); idx != -1 {
				grandParentPath = grandParentPath[:idx]
			}
			if grandParentNode, gpExists := nodeMap[grandParentPath]; gpExists {
				grandParentNode.Children = append(grandParentNode.Children, parentNode)
			}
		}

		// Append current node to parent's children
		parentNode.Children = append(parentNode.Children, node)
	}

	return root, nil
}

// ConvertTreePathsToUserFacing recursively converts mount paths to user-facing paths in the file tree.
// This transforms internal paths like /mnt/workspace/test.txt to /workspace/test.txt for API responses.
func ConvertTreePathsToUserFacing(node *FileTreeNode) {
	if node == nil {
		return
	}

	// Convert mount paths to user-facing paths
	node.Path = strings.Replace(node.Path, VOLUME_MOUNT_PATH, USER_WORKSPACE_PATH, 1)
	node.Path = strings.Replace(node.Path, S3_MOUNT_PATH, S3_USER_PATH, 1)

	// Recursively convert all children
	for _, child := range node.Children {
		ConvertTreePathsToUserFacing(child)
	}
}

func (s *SandboxService) ForceVolumeSync(ctx context.Context,
	sandboxInfo *modal.SandboxInfo,
	_ *sandbox.Sandbox,
) error {
	syncProcess, err := sandboxInfo.Sandbox.Exec(ctx, []string{"sync", VOLUME_MOUNT_PATH}, nil)
	if err != nil {
		return nil
	}

	exitCode, err := syncProcess.Wait(ctx)
	if err != nil {
		return errors.WithMessagef(err, "failed waiting for sync process in sandbox %s", sandboxInfo.SandboxID)
	}
	if exitCode != 0 {
		return errors.Errorf("sync process exited with code %d", exitCode)
	}

	return nil
}
