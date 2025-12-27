# Sandbox Database Migration - Design Document

## Overview

This design migrates the Modal sandbox system from in-memory cache storage (`sync.Map`) to persistent database storage using the existing `SandboxModel`. The migration enables sandbox persistence across server restarts, multi-instance deployments, and provides proper multi-tenant isolation. The UI will be enhanced to support selecting existing sandboxes or creating new ones.

### Goals

1. Replace in-memory `sandboxCache` with database persistence
2. Store sandbox creation timestamp and last S3 sync info in `meta_data` JSONB column
3. Use `external_id` column to store Modal sandbox ID for direct querying
4. Use existing `status` column for sandbox status (running, terminated, error)
5. Implement account-scoped queries for multi-tenant security
6. Update controllers to use premade sandbox templates based on provider/agent types
7. Update UI to display and select from existing sandboxes
8. Remove all Phase 1/Phase 2 TODO comments related to cache migration

### Non-Goals

- Exposing sandbox configuration to frontend (templates are backend-only)
- Adding joined query functions (not needed for this model)
- Changing the Modal integration layer
- Implementing sandbox lifecycle management (auto-cleanup, billing)
- Adding admin CRUD endpoints (already auto-generated)

## Architecture

### Current Architecture (Phase 1)

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTP
       ▼
┌─────────────────────────────────┐
│      Controllers                │
│  - createSandbox()              │
│  - getSandbox()                 │
│  - deleteSandbox()              │
│  - streamClaude()               │
└──────┬─────────────┬────────────┘
       │             │
       │             └──────────┐
       ▼                        ▼
┌──────────────┐         ┌─────────────┐
│ SandboxService│        │ sandboxCache│
│   Layer      │         │  (sync.Map) │
└──────┬───────┘         └─────────────┘
       │                  In-Memory Only
       ▼
┌─────────────────┐
│ Modal Integration│
│    (API Client)  │
└──────────────────┘
```

### New Architecture (Phase 2)

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTP
       ▼
┌────────────────────────────────────────┐
│         Controllers                     │
│  - createSandbox()    (DB save)        │
│  - getSandbox()       (DB query)       │
│  - listSandboxes()    (NEW)            │
│  - deleteSandbox()    (DB delete)      │
│  - streamClaude()     (DB lookup)      │
└──────┬─────────────────────────────────┘
       │
       ├──────────────────┬───────────────┐
       │                  │               │
       ▼                  ▼               ▼
┌──────────────┐   ┌────────────┐  ┌──────────────┐
│SandboxService│   │  Sandbox   │  │   Database   │
│   Layer      │   │   Model    │  │  (Postgres)  │
└──────┬───────┘   └────────────┘  └──────────────┘
       │                Persistent      Multi-Tenant
       ▼                 Storage           Scoped
┌─────────────────┐
│ Modal Integration│
│    (API Client)  │
└──────────────────┘
```

### Data Flow

#### Create Sandbox Flow
```
1. User → POST /sandbox → createSandbox(agentID, providerType)
2. Extract accountID from userSession
3. Service.GetSandboxTemplate(providerType, agentID) → premade config
4. Call SandboxService.CreateSandbox() → Modal API
5. Receive SandboxInfo from Modal
6. Create Sandbox model instance
7. Set AccountID, AgentID, Provider, ExternalID (sandbox_id), Status
8. Set MetaData with minimal info (created_at, last_sync)
9. Save to database
10. Return SandboxID to client
```

#### Get Sandbox Flow
```
1  User → GET /{id}  → authGet()
2. Extract accountID from userSession
3. Query database: Get by ID with AccountID verification
4. Return sandbox details to client
Note: Auto-generated auth endpoint handles ownership verificationa
6. Return sandbox details to client
```

#### List Sandboxes Flow
```
1. User → GET /sandbox → authIndex()
2. Extract accountID from userSession
3. Query database: FindAll by AccountID, order by CreatedAt DESC
4. Filter out deleted/disabled sandboxes (automatic)
5. Return list of sandboxes
Note: Auto-generated auth endpoint handles this
```

