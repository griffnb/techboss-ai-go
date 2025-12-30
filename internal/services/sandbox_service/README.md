# Sandbox Service Documentation

## Overview

The Sandbox Service manages the lifecycle of sandboxed execution environments with support for flexible lifecycle hooks and intelligent S3 synchronization. It provides a conversation-centric architecture where sandboxes are tied to conversations and messages are tracked through their entire lifecycle.

## Architecture

The service is organized into focused subfolders and files:

```
internal/services/sandbox_service/
├── state_files/              # State file management (subfolder)
│   ├── types.go              # State file data structures
│   ├── reader.go             # Read state files from disk/S3
│   ├── writer.go             # Write state files atomically
│   ├── comparator.go         # Compare states, generate diffs
│   └── *_test.go             # Tests for state file operations
├── lifecycle/                # Lifecycle hooks (subfolder)
│   ├── types.go              # Hook signatures and context
│   ├── executor.go           # Hook execution with logging
│   ├── defaults.go           # Default hook implementations
│   └── *_test.go             # Tests for lifecycle operations
├── sandbox_service.go        # Core sandbox operations
├── templates.go              # Provider-specific templates
├── lifecycle_coordination.go # Hook coordination
└── reconstruct.go            # Sandbox info reconstruction
```

## State Files Subfolder

### Purpose

The `state_files/` subfolder provides state management for tracking file synchronization between local sandbox volumes and S3 storage. It works like `.git` to maintain perfect sync state, enabling:

- Efficient incremental synchronization
- Detection of deleted files
- Checksum-based change detection
- Atomic updates to prevent corruption

### State File Format

State files (`.sandbox-state`) are JSON files tracking the state of all files in a directory:

```json
{
  "version": "1.0",
  "last_synced_at": 1704067200,
  "files": [
    {
      "path": "src/main.py",
      "checksum": "098f6bcd4621d373cade4e832627b4f6",
      "size": 1024,
      "modified_at": 1704067100
    }
  ]
}
```

**Fields:**
- `version`: Schema version for forward compatibility
- `last_synced_at`: Unix timestamp of last sync operation
- `files`: Array of tracked files with metadata

**FileEntry Fields:**
- `path`: Relative path from volume root
- `checksum`: MD5 hash matching S3 ETag format
- `size`: File size in bytes
- `modified_at`: Unix timestamp of last modification

### File Organization

**`types.go`** - Data structures
- `StateFile`: Main state file structure
- `FileEntry`: Individual file metadata
- `StateDiff`: Comparison results
- `SyncStats`: Sync operation metrics

**`reader.go`** - Reading state files
- `ReadLocalStateFile()`: Read from sandbox local volume
- `ReadS3StateFile()`: Read from S3 mount
- `ParseStateFile()`: Parse JSON with version validation

**`writer.go`** - Writing state files atomically
- `WriteLocalStateFile()`: Atomic write to local volume
- `WriteS3StateFile()`: Atomic write to S3
- `GenerateStateFile()`: Scan directory and generate state

**`comparator.go`** - State comparison
- `CompareStateFiles()`: Determine sync actions needed
- `CheckIfStale()`: Determine if state file is outdated

### Usage Example

```go
import "github.com/griffnb/techboss-ai-go/internal/services/sandbox_service/state_files"

// Read current states
localState, err := state_files.ReadLocalStateFile(ctx, sandboxInfo, "/volume")
s3State, err := state_files.ReadS3StateFile(ctx, sandboxInfo, "/s3")

// Compare to determine actions
diff := state_files.CompareStateFiles(localState, s3State)

// Process differences
for _, file := range diff.FilesToDownload {
    // Download file.Path from S3
}
for _, path := range diff.FilesToDelete {
    // Delete path from local volume
}

// Generate new state after changes
newState, err := state_files.GenerateStateFile(ctx, sandboxInfo, "/volume")

// Save state atomically
err = state_files.WriteLocalStateFile(ctx, sandboxInfo, "/volume", newState)
```

### State Comparison Logic

The comparator uses the following logic to determine actions:

1. **Files to Download**: Files in S3 state that are either:
   - Not in local state (new files)
   - Have different checksums (modified files)

