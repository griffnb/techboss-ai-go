# Modal Sandbox System - Design Document

## Overview

This design document outlines the implementation of a powerful, configurable sandbox system using Modal that enables isolated Claude Code agent environments with persistent storage and S3 integration. The system will be implemented in two phases:

**Phase 1**: Core sandbox API infrastructure with comprehensive integration tests
**Phase 2**: HTTP endpoints and web UI for end-to-end user interaction

The design leverages the existing Modal Go SDK (`libmodal/modal-go`) and follows established patterns in the techboss-ai-go codebase for integration clients, controllers, testing, and streaming responses.

### Design Principles

1. **Follow TDD**: All code will be test-driven (RED → GREEN → REFACTOR)
2. **Idiomatic Go**: Use established patterns from existing codebase
3. **Proper Error Handling**: Wrap errors with context using `errors.Wrapf`
4. **Singleton Pattern**: Use `sync.Once` for client instantiation
5. **Resource Management**: Proper cleanup using defer and context cancellation
6. **Streaming First**: Support real-time interaction with Claude agent

---

## Architecture

### High-Level Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         Web UI (Phase 2)                    │
│                    (Simple HTML/CSS/JS)                     │
└───────────────────────┬─────────────────────────────────────┘
                        │ HTTP/SSE
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                    HTTP Controllers (Phase 2)                │
│   /api/modal/sandbox - Create, Get, Delete                 │
│   /api/modal/sandbox/:id/claude - Stream Claude output     │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│              Modal Integration Service (Phase 1)            │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Sandbox Management                      │  │
│  │  - CreateSandbox(config) → Sandbox                   │  │
│  │  - TerminateSandbox(sandbox)                         │  │
│  │  - GetSandboxStatus(id) → Status                     │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Storage Operations                      │  │
│  │  - SyncVolumeToS3(volume, bucket)                    │  │
│  │  - InitVolumeFromS3(volume, bucket)                  │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Claude Execution                        │  │
│  │  - ExecClaude(sandbox, prompt) → Process            │  │
│  │  - StreamClaudeOutput(process, writer)               │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
└───────────────────────┬─────────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────────┐
│                Modal Go SDK (libmodal/modal-go)             │
│  - Apps, Sandboxes, Volumes, CloudBucketMounts, Secrets    │
└─────────────────────────────────────────────────────────────┘
```

### Directory Structure

```
internal/
├── integrations/
│   └── modal/
│       ├── client.go              # Singleton client with Modal SDK (already exists)
│       ├── sandbox.go             # Sandbox creation/management (Phase 1)
│       ├── sandbox_test.go        # Real integration tests (Phase 1)
│       ├── storage.go             # S3/Volume sync operations (Phase 1)
│       ├── storage_test.go        # Storage integration tests (Phase 1)
│       ├── claude.go              # Claude execution & streaming (Phase 1)
│       └── claude_test.go         # Claude integration tests (Phase 1)
│
├── services/
│   └── modal/                     # Service layer (Phase 2)
│       └── sandbox_service.go     # Business logic layer between controller and integration
│
├── controllers/
│   └── sandbox/                   # Phase 2 controllers
│       ├── setup.go               # Route definitions
│       ├── sandbox.go             # Sandbox CRUD endpoints
│       └── claude.go              # Claude streaming endpoint
│
└── environment/
    └── config.go                  # Modal configuration struct (already exists)

static/                            # Phase 2 UI
└── modal-sandbox-ui.html          # Simple HTML UI
```

---

## Components and Interfaces

### Phase 1: Core Integration Service

#### 1. Modal API Client (`client.go`)

**Purpose**: Singleton client for Modal SDK operations

**Pattern**: Follows existing integration client pattern

```go
package modal

import (
    "sync"
    "github.com/modal-labs/libmodal/modal-go"
    "github.com/griffnb/techboss-ai-go/internal/environment"
    "github.com/griffnb/core/lib/tools"
)

var (
    instance *APIClient
    once     sync.Once
)

type APIClient struct {
    tokenID     string
    tokenSecret string
    client      *modal.Client
}

// Client returns the singleton instance
func Client() *APIClient {
    once.Do(func() {
        if !tools.Empty(environment.GetConfig().Modal) && 
           !tools.Empty(environment.GetConfig().Modal.TokenSecret) {
            apiConfig := environment.GetConfig().Modal
            instance = NewClient(apiConfig)
        }
    })
    return instance
}

// Configured checks if Modal is properly configured
func Configured() bool {
    return !tools.Empty(environment.GetConfig().Modal) && 
           !tools.Empty(environment.GetConfig().Modal.TokenSecret)
}

// NewClient creates a new Modal client instance
func NewClient(config *environment.Modal) *APIClient {
    client := &APIClient{
        tokenID:     config.TokenID,
        tokenSecret: config.TokenSecret,
    }
    
    mc, err := modal.NewClientWithOptions(&modal.ClientParams{
        TokenID:     client.tokenID,
        TokenSecret: client.tokenSecret,
    })
    if err != nil {
        log.Fatalf("Failed to create Modal client: %v", err)
    }
    client.client = mc
    
    return client
}
```

**Key Design Decisions**:
- Singleton pattern prevents multiple client instances
- Lazy initialization with `sync.Once` for thread safety
- `Configured()` function for graceful test skipping
- Follows exact pattern from existing `cloudflare`, `sendpulse` integrations

---

#### 2. Sandbox Management (`sandbox.go`)

**Purpose**: Create, configure, and manage Modal sandboxes

**Key Types**:

```go
// SandboxConfig holds configuration for creating a sandbox
type SandboxConfig struct {
    AccountID         types.UUID                      // Required: Account scoping
    Image             *ImageConfig                     // Docker image configuration
    VolumeName        string                          // Volume name for persistence
    VolumeMountPath   string                          // Where to mount volume (e.g., "/mnt/workspace")
    S3Config          *S3MountConfig                  // Optional S3 bucket mount
    Workdir           string                          // Working directory for processes
    Secrets           map[string]string               // Additional secrets to inject
    EnvironmentVars   map[string]string               // Custom environment variables
}

