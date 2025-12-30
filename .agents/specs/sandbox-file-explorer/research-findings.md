# Sandbox File Explorer - Research Findings

## Executive Summary

This document contains comprehensive research findings for implementing a Sandbox File Explorer feature in the Go backend. The research covers existing patterns for Modal sandbox integration, S3 operations, controller architecture, model patterns, and API response structures.

---

## 1. Modal Sandbox Integration

### 1.1 Sandbox Model

**Location**: `/Users/griffnb/projects/techboss/techboss-ai-go/internal/models/sandbox/sandbox.go`

```go
type DBColumns struct {
    base.Structure
    OrganizationID *fields.UUIDField                  `column:"organization_id" type:"uuid"     default:"null" index:"true" null:"true" public:"view"`
    AccountID      *fields.UUIDField                  `column:"account_id"      type:"uuid"     default:"null" index:"true" null:"true" public:"view"`
    MetaData       *fields.StructField[*MetaData]     `column:"meta_data"       type:"jsonb"    default:"{}"`
    Provider       *fields.IntConstantField[Provider] `column:"type"            type:"smallint" default:"0"`
    AgentID        *fields.UUIDField                  `column:"agent_id"        type:"uuid"     default:"null" index:"true" null:"true" public:"view"`
    ExternalID     *fields.StringField                `column:"external_id"     type:"text"     default:"null" index:"true" null:"true" public:"view"`
}
```

**Key Insights**:
- `ExternalID` stores the Modal sandbox ID (e.g., "sb-xxx") for API operations
- `MetaData` is JSONB field for storing sync statistics and custom data
- `AccountID` provides multi-tenant scoping
- Public fields are marked with `public:"view"` tag

### 1.2 Modal API Client

**Location**: `/Users/griffnb/projects/techboss/techboss-ai-go/internal/integrations/modal/sandbox.go`

**SandboxInfo Structure**:
```go
type SandboxInfo struct {
    SandboxID string         // Modal sandbox ID
    Sandbox   *modal.Sandbox // Underlying Modal sandbox
    Config    *SandboxConfig // Original configuration
    CreatedAt time.Time      // Creation timestamp
    Status    SandboxStatus  // Current status
}

type SandboxConfig struct {
    AccountID       types.UUID        // Account scoping
    Image           *ImageConfig      // Docker image config
    VolumeName      string            // Volume name
    VolumeMountPath string            // Where to mount volume (e.g., "/mnt/workspace")
    S3Config        *S3MountConfig    // S3 bucket mount
    Workdir         string            // Working directory
    Secrets         map[string]string // Additional secrets
    EnvironmentVars map[string]string // Custom env vars
}
```

### 1.3 Executing Commands in Sandboxes

**Pattern for executing commands** (`sandboxInfo.Sandbox.Exec`):

```go
// Execute command in sandbox
cmd := []string{
    "sh", "-c",
    fmt.Sprintf("ls -la %s", path),
}

process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, &modal.SandboxExecParams{
    Workdir: "/",
})
if err != nil {
    return errors.Wrapf(err, "failed to execute command")
}

// Read output
var output bytes.Buffer
scanner := bufio.NewScanner(process.Stdout)
for scanner.Scan() {
    output.Write(scanner.Bytes())
    output.WriteByte('\n')
}

// Wait for completion
exitCode, err := process.Wait(ctx)
if err != nil {
    return errors.Wrapf(err, "failed to wait for process")
}

if exitCode != 0 {
    return errors.Errorf("command failed with exit code %d", exitCode)
}
```

### 1.4 File Listing Pattern

**From**: `/Users/griffnb/projects/techboss/techboss-ai-go/internal/services/sandbox_service/state_files/writer.go`

```go
// List all files recursively (excluding specific files)
cmd := []string{
    "sh", "-c",
    fmt.Sprintf("cd %s && find . -type f ! -name '%s' -print", directoryPath, excludeFile),
}

// Output format: "./path/to/file.txt"
```

