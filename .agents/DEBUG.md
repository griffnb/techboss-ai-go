# Modal Sandbox System - Debug Session Summary

**Date**: December 27, 2024

## Issues Resolved

### 1. S3 Sync Permission Error with sudo

**Problem**: S3 sync was failing with error:
```
sudo: The "no new privileges" flag is set, which prevents sudo from running as root.
```

**Root Cause**: Modal sandboxes run with the "no new privileges" flag set, which prevents `sudo` from escalating privileges. The S3 sync command was trying to use `sudo aws` which failed.

**Solution**:
- Removed `sudo` from the AWS CLI invocation in `internal/integrations/modal/storage.go:150`
- AWS CLI now runs directly as `claudeuser` with credentials passed via Modal secrets
- Removed sudo package and sudoers configuration from `internal/integrations/modal/image_templates.go`

**Files Modified**:
- `internal/integrations/modal/storage.go` - Line 150: Changed `sudo aws s3 sync` to `aws s3 sync`
- `internal/integrations/modal/image_templates.go` - Removed sudo installation and sudoers config (lines 37, 59-61)

---

### 2. Workspace Permission Denied Errors

**Problem**: Claude inside the sandbox was getting "Permission denied" errors when trying to create files/directories in `/mnt/workspace`.

**Error Example**:
```
mkdir: can't create directory 'marketing': Permission denied
EACCES: permission denied, open '/__modal/volumes/.../file.md'
```

**Root Cause**: Modal volumes are mounted at runtime as root-owned directories. Even though we created `/mnt/workspace` and set ownership to `claudeuser` during Docker image build, when Modal mounts the actual volume over that directory at runtime, it becomes root-owned again.

**First Attempt (Failed)**: Added a CMD to the Dockerfile to fix ownership at container startup. This failed because the CMD either runs before the volume is mounted, or Modal doesn't allow ownership changes on mounted volumes at container startup.

**Second Attempt (Failed)**: Combined chown and Claude execution in a single command with nested shell execution. This had complex quoting issues and Modal was running commands as `claudeuser` by default (not root), so chown failed.

**Third Attempt (Failed)**: Split into two exec calls with USER root, but chown still didn't work because `/mnt/workspace` is a **symlink** to `/__modal/volumes/vo-...`. Running `chown -R /mnt/workspace` only changed the symlink ownership, not the actual volume directory.

**Working Solution**: Resolve symlink before chowning, split into TWO separate `Sandbox.Exec()` calls:
1. Added `USER root` to Dockerfile so Modal runs exec commands as root by default
2. **STEP 1**: First exec call resolves symlink and fixes permissions
   ```bash
   REAL_PATH=$(readlink -f /mnt/workspace) && chown -R claudeuser:claudeuser $REAL_PATH
   ```
   - Uses `readlink -f` to resolve `/mnt/workspace` → `/__modal/volumes/vo-...`
   - Chowns the actual volume directory, not the symlink
   - Waits for completion and validates exit code
3. **STEP 2**: Second exec call runs Claude as `claudeuser` using `runuser`
   - Full API access with secrets
   - PTY enabled for Claude CLI
   - Workspace now has proper permissions

**Key Insights**:
- Modal's `Sandbox.Exec()` respects the USER directive in the Dockerfile
- Modal volume mounts are **symlinks** - must resolve them before chowning
- By keeping USER as root and explicitly switching to claudeuser only for Claude execution, we can run setup commands with proper privileges

**Files Modified**:
- `internal/integrations/modal/image_templates.go` - Lines 57-59: Added `USER root` directive
- `internal/integrations/modal/claude.go` - Lines 83-154: Split into two exec calls (setup + execution)

---

### 3. Claude Responses Not Displaying in UI

**Problem**: When Claude executed in the sandbox, the web UI showed empty responses. The typing indicator would appear but no actual content was displayed.

**Root Cause**: The UI was expecting a different response format than what Claude actually sends. Claude sends responses with nested `message` objects containing `content` arrays with `type`, `text`, and `tool_use` fields, but the UI was looking for a simpler flat structure.

**Example Claude Response Format**:
```json
{
  "type": "system",
  "subtype": "init",
  "cwd": "/mnt/workspace",
  "session_id": "..."
}

{
  "type": "assistant",
  "message": {
    "content": [
      {"type": "text", "text": "I'll help you..."},
      {"type": "tool_use", "id": "...", "name": "Bash", "input": {...}}
    ]
  }
}

{
  "type": "user",
  "message": {
    "content": [
      {"type": "tool_result", "content": "...", "is_error": false}
    ]
  }
}
```

**Solution**: Completely rewrote the `appendStreamContent()` function to handle the correct format:
1. Handle `system/init` messages (show initialization)
2. Process `assistant` messages with content arrays containing `text` and `tool_use` blocks
3. Process `user` messages with `tool_result` blocks
4. Handle `result` messages for completion
5. Added automatic scrolling after each content append
6. Improved text display with `white-space: pre-wrap` and `word-break: break-word`
7. Added error styling for failed tool executions

**Files Modified**:
- `static/modal-sandbox-ui.html`:
  - Lines 792-839: Rewrote `appendStreamContent()` to parse correct format
  - Lines 842-854: Updated `appendTextContent()` to add proper styling and scrolling
  - Lines 856-877: Updated `appendToolUse()` to trigger scrolling
  - Lines 880-900: Updated `appendToolResult()` to show errors and trigger scrolling
  - Lines 245-258: Added CSS for `.tool-result-error` styling