// ImageConfig defines Docker image to use
type ImageConfig struct {
    BaseImage        string                           // e.g., "alpine:3.21"
    DockerfileCommands []string                       // Custom Dockerfile commands
}

// S3MountConfig defines S3 bucket mount configuration
type S3MountConfig struct {
    BucketName       string                           // S3 bucket name
    SecretName       string                           // Modal secret for AWS credentials
    KeyPrefix        string                           // S3 key prefix (e.g., "docs/{accountID}/{timestamp}/")
    MountPath        string                           // Where to mount in container
    ReadOnly         bool                             // Mount as read-only
    Timestamp        int64                            // Unix timestamp for versioning
}

// SandboxInfo contains sandbox metadata and state
type SandboxInfo struct {
    SandboxID        string                           // Modal sandbox ID
    Sandbox          *modal.Sandbox                   // Underlying Modal sandbox
    Config           *SandboxConfig                   // Original configuration
    CreatedAt        time.Time                        // Creation timestamp
    Status           SandboxStatus                    // Current status
}

// SandboxStatus represents sandbox state
type SandboxStatus string

const (
    SandboxStatusRunning    SandboxStatus = "running"
    SandboxStatusTerminated SandboxStatus = "terminated"
    SandboxStatusError      SandboxStatus = "error"
)
```

**Core Methods**:

```go
// CreateSandbox creates a new Modal sandbox with the given configuration
func (c *APIClient) CreateSandbox(
    ctx context.Context,
    config *SandboxConfig,
) (*SandboxInfo, error)

// TerminateSandbox terminates a sandbox and optionally syncs data
func (c *APIClient) TerminateSandbox(
    ctx context.Context,
    sandboxInfo *SandboxInfo,
    syncToS3 bool,
) error

// GetSandboxStatus returns the current status of a sandbox
func (c *APIClient) GetSandboxStatus(
    ctx context.Context,
    sandboxID string,
) (SandboxStatus, error)

// WaitForSandbox blocks until sandbox exits and returns exit code
func (c *APIClient) WaitForSandbox(
    ctx context.Context,
    sandboxInfo *SandboxInfo,
) (int, error)

// GetLatestVersion retrieves the most recent timestamp version for an account
func (c *APIClient) GetLatestVersion(
    ctx context.Context,
    accountID types.UUID,
    bucketName string,
) (int64, error)
```

**Implementation Details**:

1. **App Creation**: Each sandbox gets or creates a Modal app scoped by `accountID`
   ```go
   appName := fmt.Sprintf("app-%s", config.AccountID.String())
   app, err := c.client.Apps.FromName(ctx, appName, &modal.AppFromNameParams{
       CreateIfMissing: true,
   })
   ```

2. **Image Building**: Support both registry images and custom Dockerfile commands
   ```go
   image := c.client.Images.FromRegistry(config.Image.BaseImage, nil)
   if len(config.Image.DockerfileCommands) > 0 {
       image = image.DockerfileCommands(config.Image.DockerfileCommands, nil)
   }
   ```

3. **Volume Setup**: Create or retrieve named volume
   ```go
   volumeName := config.VolumeName
   if tools.Empty(volumeName) {
       volumeName = fmt.Sprintf("volume-%s", config.AccountID.String())
   }
   volume, err := c.client.Volumes.FromName(ctx, volumeName, &modal.VolumeFromNameParams{
       CreateIfMissing: true,
   })
   ```

4. **S3 Mount** (optional): Create cloud bucket mount if configured
   ```go
   if config.S3Config != nil {
       secret, err := c.client.Secrets.FromName(ctx, config.S3Config.SecretName, nil)
       // Create cloud bucket mount
       cloudBucketMount, err := c.client.CloudBucketMounts.New(
           config.S3Config.BucketName,
           &modal.CloudBucketMountParams{
               Secret:    secret,
               KeyPrefix: &config.S3Config.KeyPrefix,
               ReadOnly:  config.S3Config.ReadOnly,
           },
       )
   }
   ```

5. **Sandbox Creation**: Assemble all components
   ```go
   sb, err := c.client.Sandboxes.Create(ctx, app, image, &modal.SandboxCreateParams{
       Volumes: map[string]*modal.Volume{
           config.VolumeMountPath: volume,
       },
       CloudBucketMounts: cloudBucketMounts,
       Workdir: &config.Workdir,
   })
   ```

**Error Handling**: All errors wrapped with `errors.Wrapf` including context

---

#### 3. Storage Operations (`storage.go`)

**Purpose**: Sync data between Modal volumes and S3 buckets

**Core Methods**:

```go
// InitVolumeFromS3 copies files from S3 bucket to volume on sandbox startup
func (c *APIClient) InitVolumeFromS3(
    ctx context.Context,
    sandboxInfo *SandboxInfo,
) (*SyncStats, error)

// SyncVolumeToS3 copies files from volume to S3 bucket
func (c *APIClient) SyncVolumeToS3(
    ctx context.Context,
    sandboxInfo *SandboxInfo,
) (*SyncStats, error)