#### Delete Sandbox Flow
```
1. User → DELETE /sandbox/{id} → authDelete()
2. Extract accountID from userSession
3. Query database: Get by ID with AccountID verification
4. Reconstruct SandboxInfo from model fields
5. Call SandboxService.TerminateSandbox() → Modal API
6. Soft-delete database record (set deleted=1)
7. Return success response
```

#### Stream Claude Flow
```
1. User → POST /sandbox/{sandboxID}/claude → streamClaude()
2. Extract accountID fromid}/claude → streamClaude()
2. Extract accountID from userSession
3. Query database: Get by ID with AccountID verification
4. Reconstruct SandboxInfo from model fields (external_id, provider)
5. Call SandboxService.ExecuteClaudeStream() with SandboxInfo
6``

## Components and Interfaces

### 1. MetaData Structure (Simplified)

**File**: `internal/models/sandbox/meta_data.go`

**Current**:
```go
type MetaData struct {
    SandboxID string               `json:"sandbox_id"` // Modal sandbox ID
    Status    *modal.SandboxStatus `json:"status"`     // Current status
}
```

**Updated**:
```go
package sandbox

import (
    "time"
)

// MetaData stores minimal sandbox metadata in JSONB format
// Most data is now stored in dedicated columns (external_id, status, provider, agent_id)
// All fields use snake_case for JSON marshaling
type MetaData struct {
    LastS3Sync  *int64      `json:"last_s3_sync"`   // Last S3 sync unix timestamp (nullable)
    SyncStats   *SyncStats  `json:"sync_stats"`     // Last sync statistics (nullable)
}

// SyncStats stores the results of the last S3 sync operation
type SyncStats struct {
    FilesProcessed   int   `json:"files_processed"`
    BytesTransferred int64 `json:"bytes_transferred"`
    DurationMs       int64 `json:"duration_ms"`
}

// UpdateLastSync updates the last sync timestamp and stats
func (m *MetaData) UpdateLastSync(filesProcessed int, bytesTransferred int64, durationMs int64) {
    now := time.Now().Unix()
    m.LastS3Sync = &now
    m.SyncStats = &SyncStats{
        FilesProcessed:   filesProcessed,
        BytesTransferred: bytesTransferred,
        DurationMs:       durationMs,
    }
}
```

### 2. Sandbox Model Fields

The sandbox model now uses dedicated columns instead of storing everything in metadata:

```go
type DBColumns struct {
    base.Structure                                          // ID, URN, Status, Deleted, Disabled, CreatedAt, UpdatedAt
    OrganizationID *fields.UUIDField                       `column:"organization_id" type:"uuid"     ...`
    AccountID      *fields.UUIDField                       `column:"account_id"      type:"uuid"     ...`
    AgentID        *fields.UUIDField                       `column:"agent_id"        type:"uuid"     ...`
    Provider       *fields.IntConstantField[Provider]      `column:"type"            type:"smallint" ...`
    ExternalID     *fields.StringField                     `column:"external_id"     type:"text"     ...` // Modal sandbox ID
    MetaData       *fields.StructField[*MetaData]          `column:"meta_data"       type:"jsonb"    ...`
}
```

**Key Fields**:
- `ExternalID`: Stores the Modal sandbox ID (e.g., "sb-abc123") for direct querying
- `Status`: Uses base.Structure.Status for sandbox state (0=active, 1=terminated, etc.)
- `Provider`: Enum for provider type (PROVIDER_CLAUDE_CODE, etc.)
- `AgentID`: Links to the agent configuration
- `MetaData`: Minimal JSONB for sync timestamps and stats only

### 3. Database Query Functions

**File**: `internal/models/sandbox/queries.go`

```go
package sandbox

import (
    "context"
    "github.com/griffnb/core/lib/model"
    "github.com/griffnb/core/lib/types"
    "github.com/pkg/errors"
)

