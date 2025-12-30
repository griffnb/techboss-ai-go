# Sandbox File Explorer - Design Document

## Overview

The Sandbox File Explorer feature provides APIs to browse and retrieve files from Modal sandbox volumes and S3-synced storage. The design follows existing patterns in the codebase, leveraging the Modal sandbox command execution infrastructure and building upon proven file operation patterns from the state file system.

This feature will be implemented as a new controller with supporting service layer logic, following the project's standard controller/service architecture pattern.

## Architecture

### High-Level Architecture

```
┌─────────────┐
│   Client    │
│ (Frontend)  │
└──────┬──────┘
       │ HTTP Requests
       │
┌──────▼────────────────────────────────────────────────────┐
│                    API Layer                               │
│  ┌─────────────────────────────────────────────────────┐  │
│  │   SandboxFileController                             │  │
│  │   - ListFiles(sandboxID, source, path, pagination)  │  │
│  │   - GetFileContent(sandboxID, source, filePath)     │  │
│  │   - GetFileTree(sandboxID, source, path)            │  │
│  └─────────────────┬───────────────────────────────────┘  │
└────────────────────┼──────────────────────────────────────┘
                     │
┌────────────────────▼──────────────────────────────────────┐
│                Service Layer                               │
│  ┌─────────────────────────────────────────────────────┐  │
│  │   SandboxFileService                                │  │
│  │   - ListSandboxFiles(sandbox, opts)                 │  │
│  │   - ListS3Files(sandbox, opts)                      │  │
│  │   - GetFileContent(sandbox, source, path)           │  │
│  │   - BuildFileTree(files)                            │  │
│  └─────────────────┬───────────────────────────────────┘  │
└────────────────────┼──────────────────────────────────────┘
                     │
         ┌───────────┴───────────┐
         │                       │
┌────────▼────────┐    ┌─────────▼────────┐
│ Modal Sandbox   │    │   S3 Storage     │
│   Volume Ops    │    │   (via mount)    │
│                 │    │                  │
│ - find commands │    │ - find commands  │
│ - stat metadata │    │ - stat metadata  │
│ - cat/read file │    │ - cat/read file  │
└─────────────────┘    └──────────────────┘
```

### Design Rationale

1. **Command-Based File Operations**: Following the existing pattern in `state_service.go`, we'll use shell commands (`find`, `stat`, `md5sum`, `cat`) executed via `sandbox.ExecCommand()` rather than direct API calls. This approach:
   - Works consistently for both volume and S3 (mounted via volume)
   - Leverages proven patterns already in production
   - Simplifies implementation and maintenance

2. **Service Layer Abstraction**: Business logic will reside in `SandboxFileService` to:
   - Keep controllers thin (following existing pattern)
   - Enable reusability across multiple endpoints
   - Facilitate comprehensive unit testing

3. **Unified API with Source Parameter**: A single set of endpoints will handle both volume and S3 files using a `source` query parameter:
   - Simplifies frontend integration
   - Reduces endpoint proliferation
   - Follows RESTful principles

## Components and Interfaces

### 1. Sandbox Files Controller

**Location**: `internal/controllers/sandboxes/files.go`

**Responsibilities**:
- Handle HTTP requests for file operations
- Validate input parameters
- Enforce authentication and authorization via role handlers
- Format responses using standard patterns

**Endpoints**:

```go
// Admin Routes (full access)
GET    /admin/sandbox/:id/files              // List files
GET    /admin/sandbox/:id/files/content      // Get file content
GET    /admin/sandbox/:id/files/tree         // Get file tree

// Public Routes (owner-only access)
GET    /sandbox/:id/files                    // List files
GET    /sandbox/:id/files/content            // Get file content
GET    /sandbox/:id/files/tree               // Get file tree
```

**Query Parameters**:
- `source` (string): "volume" or "s3" (default: "volume")
- `path` (string): Root path to list from (default: "/")
- `page` (int): Page number for pagination (default: 1)
- `per_page` (int): Items per page (default: 100, max: 1000)
- `file_path` (string): Specific file path for content retrieval
- `recursive` (bool): Include subdirectories (default: true)