// SyncStats contains statistics about sync operation
type SyncStats struct {
    FilesProcessed int           // Number of files synced
    BytesTransferred int64        // Total bytes transferred
    Duration       time.Duration // Time taken
    Errors         []error       // Any non-fatal errors
}
```

**Implementation Strategy**:

Since Modal sandboxes provide direct access to both volumes and S3 mounts, we'll use shell commands executed within the sandbox for efficient data transfer:

**InitVolumeFromS3 Implementation**:
```go
// Execute copy command inside sandbox
cmd := []string{
    "sh", "-c",
    fmt.Sprintf(
        "cp -r %s/* %s/ && echo 'Sync complete'",
        config.S3Config.MountPath,   // Source: S3 mount (read-only)
        config.VolumeMountPath,       // Destination: writable volume
    ),
}

process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, &modal.SandboxExecParams{
    Workdir: "/",
})
```

**SyncVolumeToS3 Implementation**:
```go
// Create new timestamped version in S3
timestamp := time.Now().Unix()
s3Path := fmt.Sprintf("%s/%d", config.AccountID, timestamp)

// Use AWS CLI to sync (requires AWS CLI in image)
cmd := []string{
    "sh", "-c",
    fmt.Sprintf(
        "aws s3 sync %s s3://%s/%s/",
        config.VolumeMountPath,
        config.S3Config.BucketName,
        s3Path,
    ),
}

// Execute with AWS credentials from secrets
secrets, err := c.client.Secrets.FromName(ctx, config.S3Config.SecretName, nil)
process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, &modal.SandboxExecParams{
    Secrets: []*modal.Secret{secrets},
    Workdir: "/",
})
```

**GetLatestVersion Implementation**:
```go
// List all versions (timestamps) for an account
cmd := []string{
    "sh", "-c",
    fmt.Sprintf(
        "aws s3 ls s3://%s/%s/ | awk '{print $2}' | sed 's/\\///g' | sort -n | tail -1",
        bucketName,
        accountID,
    ),
}

// Parse the latest timestamp from output
// Returns the most recent unix timestamp
```

**Key Design Decisions**:
- Use unix timestamps for versioning: `{account}/{timestamp}/files/...`
- Easy to identify latest version (highest timestamp)
- Preserves history without overwriting
- Use in-sandbox execution for efficiency (no data transfer to/from API)
- Requires AWS CLI in Docker image for S3 sync
- Read sync stats from command output
- Non-blocking with timeout protection

---

#### 4. Claude Execution (`claude.go`)

**Purpose**: Execute Claude Code CLI and stream output

**Core Methods**:

```go
// ClaudeExecConfig holds configuration for Claude execution
type ClaudeExecConfig struct {
    Prompt              string            // User prompt for Claude
    Workdir             string            // Working directory (default: volume mount)
    OutputFormat        string            // "stream-json" or "text"
    SkipPermissions     bool              // --dangerously-skip-permissions
    Verbose             bool              // Enable verbose output
    AdditionalFlags     []string          // Any additional CLI flags
}

// ClaudeProcess represents a running Claude process
type ClaudeProcess struct {
    Process     *modal.ContainerProcess  // Underlying Modal process
    Config      *ClaudeExecConfig        // Execution configuration
    StartedAt   time.Time                // When process started
}

// ExecClaude starts Claude Code CLI in the sandbox
func (c *APIClient) ExecClaude(
    ctx context.Context,
    sandboxInfo *SandboxInfo,
    config *ClaudeExecConfig,
) (*ClaudeProcess, error)

// StreamClaudeOutput streams Claude output to http.ResponseWriter using SSE
func (c *APIClient) StreamClaudeOutput(
    ctx context.Context,
    claudeProcess *ClaudeProcess,
    responseWriter http.ResponseWriter,
) error

// WaitForClaude blocks until Claude process completes
func (c *APIClient) WaitForClaude(
    ctx context.Context,
    claudeProcess *ClaudeProcess,
) (int, error)
```

**Implementation Details**:

1. **Build Claude Command**:
```go
cmd := []string{"claude"}

// Add flags
if config.SkipPermissions {
    cmd = append(cmd, "--dangerously-skip-permissions")
}
if config.Verbose {
    cmd = append(cmd, "--verbose")
}
if !tools.Empty(config.OutputFormat) {
    cmd = append(cmd, "--output-format", config.OutputFormat)
}

// Add prompt
cmd = append(cmd, "-c", "-p", config.Prompt)

// Add any additional flags
cmd = append(cmd, config.AdditionalFlags...)
```

2. **Get Claude Secrets**:
```go
// Retrieve Anthropic API key and AWS Bedrock credentials
secrets, err := c.client.Secrets.FromMap(ctx, map[string]string{
    "ANTHROPIC_API_KEY":       environment.GetConfig().Anthropic.APIKey,
    "AWS_BEDROCK_API_KEY":     environment.GetConfig().AWS.BedrockAPIKey,
    "CLAUDE_CODE_USE_BEDROCK": "1",
    "AWS_REGION":              environment.GetConfig().AWS.Region,
}, nil)
```

3. **Execute with PTY** (required by Claude CLI):
```go
workdir := config.Workdir
if tools.Empty(workdir) {
    workdir = sandboxInfo.Config.VolumeMountPath // Default to volume
}

process, err := sandboxInfo.Sandbox.Exec(ctx, cmd, &modal.SandboxExecParams{
    PTY:     true,  // CRITICAL: Claude requires PTY
    Secrets: []*modal.Secret{secrets},
    Workdir: workdir,
})
```

4. **Stream Output with SSE**:
```go
// Set SSE headers (pattern from openai/client.go)
responseWriter.Header().Set("Content-Type", "text/event-stream")
responseWriter.Header().Set("Cache-Control", "no-cache")
responseWriter.Header().Set("Connection", "keep-alive")
responseWriter.Header().Set("Access-Control-Allow-Origin", "*")