2. **Files to Delete**: Files in local state but not in S3 state (removed files)

3. **Files to Skip**: Files present in both with matching checksums (unchanged)

### Atomic Writes

State file writes use atomic operations to prevent corruption:

1. Write to temporary file (`.sandbox-state.tmp`)
2. Rename to final file (`.sandbox-state`)
3. Rename is atomic on both local filesystems and S3 mounts

This ensures state files are never partially written, even if the process is interrupted.

## Lifecycle Hooks Subfolder

### Purpose

The `lifecycle/` subfolder provides a composable hook system for sandbox lifecycle management. Hooks allow providers to customize behavior at key lifecycle points without modifying core logic.

### Hook Types

Four lifecycle hooks are available:

1. **OnColdStart** - Executed when new sandbox is created (CRITICAL)
2. **OnMessage** - Executed when messages are saved (non-critical)
3. **OnStreamFinish** - Executed after streaming completes (non-critical)
4. **OnTerminate** - Executed when sandbox is terminated (non-critical)

### File Organization

**`types.go`** - Hook definitions
- `HookFunc`: Function signature for hooks
- `HookData`: Context passed to hooks
- `TokenUsage`: Token consumption data
- `LifecycleHooks`: Complete hook set

**`executor.go`** - Hook execution
- `ExecuteHook()`: Execute hook with logging and timing
- Handles nil hooks gracefully
- Logs execution time and errors

**`defaults.go`** - Default implementations
- `DefaultOnColdStart()`: S3 sync on initialization
- `DefaultOnMessage()`: Message persistence
- `DefaultOnStreamFinish()`: S3 sync and stats update
- `DefaultOnTerminate()`: Cleanup operations

### Hook Context (HookData)

Every hook receives a `HookData` struct with context:

```go
type HookData struct {
    ConversationID types.UUID          // Conversation being processed
    SandboxInfo    *modal.SandboxInfo  // Sandbox environment details
    Message        *message.Message    // For OnMessage hook
    TokenUsage     *TokenUsage         // For OnStreamFinish hook
}
```

### Error Handling Philosophy

Hooks determine their own criticality by their return value:

**Critical Hook:**
```go
func MyCriticalHook(ctx context.Context, hookData *HookData) error {
    err := doSomethingImportant()
    if err != nil {
        return err // Propagate error - caller will fail operation
    }
    return nil
}
```

**Non-Critical Hook:**
```go
func MyNonCriticalHook(ctx context.Context, hookData *HookData) error {
    err := doSomethingOptional()
    if err != nil {
        log.Errorf("Optional operation failed but continuing: %v", err)
        return nil // Swallow error - caller continues normally
    }
    return nil
}
```

### Hook Execution Rules

| Hook | When Executed | Criticality | Error Behavior |
|------|---------------|-------------|----------------|
| OnColdStart | New sandbox creation | CRITICAL | Fails sandbox creation |
| OnMessage | Message save | Non-critical | Logged, operation continues |
| OnStreamFinish | After streaming | Non-critical | Logged, operation continues |
| OnTerminate | Sandbox termination | Non-critical | Logged, termination continues |

### Creating Custom Hooks

To create custom hooks for a new provider:

```go
// 1. Define custom hook function
func CustomOnColdStart(ctx context.Context, hookData *HookData) error {
    // Custom initialization logic
    log.Infof("Custom cold start for conversation %s", hookData.ConversationID)

    // Call default hook for standard S3 sync
    err := lifecycle.DefaultOnColdStart(ctx, hookData)
    if err != nil {
        return err // Propagate critical errors
    }

    // Additional custom logic
    err = doCustomInitialization(hookData.SandboxInfo)
    if err != nil {
        return err // This is critical too
    }

    return nil
}

// 2. Register hooks in template
template := &SandboxTemplate{
    Provider: sandbox.PROVIDER_CUSTOM,
    Hooks: &lifecycle.LifecycleHooks{
        OnColdStart:    CustomOnColdStart,
        OnMessage:      lifecycle.DefaultOnMessage,      // Use default
        OnStreamFinish: lifecycle.DefaultOnStreamFinish, // Use default
        OnTerminate:    lifecycle.DefaultOnTerminate,    // Use default
    },
}
```