**For file metadata**:
```go
// Get size, modified time, and checksum
metaCmd := []string{
    "sh", "-c",
    fmt.Sprintf("stat -c '%%s|%%Y' %s && md5sum %s | cut -d' ' -f1", fullPath, fullPath),
}

// Output format:
// 1234|1234567890
// abc123def456...
```

### 1.5 Reading File Contents

**Pattern** (from `storage_state.go`):
```go
// Read file contents
cmd := []string{
    "sh", "-c",
    fmt.Sprintf("test -f %s && cat %s || echo 'FILE_NOT_FOUND'", filePath, filePath),
}

process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, nil)
// ... read output ...
```

---

## 2. S3 Integration

### 2.1 S3 Client Access

**Location**: `/Users/griffnb/projects/techboss/techboss-ai-go/internal/environment/helpers.go`

```go
func GetS3() *s3.S3 {
    return env.Env().(*SysEnvironment).S3
}
```

**S3 Client Type**: `github.com/griffnb/core/lib/s3`

### 2.2 S3 Configuration

**From SandboxConfig**:
```go
type S3MountConfig struct {
    BucketName string // S3 bucket name
    SecretName string // Modal secret for AWS credentials
    KeyPrefix  string // S3 key prefix (e.g., "docs/{account}/{timestamp}/")
    MountPath  string // Where to mount in container
    ReadOnly   bool   // Mount as read-only
    Timestamp  int64  // Unix timestamp for versioning
}
```

**Default S3 Setup** (from `sandbox_service.go`):
```go
// Always include S3 config for workspace persistence
envConfig := environment.GetConfig()
if envConfig != nil && envConfig.S3Config != nil && envConfig.S3Config.Buckets != nil {
    if bucketName, ok := envConfig.S3Config.Buckets["agent-docs"]; ok {
        config.S3Config = &modal.S3MountConfig{
            BucketName: bucketName,
            SecretName: "s3-bucket",
            KeyPrefix:  fmt.Sprintf("docs/%s/", accountID), // Must end with /
            MountPath:  "/mnt/s3-bucket",
            ReadOnly:   true,
        }
    }
}
```

### 2.3 S3 Mount Access Pattern

**S3 files are accessed via mount path in sandbox** (not direct S3 API):

```go
// S3 is mounted at /mnt/s3-bucket in the sandbox
// Files can be listed/read using standard shell commands:
cmd := []string{
    "sh", "-c", 
    fmt.Sprintf("ls -la %s", s3MountPath),
}
```

**Timestamp-based versioning**:
```go
// Each sync creates a new timestamped path:
// s3://bucket/docs/{account_id}/{unix_timestamp}/
timestamp := time.Now().Unix()
s3Path := fmt.Sprintf("s3://%s/docs/%s/%d/",
    bucketName,
    accountID,
    timestamp,
)
```

### 2.4 S3 File Operations

**From**: `/Users/griffnb/projects/techboss/techboss-ai-go/internal/integrations/modal/storage.go`

```go
// Copy from S3 to volume
cmd := []string{
    "runuser", "-u", ClaudeUserName, "--",
    "sh", "-c",
    fmt.Sprintf("cp -rv %s/* %s/ 2>&1 || true",
        s3MountPath,
        volumeMountPath,
    ),
}

// Sync to S3 using AWS CLI
syncCmd := fmt.Sprintf("aws s3 sync %s %s --exact-timestamps 2>&1", 
    volumeMountPath, s3Path)
cmd := []string{
    "runuser", "-u", ClaudeUserName, "--",
    "sh", "-c", syncCmd,
}
```

---

## 3. Controller Patterns

### 3.1 Controller Structure

**Setup File**: `/Users/griffnb/projects/techboss/techboss-ai-go/internal/controllers/sandboxes/setup.go`

