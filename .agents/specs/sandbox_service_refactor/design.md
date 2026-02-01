# Sandbox Service Refactor - Design Document

## Overview

This document provides a comprehensive design for refactoring `internal/services/sandbox_service` into a stateful service with a GetOrCreate pattern. The design focuses on moving commonly-threaded parameters (sandboxInfo, sandboxModel, user) into the service struct itself, simplifying method signatures while maintaining all existing functionality.

### Design Goals

1. **Universal GetOrCreate Pattern**: Single constructor that handles both existing and new sandboxes
2. **Stateful Service**: All context stored in service struct, no parameter threading
3. **Simplified API**: Clean, intuitive method signatures
4. **Backward Compatible Behavior**: All functionality works identically to current implementation
5. **Maintainable**: Clear separation of concerns, easy to extend

---

## API Summary

### Public API (12 methods)

**Constructor:**
- `NewSandboxService(ctx, sandboxID, account, config) (*SandboxService, error)`

**Core Operations:**
- `TerminateSandbox(ctx, syncToS3) error`
- `ExecuteClaudeStream(ctx, config, responseWriter) error`

**File Operations:**
- `ListFiles(ctx, opts) ([]FileInfo, error)`
- `GetFileContent(ctx, source, filePath) (string, error)`
- `BuildFileTree(files, rootPath) *FileTreeNode`

**Sync:**
- `SyncFiles(ctx) error` ← replaces SyncToS3, ForceVolumeSync, orchestrators

**Optional:**
- `SetConversationID(conversationID)` ← for future conversation context
- `GetSandboxInfo() *modal.SandboxInfo` ← testing/debugging
- `GetSandboxModel() *sandbox.Sandbox` ← testing/debugging
- `GetVolume() *modal.VolumeInfo` ← access to volume info

### Private Methods (18 methods)

**Initialization:**
- `loadOrCreateSandbox(ctx, sandboxID) error`
- `reconstructSandboxInfo(ctx, model) (*modal.SandboxInfo, error)`
- `ensureSandboxAccessible(ctx) error`
- `loadTemplate() error`
- `validateParams(sandboxID, account, config) error`

**State Management:**
- `ensureSandboxRunning(ctx, operation) error`
- `updateSandboxState(ctx, newSandboxInfo) error`

**Command Builders:**
- `buildListFilesCommand(opts) string`
- `buildReadFileCommand(source, filePath, maxSize) string`

**Lifecycle Hooks:**
- `executeColdStartHook(ctx) error` ← automatic on init (CRITICAL)
- `executeMessageHook(ctx, msg) error` ← future use (NON-CRITICAL)
- `executeStreamFinishHook(ctx, tokenUsage) error` ← future use (NON-CRITICAL)
- `executeTerminateHook(ctx) error` ← automatic on terminate (NON-CRITICAL)

**Sync Operations:**
- `initFromS3(ctx) error`
- `orchestratePullSync(ctx, staleThresholdSeconds) error`
- `orchestratePushSync(ctx) error`

**Utilities:**
- `generateAppName(accountID) string`
- `generateVolumeName(accountID) string`
- `buildPermissionFixCommand(workdir, username) string`

### Package-Level Functions (4 functions)

**Template Factories:**
- `GetSandboxTemplate(sandboxType, agentID) *SandboxTemplate`
- `GetGitHubTemplate(config) *SandboxTemplate`
- `GetGitHubImage() string`

**Tree Utilities:**
- `ConvertTreePathsToUserFacing(node) *FileTreeNode`

---

## Architecture

### Service Struct Design

```go
// SandboxService provides stateful operations for managing sandboxes.
// The service maintains all necessary context (sandbox info, database model, account)
// internally, eliminating the need to pass these parameters to every method.
//
// The service implements a GetOrCreate pattern - when initialized with a sandbox ID,
// it will load the existing sandbox if it exists, or create a new one if it doesn't.
// The constructor ensures the sandbox is accessible and in a usable state.
type SandboxService struct {
    // Core dependencies
    client modal.APIClientInterface

    // Sandbox context - loaded or created during initialization
    sandboxInfo  *modal.SandboxInfo  // Modal API sandbox information
    sandboxModel *sandbox.Sandbox     // Database model
    volume       *modal.VolumeInfo    // Volume information (persists across reboots)

    // Account context
    account   *account.Account  // Account owning the sandbox (for DB updates)
    accountID types.UUID        // Account ID (cached for convenience)

    // Configuration and template
    config   *modal.SandboxConfig  // Sandbox configuration (for creation/recreation)
    template *SandboxTemplate      // Lifecycle hook template (optional)

    // Optional conversation context
    conversationID types.UUID  // Conversation ID for lifecycle hooks (optional)
}
```

### State Management Strategy

**Initialization State Flow:**
```
NewSandboxService()
    ↓
validateParams()
    ↓
Load sandbox by ID from DB
    ↓
    ├─→ Exists? → Load model → reconstructSandboxInfo() → Ensure accessible → Ready
    └─→ Not exists? → Create sandbox → Save model → Construct SandboxInfo → Ready
```

**Operation State Flow:**
```
Method Call (e.g., ListFiles)
    ↓
ensureSandboxRunning() [private]
    ↓
    ├─→ Running? → Proceed
    └─→ Terminated? → Create new sandbox → Update state → Proceed
```