// FindByExternalID finds a sandbox by its Modal external ID and AccountID
// This ensures users can only access their own sandboxes
func FindByExternalID(ctx context.Context, externalID string, accountID types.UUID) (*Sandbox, error) {
    options := model.NewOptions().
        WithCondition(
            "%s = :external_id: AND %s = :account_id: AND %s = 0 AND %s = 0",
            Columns.ExternalID.Column(),
            Columns.AccountID.Column(),
            Columns.Deleted.Column(),
            Columns.Disabled.Column(),
        ).
        WithParam(":external_id:", externalID).
        WithParam(":account_id:", accountID)
    
    sandbox, err := FindFirst(ctx, options)
    if err != nil {
        return nil, errors.Wrapf(err, "sandbox not found: %s", externalID)
    }
    
    if sandbox == nil {
        return nil, errors.Errorf("sandbox not found: %s", externalID)
    }
    
    return sandbox, nil
}

// FindAllByAccount returns all active sandboxes for a specific account
// Excludes deleted and disabled sandboxes
func FindAllByAccount(ctx context.Context, accountID types.UUID) ([]*Sandbox, error) {
    options := model.NewOptions().
        WithCondition("%s = :account_id: AND %s = 0 AND %s = 0", 
            Columns.AccountID.Column(), 
            Columns.Deleted.Column(),
            Columns.Disabled.Column()).
        WithParam(":account_id:", accountID).
        WithOrderBy("created_at DESC")
    
    return FindAll(ctx, options)
}

// CountByAccount returns the count of active sandboxes for an account
func CountByAccount(ctx context.Context, accountID types.UUID) (int64, error) {
    options := model.NewOptions().
        WithCondition("%s = :account_id: AND %s = 0 AND %s = 0",
            Columns.AccountID.Column(),
            Columns.Deleted.Column(),
            Columns.Disabled.Column()).
        WithParam(":account_id:", accountID)
    
    return FindResultsCount(ctx, options)
}
```

### 4. Sandbox Templates Service

**File**: `internal/services/sandbox_service/templates.go` (NEW)

This file defines premade sandbox configurations based on provider and agent type.

```go
package sandbox_service

import (
    "github.com/griffnb/core/lib/types"
    "github.com/griffnb/techboss-ai-go/internal/integrations/modal"
    "github.com/griffnb/techboss-ai-go/internal/models/sandbox"
    "github.com/pkg/errors"
)

// SandboxTemplate defines a premade sandbox configuration
type SandboxTemplate struct {
    Provider       sandbox.Provider
    ImageConfig    *modal.ImageConfig
    VolumeName     string
    S3BucketName   string
    S3KeyPrefix    string
    InitFromS3     bool
}

// GetSandboxTemplate returns a premade template based on provider and agent
func GetSandboxTemplate(provider sandbox.Provider, agentID types.UUID) (*SandboxTemplate, error) {
    switch provider {
    case sandbox.PROVIDER_CLAUDE_CODE:
        return getClaudeCodeTemplate(agentID), nil
    default:
        return nil, errors.Errorf("unsupported provider: %d", provider)
    }
}

// getClaudeCodeTemplate returns the Claude Code sandbox template
func getClaudeCodeTemplate(agentID types.UUID) *SandboxTemplate {
    return &SandboxTemplate{
        Provider:    sandbox.PROVIDER_CLAUDE_CODE,
        ImageConfig: modal.GetImageConfigFromTemplate("claude"),
        VolumeName:  "", // Will be auto-generated per account
        S3BucketName: "", // Optional: configure if S3 persistence needed
        S3KeyPrefix:  "", // Optional: configure if S3 persistence needed
        InitFromS3:   false,
    }
}