```go
func Setup(coreRouter *router.CoreRouter) {
    // Admin routes - full CRUD access
    coreRouter.AddMainRoute(tools.BuildString("/admin/", ROUTE), func(r chi.Router) {
        r.Group(func(adminR chi.Router) {
            adminR.Get("/", helpers.RoleHandler(helpers.RoleHandlerMap{
                constants.ROLE_READ_ADMIN: response.StandardRequestWrapper(adminIndex),
            }))
            adminR.Get("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
                constants.ROLE_READ_ADMIN: response.StandardRequestWrapper(adminGet),
            }))
        })
    })

    // Public authenticated routes - restricted access
    coreRouter.AddMainRoute(tools.BuildString("/", ROUTE), func(r chi.Router) {
        r.Group(func(authR chi.Router) {
            authR.Get("/", helpers.RoleHandler(helpers.RoleHandlerMap{
                constants.ROLE_ANY_AUTHORIZED: response.StandardPublicRequestWrapper(authIndex),
            }))
            authR.Get("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
                constants.ROLE_ANY_AUTHORIZED: response.StandardPublicRequestWrapper(authGet),
            }))
        })
    })
}
```

### 3.2 Controller Function Signature

**Standard pattern**:
```go
func controllerFunction(_ http.ResponseWriter, req *http.Request) (*ReturnType, int, error) {
    // Get user session
    userSession := request.GetReqSession(req)
    accountID := userSession.User.ID()
    
    // Get URL parameters
    id := chi.URLParam(req, "id")
    
    // Business logic...
    
    // Return responses
    return response.Success(data)
    return response.AdminBadRequestError[*Type](err)
    return response.PublicBadRequestError[*Type]()
}
```

### 3.3 Request Wrappers

**Admin Wrapper** (`StandardRequestWrapper`):
- Returns internal errors directly
- Full field visibility
- For admin routes only

**Public Wrapper** (`StandardPublicRequestWrapper`):
- Masks internal errors
- Only shows fields marked `public:"view"` or `public:"edit"`
- For authenticated user routes

### 3.4 Query Parameters

**Auto query building** (`request.BuildIndexParams`):

```go
parameters := request.BuildIndexParams(req.Context(), req.URL.Query(), TABLE_NAME)
```

**Supported patterns**:
| Query | SQL Result | Description |
|-------|-----------|-------------|
| `?name=john` | `WHERE table.name = 'john'` | Exact match |
| `?name[]=john&name[]=jane` | `WHERE table.name IN('john','jane')` | Multiple values |
| `?not:name=john` | `WHERE table.name != 'john'` | Not equal |
| `?q:name=john` | `WHERE LOWER(table.name) ILIKE '%john%'` | Like search |
| `?gt:age=25` | `WHERE table.age > 25` | Greater than |
| `?lt:age=65` | `WHERE table.age < 65` | Less than |
| `?limit=10` | `LIMIT 10` | Result limit |
| `?offset=20` | `OFFSET 20` | Result offset |
| `?order=name,created_at desc` | `ORDER BY table.name asc, table.created_at desc` | Custom ordering |

### 3.5 Example Controller

**From**: `/Users/griffnb/projects/techboss/techboss-ai-go/internal/controllers/sandboxes/sandbox.go`

```go
func createSandbox(_ http.ResponseWriter, req *http.Request) (*sandbox.Sandbox, int, error) {
    userSession := request.GetReqSession(req)
    accountID := userSession.User.ID()

    // Parse request body
    data, err := request.GetJSONPostAs[*CreateSandboxTemplateRequest](req)
    if err != nil {
        log.ErrorContext(err, req.Context())
        return response.AdminBadRequestError[*sandbox.Sandbox](err)
    }

    // Business logic
    service := sandbox_service.NewSandboxService()
    sandboxInfo, err := service.CreateSandbox(req.Context(), accountID, config)
    if err != nil {
        log.ErrorContext(err, req.Context())
        return response.AdminBadRequestError[*sandbox.Sandbox](err)
    }

    // Save to database
    sandboxModel := sandbox.New()
    sandboxModel.AccountID.Set(accountID)
    sandboxModel.ExternalID.Set(sandboxInfo.SandboxID)
    err = sandboxModel.Save(userSession.User)
    if err != nil {
        log.ErrorContext(err, req.Context())
        return response.AdminBadRequestError[*sandbox.Sandbox](err)
    }

    return response.Success(sandboxModel)
}
```