### Composing Hooks

Hooks can be composed to combine behaviors:

```go
func ComposedHook(ctx context.Context, hookData *HookData) error {
    // Execute multiple operations
    if err := FirstHook(ctx, hookData); err != nil {
        return err // Critical failure
    }

    // Non-critical operation
    if err := SecondHook(ctx, hookData); err != nil {
        log.Errorf("Non-critical operation failed: %v", err)
        // Continue anyway
    }

    return nil
}
```

### Default Hook Implementations

**DefaultOnColdStart** (CRITICAL):
- Syncs files from S3 to local volume using state files
- Returns errors that fail sandbox creation
- Ensures sandbox starts with latest files

**DefaultOnMessage** (non-critical):
- Saves message to DynamoDB
- Updates conversation message counter
- Swallows errors and logs them

**DefaultOnStreamFinish** (non-critical):
- Syncs local volume changes to S3 with timestamp versioning
- Updates conversation token usage statistics
- Swallows errors and logs them

**DefaultOnTerminate** (non-critical):
- Default implementation is a no-op
- Custom providers can override for cleanup

## Hook Coordination

The `lifecycle_coordination.go` file integrates hooks into the sandbox service:

### Hook Execution Methods

```go
// ExecuteColdStartHook - CRITICAL, returns errors
func (s *SandboxService) ExecuteColdStartHook(
    ctx context.Context,
    conversationID types.UUID,
    sandboxInfo *modal.SandboxInfo,
    template *SandboxTemplate,
) error

// ExecuteMessageHook - Non-critical, ignores return
func (s *SandboxService) ExecuteMessageHook(
    ctx context.Context,
    conversationID types.UUID,
    sandboxInfo *modal.SandboxInfo,
    template *SandboxTemplate,
    msg *message.Message,
)

// ExecuteStreamFinishHook - Non-critical, ignores return
func (s *SandboxService) ExecuteStreamFinishHook(
    ctx context.Context,
    conversationID types.UUID,
    sandboxInfo *modal.SandboxInfo,
    template *SandboxTemplate,
    tokenUsage *lifecycle.TokenUsage,
)

// ExecuteTerminateHook - Returns errors if hook determines critical
func (s *SandboxService) ExecuteTerminateHook(
    ctx context.Context,
    conversationID types.UUID,
    sandboxInfo *modal.SandboxInfo,
    template *SandboxTemplate,
) error
```

### Usage in Controllers

Hooks are automatically executed by the conversation service. Controllers don't need to call hooks directly:

```go
// In controller
service := conversation_service.NewConversationService()

// Ensure sandbox exists - OnColdStart hook runs here
sandboxInfo, template, err := service.EnsureSandbox(ctx, conv, provider)
if err != nil {
    // OnColdStart failure returns error here
    return err
}

// Stream with hooks - OnMessage, OnStreamFinish run here
err = service.StreamClaudeWithHooks(ctx, conversationID, sandboxInfo, template, prompt, w)
```

## Template System

Templates define provider-specific sandbox configurations including hooks:

### Template Structure

```go
type SandboxTemplate struct {
    Provider      sandbox.Provider      // Provider type (e.g., PROVIDER_CLAUDE_CODE)
    ImageConfig   *modal.ImageConfig    // Container image configuration
    VolumeName    string                // Volume mount name
    S3BucketName  string                // S3 bucket for file storage
    S3KeyPrefix   string                // S3 key prefix pattern
    InitFromS3    bool                  // Whether to sync from S3 on cold start
    Hooks         *lifecycle.LifecycleHooks // Lifecycle hooks
}
```

### Registering Templates

Templates are registered in `templates.go`:

```go
func getClaudeCodeTemplate(agentID types.UUID) *SandboxTemplate {
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
            OnTerminate:    lifecycle.DefaultOnTerminate,
        },
    }
}
```

### Adding New Providers

To add a new sandbox provider:

1. Define provider constant in `internal/models/sandbox/sandbox.go`:
```go
const PROVIDER_MY_CUSTOM Provider = 3
```