// Get flusher for real-time streaming
flusher, ok := responseWriter.(http.Flusher)
if !ok {
    return errors.New("response writer does not support flushing")
}

// Stream from Claude stdout
scanner := bufio.NewScanner(claudeProcess.Process.Stdout)
for scanner.Scan() {
    line := scanner.Text()
    
    // Write SSE formatted output
    _, err := fmt.Fprintf(responseWriter, "data: %s\n\n", line)
    if err != nil {
        return errors.Wrap(err, "failed to write streaming response")
    }
    
    // Flush immediately
    flusher.Flush()
}

// Check for scan errors
if err := scanner.Err(); err != nil {
    return errors.Wrap(err, "error reading Claude output")
}

// Send completion event
fmt.Fprintf(responseWriter, "data: [DONE]\n\n")
flusher.Flush()
```

**Key Design Decisions**:
- PTY required for Claude CLI interaction
- SSE for streaming (follows existing `openai` proxy pattern)
- Supports both JSON and text output formats
- Context cancellation propagates to Claude process
- All output logged for debugging

---

### Phase 2: HTTP Controllers & Web UI

#### 5. Service Layer (`services/modal/sandbox_service.go`)

**Purpose**: Business logic layer between controllers and Modal integration

**Pattern**: Follows existing service layer pattern from `ai_proxies/openai`

```go
package modal

import (
    "context"
    "net/http"
    
    "github.com/griffnb/core/lib/types"
    "github.com/griffnb/techboss-ai-go/internal/integrations/modal"
    "github.com/pkg/errors"
)

// SandboxService handles business logic for sandbox operations
type SandboxService struct {
    client *modal.APIClient
}

// NewSandboxService creates a new sandbox service
func NewSandboxService() *SandboxService {
    return &SandboxService{
        client: modal.Client(),
    }
}

// CreateSandbox creates a new sandbox with the given configuration
func (s *SandboxService) CreateSandbox(
    ctx context.Context,
    accountID types.UUID,
    config *modal.SandboxConfig,
) (*modal.SandboxInfo, error) {
    // Add account ID to config
    config.AccountID = accountID
    
    // Create sandbox via integration
    return s.client.CreateSandbox(ctx, config)
}

// TerminateSandbox terminates a sandbox
func (s *SandboxService) TerminateSandbox(
    ctx context.Context,
    sandboxInfo *modal.SandboxInfo,
    syncToS3 bool,
) error {
    return s.client.TerminateSandbox(ctx, sandboxInfo, syncToS3)
}

// ExecuteClaudeStream executes Claude and streams output to HTTP response
func (s *SandboxService) ExecuteClaudeStream(
    ctx context.Context,
    sandboxInfo *modal.SandboxInfo,
    config *modal.ClaudeExecConfig,
    responseWriter http.ResponseWriter,
) error {
    // Execute Claude
    claudeProcess, err := s.client.ExecClaude(ctx, sandboxInfo, config)
    if err != nil {
        return err
    }
    
    // Stream output
    return s.client.StreamClaudeOutput(ctx, claudeProcess, responseWriter)
}

// InitFromS3 initializes volume from S3 bucket
func (s *SandboxService) InitFromS3(
    ctx context.Context,
    sandboxInfo *modal.SandboxInfo,
) (*modal.SyncStats, error) {
    return s.client.InitVolumeFromS3(ctx, sandboxInfo)
}

// SyncToS3 syncs volume to S3 bucket
func (s *SandboxService) SyncToS3(
    ctx context.Context,
    sandboxInfo *modal.SandboxInfo,
) (*modal.SyncStats, error) {
    return s.client.SyncVolumeToS3(ctx, sandboxInfo)
}
```

**Key Design Decisions**:
- Service layer provides clean abstraction for controllers
- Handles business logic like adding account ID to config
- Even if pass-through now, allows for future enhancements:
  - Validation logic
  - Database persistence
  - Usage tracking
  - Rate limiting
- Follows pattern from `ai_proxies/openai/service.go`

---

#### 6. HTTP Controllers (`controllers/sandbox/`)

**Route Setup** (`setup.go`):

```go
package sandbox

import (
    "github.com/go-chi/chi/v5"
    "github.com/griffnb/core/lib/router"
    "github.com/griffnb/core/lib/tools"
    "github.com/griffnb/techboss-ai-go/internal/constants"
    "github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
)

const ROUTE string = "sandbox"

// Setup configures sandbox routes
func Setup(coreRouter *router.CoreRouter) {
    coreRouter.AddMainRoute(tools.BuildString("/", ROUTE), func(r chi.Router) {
        r.Group(func(authR chi.Router) {
            // Sandbox CRUD
            authR.Post("/", helpers.RoleHandler(helpers.RoleHandlerMap{
                constants.ROLE_ANY_AUTHORIZED: response.StandardRequestWrapper(createSandbox),
            }))
            
            authR.Get("/{sandboxID}", helpers.RoleHandler(helpers.RoleHandlerMap{
                constants.ROLE_ANY_AUTHORIZED: response.StandardRequestWrapper(getSandbox),
            }))
            
            authR.Delete("/{sandboxID}", helpers.RoleHandler(helpers.RoleHandlerMap{
                constants.ROLE_ANY_AUTHORIZED: response.StandardRequestWrapper(deleteSandbox),
            }))
            
            // Claude streaming endpoint
            authR.Post("/{sandboxID}/claude", helpers.RoleHandler(helpers.RoleHandlerMap{
                constants.ROLE_ANY_AUTHORIZED: router.NoTimeoutStreamingMiddleware(streamClaude),
            }))
        })
    })
}
```

**Sandbox CRUD Endpoints** (`sandbox.go`):

```go
// CreateSandboxRequest holds request data for sandbox creation
type CreateSandboxRequest struct {
    ImageBase           string            `json:"image_base"`
    DockerfileCommands  []string          `json:"dockerfile_commands"`
    VolumeName          string            `json:"volume_name"`
    S3BucketName        string            `json:"s3_bucket_name"`
    S3KeyPrefix         string            `json:"s3_key_prefix"`
    InitFromS3          bool              `json:"init_from_s3"`
}