---

## 4. Model Patterns

### 4.1 Model Generation

**ALWAYS use code generator tool** (`#code_tools make_object` or `make_public_object`):
- Generates model struct
- Creates controller boilerplate
- Generates database migration
- Creates query helpers

### 4.2 Model Structure

**Base pattern**:
```go
//go:generate core_gen model ModelName
package modelname

const (
    TABLE        = "tablename"
    CHANGE_LOGS  = true
    CLIENT       = environment.CLIENT_DEFAULT
    IS_VERSIONED = false
)

type Structure struct {
    DBColumns
    JoinData
}

type DBColumns struct {
    base.Structure
    Field1 *fields.StringField  `column:"field1" type:"text" default:"" public:"view"`
    Field2 *fields.UUIDField    `column:"field2" type:"uuid" default:"null" null:"true"`
}

type ModelName struct {
    model.BaseModel
    DBColumns
}
```

### 4.3 Field Access

```go
// Set values
model.Field.Set("value")
model.UUID.Set(types.UUID("uuid-string"))

// Get values
value := model.Field.Get()
uuid := model.UUID.Get()

// For struct/JSON fields
model.MetaData.Set(&MetaDataStruct{...})
data := model.MetaData.Get()
```

### 4.4 Query Building

```go
// Build query options
options := model.NewOptions().
    WithCondition("%s = :id:", Columns.ID_.Column()).
    WithParam(":id:", id).
    WithLimit(100).
    WithOrder("created_at DESC")

// Execute queries
models, err := modelname.FindAll(ctx, options)
model, err := modelname.FindFirst(ctx, options)
joined, err := modelname.GetJoined(ctx, id)
```

### 4.5 Model Lifecycle

```go
func (this *ModelName) beforeSave(ctx context.Context) error {
    this.BaseBeforeSave(ctx)
    common.GenerateURN(this)
    common.SetDisabledDeleted(this)
    return this.ValidateSubStructs()
}

func (this *ModelName) afterSave(ctx context.Context) {
    this.BaseAfterSave(ctx)
    // Post-save operations (caching, etc.)
}
```

### 4.6 Public JSON Response

**Location**: `/Users/griffnb/projects/techboss/techboss-ai-go/internal/models/sandbox/response.go`

```go
func (this *Sandbox) ToPublicJSON() any {
    return sanitize.SanitizeModel(this, &Structure{})
}
```

**Sanitization**:
- Only returns fields marked with `public:"view"` or `public:"edit"`
- Used by `StandardPublicRequestWrapper`

---

## 5. API Response Patterns

### 5.1 Success Response

```go
return response.Success(data)
// Returns: (data, http.StatusOK, nil)
```

### 5.2 Error Responses

**Admin errors** (full visibility):
```go
return response.AdminBadRequestError[*Type](err)
// Returns: (zeroValue, http.StatusBadRequest, err)
```

**Public errors** (masked):
```go
return response.PublicBadRequestError[*Type]()
// Returns: (zeroValue, http.StatusBadRequest, publicError)
```

### 5.3 Response Structure

**Standard format**:
```json
{
  "success": true,
  "data": { /* model or array */ }
}
```

**Error format**:
```json
{
  "success": false,
  "error": "Error message"
}
```

### 5.4 Pagination

**Frontend uses**:
- `limit` - Number of results per page
- `offset` - Skip N results
- `page` - Calculated as `Math.ceil(offset / limit) + 1`

**Backend** (from BuildIndexParams):
- Automatically converts query params to SQL
- Supports limit/offset for pagination
- No special pagination wrapper needed

---

## 6. Service Layer Pattern

### 6.1 Service Structure