2. Create template function in `templates.go`:
```go
func getMyCustomTemplate(agentID types.UUID) *SandboxTemplate {
    return &SandboxTemplate{
        Provider:     sandbox.PROVIDER_MY_CUSTOM,
        ImageConfig:  modal.GetImageConfigFromTemplate("custom"),
        VolumeName:   fmt.Sprintf("custom-%s", agentID),
        S3BucketName: "my-custom-bucket",
        S3KeyPrefix:  "custom/%s/",
        InitFromS3:   true,
        Hooks: &lifecycle.LifecycleHooks{
            OnColdStart:    CustomOnColdStart,    // Your custom hook
            OnMessage:      lifecycle.DefaultOnMessage,
            OnStreamFinish: lifecycle.DefaultOnStreamFinish,
            OnTerminate:    lifecycle.DefaultOnTerminate,
        },
    }
}
```

3. Register in `GetSandboxTemplate()`:
```go
func GetSandboxTemplate(provider sandbox.Provider, agentID types.UUID) (*SandboxTemplate, error) {
    switch provider {
    case sandbox.PROVIDER_CLAUDE_CODE:
        return getClaudeCodeTemplate(agentID), nil
    case sandbox.PROVIDER_MY_CUSTOM:
        return getMyCustomTemplate(agentID), nil
    default:
        return nil, errors.Errorf("unsupported provider: %d", provider)
    }
}
```

## Error Handling Patterns

### Critical vs Non-Critical Errors

**Critical Errors** - Operations that must succeed:
- OnColdStart failures (sandbox would be inconsistent)
- Sandbox creation failures
- Streaming execution failures

**Non-Critical Errors** - Operations that can fail gracefully:
- Message persistence failures (content already streamed)
- S3 sync failures (can retry later)
- Statistics update failures (informational only)

### Hook Error Handling Example

```go
// Critical hook - returns errors
func MyCriticalHook(ctx context.Context, hookData *HookData) error {
    if err := mustSucceedOperation(); err != nil {
        return errors.Wrapf(err, "critical operation failed")
    }
    return nil
}

// Non-critical hook - swallows errors
func MyNonCriticalHook(ctx context.Context, hookData *HookData) error {
    if err := optionalOperation(); err != nil {
        log.Errorf("Optional operation failed but continuing: %v", err)
        return nil // Swallow error
    }
    return nil
}
```

## Testing

### Unit Tests

Each subfolder contains comprehensive unit tests:

```
state_files/
├── reader_test.go      # Test reading state files
├── writer_test.go      # Test writing state files
├── comparator_test.go  # Test state comparison
└── state_files_test.go # Integration tests

lifecycle/
├── executor_test.go    # Test hook execution
└── defaults_test.go    # Test default hooks
```

### Test Coverage

- Target: ≥90% coverage for all new code
- Use table-driven tests for multiple scenarios
- Mock external dependencies (S3, Modal API)
- Test error paths and edge cases

### Running Tests

```bash
# Run all sandbox service tests
#code_tools run_tests --package=./internal/services/sandbox_service

# Run specific subfolder tests
#code_tools run_tests --package=./internal/services/sandbox_service/state_files
#code_tools run_tests --package=./internal/services/sandbox_service/lifecycle
```

## Best Practices

### State File Management

1. **Always use atomic writes** - Never write directly to `.sandbox-state`
2. **Handle missing state files** - Treat as empty state, not error
3. **Validate versions** - Check state file schema version
4. **Log state changes** - Log files downloaded/deleted for debugging

### Lifecycle Hooks

1. **Keep hooks focused** - Each hook has one clear responsibility
2. **Determine criticality carefully** - Only return errors if operation truly must succeed
3. **Log all errors** - Even non-critical errors should be logged
4. **Compose default hooks** - Call `DefaultOnX` then add custom logic
5. **Test error paths** - Test both success and failure scenarios

### Performance

1. **State comparison is O(n)** - Uses maps for efficient lookups
2. **Parallel file operations** - State file design supports parallel sync
3. **Incremental sync** - Only sync changed files, not entire directories
4. **Checksum caching** - State files cache checksums to avoid recalculation