// CreateSandboxResponse holds response data
type CreateSandboxResponse struct {
    SandboxID    string    `json:"sandbox_id"`
    Status       string    `json:"status"`
    CreatedAt    time.Time `json:"created_at"`
}

// createSandbox creates a new sandbox for the authenticated user
func createSandbox(w http.ResponseWriter, req *http.Request) (*CreateSandboxResponse, int, error) {
    // Get authenticated user session
    userSession := request.GetReqSession(req)
    accountID := userSession.User.AccountID
    
    // Parse request body
    rawData := request.GetJSONPostData(req)
    data := request.ConvertPost[CreateSandboxRequest](rawData)
    
    // Build sandbox config
    config := &modal.SandboxConfig{
        AccountID: accountID,
        Image: &modal.ImageConfig{
            BaseImage:          data.ImageBase,
            DockerfileCommands: data.DockerfileCommands,
        },
        VolumeName:      data.VolumeName,
        VolumeMountPath: "/mnt/workspace",
        Workdir:         "/mnt/workspace",
    }
    
    // Add S3 config if provided
    if !tools.Empty(data.S3BucketName) {
        config.S3Config = &modal.S3MountConfig{
            BucketName:  data.S3BucketName,
            SecretName:  "s3-bucket", // Default secret name
            KeyPrefix:   data.S3KeyPrefix,
            MountPath:   "/mnt/s3-bucket",
            ReadOnly:    true,
        }
    }
    
    // Create sandbox via service
    service := modal.NewSandboxService()
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
        }
    }
    
    // TODO: Store sandboxInfo in database/cache for retrieval
    
    // Return response
    resp := &CreateSandboxResponse{
        SandboxID: sandboxInfo.SandboxID,
        Status:    string(sandboxInfo.Status),
        CreatedAt: sandboxInfo.CreatedAt,
    }
    
    return response.Success(resp)
}

// getSandbox retrieves sandbox status
func getSandbox(w http.ResponseWriter, req *http.Request) (*CreateSandboxResponse, int, error) {
    sandboxID := chi.URLParam(req, "sandboxID")
    
    // TODO: Retrieve sandboxInfo from database/cache
    // For now, return error
    
    return response.AdminBadRequestError[*CreateSandboxResponse](
        errors.New("sandbox retrieval not yet implemented"),
    )
}

// deleteSandbox terminates a sandbox
func deleteSandbox(w http.ResponseWriter, req *http.Request) (*CreateSandboxResponse, int, error) {
    sandboxID := chi.URLParam(req, "sandboxID")
    
    // TODO: Retrieve sandboxInfo from database/cache
    // Terminate sandbox
    // client.TerminateSandbox(req.Context(), sandboxInfo, true)
    
    return response.AdminBadRequestError[*CreateSandboxResponse](
        errors.New("sandbox deletion not yet implemented"),
    )
}
```

**Claude Streaming Endpoint** (`claude.go`):

```go
// ClaudeRequest holds request data for Claude execution
type ClaudeRequest struct {
    Prompt string `json:"prompt"`
}

// streamClaude executes Claude and streams output
func streamClaude(w http.ResponseWriter, req *http.Request) {
    sandboxID := chi.URLParam(req, "sandboxID")
    
    // Parse request body
    rawData := request.GetJSONPostData(req)
    data := request.ConvertPost[ClaudeRequest](rawData)
    
    if tools.Empty(data.Prompt) {
        http.Error(w, "prompt is required", http.StatusBadRequest)
        return
    }
    
    // TODO: Retrieve sandboxInfo from database/cache
    // For Phase 1, we'll pass sandboxInfo directly in tests
    
    // Create Claude execution config
    claudeConfig := &modal.ClaudeExecConfig{
        Prompt:          data.Prompt,
        OutputFormat:    "stream-json",
        SkipPermissions: true,
        Verbose:         true,
    }
    
    // Execute Claude and stream via service
    service := modal.NewSandboxService()
    err = service.ExecuteClaudeStream(req.Context(), sandboxInfo, claudeConfig, w)
    if err != nil {
        log.ErrorContext(err, req.Context())
        http.Error(w, "Failed to execute Claude", http.StatusInternalServerError)
        return
    }
}
```

**Key Design Decisions**:
- Standard controller patterns from existing codebase
- Role-based access control (any authorized user)
- Request/response structs with JSON tags
- Streaming endpoint uses `NoTimeoutStreamingMiddleware`
- TODO markers for Phase 2 implementation (database storage)

---

#### 6. Web UI (`static/modal-sandbox-ui.html`)

**Purpose**: Simple HTML interface for creating sandboxes and chatting with Claude

**Features**:
- Create sandbox button with loading state
- Chat interface with message history
- Real-time streaming of Claude responses
- Error handling and display
- No build process required (vanilla JS)

**Architecture**:
```html
<!DOCTYPE html>
<html>
<head>
    <title>Modal Sandbox - Claude Agent</title>
    <style>
        /* Simple, clean CSS */
        /* Chat interface styling */
        /* Loading states */
    </style>