// BuildSandboxConfig creates a modal.SandboxConfig from a template
func (t *SandboxTemplate) BuildSandboxConfig(accountID types.UUID) *modal.SandboxConfig {
    config := &modal.SandboxConfig{
        AccountID:       accountID,
        Image:           t.ImageConfig,
        VolumeName:      t.VolumeName,
        VolumeMountPath: "/mnt/workspace",
        Workdir:         "/mnt/workspace",
    }
    
    // Add S3 config if specified
    if t.S3BucketName != "" {
        config.S3Config = &modal.S3MountConfig{
            BucketName: t.S3BucketName,
            SecretName: "s3-bucket",
            KeyPrefix:  t.S3KeyPrefix,
            MountPath:  "/mnt/s3-bucket",
            ReadOnly:   true,
        }
    }
    
    return config
}
```

### 5. Controller Updates

**File**: `internal/controllers/sandboxes/sandbox.go`

#### Request Types

```go
// CreateSandboxRequest holds request data for sandbox creation
// Frontend only specifies the provider type and agent, not configuration details
type CreateSandboxRequest struct {
    Provider sandbox.Provider `json:"provider"` // Provider type (1=Claude Code)
    AgentID  types.UUID       `json:"agent_id"` // Agent ID (optional, for future use)
}

// SyncSandboxRequest holds request data for S3 sync operations
type SyncSandboxRequest struct {
    // No parameters needed - syncs the sandbox's configured S3 bucket
}
```

#### Updated createSandbox

```go
// createSandbox creates a new sandbox using a premade template based on provider/agent
func createSandbox(_ http.ResponseWriter, req *http.Request) (*sandbox.Sandbox, int, error) {
    // Get authenticated user session
    userSession := request.GetReqSession(req)
    accountID := userSession.User.ID()

    // Parse request body
    data, err := request.GetJSONPostAs[*CreateSandboxRequest](req)
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

    // Save to database
    sandboxModel := sandbox.New()
    sandboxModel.AccountID.Set(accountID)
    sandboxModel.AgentID.Set(data.AgentID)
    sandboxModel.Provider.Set(data.Provider)
    sandboxModel.ExternalID.Set(sandboxInfo.SandboxID)
    sandboxModel.Status.Set(constants.STATUS_ACTIVE) // 0 = active/running
    sandboxModel.MetaData.Set(&sandbox.MetaData{}) // Empty metadata initially
    
    err = sandboxModel.Save(userSession.User)
    if err != nil {
        log.ErrorContext(err, req.Context())
        // Note: Sandbox was created in Modal but DB save failed
        // Consider adding cleanup logic here or async cleanup task
        return response.AdminBadRequestError[*sandbox.Sandbox](err)
    }

    log.Infof("Created sandbox %s (external_id: %s) for account %s",
        sandboxModel.ID(), sandboxInfo.SandboxID, accountID)

    return response.Success(sandboxModel)
}
```

#### Updated authDelete (soft delete with Modal termination)

```go
// authDelete terminates a sandbox and soft-deletes the database record
// Note: The auth framework already handles ownership verification
func authDelete(_ http.ResponseWriter, req *http.Request) (*sandbox.Sandbox, int, error) {
    userSession := request.GetReqSession(req)
    accountID := userSession.User.ID()
    id := chi.URLParam(req, "id")

    log.Infof("authDelete called for sandbox ID: %s", id)

    // Get sandbox (auth framework ensures ownership)
    sandboxModel, err := sandbox.Get(req.Context(), types.UUID(id))
    if err != nil {
        log.ErrorContext(err, req.Context())
        return response.AdminBadRequestError[*sandbox.Sandbox](err)
    }

    // Reconstruct SandboxInfo for Modal termination
    sandboxInfo := reconstructSandboxInfo(sandboxModel, accountID)

    // Terminate sandbox via service with S3 sync
    service := sandbox_service.NewSandboxService()
    err = service.TerminateSandbox(req.Context(), sandboxInfo, true)
    if err != nil {
        log.ErrorContext(err, req.Context())
        // Log error but continue with soft delete
        log.Warnf("Failed to terminate Modal sandbox %s: %v", sandboxModel.ExternalID.GetI(), err)
    }

    // Update status and soft delete
    sandboxModel.Status.Set(constants.STATUS_TERMINATED) // Or appropriate constant
    sandboxModel.Deleted.Set(1)
    err = sandboxModel.Save(userSession.User)
    if err != nil {
        log.ErrorContext(err, req.Context())
        return response.AdminBadRequestError[*sandbox.Sandbox](err)
    }

    log.Infof("Terminated and deleted sandbox %s (external_id: %s)",
        sandboxModel.ID(), sandboxModel.ExternalID.GetI())

    return response.Success(sandboxModel)
}
```

#### Updated syncSandbox

```go
// syncSandbox syncs the sandbox volume to S3 without terminating
func syncSandbox(_ http.ResponseWriter, req *http.Request) (*sandbox.Sandbox, int, error) {
    userSession := request.GetReqSession(req)
    accountID := userSession.User.ID()
    id := chi.URLParam(req, "id")

    log.Infof("syncSandbox called for sandbox ID: %s", id)

    // Get sandbox (auth framework ensures ownership)
    sandboxModel, err := sandbox.Get(req.Context(), types.UUID(id))
    if err != nil {
        log.ErrorContext(err, req.Context())
        return response.AdminBadRequestError[*sandbox.Sandbox](err)
    }

    // Reconstruct SandboxInfo for Modal sync
    sandboxInfo := reconstructSandboxInfo(sandboxModel, accountID)

    // Sync to S3 via service layer
    service := sandbox_service.NewSandboxService()
    stats, err := service.SyncToS3(req.Context(), sandboxInfo)
    if err != nil {
        log.ErrorContext(err, req.Context())
        return response.AdminBadRequestError[*sandbox.Sandbox](err)
    }

    log.Infof("Synced sandbox %s to S3: %d files, %d bytes, %dms",
        id, stats.FilesProcessed, stats.BytesTransferred, stats.Duration.Milliseconds())

    // Update metadata with sync stats
    metadata := sandboxModel.MetaData.GetI()
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
        log.Warnf("Failed to update metadata after sync: %v", err)
    }

    return response.Success(sandboxModel)
}
```

**File**: `internal/controllers/sandboxes/claude.go`

#### Updated streamClaude

```go
// streamClaude executes Claude Code CLI in a sandbox and streams output using SSE
func streamClaude(w http.ResponseWriter, req *http.Request) {
    userSession := request.GetReqSession(req)
    accountID := userSession.User.ID()
    id := chi.URLParam(req, "id")

    // Parse request body for prompt
    data, err := request.GetJSONPostAs[*ClaudeRequest](req)
    if err != nil {
        log.ErrorContext(err, req.Context())
        http.Error(w, "failed to parse request body", http.StatusBadRequest)
        return
    }

    // Validate prompt not empty
    if tools.Empty(data.Prompt) {
        http.Error(w, "prompt is required", http.StatusBadRequest)
        return
    }

    log.Infof("streamClaude called for sandbox ID: %s, prompt: %s", id, data.Prompt)

    // Get sandbox (verify ownership through AccountID)
    sandboxModel, err := sandbox.Get(req.Context(), types.UUID(id))
    if err != nil || sandboxModel.AccountID.GetI() != accountID {
        log.ErrorContext(err, req.Context())
        http.Error(w, "sandbox not found", http.StatusNotFound)
        return
    }

    // Reconstruct SandboxInfo from model
    sandboxInfo := reconstructSandboxInfo(sandboxModel, accountID)

    // Build ClaudeExecConfig
    claudeConfig := &modal.ClaudeExecConfig{
        Prompt:          data.Prompt,
        OutputFormat:    "stream-json",
        SkipPermissions: true,
        Verbose:         true,
    }

    // Execute Claude and stream via service layer
    service := sandbox_service.NewSandboxService()
    err = service.ExecuteClaudeStream(req.Context(), sandboxInfo, claudeConfig, w)
    if err != nil {
        log.ErrorContext(err, req.Context())
        http.Error(w, "failed to execute Claude", http.StatusInternalServerError)
        return
    }
}
```

#### Helper Function

```go
// reconstructSandboxInfo creates a modal.SandboxInfo from database model
// This is needed for operations that interact with the Modal API
func reconstructSandboxInfo(model *sandbox.Sandbox, accountID types.UUID) *modal.SandboxInfo {
    // Get template to reconstruct config
    template, _ := sandbox_service.GetSandboxTemplate(
        model.Provider.GetI(),
        model.AgentID.GetI(),
    )
    
    var config *modal.SandboxConfig
    if template != nil {
        config = template.BuildSandboxConfig(accountID)
    } else {
        // Fallback basic config if template not found
        config = &modal.SandboxConfig{
            AccountID:       accountID,
            Image:           modal.GetImageConfigFromTemplate("claude"),
            VolumeName:      "",
            VolumeMountPath: "/mnt/workspace",
            Workdir:         "/mnt/workspace",
        }
    }
    
    // Map database status to Modal status
    var modalStatus modal.SandboxStatus
    if model.Deleted.GetI() == 1 || model.Status.GetI() != constants.STATUS_ACTIVE {
        modalStatus = modal.SandboxStatusTerminated
    } else {
        modalStatus = modal.SandboxStatusRunning
    }
    
    return &modal.SandboxInfo{
        SandboxID: model.ExternalID.GetI(),
        Config:    config,
        CreatedAt: model.CreatedAt.GetI(),
        Status:    modalStatus,
        Sandbox:   nil, // Not reconstructed from DB
    }
}
```

## Error Handling

### Error Categories

1. **Database Errors** (500)
   - Connection failures
   - Query syntax errors
   - Transaction rollback failures

2. **Not Found Errors** (404)
   - Sandbox not found by ID or external_id
   - Record deleted or doesn't exist

3. **Authorization Errors** (403)
   - User attempting to access another user's sandbox
   - AccountID mismatch

4. **Validation Errors** (400)
   - Invalid provider type
   - Missing required fields
   - Invalid status transitions

5. **Modal Integration Errors** (500/502)
   - Modal API failures
   - Sandbox creation failures
   - Network timeouts

### Error Handling Patterns

```go
// Database query with proper error wrapping
sandboxModel, err := sandbox.Get(ctx, types.UUID(id))
if err != nil {
    log.ErrorContext(err, ctx)
    return response.AdminBadRequestError[*Sandbox](
        errors.Wrapf(err, "failed to retrieve sandbox: %s", id),
    )
}