---

## API Surface - Public Methods

### Constructor

#### `NewSandboxService`

```go
// NewSandboxService creates a new stateful sandbox service instance.
//
// This constructor implements a GetOrCreate pattern:
// - If the sandbox ID exists in the database, it loads the existing sandbox
// - If the sandbox ID does not exist, it creates a new sandbox using the provided config
// - Ensures the sandbox is accessible and in a usable state
//
// The returned service is fully initialized and ready to use. All subsequent
// method calls will use the internal state (sandboxInfo, sandboxModel, volume, account)
// without requiring these as parameters.
//
// Parameters:
//   - ctx: Context for database and API operations
//   - sandboxID: UUID of the sandbox to load or create
//   - account: Account owning the sandbox (for DB updates and authorization)
//   - config: Sandbox configuration (used if creating a new sandbox)
//
// Returns:
//   - *SandboxService: Fully initialized service ready to use
//   - error: Validation errors, database errors, or sandbox creation errors
//
// Example:
//   service, err := NewSandboxService(ctx, sandboxID, account, config)
//   if err != nil {
//       return err
//   }
//   // Service is now ready - sandbox exists and is accessible
//   files, err := service.ListFiles(ctx, opts)
func NewSandboxService(
    ctx context.Context,
    sandboxID types.UUID,
    account *account.Account,
    config *modal.SandboxConfig,
) (*SandboxService, error)
```

**Responsibilities:**
1. Validate input parameters (non-zero UUID, non-nil account and config)
2. Initialize Modal API client
3. Call `loadOrCreateSandbox()` to get or create sandbox
4. Load volume information from sandbox
5. Load appropriate template based on sandbox type/agent ID
6. Ensure sandbox is accessible and ready to use
7. Return fully initialized service

**Error Cases:**
- Invalid parameters (zero UUID, nil account/config)
- Database query failures
- Sandbox creation failures (API errors)
- Sandbox not accessible
- Volume information unavailable

---

### Core Sandbox Operations

#### `TerminateSandbox`

```go
// TerminateSandbox stops the sandbox and optionally syncs data to S3.
//
// If syncToS3 is true, this method will push all sandbox files to S3 before
// terminating. After termination, the sandbox model is updated in the database
// to reflect the terminated state.
//
// Parameters:
//   - ctx: Context for API operations
//   - syncToS3: Whether to sync files to S3 before terminating
//
// Returns:
//   - error: API errors, sync errors, or database update errors
func (s *SandboxService) TerminateSandbox(ctx context.Context, syncToS3 bool) error
```

**Responsibilities:**
1. Optionally sync to S3 if requested
2. Call Modal API to terminate sandbox
3. Update database model with terminated state
4. Update internal state to reflect termination

**State Changes:**
- sandboxInfo status changes to terminated
- sandboxModel.Status updated in DB

---

#### `ExecuteClaudeStream`

```go
// ExecuteClaudeStream executes a Claude command in the sandbox and streams the response.
//
// This method ensures the sandbox is running before executing the command.
// If the sandbox is terminated, it will automatically create a new one.
//
// Parameters:
//   - ctx: Context for API operations
//   - config: Execution configuration (command, environment, etc.)
//   - responseWriter: Writer for streaming response data
//
// Returns:
//   - error: API errors, execution errors, or stream errors
func (s *SandboxService) ExecuteClaudeStream(
    ctx context.Context,
    config *claude.ExecutionConfig,
    responseWriter io.Writer,
) error
```

**Responsibilities:**
1. Ensure sandbox is running (calls `ensureSandboxRunning()`)
2. Execute command via Modal API
3. Stream response to writer
4. Handle any state changes from auto-restart

---

### File Operations

#### `ListFiles`

```go
// ListFiles lists files in the sandbox with optional filtering.
//
// This method ensures the sandbox is running before listing files.
//
// Parameters:
//   - ctx: Context for API operations
//   - opts: File listing options (path filter, max depth, etc.)
//
// Returns:
//   - []FileInfo: List of files matching the options
//   - error: API errors or execution errors
func (s *SandboxService) ListFiles(ctx context.Context, opts *ListFilesOptions) ([]FileInfo, error)
```

**Responsibilities:**
1. Ensure sandbox is running
2. Build list files command based on options
3. Execute command in sandbox
4. Parse and return file list

---

#### `GetFileContent`

```go
// GetFileContent reads the content of a file from the sandbox.
//
// This method ensures the sandbox is running before reading the file.
//
// Parameters:
//   - ctx: Context for API operations
//   - source: File source ("working_dir" or "home_dir")
//   - filePath: Path to the file relative to source
//
// Returns:
//   - string: File content
//   - error: API errors, file not found, or execution errors
func (s *SandboxService) GetFileContent(
    ctx context.Context,
    source string,
    filePath string,
) (string, error)
```

**Responsibilities:**
1. Ensure sandbox is running
2. Build read file command
3. Execute command in sandbox
4. Return file content

---

#### `BuildFileTree`

