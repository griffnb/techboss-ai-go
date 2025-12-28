# Sandbox Lifecycle Hooks with Conversation Integration - Design Document

## Overview

This design document outlines the architecture for implementing a conversation-centric sandbox system with flexible lifecycle hooks. The system refactors sandbox operations to be driven by conversations, with message tracking, intelligent S3 synchronization using state files, and composable lifecycle hooks.

### Key Architectural Principles

**File Organization Philosophy:**
- **Single Responsibility**: Each file contains functions related to a specific concern
- **Cohesive Grouping**: Related functionality is grouped together but files remain focused
- **No God Files**: Avoid large files with mixed responsibilities - prefer multiple smaller focused files
- **Clear Naming**: File names clearly indicate their purpose and scope

**Design Goals:**
- Move from sandbox-centric to conversation-centric architecture
- Track all messages and token usage within conversations
- Use state files (`.sandbox-state`) for reliable S3 synchronization
- Provide composable hooks: `OnColdStart`, `OnMessage`, `OnStreamFinish`, `OnTerminate`
- Maintain backward compatibility while migrating endpoints
- Keep files small, focused, and testable

## Architecture

### High-Level Flow

```
User Request → Conversation Controller → Conversation Service → Sandbox Service → Modal Integration
                       ↓                        ↓                      ↓
                  Message Storage         Hook Execution        State File Sync
                   (DynamoDB)            (Template Hooks)         (S3 + Local)
                       ↓                        ↓
                Conversation Stats       Lifecycle Events
                 (PostgreSQL)           (Cold Start, Stream, etc.)
```

### Component Layers

1. **Models Layer** (`internal/models/`)
   - `conversation/`: Conversation model with stats (PostgreSQL)
   - `message/`: Message model with tool calls (DynamoDB)
   - `sandbox/`: Sandbox model with metadata

2. **Integration Layer** (`internal/integrations/modal/`)
   - Claude execution and streaming
   - S3 sync operations with state file management
   - Low-level Modal API interactions

3. **Service Layer** (`internal/services/`)
   - `conversation_service/`: Conversation business logic and message management
   - `sandbox_service/`: Sandbox lifecycle with hook orchestration
     - `state_files/`: State file parsing, comparison, and management (subfolder)
     - `lifecycle/`: Lifecycle hooks system (subfolder)

4. **Controller Layer** (`internal/controllers/`)
   - `conversations/`: Conversation CRUD and streaming endpoints
   - `sandboxes/`: Sandbox management (existing, to be integrated)

## Components and Interfaces

### Phase 1: Conversation Model Extensions

#### 1.1. Conversation Stats Structure (`internal/models/conversation/stats.go`)

**Purpose**: Extend conversation stats to track detailed token usage.

**File Organization**: Already exists, just needs updates.

```go
package conversation

type ConversationStats struct {
    MessagesExchanged  int   `json:"messages_exchanged"`
    TotalInputTokens   int64 `json:"total_input_tokens"`   // NEW
    TotalOutputTokens  int64 `json:"total_output_tokens"`  // NEW
    TotalCacheTokens   int64 `json:"total_cache_tokens"`   // NEW
    TotalTokensUsed    int64 `json:"total_tokens_used"`    // DEPRECATED - kept for compatibility
}
```

**Key Methods to Add**:
- `AddTokenUsage(inputTokens, outputTokens, cacheTokens int64)` - Accumulates token counts
- `IncrementMessages()` - Increments message counter

#### 1.2. Message Model Extensions (`internal/models/message/message.go`)

**Purpose**: Extend message model to support tool calls within the same message.

**Updated Structure**:
```go
type Message struct {
    Key            string     `json:"key"`            // Auto-generated unique ID
    ConversationID types.UUID `json:"conversation_id"` // Links to conversation
    Body           string     `json:"body"`            // Message content
    Role           int64      `json:"role"`            // 1=user, 2=assistant, 3=tool
    Timestamp      int64      `json:"timestamp"`       // Unix timestamp
    Tokens         int64      `json:"tokens"`          // Token usage for this message
    ToolCalls      []ToolCall `json:"tool_calls"`      // NEW - Tool calls within message
}

// ToolCall represents a tool invocation within a message
type ToolCall struct {
    ID       string                 `json:"id"`        // Tool call ID
    Type     string                 `json:"type"`      // e.g., "function", "bash_command"
    Name     string                 `json:"name"`      // Tool/function name
    Input    map[string]interface{} `json:"input"`     // Tool input parameters
    Output   string                 `json:"output"`    // Tool execution result
    Status   string                 `json:"status"`    // "pending", "success", "error"
    Error    string                 `json:"error"`     // Error message if failed
}
```

**Why Tool Calls in Same Message**: Keeps conversation flow intact - a single assistant message can contain multiple tool invocations and their results, maintaining logical grouping.

---

### Phase 2: State File Management (Sandbox Service Subfolder)

**Design Principle**: State files are only used by sandbox service, so they live as a subfolder within it. This keeps related functionality together without creating unnecessary top-level services.

#### 2.1. State File Types (`internal/services/sandbox_service/state_files/types.go`)

**Purpose**: Define data structures for state file format.

**Why Separate File**: Type definitions are foundational and referenced by multiple files.

```go
package state_files

import "time"

// StateFile represents the .sandbox-state file format
type StateFile struct {
    Version      string      `json:"version"`       // Schema version (e.g., "1.0")
    LastSyncedAt int64       `json:"last_synced_at"` // Unix timestamp
    Files        []FileEntry `json:"files"`          // Array of tracked files
}

// FileEntry represents a single file in the state
type FileEntry struct {
    Path       string `json:"path"`        // Relative path from volume root
    Checksum   string `json:"checksum"`    // MD5 hash (matches S3 ETag)
    Size       int64  `json:"size"`        // File size in bytes
    ModifiedAt int64  `json:"modified_at"` // Unix timestamp
}

// StateDiff represents differences between two state files
type StateDiff struct {
    FilesToDownload []FileEntry // Files to download from S3
    FilesToDelete   []string    // Local file paths to delete
    FilesToSkip     []FileEntry // Files that match (no action needed)
}

// SyncStats tracks sync operation results
type SyncStats struct {
    FilesDownloaded  int           // Files downloaded from S3
    FilesDeleted     int           // Local files deleted
    FilesSkipped     int           // Files unchanged (skipped)
    BytesTransferred int64         // Total bytes transferred
    Duration         time.Duration // Operation duration
    Errors           []error       // Non-fatal errors
}
```