// Not found check
if sandboxModel == nil {
    err := errors.Errorf("sandbox not found: %s", id)
    log.ErrorContext(err, ctx)
    return response.AdminNotFoundError[*Sandbox](err)
}

// Ownership verification
if sandboxModel.AccountID.GetI() != accountID {
    err := errors.New("unauthorized access to sandbox")
    log.ErrorContext(err, ctx)
    return response.AdminForbiddenError[*Sandbox](err)
}

// Modal integration error with cleanup consideration
sandboxInfo, err := service.CreateSandbox(ctx, accountID, config)
if err != nil {
    log.ErrorContext(err, ctx)
    return response.AdminBadRequestError[*Sandbox](
        errors.Wrap(err, "failed to create Modal sandbox"),
    )
}

// Database save error after Modal creation
err = sandboxModel.Save(user)
if err != nil {
    log.ErrorContext(err, ctx)
    // TODO: Consider async cleanup of orphaned Modal sandbox
    log.Warnf("Sandbox %s created in Modal but DB save failed", sandboxInfo.SandboxID)
    return response.AdminBadRequestError[*Sandbox](
        errors.Wrap(err, "failed to persist sandbox to database"),
    )
}
```

## Testing Strategy

### Unit Tests

**File**: `internal/models/sandbox/sandbox_test.go`

Use table-driven tests following TDD principles. Always write tests FIRST (RED → GREEN → REFACTOR).

```go
func TestSandbox_SaveWithMetaData(t *testing.T) {
    t.Run("Saves sandbox with minimal metadata", func(t *testing.T) {
        // Arrange
        obj := sandbox.New()
        obj.AccountID.Set(types.UUID("test-account-123"))
        obj.ExternalID.Set("sb-test-123")
        obj.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
        obj.Status.Set(constants.STATUS_ACTIVE)
        obj.MetaData.Set(&sandbox.MetaData{})
        
        // Act
        err := obj.Save(nil)
        
        // Assert
        assert.NoError(t, err)
        defer testtools.CleanupModel(obj)
        
        retrieved, err := sandbox.Get(context.Background(), obj.ID())
        assert.NoError(t, err)
        assert.Equal(t, "sb-test-123", retrieved.ExternalID.GetI())
        assert.Equal(t, sandbox.PROVIDER_CLAUDE_CODE, retrieved.Provider.GetI())
    })
}