```go
// BuildFileTree builds a hierarchical tree structure from a flat list of files.
//
// This is a utility method that doesn't require sandbox interaction.
//
// Parameters:
//   - files: Flat list of file paths
//   - rootPath: Root path for the tree
//
// Returns:
//   - *FileTreeNode: Root node of the tree structure
func (s *SandboxService) BuildFileTree(files []string, rootPath string) *FileTreeNode
```

**Responsibilities:**
1. Parse file paths
2. Build hierarchical tree structure
3. Return root node

**Note:** This is stateless and doesn't interact with the sandbox.

---

### File Sync Operations

#### `SyncFiles`

```go
// SyncFiles intelligently syncs files between the sandbox and S3.
//
// This method automatically detects the appropriate sync operation:
// - Pull from S3 if local state is stale
// - Push to S3 if local changes need to be persisted
// - Bidirectional sync if needed
//
// The method uses state files for incremental syncing to minimize data transfer.
// This replaces the previous SyncToS3 and ForceVolumeSync methods with a single
// intelligent sync operation.
//
// Parameters:
//   - ctx: Context for API operations
//
// Returns:
//   - error: API errors or sync errors
func (s *SandboxService) SyncFiles(ctx context.Context) error
```

**Responsibilities:**
1. Ensure sandbox is running
2. Analyze current state vs S3 state
3. Determine if pull, push, or bidirectional sync is needed
4. Perform incremental sync using state files
5. Update state tracking
6. Handle sync conflicts if any

**Note:** This consolidates `SyncToS3`, `ForceVolumeSync`, and the orchestrator logic into one intelligent method.

---

---

### Optional Setters

#### `SetConversationID`

```go
// SetConversationID sets the conversation ID for lifecycle hooks.
//
// This is optional - if not set, lifecycle hooks will use a zero UUID.
//
// Parameters:
//   - conversationID: Conversation ID to associate with lifecycle hooks
func (s *SandboxService) SetConversationID(conversationID types.UUID)
```

**Responsibilities:**
1. Update internal conversationID field

**Use Case:** When the same service instance needs to be used for multiple conversations.

---

### Getters (for testing/debugging)

#### `GetSandboxInfo`

```go
// GetSandboxInfo returns the current sandbox info.
//
// This is primarily useful for testing and debugging to verify service state.
//
// Returns:
//   - *modal.SandboxInfo: Current sandbox info (may change after auto-restart)
func (s *SandboxService) GetSandboxInfo() *modal.SandboxInfo
```

---

#### `GetSandboxModel`

```go
// GetSandboxModel returns the current sandbox database model.
//
// This is primarily useful for testing and debugging to verify service state.
//
// Returns:
//   - *sandbox.Sandbox: Current sandbox model (may change after auto-restart)
func (s *SandboxService) GetSandboxModel() *sandbox.Sandbox
```

---

#### `GetVolume`

```go
// GetVolume returns the current volume information.
//
// The volume persists across sandbox reboots and is important for maintaining
// data continuity. This is useful for operations that need to reference the
// volume directly or ensure volume consistency.
//
// Returns:
//   - *modal.VolumeInfo: Current volume info
func (s *SandboxService) GetVolume() *modal.VolumeInfo
```

---

## API Surface - Private Methods

### Initialization Helpers

#### `loadOrCreateSandbox`

```go
// loadOrCreateSandbox implements the GetOrCreate pattern for sandbox initialization.
//
// This method attempts to load the sandbox by ID from the database. If it exists,
// it reconstructs the SandboxInfo. If it doesn't exist, it creates a new sandbox
// via the Modal API and saves it to the database.
//
// Parameters:
//   - ctx: Context for database and API operations
//   - sandboxID: UUID of the sandbox to load or create
//
// Returns:
//   - error: Database errors, API errors, or reconstruction errors
//
// Side effects:
//   - Sets s.sandboxModel
//   - Sets s.sandboxInfo
//   - Sets s.volume
func (s *SandboxService) loadOrCreateSandbox(ctx context.Context, sandboxID types.UUID) error
```

**Responsibilities:**
1. Query database for sandbox by ID
2. If found:
   - Load sandbox model
   - Call `reconstructSandboxInfo()` to build SandboxInfo
   - Extract volume information
3. If not found:
   - Generate sandbox names (using `generateAppName`, `generateVolumeName`)
   - Create sandbox via Modal API (using s.config)
   - Save new sandbox model to database
   - Construct SandboxInfo from creation response
   - Extract volume information
4. Set s.sandboxModel, s.sandboxInfo, and s.volume
5. Ensure sandbox is accessible (call `ensureSandboxAccessible()`)

**Error Cases:**
- Database query failures
- Sandbox creation API failures
- Database insert failures
- SandboxInfo reconstruction failures
- Sandbox not accessible

---

#### `reconstructSandboxInfo`

```go
// reconstructSandboxInfo reconstructs SandboxInfo from a database model.
//
// This is used during service initialization when loading an existing sandbox.
//
// Parameters:
//   - ctx: Context for operations
//   - model: Sandbox database model
//
// Returns:
//   - *modal.SandboxInfo: Reconstructed sandbox info
//   - error: Reconstruction errors (invalid data, missing fields, etc.)
func (s *SandboxService) reconstructSandboxInfo(
    ctx context.Context,
    model *sandbox.Sandbox,
) (*modal.SandboxInfo, error)
```