### Security

1. **Validate file paths** - Prevent path traversal attacks
2. **Sanitize checksums** - Validate MD5 format
3. **Limit file sizes** - Prevent memory exhaustion
4. **Use context timeouts** - Prevent indefinite operations

## Configuration

### Environment Variables

```bash
# Sync behavior
SANDBOX_SYNC_STALE_THRESHOLD=3600  # seconds (1 hour default)

# S3 Configuration
S3_BUCKET_AGENT_DOCS=techboss-agent-docs

# Modal Configuration
MODAL_TOKEN_ID=...
MODAL_TOKEN_SECRET=...
```

### Template Configuration

Templates can be customized per agent or provider:

```go
func getCustomTemplate(agentID types.UUID) *SandboxTemplate {
    return &SandboxTemplate{
        Provider:     sandbox.PROVIDER_CUSTOM,
        S3KeyPrefix:  fmt.Sprintf("agents/%s/", agentID), // Per-agent prefix
        InitFromS3:   true,
        Hooks: &lifecycle.LifecycleHooks{
            OnColdStart: func(ctx context.Context, hookData *HookData) error {
                // Agent-specific initialization
                return lifecycle.DefaultOnColdStart(ctx, hookData)
            },
        },
    }
}
```

## Troubleshooting

### State File Issues

**Problem**: Files not syncing correctly

**Solution**: Check state file validity:
```bash
# In sandbox, check local state
cat /volume/.sandbox-state

# Check S3 state
cat /s3/.sandbox-state

# Look for checksum mismatches
```

**Problem**: State file corruption

**Solution**: Regenerate state file:
```go
newState, err := state_files.GenerateStateFile(ctx, sandboxInfo, "/volume")
err = state_files.WriteLocalStateFile(ctx, sandboxInfo, "/volume", newState)
```

### Hook Issues

**Problem**: OnColdStart failing

**Solution**: Check logs for specific error:
- S3 permissions
- State file read errors
- File download failures

**Problem**: OnStreamFinish not executing

**Solution**: Verify hook is registered in template:
```go
template.Hooks.OnStreamFinish != nil
```

### Performance Issues

**Problem**: Slow sync operations

**Solution**: Check file counts and sizes:
```go
stats := state_files.CompareStateFiles(local, s3)
log.Infof("Files to sync: %d", len(stats.FilesToDownload))
```

## Migration Guide

### From Legacy Sandbox System

**Old Architecture:**
- Sandbox-centric (sandboxes are primary)
- No message tracking
- Manual S3 sync
- No lifecycle hooks

**New Architecture:**
- Conversation-centric (conversations are primary)
- Automatic message tracking
- State-based S3 sync
- Lifecycle hooks for extensibility

**Migration Steps:**

1. **Update endpoint calls:**
```go
// Old
POST /sandboxes/{id}/claude
{ "prompt": "..." }

// New
POST /conversation/{conversationId}/sandbox/{sandboxId}
{ "prompt": "...", "agent_id": "..." }
```

2. **Handle conversation IDs:**
```javascript
// Generate conversation ID on client
const conversationId = generateUUID();
```

3. **Update response handling:**
```javascript
// Response format is similar, but includes conversation metadata
eventSource.onmessage = (event) => {
    const data = JSON.parse(event.data);
    // Handle as before
};
```

4. **Migrate existing sandboxes:**
- Existing sandboxes continue to work
- Create conversation records for existing sandboxes
- Link sandboxes to conversations via `conversation.sandbox_id`

## Summary

The Sandbox Service provides a robust, extensible system for managing sandboxed execution environments:

- **State Files**: Reliable file synchronization with checksum-based change detection
- **Lifecycle Hooks**: Composable, provider-specific customization at key lifecycle points
- **Error Handling**: Clear distinction between critical and non-critical failures
- **Conversation-Centric**: Messages and sandboxes are tied to conversations for better tracking

For more information, see:
- `/docs/CONTROLLERS.md` - Controller endpoint documentation
- `/docs/PRD.md` - Product requirements and design decisions
- Design document: `.agents/specs/sandbox-lifecycle-hooks/design.md`