func TestSandbox_FindByExternalID(t *testing.T) {
    tests := []struct {
        name      string
        setupFn   func() (*sandbox.Sandbox, types.UUID)
        searchID  string
        accountID types.UUID
        expectErr bool
    }{
        {
            name: "finds sandbox with correct account",
            setupFn: func() (*sandbox.Sandbox, types.UUID) {
                accountID := types.UUID("test-account-456")
                obj := sandbox.New()
                obj.AccountID.Set(accountID)
                obj.ExternalID.Set("sb-find-test")
                obj.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
                obj.Save(nil)
                return obj, accountID
            },
            searchID:  "sb-find-test",
            expectErr: false,
        },
        {
            name: "returns error with wrong account",
            setupFn: func() (*sandbox.Sandbox, types.UUID) {
                obj := sandbox.New()
                obj.AccountID.Set(types.UUID("different-account"))
                obj.ExternalID.Set("sb-find-test-2")
                obj.Save(nil)
                return obj, types.UUID("wrong-account")
            },
            searchID:  "sb-find-test-2",
            expectErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            obj, accountID := tt.setupFn()
            defer testtools.CleanupModel(obj)
            
            // Act
            found, err := sandbox.FindByExternalID(context.Background(), tt.searchID, accountID)
            
            // Assert
            if tt.expectErr {
                assert.Error(t, err)
                assert.Empty(t, found)
            } else {
                assert.NoError(t, err)
                assert.NEmpty(t, found)
            }
        })
    }
}