**Responsibilities:**
1. Extract all fields from database model
2. Build Modal SandboxInfo structure
3. Validate reconstructed data
4. Return constructed SandboxInfo

**Note:** Moved from package-level function to private method since it only used internally.

---

#### `ensureSandboxAccessible`

```go
// ensureSandboxAccessible verifies the sandbox is accessible and ready to use.
//
// This is called after loading or creating a sandbox to ensure it's in a usable state.
//
// Parameters:
//   - ctx: Context for API operations
//
// Returns:
//   - error: Errors if sandbox is not accessible
func (s *SandboxService) ensureSandboxAccessible(ctx context.Context) error
```

**Responsibilities:**
1. Check sandbox status via Modal API
2. Verify sandbox is reachable
3. Validate volume is attached
4. Return error if not accessible

---

#### `loadTemplate`

```go
// loadTemplate loads the appropriate sandbox template based on sandbox type and agent ID.
//
// If the sandbox type doesn't map to a template, this sets s.template to nil (not an error).
//
// Returns:
//   - error: Only returns error if template loading logic itself fails (not "no template")
//
// Side effects:
//   - Sets s.template (may be nil if no template applicable)
func (s *SandboxService) loadTemplate() error
```

**Responsibilities:**
1. Check sandbox type (from s.sandboxModel or s.config)
2. Call `GetSandboxTemplate(sandboxType, agentID)`
3. Set s.template (may be nil)

**Note:** This is not an error if no template exists - some sandboxes don't use templates.

---

#### `validateParams`

```go
// validateParams validates constructor parameters before initialization.
//
// Parameters:
//   - sandboxID: Sandbox UUID to validate
//   - account: Account to validate
//   - config: Sandbox config to validate
//
// Returns:
//   - error: Validation error with clear message about what's invalid
func validateParams(
    sandboxID types.UUID,
    account *account.Account,
    config *modal.SandboxConfig,
) error
```

**Responsibilities:**
1. Check sandboxID is not zero
2. Check account is not nil
3. Check account.ID is not zero
4. Check config is not nil
5. Return descriptive error if any validation fails

---

### State Management Helpers

#### `ensureSandboxRunning`

```go
// ensureSandboxRunning ensures the sandbox is in a running state before an operation.
//
// If the sandbox is already running, this is a no-op. If the sandbox is terminated,
// this method creates a new sandbox with the same configuration, updates the database,
// and updates internal state.
//
// Parameters:
//   - ctx: Context for API and database operations
//   - operation: Name of the operation requiring a running sandbox (for error messages)
//
// Returns:
//   - error: API errors, database errors, or state sync errors
//
// Side effects (if sandbox recreated):
//   - Creates new sandbox via Modal API
//   - Updates s.sandboxModel in database
//   - Updates s.sandboxInfo
func (s *SandboxService) ensureSandboxRunning(ctx context.Context, operation string) error
```

**Responsibilities:**
1. Check current sandbox status from s.sandboxInfo
2. If running: return immediately (no-op)
3. If terminated:
   - Create new sandbox via Modal API (using s.config)
   - Update database model with new sandbox ID and status
   - Update s.sandboxModel to point to updated DB record
   - Update s.sandboxInfo with new sandbox info
4. If other status: handle appropriately

**State Changes:**
- May update s.sandboxInfo (new sandbox)
- May update s.sandboxModel (new DB record)

**Error Cases:**
- Sandbox creation API failures
- Database update failures
- State synchronization failures

---

#### `updateSandboxState`

```go
// updateSandboxState updates internal state and database after sandbox changes.
//
// This method ensures atomicity - either both the database and internal state are updated,
// or neither is updated (on error).
//
// Parameters:
//   - ctx: Context for database operations
//   - newSandboxInfo: New sandbox info to set
//
// Returns:
//   - error: Database errors
//
// Side effects:
//   - Updates s.sandboxInfo
//   - Updates s.sandboxModel in database
func (s *SandboxService) updateSandboxState(
    ctx context.Context,
    newSandboxInfo *modal.SandboxInfo,
) error
```

**Responsibilities:**
1. Update sandbox model in database
2. Update s.sandboxModel in memory
3. Update s.sandboxInfo in memory
4. Ensure atomicity (rollback on failure)

---

### Command Building Helpers

#### `buildListFilesCommand`

```go
// buildListFilesCommand builds the shell command for listing files.
//
// Parameters:
//   - opts: File listing options
//
// Returns:
//   - string: Shell command to execute
func (s *SandboxService) buildListFilesCommand(opts *ListFilesOptions) string
```

**Responsibilities:**
1. Build find/ls command based on options
2. Handle path filters, depth limits, etc.
3. Return command string

**Note:** Existing logic, just moved to receiver method.

---

#### `buildReadFileCommand`

```go
// buildReadFileCommand builds the shell command for reading a file.
//
// Parameters:
//   - source: File source ("working_dir" or "home_dir")
//   - filePath: Path to the file
//   - maxSize: Maximum file size to read (optional)
//
// Returns:
//   - string: Shell command to execute
func (s *SandboxService) buildReadFileCommand(
    source string,
    filePath string,
    maxSize int64,
) string
```