**Controller Functions** (following existing pattern):

```go
// adminListFiles handles GET /admin/sandbox/:id/files
func adminListFiles(_ http.ResponseWriter, req *http.Request) (*sandbox_service.FileListResponse, int, error) {
    // 1. Get user session
    userSession := request.GetReqSession(req)
    
    // 2. Extract and validate parameters
    id := chi.URLParam(req, "id")
    opts := parseFileListOptions(req)
    
    // 3. Load sandbox
    sandboxModel, err := sandbox.Get(req.Context(), types.UUID(id))
    if err != nil {
        return response.AdminBadRequestError[*sandbox_service.FileListResponse](err)
    }
    
    // 4. Call service layer
    service := sandbox_service.NewSandboxService()
    files, err := service.ListFiles(req.Context(), sandboxModel, opts)
    if err != nil {
        return response.AdminBadRequestError[*sandbox_service.FileListResponse](err)
    }
    
    return response.Success(files)
}

// authListFiles handles GET /sandbox/:id/files (owner-only)
func authListFiles(_ http.ResponseWriter, req *http.Request) (*sandbox_service.FileListResponse, int, error) {
    // Similar to adminListFiles but with auth framework ownership verification
}

// adminGetFileContent handles GET /admin/sandbox/:id/files/content
func adminGetFileContent(w http.ResponseWriter, req *http.Request) (*sandbox_service.FileContent, int, error) {
    // Implementation details below
}

// adminGetFileTree handles GET /admin/sandbox/:id/files/tree
func adminGetFileTree(_ http.ResponseWriter, req *http.Request) (*sandbox_service.FileTreeNode, int, error) {
    // Implementation details below
}
```

### 2. Sandbox File Service

**Location**: `internal/services/sandbox_service/files.go`

**Responsibilities**:
- Execute file operations on sandbox volumes and S3
- Build file metadata structures
- Handle pagination logic
- Construct hierarchical file trees
- Manage error handling and retries

**Service Structure** (following existing pattern with methods on SandboxService):

```go
// Add to existing SandboxService in internal/services/sandbox_service/sandbox_service.go

type FileListOptions struct {
    Source    string // "volume" or "s3"
    Path      string
    Recursive bool
    Page      int
    PerPage   int
}

type FileInfo struct {
    Name         string    `json:"name"`
    Path         string    `json:"path"`
    Size         int64     `json:"size"`
    ModifiedAt   time.Time `json:"modified_at"`
    IsDirectory  bool      `json:"is_directory"`
    Checksum     string    `json:"checksum,omitempty"`
}

type FileListResponse struct {
    Files      []FileInfo `json:"files"`
    TotalCount int        `json:"total_count"`
    Page       int        `json:"page"`
    PerPage    int        `json:"per_page"`
    TotalPages int        `json:"total_pages"`
}

type FileTreeNode struct {
    Name        string         `json:"name"`
    Path        string         `json:"path"`
    IsDirectory bool           `json:"is_directory"`
    Size        int64          `json:"size,omitempty"`
    ModifiedAt  time.Time      `json:"modified_at,omitempty"`
    Children    []FileTreeNode `json:"children,omitempty"`
}

type FileContent struct {
    Content     []byte `json:"content"`
    ContentType string `json:"content_type"`
    FileName    string `json:"file_name"`
}

// ListFiles retrieves files from sandbox volume or S3
func (s *SandboxService) ListFiles(
    ctx context.Context,
    sandboxModel *sandbox.Sandbox,
    opts FileListOptions,
) (*FileListResponse, error) {
    // Implementation details below
}

// GetFileContent retrieves the content of a specific file
func (s *SandboxService) GetFileContent(
    ctx context.Context,
    sandboxModel *sandbox.Sandbox,
    source string,
    filePath string,
) (*FileContent, error) {
    // Implementation details below
}

// BuildFileTree converts flat file list to hierarchical tree
func (s *SandboxService) BuildFileTree(
    files []FileInfo,
    rootPath string,
) (*FileTreeNode, error) {
    // Implementation details below
}
```