</head>
<body>
    <div id="app">
        <!-- Sandbox creation section -->
        <div id="create-section">
            <h1>Create Modal Sandbox</h1>
            <button id="create-btn">Create Sandbox</button>
            <div id="status"></div>
        </div>
        
        <!-- Chat section (hidden until sandbox ready) -->
        <div id="chat-section" style="display: none;">
            <h1>Chat with Claude</h1>
            <div id="messages"></div>
            <form id="chat-form">
                <input type="text" id="prompt-input" placeholder="Ask Claude..." />
                <button type="submit">Send</button>
            </form>
        </div>
    </div>
    
    <script>
        let sandboxID = null;
        
        // Create sandbox
        document.getElementById('create-btn').addEventListener('click', async () => {
            const statusEl = document.getElementById('status');
            statusEl.textContent = 'Creating sandbox...';
            
            try {
                const response = await fetch('/api/sandbox', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    credentials: 'include',
                    body: JSON.stringify({
                        image_base: 'alpine:3.21',
                        dockerfile_commands: [
                            'RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep aws-cli',
                            'RUN curl -fsSL https://claude.ai/install.sh | bash',
                            'ENV PATH=/root/.local/bin:$PATH USE_BUILTIN_RIPGREP=0'
                        ],
                        init_from_s3: true
                    })
                });
                
                if (!response.ok) throw new Error('Failed to create sandbox');
                
                const data = await response.json();
                sandboxID = data.sandbox_id;
                
                // Show chat interface
                document.getElementById('create-section').style.display = 'none';
                document.getElementById('chat-section').style.display = 'block';
                
            } catch (error) {
                statusEl.textContent = `Error: ${error.message}`;
            }
        });
        
        // Send message to Claude
        document.getElementById('chat-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const input = document.getElementById('prompt-input');
            const prompt = input.value.trim();
            if (!prompt) return;
            
            // Add user message to UI
            addMessage('user', prompt);
            input.value = '';
            
            // Create assistant message container
            const assistantMsg = addMessage('assistant', '');
            
            // Stream Claude response using EventSource
            const url = `/api/sandbox/${sandboxID}/claude`;
            const response = await fetch(url, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify({ prompt })
            });
            
            // Read streaming response
            const reader = response.body.getReader();
            const decoder = new TextDecoder();
            
            while (true) {
                const { done, value } = await reader.read();
                if (done) break;
                
                const chunk = decoder.decode(value);
                const lines = chunk.split('\n');
                
                for (const line of lines) {
                    if (line.startsWith('data: ')) {
                        const data = line.slice(6);
                        if (data === '[DONE]') break;
                        
                        // Append to assistant message
                        assistantMsg.textContent += data;
                    }
                }
            }
        });
        
        function addMessage(role, content) {
            const messagesEl = document.getElementById('messages');
            const msgEl = document.createElement('div');
            msgEl.className = `message message-${role}`;
            msgEl.textContent = content;
            messagesEl.appendChild(msgEl);
            messagesEl.scrollTop = messagesEl.scrollHeight;
            return msgEl;
        }
    </script>
</body>
</html>
```

**Key Design Decisions**:
- Zero dependencies (vanilla JS)
- Server-Sent Events for streaming
- Progressive enhancement (works without JS for basic functionality)
- Responsive design with CSS Grid/Flexbox
- Error boundaries for graceful degradation

---

## Data Models

### Configuration Structures

```go
// Environment configuration (in environment/config.go)
type Config struct {
    config.DefaultConfig
    Modal      *Modal      `json:"modal"`
    Anthropic  *Anthropic  `json:"anthropic"`
    AWS        *AWS        `json:"aws"`
    // ... other integrations
}

type Modal struct {
    TokenID     string `json:"token_id"`
    TokenSecret string `json:"token_secret"`
}

type Anthropic struct {
    APIKey string `json:"api_key"`
}

type AWS struct {
    BedrockAPIKey string `json:"bedrock_api_key"`
    Region        string `json:"region"`
}
```

### Runtime Data Structures

**Phase 1** stores data in memory (for testing):
- `SandboxInfo` struct holds all sandbox metadata
- Passed directly between functions in tests

**Phase 2** (future enhancement):
- Store `SandboxInfo` in database or cache (Redis)
- Associate sandboxes with user accounts
- Track sandbox lifecycle and usage

**Database Schema** (Phase 2 - not implemented yet):
```sql
CREATE TABLE modal_sandboxes (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id),
    sandbox_id VARCHAR(255) NOT NULL,
    config JSONB NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    terminated_at TIMESTAMP,
    INDEX idx_account_id (account_id),
    INDEX idx_sandbox_id (sandbox_id)
);
```

---

## Error Handling

### Error Wrapping Pattern

All errors follow the established pattern:

```go
if err != nil {
    return errors.Wrapf(err, "failed to {operation}: additional context")
}
```

### Error Categories

1. **Configuration Errors**: Missing credentials, invalid config
   - Fail fast with descriptive messages
   - Check with `Configured()` before operations

2. **Modal SDK Errors**: API failures, network issues
   - Wrap with operation context
   - Include sandbox ID, account ID where relevant

3. **Execution Errors**: Claude process failures, command errors
   - Capture exit codes
   - Stream error output to client

4. **Streaming Errors**: Connection issues, flush failures
   - Log but don't necessarily fail (client may have disconnected)
   - Cleanup resources in defer

### Error Logging

```go
// Always log errors with context
log.ErrorContext(err, ctx)