**Responsibilities:**
1. Build cat/head command based on parameters
2. Handle file path escaping
3. Handle max size limits
4. Return command string

**Note:** Existing logic, just moved to receiver method.

---


### Private Lifecycle Hooks

These hooks are called automatically by the service and should not be invoked externally.

**Hook System Overview:**

The lifecycle hook system uses a registry pattern where hooks are registered in the `SandboxTemplate`:

```go
// Hooks are function pointers registered in templates
type HookFunc func(ctx context.Context, hookData *HookData) error

type LifecycleHooks struct {
    OnColdStart    HookFunc  // Sandbox initialization
    OnMessage      HookFunc  // Message persistence
    OnStreamFinish HookFunc  // Post-streaming operations
    OnTerminate    HookFunc  // Cleanup operations
}
```

**Hook Execution Pattern:**
1. Service checks if template and specific hook exist
2. Builds `lifecycle.HookData` from internal state
3. Calls `lifecycle.ExecuteHook()` which handles logging, timing, and execution
4. Returns error (critical) or nil (success/no hook)

**Hook Criticality:**
- **CRITICAL** (OnColdStart): Errors propagate → fail operations
- **NON-CRITICAL** (OnMessage, OnStreamFinish, OnTerminate): Errors swallowed by hook implementations → operations continue

**Default Implementations:**
- `lifecycle.DefaultOnColdStart`: S3 sync with state files
- `lifecycle.DefaultOnMessage`: DynamoDB message save + stats
- `lifecycle.DefaultOnStreamFinish`: S3 sync + token stats
- `lifecycle.DefaultOnTerminate`: Logging only

#### `executeColdStartHook`

```go
// executeColdStartHook executes the cold start lifecycle hook.
//
// Called automatically during sandbox initialization after sandbox is created.
// This is a CRITICAL hook - errors are propagated and will fail initialization.
//
// The hook is registered in the template's LifecycleHooks struct. If no hook
// is registered (nil), this is a no-op. The default implementation (DefaultOnColdStart)
// performs S3 sync with state file tracking.
//
// Parameters:
//   - ctx: Context for API operations
//
// Returns:
//   - error: Hook execution errors (nil if no hook exists or hook succeeds)
func (s *SandboxService) executeColdStartHook(ctx context.Context) error
```

**Responsibilities:**
1. Check if s.template exists and has OnColdStart hook
2. Build lifecycle.HookData from internal state (s.sandboxInfo, s.conversationID)
3. Call lifecycle.ExecuteHook() with the template's OnColdStart hook
4. Return error if hook fails (CRITICAL - will fail sandbox initialization)

**Implementation:**
```go
if s.template == nil || s.template.Hooks == nil || s.template.Hooks.OnColdStart == nil {
    return nil // No hook registered
}

hookData := &lifecycle.HookData{
    ConversationID: s.conversationID,
    SandboxInfo:    s.sandboxInfo,
}

return lifecycle.ExecuteHook(ctx, "OnColdStart", s.template.Hooks.OnColdStart, hookData)
```

---

#### `executeMessageHook`

```go
// executeMessageHook executes the message lifecycle hook.
//
// Currently unused since conversation_service is not in use.
// May be called in the future when message handling is implemented.
//
// Parameters:
//   - ctx: Context for API operations
//   - msg: Message model to pass to the hook
//
// Returns:
//   - error: Hook execution errors (nil if no hook exists)
func (s *SandboxService) executeMessageHook(ctx context.Context, msg *message.Message) error
```

---

#### `executeStreamFinishHook`

```go
// executeStreamFinishHook executes the stream finish lifecycle hook.
//
// Currently unused since conversation_service is not in use.
// May be called in the future when stream handling is implemented.
//
// Parameters:
//   - ctx: Context for API operations
//   - tokenUsage: Token usage information
//
// Returns:
//   - error: Hook execution errors (nil if no hook exists)
func (s *SandboxService) executeStreamFinishHook(ctx context.Context, tokenUsage *TokenUsage) error
```

---

#### `executeTerminateHook`

```go
// executeTerminateHook executes the terminate lifecycle hook.
//
// Called automatically during sandbox termination before the sandbox is stopped.
// This is a NON-CRITICAL hook - errors are logged but swallowed by the hook implementation.
//
// The hook is registered in the template's LifecycleHooks struct. If no hook
// is registered (nil), this is a no-op. The default implementation (DefaultOnTerminate)
// just logs the termination.
//
// Parameters:
//   - ctx: Context for API operations
//
// Returns:
//   - error: Always nil (hook swallows errors internally)
func (s *SandboxService) executeTerminateHook(ctx context.Context) error
```

**Responsibilities:**
1. Check if s.template exists and has OnTerminate hook
2. Build lifecycle.HookData from internal state
3. Call lifecycle.ExecuteHook() with the template's OnTerminate hook
4. Return nil (NON-CRITICAL - errors swallowed by hook implementation)

**Implementation:**
```go
if s.template == nil || s.template.Hooks == nil || s.template.Hooks.OnTerminate == nil {
    return nil // No hook registered
}

hookData := &lifecycle.HookData{
    ConversationID: s.conversationID,
    SandboxInfo:    s.sandboxInfo,
}

// Hook implementation swallows errors if non-critical
_ = lifecycle.ExecuteHook(ctx, "OnTerminate", s.template.Hooks.OnTerminate, hookData)
return nil
```