func TestMetaData_UpdateLastSync(t *testing.T) {
    t.Run("Updates sync timestamp and stats", func(t *testing.T) {
        // Arrange
        metadata := &sandbox.MetaData{}
        
        // Act
        metadata.UpdateLastSync(10, 1024, 500)
        
        // Assert
        assert.NEmpty(t, metadata.LastS3Sync)
        assert.NEmpty(t, metadata.SyncStats)
        assert.Equal(t, 10, metadata.SyncStats.FilesProcessed)
        assert.Equal(t, int64(1024), metadata.SyncStats.BytesTransferred)
        assert.Equal(t, int64(500), metadata.SyncStats.DurationMs)
    })
}
```

### Controller Tests

**File**: `internal/controllers/sandboxes/sandbox_test.go`

Use the `testing_service.TestRequest` pattern as shown in the testing skill:

```go
package sandboxes

import (
    "net/http"
    "testing"

    "github.com/griffnb/techboss-ai-go/internal/common/system_testing"
    "github.com/griffnb/techboss-ai-go/internal/models/sandbox"
    "github.com/griffnb/techboss-ai-go/internal/services/testing_service"
    "github.com/griffnb/core/lib/testtools/assert"
    "github.com/griffnb/core/lib/testtools"
)

func init() {
    system_testing.BuildSystem()
}

func TestCreateSandbox(t *testing.T) {
    skipIfNotConfigured(t)
    
    t.Run("Creates sandbox with premade template", func(t *testing.T) {
        // Arrange
        body := map[string]any{
            "data": map[string]any{
                "provider": sandbox.PROVIDER_CLAUDE_CODE,
                "agent_id": "",
            },
        }
        
        req, err := testing_service.NewPOSTRequest[*sandbox.Sandbox]("/", nil, body)
        assert.NoError(t, err)
        
        err = req.WithAccount() // Use account user for auth endpoint
        assert.NoError(t, err)
        
        // Act
        resp, errCode, err := req.Do(createSandbox)
        
        // Assert
        assert.NoError(t, err)
        assert.Equal(t, http.StatusOK, errCode)
        assert.NEmpty(t, resp)
        assert.NEmpty(t, resp.ExternalID.GetI())
        
        defer testtools.CleanupModel(resp)
        
        // Cleanup Modal sandbox
        // ... termination logic
    })
}