### 3. Command Execution Patterns

Following the pattern from `state_service.go`:

**List Files Command**:
```bash
# For volume files
find /workspace -type f -o -type d | head -n {limit}

# For S3 files (mounted at /s3-bucket)
find /s3-bucket -type f -o -type d | head -n {limit}
```

**Get File Metadata Command**:
```bash
# Get detailed info for each file
stat -c "%n|%s|%Y" {file_path}
# Output format: path|size|timestamp

# Get checksum (optional, for files only)
md5sum {file_path} 2>/dev/null || echo ""
```

**Get File Content Command**:
```bash
# Read file content
cat {file_path}

# For large files, use head/tail or streaming
head -c 10485760 {file_path}  # First 10MB
```

### 4. Response Formats

**List Files Response**:
```json
{
  "success": true,
  "data": {
    "files": [
      {
        "name": "main.go",
        "path": "/workspace/cmd/server/main.go",
        "size": 2048,
        "modified_at": "2025-12-30T10:30:00Z",
        "is_directory": false,
        "checksum": "5d41402abc4b2a76b9719d911017c592"
      },
      {
        "name": "controllers",
        "path": "/workspace/internal/controllers",
        "size": 4096,
        "modified_at": "2025-12-30T09:15:00Z",
        "is_directory": true
      }
    ],
    "total_count": 245,
    "page": 1,
    "per_page": 100,
    "total_pages": 3
  }
}
```

**File Tree Response**:
```json
{
  "success": true,
  "data": {
    "name": "workspace",
    "path": "/workspace",
    "is_directory": true,
    "children": [
      {
        "name": "cmd",
        "path": "/workspace/cmd",
        "is_directory": true,
        "children": [
          {
            "name": "main.go",
            "path": "/workspace/cmd/main.go",
            "is_directory": false,
            "size": 2048,
            "modified_at": "2025-12-30T10:30:00Z"
          }
        ]
      }
    ]
  }
}
```

**File Content Response**:
```
Content-Type: application/octet-stream (or appropriate MIME type)
Content-Disposition: attachment; filename="main.go"
Content-Length: 2048

[file content bytes]
```

**Error Response**:
```json
{
  "success": false,
  "error": "File not found: /workspace/nonexistent.txt",
  "code": "FILE_NOT_FOUND"
}
```

## Data Models

### Existing Models (No Changes Required)

The feature will use existing `Sandbox` model from `internal/models/sandbox.go`:

```go
type Sandbox struct {
    ID         int64
    AccountID  int64
    UserID     *int64
    ExternalID string  // Modal sandbox ID
    Status     string
    // ... other fields
}
```

No new database models are required. File information is retrieved dynamically from the sandbox/S3 at request time.

### In-Memory Data Structures

The service layer will use Go structs (defined above in SandboxFileService) to represent file information:
- `FileInfo`: Individual file metadata
- `FileListResponse`: Paginated file list
- `FileTreeNode`: Hierarchical tree structure

## Error Handling

### Error Types

Following existing patterns in the codebase:

1. **Authorization Errors** (403):
   - User doesn't own the sandbox
   - Sandbox belongs to different account

2. **Not Found Errors** (404):
   - Sandbox ID doesn't exist
   - Requested file doesn't exist

3. **Validation Errors** (400):
   - Invalid source parameter
   - Invalid path format
   - Page/per_page out of range

4. **External Service Errors** (500):
   - Modal API unavailable
   - Command execution timeout
   - S3 mount noService) ListFiles(
    ctx context.Context,
    sandboxModel *sandbox.Sandbox,
    opts FileListOptions,
) (*FileListResponse, error) {
    // Validate options
    if opts.Source != "volume" && opts.Source != "s3" {
        return nil, errors.Errorf("invalid source: must be 'volume' or 's3'")
    }
    
    // Reconstruct SandboxInfo for command execution
    sandboxInfo := ReconstructSandboxInfo(sandboxModel, sandboxModel.AccountID.Get())
    
    // Build command
    command := s.buildListFilesCommand(opts)
    
    // Execute command with retry logic
    output, err := s.executeWithRetry(ctx, sandboxInfo, command, 3)
    if err != nil {
        return nil, errors.Wrapf(err, "failed to list files from %s", opts.Source)
    }
    
    // Parse and return
    files, err := parseFileList(output)
    if err != nil {
        return nil, errors.Wrap(err, "failed to parse file list")
    }
    
    return paginateFiles(files, opts), nil
}