---

### Private Sync Operations

These methods handle S3 synchronization internally.

#### `initFromS3`

```go
// initFromS3 initializes the sandbox by pulling files from S3.
//
// Called automatically during sandbox initialization if needed based on configuration.
//
// Parameters:
//   - ctx: Context for API operations
//
// Returns:
//   - error: API errors or sync errors
func (s *SandboxService) initFromS3(ctx context.Context) error
```

**Responsibilities:**
1. Ensure sandbox is running
2. Call lifecycle action to pull from S3
3. Wait for sync to complete

---

#### `orchestratePullSync`

```go
// orchestratePullSync orchestrates a pull sync from S3 with staleness checking.
//
// Used internally by SyncFiles to handle pull operations.
//
// Parameters:
//   - ctx: Context for API operations
//   - staleThresholdSeconds: Number of seconds before local state is considered stale
//
// Returns:
//   - error: API errors, sync errors, or state file errors
func (s *SandboxService) orchestratePullSync(ctx context.Context, staleThresholdSeconds int) error
```

**Responsibilities:**
1. Check staleness using state files
2. If stale, perform incremental pull
3. Update state files

---

#### `orchestratePushSync`

```go
// orchestratePushSync orchestrates a push sync to S3 with incremental state tracking.
//
// Used internally by SyncFiles to handle push operations.
//
// Parameters:
//   - ctx: Context for API operations
//
// Returns:
//   - error: API errors, sync errors, or state file errors
func (s *SandboxService) orchestratePushSync(ctx context.Context) error
```

**Responsibilities:**
1. Perform incremental push using state files
2. Update state files

---

### Naming and Utility Helpers

These methods generate names and perform utility operations.

#### `generateAppName`

```go
// generateAppName generates a unique app name for a sandbox.
//
// Parameters:
//   - accountID: Account ID to use in name generation
//
// Returns:
//   - string: Generated app name
func (s *SandboxService) generateAppName(accountID types.UUID) string
```

**Note:** Moved from package-level function to private method.

---

#### `generateVolumeName`

```go
// generateVolumeName generates a unique volume name for a sandbox.
//
// Parameters:
//   - accountID: Account ID to use in name generation
//
// Returns:
//   - string: Generated volume name
func (s *SandboxService) generateVolumeName(accountID types.UUID) string
```

**Note:** Moved from package-level function to private method.

---

#### `buildPermissionFixCommand`

```go
// buildPermissionFixCommand builds a command to fix file permissions.
//
// Parameters:
//   - workdir: Working directory path
//   - username: Username for permissions
//
// Returns:
//   - string: Shell command to fix permissions
func (s *SandboxService) buildPermissionFixCommand(workdir, username string) string
```

**Note:** Moved from package-level function to private method.

---

## Package-Level Utility Functions

These remain as package-level functions because they're used by external code or don't require service state.

### Template Factories

#### `GetSandboxTemplate`

```go
// GetSandboxTemplate retrieves the template for a given sandbox type and optional agent ID.
//
// Parameters:
//   - sandboxType: Type of sandbox (e.g., "github", "default")
//   - agentID: Optional agent ID for agent-specific templates
//
// Returns:
//   - *SandboxTemplate: Template if found, nil otherwise
func GetSandboxTemplate(sandboxType string, agentID types.UUID) *SandboxTemplate
```

**Note:** Remains package-level - template factory pattern.

---

#### `GetGitHubTemplate`

```go
// GetGitHubTemplate retrieves the GitHub-specific template with configuration.
//
// Parameters:
//   - config: GitHub configuration for the template
//
// Returns:
//   - *SandboxTemplate: GitHub template with config applied
func GetGitHubTemplate(config *GitHubConfig) *SandboxTemplate
```

**Note:** Remains package-level - template factory pattern.

---

#### `GetGitHubImage`

```go
// GetGitHubImage returns the Docker image for GitHub sandboxes.
//
// Returns:
//   - string: Docker image name
func GetGitHubImage() string
```

**Note:** Remains package-level - configuration utility.

---

### Tree Utilities

#### `ConvertTreePathsToUserFacing`

```go
// ConvertTreePathsToUserFacing converts internal paths to user-facing paths.
//
// Parameters:
//   - node: Root node of the file tree
//
// Returns:
//   - *FileTreeNode: Converted tree with user-facing paths
func ConvertTreePathsToUserFacing(node *FileTreeNode) *FileTreeNode
```

**Note:** Remains package-level - may be used by external code for tree manipulation.

---

## Component Interactions

### Initialization Flow