**Location**: `/Users/griffnb/projects/techboss/techboss-ai-go/internal/services/sandbox_service/`

```go
type SandboxService struct {
    client *modal.APIClient
}

func NewSandboxService() *SandboxService {
    return &SandboxService{
        client: modal.Client(),
    }
}
```

### 6.2 Service Methods

**Pattern**:
```go
func (s *SandboxService) MethodName(
    ctx context.Context,
    param1 Type1,
    param2 Type2,
) (*ReturnType, error) {
    // Validate inputs
    if param1 == nil {
        return nil, errors.New("param1 cannot be nil")
    }
    
    // Business logic
    result, err := s.client.SomeMethod(ctx, param1, param2)
    if err != nil {
        return nil, errors.Wrapf(err, "failed to do something")
    }
    
    // TODO: Add database operations
    // TODO: Emit metrics
    
    return result, nil
}
```

### 6.3 Reconstructing SandboxInfo

**Helper function**:
```go
func ReconstructSandboxInfo(sandboxModel *sandbox.Sandbox, accountID types.UUID) *modal.SandboxInfo {
    return &modal.SandboxInfo{
        SandboxID: sandboxModel.ExternalID.Get(),
        Config: &modal.SandboxConfig{
            AccountID:       accountID,
            VolumeMountPath: "/mnt/workspace",
            // ... other config
        },
    }
}
```

---

## 7. Existing File Operations

### 7.1 State Files

**Location**: `/Users/griffnb/projects/techboss/techboss-ai-go/internal/services/sandbox_service/state_files/`

**Key types**:
```go
type StateFile struct {
    Version      string      `json:"version"`
    LastSyncedAt int64       `json:"last_synced_at"`
    Files        []FileEntry `json:"files"`
}

type FileEntry struct {
    Path       string `json:"path"`        // Relative path
    Checksum   string `json:"checksum"`    // MD5 hash
    Size       int64  `json:"size"`        // Bytes
    ModifiedAt int64  `json:"modified_at"` // Unix timestamp
}
```

**Functions**:
- `GenerateStateFile(ctx, sandboxInfo, dirPath)` - Scans directory, generates state
- `ReadLocalStateFile(ctx, sandboxInfo, volumePath)` - Reads from volume
- `ReadS3StateFile(ctx, sandboxInfo, s3Path)` - Reads from S3
- `WriteLocalStateFile(ctx, sandboxInfo, volumePath, state)` - Writes atomically
- `CompareStateFiles(local, s3)` - Determines sync actions needed

### 7.2 File Listing Implementation

**From GenerateStateFile**:

```go
// 1. List all files
cmd := []string{
    "sh", "-c",
    fmt.Sprintf("cd %s && find . -type f ! -name '%s' -print", 
        directoryPath, excludeFile),
}

// 2. Get metadata for each file
metaCmd := []string{
    "sh", "-c",
    fmt.Sprintf("stat -c '%%s|%%Y' %s && md5sum %s | cut -d' ' -f1", 
        fullPath, fullPath),
}

// 3. Build FileEntry array
fileEntries := []FileEntry{
    {
        Path:       cleanPath,
        Checksum:   checksum,
        Size:       size,
        ModifiedAt: modTime,
    },
}
```

---

## 8. Architecture Insights

### 8.1 Multi-Tenant Design

- All operations scoped by `AccountID`
- Sandboxes use `account-{id}` naming for Modal apps/volumes
- S3 paths include account: `docs/{account_id}/{timestamp}/`
- Database queries filtered by account ownership

### 8.2 Security Patterns

- Authentication via `request.GetReqSession(req)`
- Authorization via `helpers.RoleHandler`
- Public endpoints use `StandardPublicRequestWrapper` to sanitize responses
- Admin endpoints use `StandardRequestWrapper` for full visibility
- Field-level visibility via `public:"view"` tags

### 8.3 Error Handling

- Use `errors.Wrapf` for context
- Log errors with `log.ErrorContext(err, req.Context())`
- Return appropriate HTTP status codes
- Distinguish between admin and public errors