// Log important operations
log.Infof("Created sandbox %s for account %s", sandboxID, accountID)

// Log Claude execution
log.Infof("Executing Claude with prompt: %s", prompt)
log.Infof("Claude exited with code %d", exitCode)
```

### HTTP Error Responses

**Phase 2 Controllers**:
```go
// Admin endpoints (detailed errors)
return response.AdminBadRequestError[T](err)

// Public endpoints (generic errors)
return response.PublicBadRequestError[T]()

// Streaming endpoints
http.Error(w, "Internal server error", http.StatusInternalServerError)
```

---

## Testing Strategy

### Test-Driven Development

**Mandatory TDD Workflow**:
1. **RED**: Write test first (it fails)
2. **GREEN**: Write minimal code to pass
3. **REFACTOR**: Clean up while keeping tests green
4. **COMMIT**: Commit with passing tests

### Phase 1 Integration Tests

All Phase 1 tests are **real integration tests** against Modal infrastructure:

**Test Setup** (`sandbox_test.go`):
```go
package modal_test

import (
    "testing"
    "github.com/griffnb/techboss-ai-go/internal/common/system_testing"
    "github.com/griffnb/techboss-ai-go/internal/integrations/modal"
)

func init() {
    system_testing.BuildSystem()  // Load config, secrets
}

func skipIfNotConfigured(t *testing.T) {
    if !modal.Configured() {
        t.Skip("Modal client is not configured, skipping test")
    }
}
```

**Test Structure** (table-driven):
```go
func TestCreateSandbox(t *testing.T) {
    skipIfNotConfigured(t)
    
    client := modal.Client()
    ctx := context.Background()
    
    t.Run("Basic sandbox with volume only", func(t *testing.T) {
        // Arrange
        config := &modal.SandboxConfig{
            AccountID: types.NewUUID(),
            Image: &modal.ImageConfig{
                BaseImage: "alpine:3.21",
            },
            VolumeName:      "test-volume",
            VolumeMountPath: "/mnt/workspace",
            Workdir:         "/mnt/workspace",
        }
        
        // Act
        sandboxInfo, err := client.CreateSandbox(ctx, config)
        defer client.TerminateSandbox(ctx, sandboxInfo, false)
        
        // Assert
        assert.NoError(t, err)
        assert.NotEmpty(t, sandboxInfo.SandboxID)
        assert.Equal(t, modal.SandboxStatusRunning, sandboxInfo.Status)
    })
    
    t.Run("Sandbox with S3 mount", func(t *testing.T) {
        // Test S3 integration
        // ...
    })
    
    t.Run("Sandbox with custom Dockerfile commands", func(t *testing.T) {
        // Test custom image building
        // ...
    })
}
```

**Test Coverage Requirements**:
- ≥90% code coverage for new code
- Test all success paths
- Test all error scenarios
- Test edge cases (empty configs, nil pointers, etc.)

**Test Files**:
1. `sandbox_test.go`: Sandbox creation, termination, status
2. `storage_test.go`: S3 sync operations
3. `claude_test.go`: Claude execution and streaming

**Critical Tests**:
- Sandbox creation with various configurations
- Volume initialization from S3
- Volume sync to S3
- Claude execution with real prompts
- Streaming output verification
- Resource cleanup (terminate sandbox)
- Error handling (invalid configs, network failures)
- Context cancellation

### Phase 2 Controller Tests

**Unit Tests** (mocked integration):
- Test request parsing
- Test response formatting
- Test error handling
- Mock Modal client calls

**Integration Tests** (optional):
- End-to-end flow from HTTP request to Claude response
- Real Modal sandbox creation
- Real streaming to HTTP client

### Running Tests

All tests MUST use `#code_tools`:

```bash
# Run all Modal tests
#code_tools run_tests internal/integrations/modal

# Run specific test
#code_tools run_tests internal/integrations/modal -test TestCreateSandbox

# With coverage
#code_tools run_tests internal/integrations/modal -coverage
```

### Test Cleanup

```go
// Always cleanup in defer
defer func() {
    if sandboxInfo != nil {
        err := client.TerminateSandbox(ctx, sandboxInfo, false)
        if err != nil {
            t.Logf("Failed to cleanup sandbox: %v", err)
        }
    }
}()
```

---

## Implementation Phases

### Phase 1: Core Integration Service

**Deliverables**:
1. ✅ Modal API client with singleton pattern
2. ✅ Sandbox creation with flexible configuration
3. ✅ Volume and S3 bucket mounting
4. ✅ S3-to-volume initialization
5. ✅ Volume-to-S3 sync
6. ✅ Claude execution with streaming
7. ✅ Comprehensive integration tests (≥90% coverage)
8. ✅ All tests passing against real Modal infrastructure

**Key Files**:
- `internal/integrations/modal/client.go` (already exists - verify configuration)
- `internal/integrations/modal/sandbox.go`
- `internal/integrations/modal/storage.go`
- `internal/integrations/modal/claude.go`
- `internal/integrations/modal/*_test.go`

**Success Criteria**:
- All tests pass with real Modal API
- Code coverage ≥90%
- All requirements from Phase 1 satisfied
- Zero hardcoded values (use config)

### Phase 2: HTTP Endpoints & Web UI

**Deliverables**:
1. ✅ Controller setup with routes
2. ✅ Sandbox CRUD endpoints
3. ✅ Claude streaming endpoint
4. ✅ Simple HTML/CSS/JS UI
5. ✅ End-to-end testing
6. ✅ Documentation updates