func TestSyncSandbox(t *testing.T) {
    skipIfNotConfigured(t)
    
    t.Run("Updates metadata with sync stats", func(t *testing.T) {
        // Arrange - create test sandbox
        sb := testing_service.Builder.BuildSandbox(t)
        defer testtools.CleanupModel(sb)
        
        req, err := testing_service.NewPOSTRequest[*sandbox.Sandbox](
            "/"+sb.ID().String()+"/sync",
            nil,
            nil,
        )
        assert.NoError(t, err)
        
        err = req.WithAccount()
        assert.NoError(t, err)
        
        // Act
        resp, errCode, err := req.Do(syncSandbox)
        
        // Assert
        assert.NoError(t, err)
        assert.Equal(t, http.StatusOK, errCode)
        assert.NEmpty(t, resp.MetaData.GetI().LastS3Sync)
    })
}
```

### Service Tests

**File**: `internal/services/sandbox_service/templates_test.go`

```go
func TestGetSandboxTemplate(t *testing.T) {
    tests := []struct {
        name      string
        provider  sandbox.Provider
        expectErr bool
    }{
        {
            name:      "returns Claude Code template",
            provider:  sandbox.PROVIDER_CLAUDE_CODE,
            expectErr: false,
        },
        {
            name:      "returns error for unsupported provider",
            provider:  sandbox.Provider(999),
            expectErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Act
            template, err := GetSandboxTemplate(tt.provider, types.UUID("test-agent"))
            
            // Assert
            if tt.expectErr {
                assert.Error(t, err)
                assert.Empty(t, template)
            } else {
                assert.NoError(t, err)
                assert.NEmpty(t, template)
                assert.Equal(t, tt.provider, template.Provider)
            }
        })
    }
}
```

### Test Coverage Requirements

- **Unit Tests**: ≥90% coverage for models and utility functions
- **Controller Tests**: Cover all endpoints with auth verification
- **Service Tests**: Cover template generation and config building
- **Edge Cases**: Non-existent IDs, ownership violations, unsupported providers
- **Error Paths**: Database failures, Modal API failures, validation errors

### Important Testing Notes

1. Always use `system_testing.BuildSystem()` in `init()`
2. Use `testing_service.NewGETRequest` / `NewPOSTRequest` / `NewPUTRequest`
3. Use `req.WithAccount()` for auth endpoints (not WithAdmin for user-facing APIs)
4. Clean up with `defer testtools.CleanupModel(obj)`
5. Extend `testing_service.Builder` if new test objects needed
6. Write tests FIRST (TDD: RED → GREEN → REFACTOR)
7. All tests must pass before committing

## Implementation Checklist

- [ ] Update `meta_data.go` with simplified MetaData structure
- [ ] Add helper method: `UpdateLastSync()`
- [ ] Add database query functions in `queries.go`: `FindByExternalID()`, `FindAllByAccount()`, `CountByAccount()`
- [ ] Create `templates.go` service with `GetSandboxTemplate()` and premade configurations
- [ ] Update `createSandbox` controller to use templates (no frontend config)
- [ ] Update `authDelete` controller to handle termination and soft-delete
- [ ] Update `syncSandbox` controller to update metadata
- [ ] Update `streamClaude` controller in `claude.go` to use external_id
- [ ] Add `reconstructSandboxInfo()` helper function
- [ ] Remove `sandboxCache` variable and all references
- [ ] Remove all Phase 2 TODO comments related to cache
- [ ] Update sandbox model to use external_id field properly
- [ ] Write unit tests for model operations and metadata helpers
- [ ] Write controller tests using testing_service pattern
- [ ] Write service tests for template generation
- [ ] Test ownership verification in all endpoints
- [ ] Update UI to load and display sandbox list by ID (not external_id)
- [ ] Add sandbox selection handlers in JavaScript
- [ ] Test UI flow: load → select → chat
- [ ] Test UI flow: load → create → chat