// Retry logic for transient failures
func (s *SandboxService) executeWithRetry(
    ctx context.Context,
    sandboxInfo *SandboxInfo,
    command string,
    maxRetries int,
) (string, error) {
    var lastErr error
    for i := 0; i < maxRetries; i++ {
        output, err := s.ExecCommand(ctx, sandboxInfo, 
    command string,
    maxRetries int,
) (string, error) {
    var lastErr error
    for i := 0; i < maxRetries; i++ {
        output,  (following existing pattern)
func adminListFiles(_ http.ResponseWriter, req *http.Request) (*sandbox_service.FileListResponse, int, error) {
    id := chi.URLParam(req, "id")
    opts := parseFileListOptions(req)
    
    // Load sandbox
    sandboxModel, err := sandbox.Get(req.Context(), types.UUID(id))
    if err != nil {
        log.ErrorContext(err, req.Context())
        return response.AdminBadRequestError[*sandbox_service.FileListResponse](err)
    }
    
    // Call service
    service := sandbox_service.NewSandboxService()
    files, err := service.ListFiles(req.Context(), sandboxModel, opts)
    if err != nil {
        log.ErrorContext(err, req.Context())
        
        // Handle specific errors with better messages
        if strings.Contains(err.Error(), "not found") {
            return nil, http.StatusNotFound, errors.New("sandbox or path not found")
        }
        
        return response.AdminBadRequestError[*sandbox_service.FileListResponse](err)
    }
    
    return response.Success(files       })
            return
        }
        
        ctx.JSON(500, gin.H{
            "success": false,
            "error": "Failed to retrieve files",
        })
        return
    }
    
    ctx.JSON(200, gin.H{
        "success": true,
        "data": files,
    })
}
```

## Testing Strategy

Following the project's TDD (Test-service/files_test.go`):

```go
func TestSandbox
**Service Layer Tests** (`sandbox_file_service_test.go`):

```go
func TestSandboxFileService_ListFiles(t *testing.T) {
    t.Run("successfully lists volume files", func(t *testing.T) {
        // Arrange: Create test sandbox with mocked ExecCommand
        // Act: Call ListFiles with volume source
        // Assert: Verify correct files returned
    })
    
    t.Run("successfully lists S3 files", func(t *testing.T) {
        // Arrange: Create test sandbox with mocked ExecCommand for S3
        // Act: Call ListFiles with s3 source
        // Assert: Verify correct files returned
    })
    
    t.Run("handles pagination correctly", func(t *testing.T) {
        // Arrange: Mock large file list
        // Act: Request different pages
        // Assert: Verify pagination math and limits
    })
    
    t.Run("returns error for invalid source", func(t *testing.T) {
        // Arrange: Create options with invalid source
        // Act: Call ListFiles
        // Assert: Verify error returned
    })
    
    t.Run("retries on transient failures", func(t *testing.T) {
        // Arrange: Mock ExecCommand to fail twice then succeed
        // Act: Call ListFiles
        // Assererify retry logic and eventual success
    })
}

func TestSandboxFileService_GetFileContent(t *testing.T) {
    t.Run("retrieves text file content", func(t *testing.T) {
        // Test implementation
    })
    
    t.Run("handles binary file content", func(t *testing.T) {
        // Test implementation
    })
    
    t.Run("returns error for non-existent file", func(t *testing.T) {
        // Test implementation
    })
}