### 8.4 Test-Driven Development

**From AGENTS.md**:
- MANDATORY: Write tests FIRST
- Use table-driven tests
- 90% coverage required
- Mock external dependencies
- Use `assert` library from `lib/testtools/assert`

---

## 9. Recommended Implementation Approach

### 9.1 Phase 1: File Listing Endpoint

**Endpoint**: `GET /sandbox/{id}/files`

**Query Parameters**:
- `source` - "volume" or "s3" (default: "volume")
- `path` - Directory path to list (default: "/mnt/workspace")
- `recursive` - Include subdirectories (default: true)
- `limit` - Max files to return
- `offset` - Pagination offset

**Response**:
```json
{
  "success": true,
  "data": {
    "files": [
      {
        "path": "relative/path/file.txt",
        "name": "file.txt",
        "size": 1234,
        "modified_at": 1234567890,
        "is_directory": false
      }
    ],
    "total": 100,
    "source": "volume"
  }
}
```

### 9.2 Phase 2: File Content Retrieval

**Endpoint**: `GET /sandbox/{id}/files/content`

**Query Parameters**:
- `path` - File path to read
- `source` - "volume" or "s3"

**Response**:
```json
{
  "success": true,
  "data": {
    "path": "file.txt",
    "content": "file contents here",
    "size": 1234,
    "mime_type": "text/plain"
  }
}
```

### 9.3 Phase 3: Tree Structure

**Endpoint**: `GET /sandbox/{id}/files/tree`

**Response**:
```json
{
  "success": true,
  "data": {
    "name": "/",
    "path": "/mnt/workspace",
    "is_directory": true,
    "children": [
      {
        "name": "folder1",
        "path": "/mnt/workspace/folder1",
        "is_directory": true,
        "children": []
      }
    ]
  }
}
```

---

## 10. Key Files to Reference

### Models
- `/Users/griffnb/projects/techboss/techboss-ai-go/internal/models/sandbox/sandbox.go` - Sandbox model
- `/Users/griffnb/projects/techboss/techboss-ai-go/internal/models/sandbox/response.go` - Public JSON

### Controllers
- `/Users/griffnb/projects/techboss/techboss-ai-go/internal/controllers/sandboxes/setup.go` - Route setup
- `/Users/griffnb/projects/techboss/techboss-ai-go/internal/controllers/sandboxes/sandbox.go` - Controller logic

### Services
- `/Users/griffnb/projects/techboss/techboss-ai-go/internal/services/sandbox_service/sandbox_service.go` - Service layer
- `/Users/griffnb/projects/techboss/techboss-ai-go/internal/services/sandbox_service/reconstruct.go` - SandboxInfo helpers

### Integration
- `/Users/griffnb/projects/techboss/techboss-ai-go/internal/integrations/modal/sandbox.go` - Modal API client
- `/Users/griffnb/projects/techboss/techboss-ai-go/internal/integrations/modal/storage.go` - File operations
- `/Users/griffnb/projects/techboss/techboss-ai-go/internal/integrations/modal/storage_state.go` - State file operations

### State Files
- `/Users/griffnb/projects/techboss/techboss-ai-go/internal/services/sandbox_service/state_files/types.go` - Data structures
- `/Users/griffnb/projects/techboss/techboss-ai-go/internal/services/sandbox_service/state_files/reader.go` - Reading state
- `/Users/griffnb/projects/techboss/techboss-ai-go/internal/services/sandbox_service/state_files/writer.go` - Writing state

### Documentation
- `/Users/griffnb/projects/techboss/techboss-ai-go/AGENTS.md` - Development guidelines
- `/Users/griffnb/projects/techboss/techboss-ai-go/docs/MODELS.md` - Model system
- `/Users/griffnb/projects/techboss/techboss-ai-go/docs/CONTROLLERS.md` - Controller patterns

---

## 11. Code Patterns to Follow

### 11.1 Listing Files in Sandbox