---

### 4. S3 Configuration Automatically Added

**Task**: Update `SandboxService.CreateSandbox()` to always include S3 configuration for workspace persistence and testing.

**Implementation**:
- Automatically adds S3 config using `environment.GetConfig().S3Config.Buckets["agent-docs"]`
- Key prefix set to `docs/{accountID}/` (trailing slash required by Modal)
- Timestamp automatically appended during sync: `docs/{accountID}/{unix_timestamp}/`
- Added proper nil checks to prevent panics when S3Config isn't configured

**Files Modified**:
- `internal/services/modal/sandbox_service.go` - Lines 52-64: Auto-configure S3

---

## Current Architecture

### Permission Model
1. **Build Time**:
   - Create `claudeuser` (UID 1000, GID 1000)
   - Install Claude globally to `/usr/local/bin/claude` (755 permissions)
   - Create `/mnt/workspace` directory (placeholder, will be replaced by Modal volume)
   - Set `USER root` - keeps container running as root for exec commands

2. **Volume Mount (Runtime)**:
   - Modal mounts volume to `/mnt/workspace` as root
   - Volume overwrites build-time directory with root ownership
   - Because USER is root, exec commands run with root privileges

3. **Execution Time - STEP 1 (Permission Fix)**:
   - First `Sandbox.Exec()`: `REAL_PATH=$(readlink -f /mnt/workspace) && chown -R claudeuser:claudeuser $REAL_PATH`
   - Resolves symlink to actual volume directory (`/__modal/volumes/vo-...`)
   - Runs as root (respects USER directive)
   - Waits for completion and validates exit code
   - Ensures claudeuser has full write access to the actual workspace

4. **Execution Time - STEP 2 (Claude)**:
   - Second `Sandbox.Exec()`: `runuser -u claudeuser -- sh -c 'cd /mnt/workspace && claude ...'`
   - Starts as root but immediately switches to `claudeuser`
   - Claude runs with full workspace access (now owned by claudeuser)
   - AWS CLI runs as `claudeuser` with credentials from secrets
   - No sudo required - everything runs with proper permissions

### S3 Integration
- **Bucket**: Retrieved from `environment.GetConfig().S3Config.Buckets["agent-docs"]`
- **Key Prefix**: `docs/{accountID}/` (set at creation)
- **Sync Path**: `docs/{accountID}/{unix_timestamp}/` (timestamp added during sync)
- **Credentials**: Passed via Modal secrets, no sudo required

### UI Streaming
- **Init Message**: Shows "🚀 Claude initialized..."
- **Text Content**: Real-time streaming with proper line breaks
- **Tool Calls**: Color-coded blocks with abbreviated parameters
  - Blue: Tool use (🔧)
  - Green: Success results (✓)
  - Red: Error results (❌)
- **Auto-scroll**: Content automatically scrolls as it arrives

---

## Testing Status

- ✅ S3 sync works without sudo
- ✅ Workspace permissions fixed at runtime
- ✅ Claude can create files/directories in workspace
- ✅ UI properly displays all Claude response types
- ✅ Real-time streaming shows progress
- ✅ Tool calls and results display correctly
- ✅ Error states properly styled

---

## Key Files

### Backend
- `internal/integrations/modal/image_templates.go` - Docker image config with runtime permission fix
- `internal/integrations/modal/storage.go` - S3 sync without sudo
- `internal/integrations/modal/claude.go` - Claude execution as non-root user
- `internal/services/modal/sandbox_service.go` - Auto S3 config

### Frontend
- `static/modal-sandbox-ui.html` - Updated streaming parser and UI display

---

## Lessons Learned

1. **Modal Sandboxes and sudo**: Modal sets "no new privileges" flag, making sudo unusable. All operations must run with proper permissions from the start.

2. **Volume Mount Timing & USER Directive**: Docker image build-time ownership changes don't persist when Modal mounts volumes at runtime. The solution is to:
   - Set `USER root` in Dockerfile so exec commands run with proper privileges
   - Use a separate exec call to fix permissions AFTER mount but BEFORE application runs
   - **CRITICAL**: Modal volume mounts are symlinks - use `readlink -f` to resolve the actual path before chowning
   - Then explicitly switch to non-root user for application execution

3. **Claude Response Format**: Claude Code CLI outputs a specific streaming JSON format with nested message structures. UI must parse this correctly to display content.

4. **Real-time UX**: Auto-scrolling and incremental content updates are crucial for good streaming UX. Every content append should trigger a scroll.

5. **AWS CLI Without Root**: AWS CLI works fine as a non-root user when credentials are properly passed via environment variables.

---

## Next Steps (If Needed)

1. **Performance**: Monitor the chown operation time for large workspaces. If it becomes a bottleneck, consider:
   - Incremental permission fixes (only fix directories as needed)
   - Caching workspace state between runs
   - Using a different volume mounting strategy

2. **Error Handling**: Add error detection if the chown command fails:
   - Log the chown output
   - Fail gracefully with a clear error message
   - Consider retry logic for transient failures

3. **UI Polish**: Add more detailed progress indicators for long-running operations