#### 2.2. State File Reader (`internal/services/sandbox_service/state_files/reader.go`)

**Purpose**: Read and parse state files from disk or S3.

**Why Separate File**: Reading/parsing is a distinct operation from writing.

```go
package state_files

import (
    "context"
    "encoding/json"
    "github.com/griffnb/techboss-ai-go/internal/integrations/modal"
)

// ReadLocalStateFile reads .sandbox-state from local volume
func ReadLocalStateFile(
    ctx context.Context,
    sandboxInfo *modal.SandboxInfo,
    volumePath string,
) (*StateFile, error) {
    // Build path: {volumePath}/.sandbox-state
    // Execute cat command in sandbox to read file
    // Parse JSON into StateFile struct
    // Return nil if file doesn't exist (treat as empty state)
    // Return error if file exists but is corrupted
}

// ReadS3StateFile reads .sandbox-state from S3 mount
func ReadS3StateFile(
    ctx context.Context,
    sandboxInfo *modal.SandboxInfo,
    s3MountPath string,
) (*StateFile, error) {
    // Similar to ReadLocalStateFile but reads from S3 mount path
    // S3 mount is already available in sandbox via CloudBucketMount
}

// ParseStateFile parses JSON bytes into StateFile struct
func ParseStateFile(data []byte) (*StateFile, error) {
    // Unmarshal JSON
    // Validate version compatibility
    // Return error if version is too new (forward compatibility issue)
    // Handle backward compatibility for older versions
}
```

#### 2.3. State File Writer (`internal/services/sandbox_service/state_files/writer.go`)

**Purpose**: Write state files atomically to prevent corruption.

**Why Separate File**: Writing is distinct from reading and has atomic update requirements.

```go
package state_files

import (
    "context"
    "encoding/json"
    "github.com/griffnb/techboss-ai-go/internal/integrations/modal"
    "time"
)

// WriteLocalStateFile writes .sandbox-state to local volume atomically
func WriteLocalStateFile(
    ctx context.Context,
    sandboxInfo *modal.SandboxInfo,
    volumePath string,
    stateFile *StateFile,
) error {
    // Update LastSyncedAt to current time
    stateFile.LastSyncedAt = time.Now().Unix()
    
    // Marshal to JSON
    // Write to temporary file: .sandbox-state.tmp
    // Rename to .sandbox-state (atomic operation)
    // This prevents corruption if process is interrupted
}

// WriteS3StateFile writes .sandbox-state to S3
func WriteS3StateFile(
    ctx context.Context,
    sandboxInfo *modal.SandboxInfo,
    s3MountPath string,
    stateFile *StateFile,
) error {
    // Similar atomic write pattern for S3
    // S3 mounts support atomic renames
}

// GenerateStateFile scans a directory and generates StateFile
func GenerateStateFile(
    ctx context.Context,
    sandboxInfo *modal.SandboxInfo,
    directoryPath string,
) (*StateFile, error) {
    // Execute find command to list all files
    // For each file, calculate MD5 checksum
    // Get file size and modification time
    // Build FileEntry array
    // Return StateFile with current timestamp
}
```

#### 2.4. State File Comparator (`internal/services/sandbox_service/state_files/comparator.go`)

**Purpose**: Compare two state files to determine sync actions.

**Why Separate File**: Comparison logic is complex and deserves its own focused file.

```go
package state_files

// CompareStateFiles compares local and S3 state to determine sync actions
func CompareStateFiles(
    localState *StateFile,
    s3State *StateFile,
) *StateDiff {
    // Create maps for O(1) lookups
    localFiles := make(map[string]FileEntry)
    s3Files := make(map[string]FileEntry)
    
    // Build file maps by path
    for _, f := range localState.Files {
        localFiles[f.Path] = f
    }
    for _, f := range s3State.Files {
        s3Files[f.Path] = f
    }
    
    diff := &StateDiff{
        FilesToDownload: []FileEntry{},
        FilesToDelete:   []string{},
        FilesToSkip:     []FileEntry{},
    }
    
    // Files in S3 but not local OR with different checksums → Download
    for path, s3File := range s3Files {
        if localFile, exists := localFiles[path]; exists {
            if localFile.Checksum == s3File.Checksum {
                diff.FilesToSkip = append(diff.FilesToSkip, s3File)
            } else {
                diff.FilesToDownload = append(diff.FilesToDownload, s3File)
            }
        } else {
            diff.FilesToDownload = append(diff.FilesToDownload, s3File)
        }
    }
    
    // Files in local but not S3 → Delete
    for path := range localFiles {
        if _, exists := s3Files[path]; !exists {
            diff.FilesToDelete = append(diff.FilesToDelete, path)
        }
    }
    
    return diff
}

// CheckIfStale determines if state file is stale based on threshold
func CheckIfStale(stateFile *StateFile, thresholdSeconds int64) bool {
    if stateFile == nil || stateFile.LastSyncedAt == 0 {
        return true
    }
    
    age := time.Now().Unix() - stateFile.LastSyncedAt
    return age > thresholdSeconds
}
```

#### 2.5. State File Tests (`internal/services/sandbox_service/state_files/state_files_test.go`)

**Purpose**: Comprehensive tests for all state file operations.

**Test Coverage**:
- Reading state files (success, missing, corrupted)
- Writing state files (atomic writes, permissions)
- Comparing state files (all diff scenarios)
- Generating state files from directories
- Staleness checking

---

### Phase 3: Enhanced Modal Integration - Storage

**Design Principle**: Modal integration is already organized by concern (claude.go, storage.go, sandbox.go). We extend storage.go with state-file-aware sync operations.

#### 3.1. Storage Types (`internal/integrations/modal/storage.go` - ADD to existing)

**Purpose**: Update SyncStats to include new metrics.

**Why Same File**: Keeps all storage-related types together.

```go
// SyncStats (UPDATED - add new fields)
type SyncStats struct {
    FilesDownloaded  int           // NEW - replaces FilesProcessed
    FilesDeleted     int           // NEW
    FilesSkipped     int           // NEW
    BytesTransferred int64         // Existing
    Duration         time.Duration // Existing
    Errors           []error       // Existing
}
```