```go
// Use find command with proper filtering
cmd := []string{
    "sh", "-c",
    fmt.Sprintf("cd %s && find . -type f -print", directoryPath),
}

process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, nil)
// ... parse output line by line ...
```

### 11.2 Getting File Metadata

```go
// Efficient single-command approach
cmd := []string{
    "sh", "-c",
    fmt.Sprintf("stat -c '%%s|%%Y|%%n' %s", filePath),
}
// Output: "1234|1234567890|filename.txt"
```

### 11.3 Reading File Contents

```go
cmd := []string{
    "sh", "-c",
    fmt.Sprintf("test -f %s && cat %s || echo 'FILE_NOT_FOUND'", path, path),
}

process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, nil)
// ... read output ...
```

### 11.4 Checking if Path is Directory

```go
cmd := []string{
    "sh", "-c",
    fmt.Sprintf("test -d %s && echo 'DIRECTORY' || echo 'NOT_DIRECTORY'", path),
}
```

---

## 12. Testing Patterns

### 12.1 Test Structure

```go
func TestFeature_Scenario(t *testing.T) {
    skipIfNotConfigured(t)
    
    ctx := context.Background()
    
    t.Run("specific test case", func(t *testing.T) {
        // Arrange
        // ... setup ...
        
        // Act
        result, err := functionUnderTest(ctx, params)
        
        // Assert
        assert.NoError(t, err)
        assert.NotEmpty(t, result)
        assert.Equal(t, expected, actual)
    })
}
```

### 12.2 Integration Test Helpers

```go
func skipIfNotConfigured(t *testing.T) {
    if os.Getenv("MODAL_TOKEN_ID") == "" {
        t.Skip("MODAL_TOKEN_ID not set, skipping integration test")
    }
}

func createTestSandbox(ctx context.Context, t *testing.T, accountID types.UUID) *modal.SandboxInfo {
    // ... create sandbox for testing ...
}

func cleanupSandbox(ctx context.Context, t *testing.T, sandboxInfo *modal.SandboxInfo) {
    err := sandboxInfo.Sandbox.Terminate(ctx)
    if err != nil {
        t.Logf("Warning: failed to cleanup sandbox: %v", err)
    }
}
```

---

## 13. Summary

### What We Have:
1. ✅ Sandbox model with External ID tracking
2. ✅ Modal integration client with command execution
3. ✅ S3 mount access via sandbox commands
4. ✅ Existing file listing/metadata patterns (GenerateStateFile)
5. ✅ Controller architecture with admin/public routes
6. ✅ Service layer pattern with validation
7. ✅ Query parameter auto-parsing
8. ✅ Public JSON sanitization

### What We Need to Build:
1. ❌ File listing endpoint with pagination
2. ❌ File content retrieval endpoint
3. ❌ Tree structure endpoint
4. ❌ File metadata endpoint
5. ❌ Support for both volume and S3 sources
6. ❌ Proper error handling for missing files
7. ❌ Security checks (ownership verification)
8. ❌ Comprehensive tests

### Design Decisions:
1. **Reuse existing patterns**: Follow GenerateStateFile's approach for file operations
2. **Service layer**: Add methods to `SandboxService`
3. **Reconstruct pattern**: Use `ReconstructSandboxInfo` for operations
4. **Public routes**: Use `StandardPublicRequestWrapper` for user-facing endpoints
5. **Pagination**: Use standard `limit`/`offset` query params
6. **Sources**: Support `?source=volume` or `?source=s3` parameter
7. **Error handling**: Return 404 for missing files, 403 for unauthorized access

---

## Next Steps

1. Review this document with the team
2. Create detailed design document with API specs
3. Break down into TDD-focused tasks
4. Implement file listing endpoint first (simplest)
5. Add file content retrieval
6. Build tree structure endpoint
7. Write comprehensive integration tests
8. Update documentation

---

**Research completed**: December 30, 2025
**Researcher**: GitHub Copilot Research Agent
**Status**: Ready for design phase