```
NewSandboxService(ctx, sandboxID, account, config)
    │
    ├─→ validateParams(sandboxID, account, config)
    │   └─→ returns error if invalid
    │
    ├─→ Initialize client: modal.Client()
    │
    ├─→ loadOrCreateSandbox(ctx, sandboxID)
    │   │
    │   ├─→ Query DB for sandbox by ID
    │   │
    │   ├─→ IF EXISTS:
    │   │   ├─→ Load model
    │   │   ├─→ reconstructSandboxInfo(ctx, model)
    │   │   ├─→ Extract volume info
    │   │   └─→ ensureSandboxAccessible(ctx)
    │   │
    │   └─→ IF NOT EXISTS:
    │       ├─→ generateAppName(accountID)
    │       ├─→ generateVolumeName(accountID)
    │       ├─→ Create sandbox via Modal API
    │       ├─→ Save model to DB
    │       ├─→ Construct SandboxInfo from API response
    │       ├─→ Extract volume info
    │       └─→ ensureSandboxAccessible(ctx)
    │
    ├─→ loadTemplate()
    │   └─→ GetSandboxTemplate(type, agentID)
    │
    ├─→ executeColdStartHook(ctx) [if template exists]
    │
    └─→ Return fully initialized service
```

---

### Operation Flow (e.g., ListFiles)

```
service.ListFiles(ctx, opts)
    │
    ├─→ ensureSandboxRunning(ctx, "list files")
    │   │
    │   ├─→ Check s.sandboxInfo.Status
    │   │
    │   ├─→ IF RUNNING:
    │   │   └─→ return (no-op)
    │   │
    │   └─→ IF TERMINATED:
    │       ├─→ Create new sandbox via Modal API
    │       ├─→ updateSandboxState(ctx, newSandboxInfo)
    │       └─→ return
    │
    ├─→ buildListFilesCommand(opts)
    │
    ├─→ Execute command in sandbox via s.client
    │
    └─→ Parse and return results
```

---

### Lifecycle Hook Flow

```
service.ExecuteMessageHook(ctx, msg)
    │
    ├─→ executeHook(ctx, "message", hookFunc)
    │   │
    │   ├─→ Check if s.template != nil
    │   │   └─→ If nil, return nil (no-op)
    │   │
    │   ├─→ Check if template has message hook
    │   │   └─→ If no hook, return nil (no-op)
    │   │
    │   └─→ Call hookFunc(ctx, s.conversationID, s.sandboxInfo, s.template)
    │
    └─→ Return result
```

---

## Data Models

### Service Struct Fields

| Field | Type | Purpose | Set When | Updated When |
|-------|------|---------|----------|--------------|
| `client` | `modal.APIClientInterface` | Modal API client | Constructor | Never |
| `sandboxInfo` | `*modal.SandboxInfo` | Sandbox metadata from Modal | Constructor | ensureSandboxRunning, updateSandboxState |
| `sandboxModel` | `*sandbox.Sandbox` | Database model | Constructor | ensureSandboxRunning, updateSandboxState |
| `volume` | `*modal.VolumeInfo` | Volume info (persists across reboots) | Constructor | ensureSandboxRunning (when sandbox recreated) |
| `account` | `*account.Account` | Account owning sandbox | Constructor | Never |
| `accountID` | `types.UUID` | Account ID (cached) | Constructor | Never |
| `config` | `*modal.SandboxConfig` | Sandbox config | Constructor | Never |
| `template` | `*SandboxTemplate` | Lifecycle hooks | Constructor | Never (could add reload method) |
| `conversationID` | `types.UUID` | Conversation context | Constructor or SetConversationID | SetConversationID |

---

### Constructor Parameters

| Parameter | Type | Required | Purpose |
|-----------|------|----------|---------|
| `ctx` | `context.Context` | Yes | Database and API operations |
| `sandboxID` | `types.UUID` | Yes | Sandbox to load or create |
| `account` | `*account.Account` | Yes | Account owning the sandbox (for DB updates and auth) |
| `config` | `*modal.SandboxConfig` | Yes | Config for sandbox creation |

---

## Error Handling

### Error Types

```go
// Validation errors
var (
    ErrInvalidSandboxID = errors.New("sandbox ID cannot be zero")
    ErrInvalidAccount   = errors.New("account cannot be nil")
    ErrInvalidConfig    = errors.New("config cannot be nil")
)

// State errors
var (
    ErrSandboxNotRunning    = errors.New("sandbox is not in running state")
    ErrSandboxNotAccessible = errors.New("sandbox is not accessible")
    ErrStateUpdateFailed    = errors.New("failed to update sandbox state")
)

// Operation errors
var (
    ErrSandboxCreationFailed = errors.New("failed to create sandbox")
    ErrReconstructionFailed  = errors.New("failed to reconstruct sandbox info")
    ErrHookExecutionFailed   = errors.New("failed to execute lifecycle hook")
    ErrSyncFailed            = errors.New("failed to sync files")
)
```

### Error Wrapping Pattern

All errors should be wrapped with context using `errors.Wrapf`:

```go
// Example
if err != nil {
    return errors.Wrapf(err, "failed to load sandbox %s", sandboxID)
}
```

---

## Testing Strategy

### Unit Tests

**Constructor Tests:**
- ✅ Valid parameters with existing sandbox (loads from DB)
- ✅ Valid parameters with non-existing sandbox (creates new)
- ✅ Invalid parameters (zero UUIDs, nil user/config)
- ✅ Database load failures
- ✅ Sandbox creation failures
- ✅ Template loading (with and without templates)

**Method Tests (per method):**
- ✅ Method with running sandbox (happy path)
- ✅ Method with terminated sandbox (auto-restart)
- ✅ Method with API failures
- ✅ Method with state update failures