#### 3.2. State-Based Sync Functions (`internal/integrations/modal/storage_state.go` - NEW)

**Purpose**: State-file-based sync operations (separate from legacy sync).

**Why New File**: New sync strategy is significantly different from existing cp-based sync. Keeping separate maintains backward compatibility and clarity.

```go
package modal

import (
    "context"
    "github.com/griffnb/techboss-ai-go/internal/services/sandbox_service/state_files"
)

// InitVolumeFromS3WithState performs intelligent sync using state files
func (c *APIClient) InitVolumeFromS3WithState(
    ctx context.Context,
    sandboxInfo *SandboxInfo,
) (*SyncStats, error) {
    // 1. Read local .sandbox-state (or treat as empty if missing)
    localState, err := state_files.ReadLocalStateFile(ctx, sandboxInfo, sandboxInfo.Config.VolumeMountPath)
    
    // 2. Read S3 .sandbox-state (or generate if missing)
    s3State, err := state_files.ReadS3StateFile(ctx, sandboxInfo, sandboxInfo.Config.S3Config.MountPath)
    if err != nil || s3State == nil {
        // Generate state from S3 directory scan
        s3State, err = state_files.GenerateStateFile(ctx, sandboxInfo, sandboxInfo.Config.S3Config.MountPath)
    }
    
    // 3. Compare states to determine actions
    diff := state_files.CompareStateFiles(localState, s3State)
    
    // 4. Execute sync actions
    stats, err := c.executeSyncActions(ctx, sandboxInfo, diff)
    
    // 5. Update local .sandbox-state
    err = state_files.WriteLocalStateFile(ctx, sandboxInfo, sandboxInfo.Config.VolumeMountPath, s3State)
    
    return stats, nil
}

// SyncVolumeToS3WithState syncs volume to S3 with timestamp versioning
func (c *APIClient) SyncVolumeToS3WithState(
    ctx context.Context,
    sandboxInfo *SandboxInfo,
) (*SyncStats, error) {
    // 1. Generate current state from local volume
    localState, err := state_files.GenerateStateFile(ctx, sandboxInfo, sandboxInfo.Config.VolumeMountPath)
    
    // 2. Use AWS CLI sync to upload to timestamped path
    timestamp := time.Now().Unix()
    s3Path := fmt.Sprintf("s3://%s/docs/%s/%d/", 
        sandboxInfo.Config.S3Config.BucketName,
        sandboxInfo.Config.AccountID,
        timestamp)
    
    // 3. Execute AWS CLI sync command
    stats, err := c.executeAWSSync(ctx, sandboxInfo, sandboxInfo.Config.VolumeMountPath, s3Path)
    
    // 4. Write .sandbox-state to both local and S3
    err = state_files.WriteLocalStateFile(ctx, sandboxInfo, sandboxInfo.Config.VolumeMountPath, localState)
    err = state_files.WriteS3StateFile(ctx, sandboxInfo, sandboxInfo.Config.S3Config.MountPath, localState)
    
    return stats, nil
}

// executeSyncActions performs download and delete operations based on diff
func (c *APIClient) executeSyncActions(
    ctx context.Context,
    sandboxInfo *SandboxInfo,
    diff *state_files.StateDiff,
) (*SyncStats, error) {
    // Build commands to download files from S3 to local
    // Build commands to delete local files
    // Execute commands in sandbox
    // Track stats and return
}

// executeAWSSync runs AWS CLI sync command
func (c *APIClient) executeAWSSync(
    ctx context.Context,
    sandboxInfo *SandboxInfo,
    sourcePath string,
    s3Path string,
) (*SyncStats, error) {
    // Similar to existing SyncVolumeToS3 but returns enhanced stats
}
```

#### 3.3. Storage Tests (`internal/integrations/modal/storage_state_test.go` - NEW)

**Purpose**: Integration tests for state-based sync.

**Why Separate File**: Keeps test files organized by functionality being tested.

---

### Phase 4: Enhanced Modal Integration - Token Tracking

**Design Principle**: Token tracking is part of Claude streaming, so it belongs in claude.go. Claude returns a final summary event at the end of streaming with complete token usage.

#### 4.1. Claude Token Tracking (`internal/integrations/modal/claude.go` - ADD to existing)

**Purpose**: Parse final token summary from Claude stream.

**Why Same File**: Token tracking is intrinsically tied to Claude execution.

**How Claude Returns Tokens**: Claude CLI outputs a final summary event in JSON format at the end of the stream containing the complete token usage. We parse this single event rather than accumulating throughout the stream.

```go
// ClaudeProcess (UPDATED - add token fields)
type ClaudeProcess struct {
    Process        *modal.ContainerProcess
    Config         *ClaudeExecConfig
    StartedAt      time.Time
    InputTokens    int64 // NEW - Set from final summary
    OutputTokens   int64 // NEW - Set from final summary
    CacheTokens    int64 // NEW - Set from final summary
}

// TokenUsage represents parsed token usage from Claude response
type TokenUsage struct {
    InputTokens  int64
    OutputTokens int64
    CacheTokens  int64
}

// StreamClaudeOutput (UPDATED - parse final token summary)
func (c *APIClient) StreamClaudeOutput(
    ctx context.Context,
    claudeProcess *ClaudeProcess,
    responseWriter http.ResponseWriter,
) error {
    // ... existing SSE setup ...
    
    for scanner.Scan() {
        line := scanner.Text()
        
        // Parse final summary event for token usage
        // Claude sends a summary event at the end with complete token counts
        if isFinalSummary(line) {
            tokens := parseTokenSummary(line)
            if tokens != nil {
                claudeProcess.InputTokens = tokens.InputTokens
                claudeProcess.OutputTokens = tokens.OutputTokens
                claudeProcess.CacheTokens = tokens.CacheTokens
            }
        }
        
        // ... existing streaming logic ...
    }
    
    return nil
}

// isFinalSummary checks if the line is Claude's final summary event
func isFinalSummary(line string) bool {
    // Check for summary event marker in JSON
    // Claude typically sends event type "summary" or "usage"
}
 (Sandbox Service Subfolder)

**Design Principle**: Lifecycle hooks are only used by sandbox service, so they live as a subfolder within it. This keeps the hook system tightly coupled with the service that uses it.

#### 5.1. Hook Types (`internal/services/sandbox_service/lifecycleokens
    // Return TokenUsage struct if found, nil otherwise
}
```

---

### Phase 5: LifeHook System

**Design Principle**: Hooks are a new concept, so they get their own dedicated service with multiple focused files.

#### 5.1. Hook Types (`internal/services/lifecycle_hooks/types.go`)

**Purpose**: Define hook function signatures and hook context.

**Why Separate File**: Type definitions are foundational.

```go
package lifecycle_hooks