**Key Files**:
- `internal/services/modal/sandbox_service.go`
- `internal/controllers/sandbox/setup.go`
- `internal/controllers/sandbox/sandbox.go`
- `internal/controllers/sandbox/claude.go`
- `static/modal-sandbox-ui.html`

**Success Criteria**:
- Users can create sandboxes via UI
- Users can chat with Claude in real-time
- Error handling works correctly
- UI is responsive and intuitive

**Future Enhancements** (not in scope):
- Database persistence for sandbox metadata
- Sandbox lifecycle management (timeouts)
- Usage tracking and billing
- Multi-user collaboration
- WebSocket for bidirectional communication

---

## Security Considerations

### Authentication & Authorization

- All endpoints require authentication (ROLE_ANY_AUTHORIZED)
- Sandboxes scoped to user's account ID
- No cross-account access

### Secrets Management

- API keys stored in environment config
- Modal secrets for AWS credentials
- Never expose secrets in API responses
- Use Modal's secret injection for Claude

### Sandbox Isolation

- Each account gets separate Modal app
- Volumes scoped to account
- S3 key prefixes enforce isolation
- No shared resources between accounts

### Input Validation

- Validate all user inputs
- Sanitize prompts before execution
- Check for malicious Dockerfile commands
- Limit resource usage (future enhancement)

---

## Performance Considerations

### Streaming Optimization

- Use `http.Flusher` for immediate delivery
- Buffer scanner for efficient I/O
- Context cancellation propagates to Modal

### Resource Management

- Cleanup sandboxes in defer blocks
- Set timeouts on long-running operations
- Limit concurrent sandbox count (future)

### Caching Strategy (Phase 2)

- Cache sandbox metadata in Redis
- Avoid database lookups on streaming endpoint
- Invalidate cache on termination

---

## Monitoring & Observability

### Logging

```go
// Key events
log.Infof("Created sandbox %s for account %s", sandboxID, accountID)
log.Infof("Synced %d files (%d bytes) to S3 in %s", stats.FilesProcessed, stats.BytesTransferred, stats.Duration)
log.Infof("Claude process exited with code %d", exitCode)

// Errors
log.ErrorContext(err, ctx)
```

### Metrics (Future Enhancement)

- Sandbox creation rate
- Average sandbox lifetime
- Claude execution duration
- Streaming connection drops
- S3 sync times

### Health Checks

```go
// Endpoint to verify Modal connectivity
func healthCheck(w http.ResponseWriter, req *http.Request) {
    if !modal.Configured() {
        http.Error(w, "Modal not configured", http.StatusServiceUnavailable)
        return
    }
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
```

---

## Dependencies

### Go Modules

```go
require (
    github.com/modal-labs/libmodal/modal-go v0.x.x
    github.com/pkg/errors v0.9.1
    github.com/go-chi/chi/v5 v5.x.x
    // ... existing dependencies
)
```

### External Services

- **Modal**: Sandbox infrastructure
- **S3**: Document storage (via Modal cloud bucket mounts)
- **Anthropic API**: Claude Code CLI
- **AWS Bedrock** (optional): Alternative Claude access

### Docker Image Requirements

For Claude Code to work, the Docker image must include:
- Bash shell
- Curl (for installation)
- Git (for repository operations)
- Ripgrep (for code search)
- Claude CLI (installed via install script)
- AWS CLI (for S3 sync operations)

**Standard Image Configuration** (using DockerfileCommands):
```go
DockerfileCommands: []string{
    "RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep aws-cli",
    "RUN curl -fsSL https://claude.ai/install.sh | bash",
    "ENV PATH=/root/.local/bin:$PATH USE_BUILTIN_RIPGREP=0",
}
```

**Note**: For testing/proof of concept, we use Modal's DockerfileCommands to build the image dynamically. This approach is simpler than maintaining a separate Dockerfile.

---

## Appendix

### Modal SDK Reference

**Key Modal SDK Concepts**:
- **Apps**: Top-level namespace for resources
- **Images**: Docker images with optional custom builds
- **Sandboxes**: Isolated compute environments
- **Volumes**: Persistent storage attached to sandboxes
- **CloudBucketMounts**: Read-only S3 bucket mounts
- **Secrets**: Secure credential injection

**SDK Documentation**: https://github.com/modal-labs/libmodal

### Existing Code References

**Patterns to Follow**:
- Client singleton: `internal/integrations/cloudflare/client.go`
- Streaming SSE: `internal/services/ai_proxies/openai/client.go`
- Controller setup: `internal/controllers/ai/setup.go`
- Integration tests: `internal/integrations/cloudflare/*_test.go`

### Related Documentation

- [AGENTS.md](/Users/griffnb/projects/techboss/techboss-ai-go/AGENTS.md) - TDD requirements, Go patterns
- [CONTROLLERS.md](/Users/griffnb/projects/techboss/techboss-ai-go/docs/CONTROLLERS.md) - Controller patterns
- [PRD.md](/Users/griffnb/projects/techboss/techboss-ai-go/docs/PRD.md) - Project requirements

---

## Open Questions & Future Work

**Phase 1 Questions**:
1. Should we add retry logic for S3 sync operations?
2. What timeout values should we use for long-running operations?
3. Should we support WebSocket in addition to SSE?

**Phase 2 Enhancements**:
1. Database model for sandbox persistence
2. Automatic sandbox cleanup after inactivity
3. Usage tracking and billing integration
4. Multiple Claude sessions per sandbox
5. File upload/download endpoints
6. Sandbox snapshot/restore functionality

**Future Features**:
1. Team collaboration (shared sandboxes)
2. Custom Docker image registry
3. GPU-enabled sandboxes
4. Integration with other LLMs
5. Sandbox templates and presets