**Lifecycle Hook Tests:**
- ✅ Hook with template (executes)
- ✅ Hook without template (no-op)
- ✅ Hook with template but no specific hook (no-op)
- ✅ Hook execution failure

**Helper Tests:**
- ✅ ensureSandboxRunning with various states
- ✅ updateSandboxState atomicity
- ✅ loadOrCreateSandbox both paths
- ✅ Command builders

### Integration Tests

- ✅ Full service lifecycle (create → operate → terminate)
- ✅ Auto-restart behavior across operations
- ✅ S3 sync operations
- ✅ Lifecycle hook execution
- ✅ Multiple operations on same service instance

### Mock Strategy

**Mock Interfaces:**
- `modal.APIClientInterface` - Mock all Modal API calls
- Database access - Use test database with builder helpers

**Test Fixtures:**
- Use `testing_service.Builder` for creating test objects
- Clean up with `testtools.CleanupModel`

---

## Migration Guide

### Code Changes Required

**Before:**
```go
// Old pattern
service := sandbox_service.NewSandboxService()

// Check if sandbox exists
sandboxModel := sandbox.LoadByID(sandboxID)
var sandboxInfo *modal.SandboxInfo
if sandboxModel != nil {
    sandboxInfo = sandbox_service.ReconstructSandboxInfo(ctx, sandboxModel, accountID)
} else {
    // Create sandbox
    sandboxInfo = createSandbox(...)
}

// Call methods with lots of parameters
files, err := service.ListFiles(ctx, sandboxInfo, sandboxModel, user, opts)
```

**After:**
```go
// New pattern - one constructor, everything automatic
service, err := sandbox_service.NewSandboxService(ctx, sandboxID, account, config)
if err != nil {
    return err
}

// Clean method calls
files, err := service.ListFiles(ctx, opts)
```

### Breaking Changes

1. **Constructor signature changed**
   - Old: `NewSandboxService() *SandboxService`
   - New: `NewSandboxService(ctx, sandboxID, account, config) (*SandboxService, error)`

2. **All method signatures changed** - removed sandboxInfo, sandboxModel, user parameters

3. **Orchestrator functions removed from public API** - now private methods called by SyncFiles

4. **SyncToS3 and ForceVolumeSync replaced** - use SyncFiles() instead

5. **Lifecycle hooks are now private** - called automatically, not by external code

### Migration Checklist

- [ ] Update all `NewSandboxService()` calls to new signature with account instead of user
- [ ] Remove sandboxInfo, sandboxModel, user parameters from all method calls
- [ ] Replace `SyncToS3()` and `ForceVolumeSync()` calls with `SyncFiles()`
- [ ] Remove orchestrator function calls (OrchestratePullSync, OrchestratePushSync)
- [ ] Remove lifecycle hook calls (now automatic)
- [ ] Update tests to use new constructor pattern
- [ ] Update mocks for new signatures
- [ ] Run full test suite

---

## Performance Considerations

### Optimizations

1. **State Caching**: SandboxInfo and template are loaded once and reused
2. **Reduced Allocations**: No repeated parameter passing reduces allocations
3. **Lazy Loading**: Template loading only happens if needed

### Performance Goals

- No regression in operation performance (file operations, sync, hooks)
- Constructor should complete in < 100ms for existing sandboxes
- Constructor may take longer for new sandbox creation (acceptable - API call required)

### Memory Impact

- Service struct is ~200 bytes (negligible)
- Holding references to sandboxInfo and sandboxModel doesn't increase memory (already held by caller before)
- Overall memory footprint should be neutral or slightly better (fewer parameter allocations)

---

## Security Considerations

### Input Validation

- All constructor parameters validated before any operations
- UUIDs checked for zero values
- Pointers checked for nil
- Config validated for required fields

### State Integrity

- Atomic updates ensure database and memory stay synchronized
- Failed operations don't leave service in inconsistent state
- State updates wrapped in transactions where applicable

### Access Control

- Account context preserved for all operations
- Account used for authorization checks and DB updates
- No privilege escalation possible through service reuse
- Volume information ensures sandbox reboots maintain correct data isolation

---

## Future Enhancements

### Potential Additions

1. **Service Pool**: Reuse service instances across requests (requires thread-safety)
2. **State Refresh**: Method to reload sandbox state from DB
3. **Template Reload**: Method to reload template if it changes
4. **Metrics**: Add instrumentation for operations
5. **Logging**: Add structured logging for state changes
6. **Webhook Support**: Add lifecycle webhooks for external notifications

### Extensibility

The stateful design makes it easy to add:
- Cross-cutting concerns (logging, metrics, tracing)
- Caching strategies
- Background operations
- Event publishing

---

## Summary

This design refactors the sandbox service into a clean, stateful API with:

- **Universal GetOrCreate pattern** - handles existing and new sandboxes transparently
- **Simplified API** - no more parameter threading
- **Clear responsibilities** - state management encapsulated in service
- **Easy to test** - stateful testing patterns
- **Easy to extend** - new methods don't require parameter threading
- **Backward compatible behavior** - all functionality preserved

The key insight is moving from "pass everything to every method" to "initialize once with context, use everywhere." This is a common pattern in service-oriented architectures and will make the sandbox service much easier to use and maintain.