import (
    "context"
    "github.com/griffnb/core/lib/types"
    "github.com/griffnb/techboss-ai-go/internal/integrations/modal"
    "github.com/griffnb/techboss-ai-go/internal/models/message"
)

// HookFunc defines the signature for lifecycle hooks
type HookFunc func(ctx context.Context, hookData *HookData) error

// HookData contains context passed to lifecycle hooks
type HookData struct {
    ConversationID types.UUID
    SandboxInfo    *modal.SandboxInfo
    Message        *message.Message // For OnMessage hook
    TokenUsage     *TokenUsage      // For OnStreamFinish hook
}

// TokenUsage from streaming
type TokenUsage struct {
    InputTokens  int64
    OutputTokens int64
    CacheTokens  int64
}sandbox_service/lifecycle/executor.go`)

**Purpose**: Execute hooks with logging - hooks decide criticality by returning or swallowing errors.

**Why Separate File**: Execution logic is separate from hook definitions.

**Design Philosophy**: Single execution function. The hook implementation decides whether to return an error (critical failure) or swallow it and return nil (non-critical). This puts the control in the hands of the hook author.

```go
package lifecycle

import (
    "context"
    "github.com/griffnb/core/lib/log"
    "time"
)

// ExecuteHook runs a lifecycle hook with logging
// The hook itself determines criticality:
// - Returning an error = critical failure, caller should handle
// - Returning nil = success or non-critical failure (hook swallowed error)
func ExecuteHook(
    ctx context.Context,
    hookName string,
    hook HookFunc,
    hookData *HookData,
) error {
    if hook == nil {
        return nil // No hook registered, skip
    }
    
    startTime := time.Now()
    log.Infof("Executing lifecycle hook: %s for conversation %s", hookName, hookData.ConversationID)
    
    err := hook(ctx, hookData)
    
    duration := time.Since(startTime)
    if err != nil {
        log.Errorf("Lifecycle hook %s failed after %v: %v", hookName, duration, err)
        return err
    }
    
    log.Infof("Lifecycle hook %s completed successfully in %v", hookName, duration)
    return nil
// ExecuteHookNonCritical runs a hook but logs errors without propagating
func ExecuteHookNonCritical(
    ctx context.Context,
    hookName string,
    hook HookFunc,
    hookData *HookData,
) {
    err := ExecuteHook(ctx, hookName, hook, hookData)
    if err != nil {
        log.Errorf("Non-critical hook %s failed but continuing: %v", hookName, err)
    }
}
```

#### 5.3. Default Hooks (`internal/services/lifecycle_hooks/defaults.go`)

**Purpose**: Provide default implementations for common hooks.

**Why Separate File**: Default implementations can be reused/composed by providers.

```go
package lifecycle_hooks
sandbox_service/lifecycle/defaults.go`)

**Purpose**: Provide default implementations for common hooks.

**Why Separate File**: Default implementations can be reused/composed by providers.

**Error Handling Strategy**:
- **DefaultOnColdStart**: Returns errors (critical - must succeed)
- **DefaultOnMessage**: Swallows errors, logs and returns nil (non-critical)
- **DefaultOnStreamFinish**: Swallows errors, logs and returns nil (non-critical)
- **DefaultOnTerminate**: Swallows errors, logs and returns nil (non-critical)

```go
package lifecycle
// DefaultOnColdStart performs S3 sync if configured
func DefaultOnColdStart(ctx context.Context, hookData *HookData) error {
    if hookData.SandboxInfo.Config.S3Config == nil { (CRITICAL - returns errors)
func DefaultOnColdStart(ctx context.Context, hookData *HookData) error {
    if hookData.SandboxInfo.Config.S3Config == nil {
        return nil // No S3 configured, skip
    }
    
    // Perform state-based sync from S3
    // This is critical - if sync fails, sandbox should not be created
    client := modal.Client()
    _, err := client.InitVolumeFromS3WithState(ctx, hookData.SandboxInfo)
    return err // Return error to caller (critical failure)
}

// DefaultOnMessage saves message to DynamoDB (NON-CRITICAL - swallows errors)
func DefaultOnMessage(ctx context.Context, hookData *HookData) error {
    if hookData.Message == nil {
        return nil
    }
    
    // Save message to DynamoDB
    err := hookData.Message.Save(ctx)
    if err != nil {
        log.Errorf("Failed to save message but continuing: %v", err)
        return nil // Swallow error - message save is non-critical
    }
    
    // Increment message counter in conversation stats
    conv, err := conversation.Get(ctx, hookData.ConversationID)
    if err != nil {
        log.Errorf("Failed to update conversation stats but continuing: %v", err)
        return nil // Swallow error
    }
    
    stats := conv.Stats.Get()
    stats.MessagesExchanged++
    conv.Stats.Set(stats)
    
    err = conv.Save(ctx)
    if err != nil {
        log.Errorf("Failed to save conversation stats but continuing: %v", err)
        return nil // Swallow error
    }
    
    return nil
}

// DefaultOnStreamFinish syncs to S3 and updates conversation stats (NON-CRITICAL - swallows errors)
func DefaultOnStreamFinish(ctx context.Context, hookData *HookData) error {
    // Sync to S3 if configured
    if hookData.SandboxInfo.Config.S3Config != nil {
        client := modal.Client()
        _, err := client.SyncVolumeToS3WithState(ctx, hookData.SandboxInfo)
        if err != nil {
            log.Errorf("Failed to sync to S3 but continuing: %v", err)
            // Continue to update stats even if sync fails
        }
    }
    
    // Update conversation stats with token usage
    if hookData.TokenUsage != nil {
        conv, err := conversation.Get(ctx, hookData.ConversationID)
        if err != nil {
            log.Errorf("Failed to get conversation for stats update but continuing: %v", err)
            return nil // Swallow error
        }
        
        stats := conv.Stats.Get()
        stats.TotalInputTokens += hookData.TokenUsage.InputTokens
        stats.TotalOutputTokens += hookData.TokenUsage.OutputTokens
        stats.TotalCacheTokens += hookData.TokenUsage.CacheTokens
        conv.Stats.Set(stats)
        
        err = conv.Save(ctx)
        if err != nil {
            log.Errorf("Failed to save conversation stats but continuing: %v", err)
            return nil // Swallow error
        }
    }
    
    return nil
}

// DefaultOnTerminate performs cleanup (NON-CRITICAL - swallows errors)
func DefaultOnTerminate(ctx context.Context, hookData *HookData) error {
    // Default: no special cleanup needed
    // If cleanup is added, swallow errors and log them
```

---

### Phase 6: Sandbox Service Integration

**Design Principle**: Sandbox service already has clear separation (templates.go, reconstruct.go, sandbox_service.go). Add new lifecycle coordination file.

#### 6.1. Template Hook Registration (`internal/services/sandbox_service/templates.go` - UPDATE)
.LifecycleHooks // NEW
}

// getClaudeCodeTemplate (UPDATED - register hooks)
func getClaudeCodeTemplate(_ types.UUID) *SandboxTemplate {
    return &SandboxTemplate{
        Provider:     sandbox.PROVIDER_CLAUDE_CODE,
        ImageConfig:  modal.GetImageConfigFromTemplate("claude"),
        VolumeName:   "",
        S3BucketName: "",
        S3KeyPrefix:  "",
        InitFromS3:   true,
        Hooks: &lifecycle.LifecycleHooks{
            OnColdStart:    lifecycle.DefaultOnColdStart,
            OnMessage:      lifecycle.DefaultOnMessage,
            OnStreamFinish: lifecycle.DefaultOnStreamFinish,
            OnTerminate:    lifecycle
// getClaudeCodeTemplate (UPDATED - register hooks)
func getClaudeCodeTemplate(_ types.UUID) *SandboxTemplate {
    return &SandboxTemplate{
        Provider:     sandbox.PROVIDER_CLAUDE_CODE,
        ImageConfig:  modal.GetImageConfigFromTemplate("claude"),
        VolumeName:   "",
        S3BucketName: "",
        S3KeyPrefix:  "",
        InitFromS3:   true,
        Hooks: &lifecycle_hooks.LifecycleHooks{
            OnColdStart:    lifecycle_hooks.DefaultOnColdStart,
            OnMessage:      lifecycle_hooks.DefaultOnMessage,
            OnStreamFinish: lifecycle_hooks.DefaultOnStreamFinish,
            OnTerminate:    lifecycle_hooks.DefaultOnTerminate,
        },
    }
}
```
sandbox_service/lifecycle"
)

// ExecuteColdStartHook runs OnColdStart hook for sandbox initialization
// Returns error since OnColdStart is critical - caller must handle failure
func (s *SandboxService) ExecuteColdStartHook(
    ctx context.Context,
    conversationID types.UUID,
    sandboxInfo *modal.SandboxInfo,
    template *SandboxTemplate,
) error {
    if template.Hooks == nil || template.Hooks.OnColdStart == nil {
        return nil
    }
    
    hookData := &lifecycle.HookData{
        ConversationID: conversationID,
        SandboxInfo:    sandboxInfo,
    }
    
    // Hook determines criticality - if it returns error, we propagate it
    return lifecycle.ExecuteHook(ctx, "OnColdStart", template.Hooks.OnColdStart, hookData)
}

// ExecuteMessageHook runs OnMessage hook
// Ignores return value - hook swallows its own errors if non-critical
func (s *SandboxService) ExecuteMessageHook(
    ctx context.Context,
    conversationID types.UUID,
    sandboxInfo *modal.SandboxInfo,
    template *SandboxTemplate,
    message *message.Message,
) {
    if template.Hooks == nil || template.Hooks.OnMessage == nil {
        return
    }
    
    hookData := &lifecycle.HookData{
        ConversationID: conversationID,
        SandboxInfo:    sandboxInfo,
        Message:        message,
    }
    
    // Hook swallows errors if non-critical, returns nil
    _ = lifecycle.ExecuteHook(ctx, "OnMessage", template.Hooks.OnMessage, hookData)
}

// ExecuteStreamFinishHook runs OnStreamFinish hook
// Ignores return value - hook swallows its own errors if non-critical
func (s *SandboxService) ExecuteStreamFinishHook(
    ctx context.Context,
    conversationID types.UUID,
    sandboxInfo *modal.SandboxInfo,
    template *SandboxTemplate,
    tokenUsage *lifecycle.TokenUsage,
) {
    if template.Hooks == nil || template.Hooks.OnStreamFinish == nil {
        return
    }
    
    hookData := &lifecycle.HookData{
        ConversationID: conversationID,
        SandboxInfo:    sandboxInfo,
        TokenUsage:     tokenUsage,
    }
    
    // Hook swallows errors if non-critical, returns nil
    _ = lifecycle.ExecuteHook(ctx, "OnStreamFinish", template.Hooks.OnStreamFinish, hookData)
}

// ExecuteTerminateHook runs OnTerminate hook
// Returns error if hook determines it's critical
func (s *SandboxService) ExecuteTerminateHook(
    ctx context.Context,
    conversationID types.UUID,
    sandboxInfo *modal.SandboxInfo,
    template *SandboxTemplate,
) error {
    if template.Hooks == nil || template.Hooks.OnTerminate == nil {
        return nil
    }
    
    hookData := &lifecycle.HookData{
        ConversationID: conversationID,
        SandboxInfo:    sandboxInfo,
    }
    
    // Hook determines criticality - if it returns error, we propagate it
    return lifecycle
    conversationID types.UUID,
    sandboxInfo *modal.SandboxInfo,
    template *SandboxTemplate,
) error {
    if template.Hooks == nil || template.Hooks.OnTerminate == nil {
        return nil
    }
    
    hookData := &lifecycle_hooks.HookData{
        ConversationID: conversationID,
        SandboxInfo:    sandboxInfo,
    }
    
    return lifecycle_hooks.ExecuteHook(ctx, "OnTerminate", template.Hooks.OnTerminate, hookData)
}
```

---

### Phase 7: Conversation Service Layer

**Design Principle**: Create new conversation service to handle conversation-specific business logic. Keep files focused by function.

#### 7.1. Conversation Service Core (`internal/services/conversation_service/conversation_service.go`)

**Purpose**: Main service struct and basic CRUD operations.

**Why New File**: New service layer for conversation-specific logic.

```go
package conversation_service

import (
    "context"
    "github.com/griffnb/core/lib/types"
    "github.com/griffnb/techboss-ai-go/internal/models/conversation"
    "github.com/griffnb/techboss-ai-go/internal/models/sandbox"
    "github.com/griffnb/techboss-ai-go/internal/services/sandbox_service"
)

// ConversationService handles conversation business logic
type ConversationService struct {
    sandboxService *sandbox_service.SandboxService
}

// NewConversationService creates a new conversation service
func NewConversationService() *ConversationService {
    return &ConversationService{
        sandboxService: sandbox_service.NewSandboxService(),
    }
}

// GetOrCreateConversation retrieves existing or creates new conversation
func (s *ConversationService) GetOrCreateConversation(
    ctx context.Context,
    conversationID types.UUID,
    accountID types.UUID,
    agentID types.UUID,
) (*conversation.Conversation, error) {
    // Try to get existing conversation
    conv, err := conversation.Get(ctx, conversationID)
    if err == nil && conv != nil {
        return conv, nil
    }
    
    // Create new conversation
    conv = conversation.New()
    conv.ID.Set(conversationID)
    conv.AccountID.Set(accountID)
    conv.AgentID.Set(agentID)
    conv.Stats.Set(&conversation.ConversationStats{
        MessagesExchanged: 0,
        TotalInputTokens:  0,
        TotalOutputTokens: 0,
        TotalCacheTokens:  0,
    })
    
    err = conv.Save(ctx)
    return conv, err
}

// EnsureSandbox ensures conversation has an active sandbox
func (s *ConversationService) EnsureSandbox(
    ctx context.Context,
    conv *conversation.Conversation,
    provider sandbox.Provider,
) (*modal.SandboxInfo, *sandbox_service.SandboxTemplate, error) {
    // If conversation already has sandbox, reconstruct info
    if !conv.SandboxID.IsEmpty() {
        sandboxModel, err := sandbox.Get(ctx, conv.SandboxID.Get())
        if err == nil && sandboxModel != nil {
            sandboxInfo := sandbox_service.ReconstructSandboxInfo(sandboxModel, conv.AccountID.Get())
            template, _ := sandbox_service.GetSandboxTemplate(provider, conv.AgentID.Get())
            return sandboxInfo, template, nil
        }
    }
    
    // Create new sandbox
    template, err := sandbox_service.GetSandboxTemplate(provider, conv.AgentID.Get())
    if err != nil {
        return nil, nil, err
    }
    
    config := template.BuildSandboxConfig(conv.AccountID.Get())
    sandboxInfo, err := s.sandboxService.CreateSandbox(ctx, conv.AccountID.Get(), config)
    if err != nil {
        return nil, nil, err
    }
    
    // Run OnColdStart hook (CRITICAL - fails if hook fails)
    err = s.sandboxService.ExecuteColdStartHook(ctx, conv.ID.Get(), sandboxInfo, template)
    if err != nil {
        // Clean up sandbox on failure
        _ = s.sandboxService.TerminateSandbox(ctx, sandboxInfo, false)
        return nil, nil, err
    }
    
    // Save sandbox to database
    sandboxModel := sandbox.New()
    sandboxModel.AccountID.Set(conv.AccountID.Get())
    sandboxModel.AgentID.Set(conv.AgentID.Get())
    sandboxModel.ExternalID.Set(sandboxInfo.SandboxID)
    sandboxModel.Provider.Set(provider)
    err = sandboxModel.Save(ctx)
    if err != nil {
        return nil, nil, err
    }
    
    // Link sandbox to conversation
    conv.SandboxID.Set(sandboxModel.ID.Get())
    err = conv.Save(ctx)
    
    return sandboxInfo, template, err
}
```

#### 7.2. Message Management (`internal/services/conversation_service/messages.go`)

**Purpose**: Handle message creation and hook execution.

**Why Separate File**: Message management is a distinct concern from conversation CRUD.

```go
package conversation_service

import (
    "context"
    "github.com/griffnb/core/lib/types"
    "github.com/griffnb/techboss-ai-go/internal/integrations/modal"
    "github.com/griffnb/techboss-ai-go/internal/models/message"
    "github.com/griffnb/techboss-ai-go/internal/services/sandbox_service"
)

const (
    ROLE_USER      = 1
    ROLE_ASSISTANT = 2
)

// SaveUserMessage creates a user message and executes OnMessage hook
func (s *ConversationService) SaveUserMessage(
    ctx context.Context,
    conversationID types.UUID,
    sandboxInfo *modal.SandboxInfo,
    template *sandbox_service.SandboxTemplate,
    prompt string,
) (*message.Message, error) {
    msg := &message.Message{
        ConversationID: conversationID,
        Body:           prompt,
        Role:           ROLE_USER,
        Tokens:         0, // User messages don't have tokens
    }
    
    // Execute OnMessage hook (non-critical)
    s.sandboxService.ExecuteMessageHook(ctx, conversationID, sandboxInfo, template, msg)
    
    return msg, nil
}

// SaveAssistantMessage creates an assistant message with token usage
func (s *ConversationService) SaveAssistantMessage(
    ctx context.Context,
    conversationID types.UUID,
    sandboxInfo *modal.SandboxInfo,
    template *sandbox_service.SandboxTemplate,
    response string,
    tokens int64,
) (*message.Message, error) {
    msg := &message.Message{
        ConversationID: conversationID,
        Body:           response,
        Role:           ROLE_ASSISTANT,
        Tokens:         tokens,
    }
    
    // Execute OnMessage hook (non-critical)
    s.sandboxService.ExecuteMessageHook(ctx, conversationID, sandboxInfo, template, msg)
    
    return msg, nil
}
```

#### 7.3. Streaming Coordination (`internal/services/conversation_service/streaming.go`)

**Purpose**: Coordinate streaming with hooks and token tracking.

**Why Separate File**: Streaming coordination is complex enough to warrant its own file.

```go
package conversation_service

import (
    "context"
    "net/http"
    "github.com/griffnb/core/lib/types"
    "github.com/griffnb/techboss-ai-go/internal/integrations/modal"
    "github.com/griffnb/techboss-ai-go/internal/services/lifecycle_hooks"
    "github.com/griffnb/techboss-ai-go/internal/services/sandbox_service"
)
parsed from final summary event)
    // TODO: Refactor ExecuteClaudeStream to return ClaudeProcess so we can access tokens
    tokenUsage := &lifecycle.TokenUsage{
        InputTokens:  0, // Get from ClaudeProcess after refactor
    conversationID types.UUID,
    sandboxInfo *modal.SandboxInfo,
    template *sandbox_service.SandboxTemplate,
    prompt string,
    responseWriter http.ResponseWriter,
) error {
    // 1. Save user message
    _, err := s.SaveUserMessage(ctx, conversationID, sandboxInfo, template, prompt)
    if err != nil {
        // Log but don't fail - message save is non-critical for streaming
    }
    
    // 2. Build Claude config
    claudeConfig := &modal.ClaudeExecConfig{
        Prompt:          prompt,
        OutputFormat:    "stream-json",
        SkipPermissions: true,
        Verbose:         true,
    }
    
    // 3. Execute and stream Claude
    err = s.sandboxService.ExecuteClaudeStream(ctx, sandboxInfo, claudeConfig, responseWriter)
    if err != nil {
        return err
    }
    
    // 4. Get token usage from Claude process (TODO: pass ClaudeProcess through)
    // For now, we'll need to refactor ExecuteClaudeStream to return ClaudeProcess
    tokenUsage := &lifecycle_hooks.TokenUsage{
        InputTokens:  0, // TODO: get from ClaudeProcess
        OutputTokens: 0,
        CacheTokens:  0,
    }
    
    // 5. Save assistant message (response body would need to be captured)
    // TODO: Capture response during streaming
    
    // 6. Execute OnStreamFinish hook (non-critical)
    s.sandboxService.ExecuteStreamFinishHook(ctx, conversationID, sandboxInfo, template, tokenUsage)
    
    return nil
}
```

---

### Phase 8: Controller Layer

**Design Principle**: Controllers are entry points. Keep them thin by delegating to service layer.

#### 8.1. Conversation Streaming Endpoint (`internal/controllers/conversations/streaming.go` - NEW)

**Purpose**: Handle conversation streaming endpoint.

**Why New File**: Streaming is a new feature separate from CRUD operations.

```go
package conversations

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/griffnb/core/lib/log"
    "github.com/griffnb/core/lib/router/request"
    "github.com/griffnb/core/lib/types"
    "github.com/griffnb/techboss-ai-go/internal/models/sandbox"
    "github.com/griffnb/techboss-ai-go/internal/services/conversation_service"
)

// StreamRequest holds request data for Claude streaming
type StreamRequest struct {
    Prompt   string           `json:"prompt"`
    Provider sandbox.Provider `json:"provider"` // Default: PROVIDER_CLAUDE_CODE
    AgentID  types.UUID       `json:"agent_id"`
}

// streamClaude handles POST /conversations/{conversationId}/sandbox/{sandboxId}
func streamClaude(w http.ResponseWriter, req *http.Request) {
    userSession := request.GetReqSession(req)
    accountID := userSession.User.ID()
    conversationID := types.UUID(chi.URLParam(req, "conversationId"))
    sandboxID := types.UUID(chi.URLParam(req, "sandboxId")) // Optional for future use
    
    // Parse request
    data, err := request.GetJSONPostAs[*StreamRequest](req)
    if err != nil || data.Prompt == "" {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }
    
    // Default provider
    if data.Provider == 0 {
        data.Provider = sandbox.PROVIDER_CLAUDE_CODE
    }
    
    // Initialize conversation service
    service := conversation_service.NewConversationService()
    
    // Get or create conversation
    conv, err := service.GetOrCreateConversation(req.Context(), conversationID, accountID, data.AgentID)
    if err != nil {
        log.ErrorContext(err, req.Context())
        http.Error(w, "failed to get conversation", http.StatusInternalServerError)
        return
    }
    
    // Ensure sandbox exists and is initialized
    sandboxInfo, template, err := service.EnsureSandbox(req.Context(), conv, data.Provider)
    if err != nil {
        log.ErrorContext(err, req.Context())
        http.Error(w, "failed to initialize sandbox", http.StatusInternalServerError)
        return
    }
    
    // Stream with full hook coordination
    err = service.StreamClaudeWithHooks(req.Context(), conversationID, sandboxInfo, template, data.Prompt, w)
    if err != nil {
        log.ErrorContext(err, req.Context())
        // Don't return error here - streaming may have started
    }
}
```

#### 8.2. Conversation Routes (`internal/controllers/conversations/setup.go` - UPDATE)

**Purpose**: Add streaming route to existing conversation controller.

**Why Same File**: Route setup is centralized here.

```go
// Add to existing Setup function:

// Streaming routes
coreRouter.AddMainRoute(tools.BuildString("/", ROUTE), func(r chi.Router) {
    r.Group(func(authR chi.Router) {
        authR.Post("/{conversationId}/sandbox/{sandboxId}", helpers.RoleHandler(helpers.RoleHandlerMap{
            constants.ROLE_ANY_AUTHORIZED: streamClaude, // No wrapper - handles HTTP directly
        }))
    })
})
```

---

## Data Flow Examples

### Example 1: New Conversation with Cold Start

```
1. User: POST /conversations/{new-id}/sandbox/{any-id}
         Body: {prompt: "Hello", provider: 1, agent_id: "..."}

2. Controller: streamClaude()
   - Get/Create conversation
   - Call service.EnsureSandbox()

3. ConversationService: EnsureSandbox()
   - No existing sandbox → Create new
   - Get template for PROVIDER_CLAUDE_CODE
   - Call sandboxService.CreateSandbox()
   - Call sandboxService.ExecuteColdStartHook() ← CRITICAL

4. SandboxService: ExecuteColdStartHook()
   - Execute template.Hooks.OnColdStart
   - DefaultOnColdStart: calls InitVolumeFromS3WithState()

5. Modal Integration: InitVolumeFromS3WithState()
   - Read local .sandbox-state (empty - new sandbox)
   - Read S3 .sandbox-state (or generate if missing)
   - Compare states → download all files
   - Execute sync actions
   - Write updated local .sandbox-state

6. ConversationService: StreamClaudeWithHooks()
   - SaveUserMessage() → OnMessage hook → DynamoDB
   - ExecuteClaudeStream() → streams to client
   - Parse tokens from stream
   - SaveAssistantMessage() → OnMessage hook → DynamoDB
   - ExecuteStreamFinishHook() → OnStreamFinish hook
   
7. DefaultOnStreamFinish:
   - Sync volume to S3 with new timestamp
   - Update conversation stats with tokens
```
andbox_service/
├── state_files/                      # State file management (subfolder)
│   ├── types.go                      # State file data structures
│   ├── reader.go                     # Read state files
│   ├── writer.go                     # Write state files atomically
│   ├── comparator.go                 # Compare states, generate diffs
│   └── state_files_test.go          # Tests
├── lifecycle/                        # Lifecycle hooks (subfolder)
│   ├── types.go                      # Hook signatures and context
│   ├── executor.go                   # Hook orchestration
│   └── defaults.go                   # Default hook implementations
└── lifecycle.go                      # Hook coordination for sandboxes

internal/integrations/modal/
├── storage_state.go                  # State-based sync operations
└── storage_state_test.go            # Integration tests

internal/services/conversation_service/
├── conversation_service.go           # Core conversation logic
├── messages.go                       # Message management
└── streaming.go                      # Streaming coordination

internal/controllers/conversations/
└── streaming.go                      # Streaming endpoint
```

### Files Modified
```
internal/models/conversation/
└── stats.go                          # Add token fields

internal/models/message/
└── message.go                        # Add tool calls support

internal/integrations/modal/
├── storage.go                        # Update SyncStats
└── claude.go                         # Add token tracking

internal/services/sandbox_service/
└── templates.go                      # Add hook registration

internal/controllers/conversations/
└── setup.go                          # Add streaming route
```

### File Count by Layer
- **Models**: 2 modifications
- **Integration**: 2 new + 2 modified = 4 files
- **Services**: 11 new + 2 modified = 13 files (state_files + lifecycle are subfolders)
- **Controllers**: 1 new + 1 modified = 2 files

**Total**: 11 new files (organized in subfolders), 7

### Files Modified
```
internal/models/conversation/
└── stats.go                          # Add token fields

internal/integrations/modal/
├── storage.go                        # Update SyncStats
└── claude.go                         # Add token tracking

internal/services/sandbox_service/
└── templates.go                      # Add hook registration

internal/controllers/conversations/
└── setup.go                          # Add streaming route
```

### File Count by Layer
- **Models**: 1 modification
- **Integration**: 2 new + 2 modified = 4 files
- **Services**: 11 new + 2 modified = 13 files
- **Controllers**: 1 new + 1 modified = 2 files

**Total**: 13 new files, 6 modified files

---

## Testing Strategy

### Unit Tests (Per File)
- Each file gets its own `_test.go` file
- Test coverage target: ≥90%
- Focus on business logic and edge cases

### Integration Tests
- `storage_state_test.go`: Actual S3 operations
- `conversation_service_test.go`: End-to-end conversation flows
- Tests use real Modal sandboxes (when configured)

### Hook Tests
- Mock implementations for each hook
- Test hook execution order
- Test error handling (critical vs non-critical)

---

## Migration Strategy

### Phase 1: Foundation (No Breaking Changes)
- Implement state file service
- Implement lifecycle hooks service
- Add new storage functions alongside existing ones
- Update conversation model

### Phase 2: Integration (Backward Compatible)
- Integrate hooks into sandbox service
- Create conversation service
- Add new endpoint (old endpoint still works)

### Phase 3: Migration
- Deprecate old `/sandboxes/{id}/claude` endpoint
- Update frontend to use new endpoint
- Monitor metrics for both endpoints

### Phase 4: Cleanup
- Remove deprecated endpoint
- Remove legacy sync functions if no longer used

---

## Error Handling Matrix

| Hook Phase | Error Type | Behavior |
|------------|-----------|----------|
| OnColdStart | Critical | Fail sandbox creation, clean up, return error |
| OnMessage | Non-Critical | Log error, continue streaming |
| OnStreamFinish | Non-Critical | Log error, return successful response |
| OnTerminate | Non-Critical | Log error, mark sandbox as terminated anyway |

---

## Configuration

### Environment Variables
```bash
# Sync behavior
SANDBOX_SYNC_STALE_THRESHOLD=3600  # seconds (1 hour default)

# S3 Configuration (existing)
S3_BUCKET_AGENT_DOCS=techboss-agent-docs

# Modal Configuration (existing)
MODAL_TOKEN_ID=...
MODAL_TOKEN_SECRET=...
```

### Template Configuration
```go
// Per-provider hook customization
template := &SandboxTemplate{
    // ... existing fields ...
    Hooks: &lifecycle_hooks.LifecycleHooks{
        OnColdStart: func(ctx context.Context, hookData *lifecycle_hooks.HookData) error {
            // Custom cold start logic
            // Can call DefaultOnColdStart for S3 sync
            return lifecycle_hooks.DefaultOnColdStart(ctx, hookData)
        },
        // ... other hooks ...
    },
}
```

---

## Success Criteria

1. ✅ All messages stored in DynamoDB with conversation linkage
2. ✅ Token usage tracked per message and aggregated in conversation stats
3. ✅ State files maintain perfect sync between local and S3
4. ✅ Files deleted locally when removed from S3
5. ✅ OnColdStart failures prevent sandbox creation
6. ✅ OnMessage/OnStreamFinish failures don't block user experience
7. ✅ Hooks are composable and reusable across providers
8. ✅ No file exceeds 400 lines (excluding generated code)
9. ✅ Test coverage ≥90% for all new code
10. ✅ Zero breaking changes to existing sandbox operations