func TestSandboxFileService_BuildFileTree(t *testing.T) {
    t.Run("builds correct tree structure", func(t *testing.T) {
        // Arrange: Create flat file list
        // Act: Build tree
        // Assererify hierarchical structure
    })
    
    t.Run("handles nested directories", func(t *testing.T) {
        // Test implementation
    })
}
```

**Controller Tests** (`sandbox_file_controller_test.go`):

```go
func TestSandboxFileController_Index(t *testing.T) {
    t.Run("returns file list fes/files_test.go`):

```go
func TestAdminListFilesse with correct data
    })
    
    t.Run("returns 403 for unauthorized user", func(t *testing.T) {
        // Arrange: Create user without access
        // Act: Make GET request
        // Assert: Verify 403 response
    })
    
    t.Run("returns 404 for non-existent sandbox", func(t *testing.T) {
        // Test implementation
    })
    
    t.Run("validates query parameters", func(t *testing.T) {
        // Test various invalid inputs
    })
}

func TestAdminGetFileContent(t *testing.T) {
    t.Run("returns file content with correct headers", func(t *testing.T) {
        // Test implementation
    })
    
    t.Run("returns 404 for non-existent file", func(t *testing.T) {
        // Test implementation
    })
}
```

### Integration Tests

**End-to-End Tests** (`sandbox_file_integration_test.go`):

```go
func TestSandboxFileExplorer_Integration(t *testing.T) {
    // Setup: Create real sandbox via testing_service.Builder
    // Create test files in sandbox
    
    t.Run("complete file listing workflow", func(t *testing.T) {
        // 1. List files
        // 2. Verify all files present
        // 3. Get file content
        // 4. Verify content matches
    })
    
    t.Run("pagination across multiple pages", func(t *testing.T) {
        // Create 250 test files
        // Request pages 1, 2, 3
        // Verify correct files on each page
    })
    
    t.Run("tree structure matches flat list", func(t *testing.T) {
        // Get flat list
        // Get tree structure
        // Verify consistency
    })
}
```

### Test Coverage Goals

- **Service Layer**: ≥95% coverage
- **Controller Layer**: ≥90% coverage
- **Integration Tests**: All major workflows covered
- **Error Paths**: All error conditions tested

### Testing Tools

- Use `assert` package from `lib/testtools/assert`
- Use `testing_service.Builder` for test data
- Clean up with `defer testtools.CleanupModel()`
- Table-driven tests for multiple scenarios

## Implementation Phases

### Phase 1: Core File Listing (MVP)

**Goal**: Basic file listing for both volume and S3 sources

**Deliverables**:
1. `SandboxService.ListFiles()` method in `sandbox_service/files.go`
2. Command execution for `find` and `stat`
3. Controller functions in `controllers/sandboxes/files.go` (`adminListFiles`, `authListFiles`)
4. Route setup in `controllers/sandboxes/setup.go`
5. Basic pagination support
6. Unit tests for service and controller
7. Integration tests for basic workflow

**Success Criteria**:
- Can list files from sandbox volume
- Can list files from S3 (via mount)
- Pagination works correctly
- Authorization enforced
- All tests pass with ≥90% coverage

### Phase 2: File Content Retrieval

**Goal**: Retrieve content of individual files

**DeSandboxService.GetFileContent()` method
2. Controller functions `adminGetFileContent`, `authGetFileContent`
2. `GetContent()` controller endpoint
3. Proper MIME type detection
4. Streaming for large files
5. Unit and integration tests

**Success Criteria**:
- Can retrieve text and binary files
- Large files handled efficiently
- Proper HTTP headers set
- Error handling for missing files

### Phase 3: Tree Structure View

**Goal**: Hierarchical tree representation of files

**DeSandboxService.BuildFileTree()` method
2. Controller functions `adminGetFileTree`, `authGetFileTree`d
2. `GetTree()` controller endpoint
3. Efficient tree building algorithm
4. Unit and integration tests

**Success Criteria**:
- Tree accurately reflects directory structure
- Performance acceptable for 10,000+ files
- Memory efficient

### Phase 4: Performance Optimizations

**Goal**: Improve performance for large file sets

**Deliverables**:
1. Response caching with TTL
2. Parallel command execution for metadata
3. Query optimization
4. Performance benchmarks

**Success Criteria**:
- Response time <5s for 10,000 files
- Efficient memory usage
- Cache hit rate >70%

## Security Considerations

### Authentication and Authorization

1. **User Authentication**: All endpoints require valid JWT token
2. **Ownership Verification**: Users can only access sandboxes they own
3. **Admin Override**: Admin users can access all sandboxes
4. **Account Isolation**: Sandboxes are account-scoped

### Implementation Pattern

```go
// In controller
func (c *SandboxFileController) Index(ctx *gin.Context) {
    // Get authenticated user from context
    user := ctx.MustGet("user").(*models.User)
    
    // Load sandbox
    sandbox := models.Sandbox.FindByID(sandboxID)
    if sandbox == nil {
        ctx.JSON(404, gin.H{"success": false, "error": "Sandbox not found"})
        return
    }
    
    // Authorization check
    if !user.IsAdmin() && sandbox.UserID != user.ID {
        ctx.JSON(403, gin.H{"success": false, "error": "Unauthorized"})
     
    }
    
    // Proceed with file operations
}
```

### Path Traversal Prevention

```go
func (s *SandboxFileService) validatePath(path string) error {
    // Prevent directory traversal
    if strings.Contains(path, "..") {
        return errors.New("invalid path: directory traversal not allowed")
    }
    
    // Ensure path is within allowed directories
    if !strings.HasPrefix(path, "/workspace") && !strings.HasPrefix(path, "/s3-bucket") {
        return errors.New("invalid path: must be within /workspace or /s3-bucket")
    }
    
    return nil
}
```

### Rate Limiting

Consider implementing rate limiting for file operations to prevent abuse:

```go
// Future enhancement: Add rate limiting middleware
// Limit: 100 requests per minute per user
// Burst: 20 requests
```

## Performance Considerations

### Expected Load

- **Files per Sandbox**: 100-10,000 files (typical), up to 50,000 (max)
- **File Sizes**: Mostly <1MB (code files), some >10MB (binaries)
- **Request Frequency**: 10-100 requests/minute (typical)

### Optimization Strategies

1. **Pagination**: Limit results to prevent large payloads
2. **Lazy Loading**: Tree structure loaded on-demand
3. **Caching**: Cache file listings with 60-second TTL
4. **Parallel Execution**: Execute metadata commands in parallel
5. **Streaming**: Use streaming for large file content

### Resource Limits

```go
const (
    MaxFilesPerRequest = 1000
    MaxFileContentSize = 100 * 1024 * 1024 // 100MB
    CommandTimeout     = 30 * time.Second
    DefaultPageSize    = 100
)
```

## Dependencies

### External Services

1. **Modal Sandbox API**: Command execution via `sandbox.ExecCommand()`
2. **S3 (Indirect)**: Accessed through mounted volumes in sandbox

### Internal Dependencies

1. **Models**: `models.Sandbox`, `models.User`, `models.Account`
2. **Services**: None (new standalone service)
3. **Controllers**: Standard Gin HTTP routing

### Go Packages

- `github.com/gin-gonic/gin`: HTTP routing
- `github.com/pkg/errors`: Error wrapping
- Standard library: `os`, `path/filepath`, `strings`, `time`

## Future Enhancements

### Potential Features (Post-MVP)

1. **Search and Filtering**: Search files by name, extension, or content
2. **File Upload**: Upload files to sandbox volumes
3. **File Operations**: Delete, rename, move files
4. **Diff View**: Compare file versions across time
5. **WebSocket Updates**: Real-time file change notifications
6. **Zip Download**: Download entire directory as zip
7. **Syntax Highlighting**: Return formatted code with highlighting
8. **File Preview**: Generate thumbnails/previews for images
9. **Access Logs**: Track which files users view
10. **Shared Links**: Generate temporary public links for files

## Conclusion

This design provides a robust, scalable foundation for the Sandbox File Explorer feature. By leveraging existing patterns and infrastructure (command execution, controller structure, authentication), the implementation will integrate seamlessly with the current codebase while maintaining high code quality standards through comprehensive testing.

The phased approach ensures incremental delivery of value, with the MVP (Phase 1) providing core functionality and subsequent phases adding enhanced capabilities based on user feedback and usage patterns.
