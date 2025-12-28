# Sandbox Lifecycle Hooks with Conversation Integration - Implementation Tasks

## CRITICAL: Task Execution Instructions for Main Agent

**DO NOT IMPLEMENT TASKS YOURSELF.** Your role is to delegate each task to a specialized sub-agent, one task at a time.

### Task Delegation Process

1. **Work Sequentially**: Execute tasks in the order listed below
2. **One Task Per Sub-Agent**: Launch a new sub-agent for each individual task
3. **Complete Context**: Provide the sub-agent with ALL necessary context from this file
4. **Wait for Completion**: Do not move to next task until current sub-agent completes
5. **Track Progress**: Mark tasks as complete (✅) after sub-agent finishes successfully
6. **Update Learnings**: If sub-agent discovers important information, add it to the Learnings section

### Sub-Agent Prompt Template

When delegating a task, use this template to ensure the sub-agent has complete context:

```
You are implementing task [TASK_NUMBER] from the Sandbox Lifecycle Hooks feature.

**Task Description**: [Copy exact task description from below]

**Requirements This Task Satisfies**: [List specific requirement numbers from requirements.md]

**Context Files to Read FIRST**:
- `/Users/griffnb/projects/techboss/techboss-ai-go/AGENTS.md` - Go development practices
- `/Users/griffnb/projects/techboss/techboss-ai-go/.agents/specs/sandbox-lifecycle-hooks/requirements.md` - Feature requirements
- `/Users/griffnb/projects/techboss/techboss-ai-go/.agents/specs/sandbox-lifecycle-hooks/design.md` - Design architecture
- `/Users/griffnb/projects/techboss/techboss-ai-go/docs/CONTROLLERS.md` - Controller patterns (if relevant)
- [Any existing files being modified]

**Existing Patterns to Follow**:
- Use pointer fields for struct members (existing pattern in models)
- Return errors with `errors.Wrapf()` for context
- Use `#code_tools` for all testing, linting, formatting
- Follow TDD: RED (write test) → GREEN (make it pass) → REFACTOR → COMMIT
- Keep files under 400 lines
- Use table-driven tests

**Critical Project Patterns**:
- All database operations use `ctx context.Context` as first parameter
- Use `types.UUID` for IDs, not string
- Error handling: wrap with context using `errors.Wrapf()`
- Validation: check nil pointers and empty values before use
- DynamoDB operations: check for throttling in return values
- PostgreSQL: use model's Save(), Get(), FindAll() methods

**TDD Workflow (MANDATORY)**:
1. Write test first (RED) - ensure it fails
2. Write minimal implementation (GREEN) - make test pass
3. Refactor if needed (REFACTOR) - keep tests passing
4. Run `#code_tools run_tests [package]` to verify
5. Run `#code_tools lint` and `#code_tools fmt`
6. Commit with passing tests

**Success Criteria**:
- Tests pass with ≥90% coverage for new code
- `#code_tools lint` reports no issues
- Code follows existing patterns in codebase
- All context files read and patterns followed
- No orphaned or unused code
- Clear, concise documentation in code comments

**Deliverables**:
- Implementation file(s)
- Comprehensive test file(s) with table-driven tests
- All tests passing
- Linting clean
```

### Example Sub-Agent Invocation

```typescript
await runSubagent({
  description: "Implement state file types",
  prompt: `You are implementing task 1.1 from the Sandbox Lifecycle Hooks feature.

**Task Description**: Create state file type definitions in internal/services/sandbox_service/state_files/types.go

[Include full template context as above]

**Specific Instructions**:
- Create types.go with StateFile, FileEntry, StateDiff, SyncStats structs
- Follow existing patterns in internal/integrations/modal/storage.go for SyncStats
- Add comprehensive Go doc comments for all exported types
- No test file needed for types-only files (types are tested through usage)
`
});
```

---

## Implementation Tasks

### Phase 1: Foundation - Message and Conversation Models

#### 1.1. Extend Message Model with Tool Calls Support
**References**: Requirements 1.2, Design Phase 1

- [ ] Read existing `internal/models/message/message.go` structure
- [ ] Write test for Message with ToolCalls field (RED)
  - Test saving message with tool calls to DynamoDB
  - Test retrieving message with tool calls
  - Test empty tool calls array
- [ ] Add `ToolCalls []ToolCall` field to Message struct
- [ ] Create ToolCall struct with fields: ID, Type, Name, Input, Output, Status, Error
- [ ] Verify DynamoDB marshaling/unmarshaling works for nested structures
- [ ] Run tests (GREEN)
- [ ] References: Requirements 1.2

#### 1.2. Update Conversation Stats Structure
**References**: Requirements 1.1, 5.7, Design Phase 1

- [ ] Read existing `internal/models/conversation/stats.go`
- [ ] Write test for ConversationStats with token fields (RED)
  - Test adding token usage
  - Test incrementing messages
- [ ] Add TotalInputTokens, TotalOutputTokens, TotalCacheTokens int64 fields
- [ ] Implement AddTokenUsage(inputTokens, outputTokens, cacheTokens int64) method
- [ ] Implement IncrementMessages() method
- [ ] Deprecate TotalTokensUsed field (keep for backward compatibility)
- [ ] Run tests (GREEN)
- [ ] References: Requirements 1.1, 5.7

---

### Phase 2: State File Management

#### 2.1. Create State File Types
**References**: Requirements 3.1-3.12, 7.1-7.8, Design Phase 2

- [ ] Create directory `internal/services/sandbox_service/state_files/`
- [ ] Create `types.go` with comprehensive Go docs
  - StateFile struct with Version, LastSyncedAt, Files
  - FileEntry struct with Path, Checksum, Size, ModifiedAt
  - StateDiff struct with FilesToDownload, FilesToDelete, FilesToSkip
  - SyncStats struct with FilesDownloaded, FilesDeleted, FilesSkipped, BytesTransferred, Duration, Errors
- [ ] No test file needed (types tested through usage)
- [ ] References: Requirements 3.1-3.12, 7.1-7.8

#### 2.2. Implement State File Reader
**References**: Requirements 3.1-3.12, 7.4-7.5, Design Phase 2

- [ ] Create `internal/services/sandbox_service/state_files/reader.go`
- [ ] Write tests first (RED):
  - Test ReadLocalStateFile with existing file
  - Test ReadLocalStateFile with missing file (returns nil, no error)
  - Test ReadLocalStateFile with corrupted JSON (returns error)
  - Test ReadS3StateFile with same scenarios
  - Test ParseStateFile with valid JSON
  - Test ParseStateFile with invalid version
- [ ] Implement ReadLocalStateFile(ctx, sandboxInfo, volumePath) function
  - Execute cat command in sandbox to read .sandbox-state
  - Return nil if file doesn't exist (not an error)
  - Return error if file is corrupted
- [ ] Implement ReadS3StateFile(ctx, sandboxInfo, s3MountPath) function
  - Similar logic but reads from S3 mount path
- [ ] Implement ParseStateFile(data []byte) function
  - Unmarshal JSON
  - Validate version compatibility
- [ ] Run tests (GREEN)
- [ ] Refactor and ensure all error paths are tested
- [ ] References: Requirements 3.1-3.12, 7.4-7.5

#### 2.3. Implement State File Writer
**References**: Requirements 3.9, 7.8, Design Phase 2

- [ ] Create `internal/services/sandbox_service/state_files/writer.go`
- [ ] Write tests first (RED):
  - Test WriteLocalStateFile creates file
  - Test WriteLocalStateFile updates LastSyncedAt
  - Test WriteLocalStateFile is atomic (temp file + rename)
  - Test WriteS3StateFile with same scenarios
  - Test GenerateStateFile scans directory correctly
  - Test GenerateStateFile calculates MD5 checksums
- [ ] Implement WriteLocalStateFile(ctx, sandboxInfo, volumePath, stateFile) function
  - Update LastSyncedAt to current time
  - Marshal to JSON
  - Write to .sandbox-state.tmp
  - Rename to .sandbox-state (atomic)
- [ ] Implement WriteS3StateFile with similar logic
- [ ] Implement GenerateStateFile(ctx, sandboxInfo, directoryPath) function
  - Execute find command to list all files
  - Calculate MD5 checksums (excluding .sandbox-state itself)
  - Get file sizes and modification times
  - Build StateFile struct
- [ ] Run tests (GREEN)
- [ ] References: Requirements 3.9, 7.8

#### 2.4. Implement State File Comparator
**References**: Requirements 3.3-3.7, Design Phase 2

- [ ] Create `internal/services/sandbox_service/state_files/comparator.go`
- [ ] Write tests first (RED):
  - Test CompareStateFiles with files only in S3 (should download)
  - Test CompareStateFiles with matching checksums (should skip)
  - Test CompareStateFiles with different checksums (should download)
  - Test CompareStateFiles with files only local (should delete)
  - Test CompareStateFiles with empty states
  - Test CheckIfStale with nil state (returns true)
  - Test CheckIfStale with fresh state (returns false)
  - Test CheckIfStale with old state (returns true)
- [ ] Implement CompareStateFiles(localState, s3State) function
  - Build maps for O(1) lookups
  - Identify files to download (in S3 but not local OR different checksum)
  - Identify files to skip (matching checksum)
  - Identify files to delete (in local but not in S3)
  - Return StateDiff
- [ ] Implement CheckIfStale(stateFile, thresholdSeconds) function
  - Return true if LastSyncedAt is 0 or older than threshold
- [ ] Run tests (GREEN)
- [ ] References: Requirements 3.3-3.7

#### 2.5. Create State File Integration Tests
**References**: Requirements 10.2, Design Phase 2

- [ ] Create `internal/services/sandbox_service/state_files/state_files_test.go`
- [ ] Write integration tests that use actual Modal sandbox (if configured):
  - Test complete cycle: generate → write → read → compare
  - Test with empty directories
  - Test with large file counts
  - Test concurrent reads (thread safety)
- [ ] Tests should skip if Modal not configured (like existing sandbox tests)
- [ ] Run tests (GREEN)
- [ ] References: Requirements 10.2

---

### Phase 3: Enhanced Modal Integration - State-Based Sync

#### 3.1. Update Storage Types
**References**: Requirements 3.10, Design Phase 3

- [ ] Read existing `internal/integrations/modal/storage.go`
- [ ] Write test for updated SyncStats (RED)
  - Test SyncStats with new fields populated
- [ ] Update SyncStats struct:
  - Rename FilesProcessed to FilesDownloaded (breaking change, needs migration)
  - Add FilesDeleted int field
  - Add FilesSkipped int field
- [ ] Run tests (GREEN)
- [ ] Update all usages of FilesProcessed in existing code
- [ ] References: Requirements 3.10

#### 3.2. Implement State-Based Sync Operations
**References**: Requirements 3.1-3.12, Design Phase 3

- [ ] Create `internal/integrations/modal/storage_state.go`
- [ ] Write tests first (RED):
  - Test InitVolumeFromS3WithState with new sandbox (full sync)
  - Test InitVolumeFromS3WithState with existing state (incremental)
  - Test SyncVolumeToS3WithState creates timestamped version
  - Test SyncVolumeToS3WithState updates state files
  - Test executeSyncActions downloads files correctly
  - Test executeSyncActions deletes files correctly
  - Test executeSyncActions skips unchanged files
- [ ] Implement InitVolumeFromS3WithState(ctx, sandboxInfo) function
  - Read local and S3 state files
  - Generate S3 state if missing
  - Compare states
  - Execute sync actions
  - Update local state file
  - Return enhanced SyncStats
- [ ] Implement SyncVolumeToS3WithState(ctx, sandboxInfo) function
  - Generate state from local volume
  - Execute AWS CLI sync to timestamped S3 path
  - Write state files to both local and S3
  - Return enhanced SyncStats
- [ ] Implement executeSyncActions(ctx, sandboxInfo, diff) helper
  - Build download commands for FilesToDownload
  - Build delete commands for FilesToDelete
  - Execute in sandbox
  - Track stats
- [ ] Implement executeAWSSync(ctx, sandboxInfo, sourcePath, s3Path) helper
  - Build AWS CLI sync command
  - Execute with secrets
  - Parse output for stats
- [ ] Run tests (GREEN)
- [ ] References: Requirements 3.1-3.12

#### 3.3. Create Storage State Integration Tests
**References**: Requirements 10.2, Design Phase 3

- [ ] Create `internal/integrations/modal/storage_state_test.go`
- [ ] Write integration tests (skip if Modal not configured):
  - Test complete cold start cycle with state files
  - Test incremental sync only downloads changed files
  - Test files deleted locally when removed from S3
  - Test concurrent sync operations (locking)
  - Test S3 timestamp versioning
- [ ] Verify state files are created and maintained correctly
- [ ] Run tests (GREEN)
- [ ] References: Requirements 10.2

---

### Phase 4: Enhanced Modal Integration - Token Tracking

#### 4.1. Add Token Tracking to Claude Execution
**References**: Requirements 5.4-5.6, Design Phase 4

- [ ] Read existing `internal/integrations/modal/claude.go`
- [ ] Write tests first (RED):
  - Test ClaudeProcess token fields are initialized to 0
  - Test parseTokenSummary extracts tokens from final summary event
  - Test isFinalSummary identifies summary events
  - Test StreamClaudeOutput updates ClaudeProcess tokens
  - Mock Claude output with final summary event
- [ ] Add InputTokens, OutputTokens, CacheTokens int64 fields to ClaudeProcess struct
- [ ] Create TokenUsage struct with token fields
- [ ] Implement isFinalSummary(line string) function
  - Check for summary event marker in Claude JSON output
- [ ] Implement parseTokenSummary(line string) function
  - Parse JSON line for token usage fields
  - Extract input_tokens, output_tokens, cache_read_tokens/cache_tokens
  - Return TokenUsage struct
- [ ] Update StreamClaudeOutput to call these functions
  - Parse final summary event
  - Update ClaudeProcess token fields
- [ ] Run tests (GREEN)
- [ ] References: Requirements 5.4-5.6

---

### Phase 5: Lifecycle Hook System

#### 5.1. Create Lifecycle Hook Types
**References**: Requirements 6.1-6.2, Design Phase 5

- [ ] Create directory `internal/services/sandbox_service/lifecycle/`
- [ ] Create `types.go` with comprehensive Go docs
  - HookFunc type signature: func(context.Context, *HookData) error
  - HookData struct with ConversationID, SandboxInfo, Message, TokenUsage fields
  - TokenUsage struct with token fields
  - LifecycleHooks struct with OnColdStart, OnMessage, OnStreamFinish, OnTerminate
- [ ] No test file needed (types tested through usage)
- [ ] References: Requirements 6.1-6.2

#### 5.2. Implement Hook Executor
**References**: Requirements 6.3-6.4, 9.1-9.3, Design Phase 5

- [ ] Create `internal/services/sandbox_service/lifecycle/executor.go`
- [ ] Write tests first (RED):
  - Test ExecuteHook with nil hook (returns nil)
  - Test ExecuteHook calls hook and logs
  - Test ExecuteHook propagates hook errors
  - Test ExecuteHook logs duration
  - Mock hook functions for testing
- [ ] Implement ExecuteHook(ctx, hookName, hook, hookData) function
  - Return nil if hook is nil
  - Log execution start
  - Call hook function
  - Log duration and result
  - Return error from hook (hook decides criticality)
- [ ] Run tests (GREEN)
- [ ] References: Requirements 6.3-6.4, 9.1-9.3

#### 5.3. Implement Default Hook Implementations
**References**: Requirements 6.7, 9.1-9.3, Design Phase 5

- [ ] Create `internal/services/sandbox_service/lifecycle/defaults.go`
- [ ] Write tests first (RED):
  - Test DefaultOnColdStart returns errors (critical)
  - Test DefaultOnColdStart calls InitVolumeFromS3WithState
  - Test DefaultOnColdStart skips if no S3Config
  - Test DefaultOnMessage swallows errors (non-critical)
  - Test DefaultOnMessage saves to DynamoDB
  - Test DefaultOnMessage updates conversation stats
  - Test DefaultOnStreamFinish swallows errors (non-critical)
  - Test DefaultOnStreamFinish syncs to S3
  - Test DefaultOnStreamFinish updates conversation token stats
  - Test DefaultOnTerminate returns nil
  - Mock DynamoDB and Modal calls
- [ ] Implement DefaultOnColdStart(ctx, hookData) function
  - Return nil if no S3Config
  - Call modal.Client().InitVolumeFromS3WithState()
  - Return error (critical - propagates to caller)
- [ ] Implement DefaultOnMessage(ctx, hookData) function
  - Save message to DynamoDB
  - Update conversation stats (messages_exchanged)
  - Log errors but return nil (swallow errors)
- [ ] Implement DefaultOnStreamFinish(ctx, hookData) function
  - Sync to S3 if configured
  - Update conversation stats with tokens
  - Log errors but return nil (swallow errors)
- [ ] Implement DefaultOnTerminate(ctx, hookData) function
  - No-op for now, returns nil
- [ ] Run tests (GREEN)
- [ ] References: Requirements 6.7, 9.1-9.3

---

### Phase 6: Sandbox Service Integration

#### 6.1. Add Hooks to Sandbox Templates
**References**: Requirements 6.1, 6.7, Design Phase 6

- [ ] Read existing `internal/services/sandbox_service/templates.go`
- [ ] Write test for SandboxTemplate with hooks (RED)
  - Test GetSandboxTemplate returns template with hooks
  - Test hooks are properly initialized
- [ ] Add Hooks *lifecycle.LifecycleHooks field to SandboxTemplate struct
- [ ] Update getClaudeCodeTemplate to register default hooks:
  - OnColdStart: lifecycle.DefaultOnColdStart
  - OnMessage: lifecycle.DefaultOnMessage
  - OnStreamFinish: lifecycle.DefaultOnStreamFinish
  - OnTerminate: lifecycle.DefaultOnTerminate
- [ ] Run tests (GREEN)
- [ ] References: Requirements 6.1, 6.7

#### 6.2. Implement Lifecycle Coordination
**References**: Requirements 6.3-6.6, Design Phase 6

- [ ] Create `internal/services/sandbox_service/lifecycle.go`
- [ ] Write tests first (RED):
  - Test ExecuteColdStartHook returns errors
  - Test ExecuteColdStartHook skips if no hook registered
  - Test ExecuteMessageHook ignores return value
  - Test ExecuteStreamFinishHook ignores return value
  - Test ExecuteTerminateHook returns errors
  - Mock hooks for testing
- [ ] Implement ExecuteColdStartHook(ctx, conversationID, sandboxInfo, template) function
  - Return nil if no hook
  - Build HookData
  - Call lifecycle.ExecuteHook
  - Return error (critical)
- [ ] Implement ExecuteMessageHook(ctx, conversationID, sandboxInfo, template, message) function
  - Return if no hook
  - Build HookData with message
  - Call lifecycle.ExecuteHook
  - Ignore return value (hook swallows errors)
- [ ] Implement ExecuteStreamFinishHook(ctx, conversationID, sandboxInfo, template, tokenUsage) function
  - Return if no hook
  - Build HookData with tokenUsage
  - Call lifecycle.ExecuteHook
  - Ignore return value (hook swallows errors)
- [ ] Implement ExecuteTerminateHook(ctx, conversationID, sandboxInfo, template) function
  - Return nil if no hook
  - Build HookData
  - Call lifecycle.ExecuteHook
  - Return error if hook determines it's critical
- [ ] Run tests (GREEN)
- [ ] References: Requirements 6.3-6.6

---

### Phase 7: Conversation Service Layer

#### 7.1. Create Conversation Service Core
**References**: Requirements 1.1-1.6, 4.1-4.10, Design Phase 7

- [ ] Create directory `internal/services/conversation_service/`
- [ ] Create `conversation_service.go`
- [ ] Write tests first (RED):
  - Test NewConversationService creates service
  - Test GetOrCreateConversation retrieves existing
  - Test GetOrCreateConversation creates new with initialized stats
  - Test EnsureSandbox with existing sandbox (reconstructs)
  - Test EnsureSandbox creates new sandbox
  - Test EnsureSandbox runs OnColdStart hook
  - Test EnsureSandbox cleans up on OnColdStart failure
  - Test EnsureSandbox links sandbox to conversation
  - Mock sandbox creation and database operations
- [ ] Implement ConversationService struct with sandboxService field
- [ ] Implement NewConversationService() constructor
- [ ] Implement GetOrCreateConversation(ctx, conversationID, accountID, agentID) function
  - Try Get existing conversation
  - If not found, create new with initialized stats (all counters = 0)
  - Save and return
- [ ] Implement EnsureSandbox(ctx, conv, provider) function
  - If conv has sandbox_id, reconstruct SandboxInfo
  - If no sandbox, create new using template
  - Run OnColdStart hook (CRITICAL - return error if fails)
  - Clean up sandbox on failure
  - Save sandbox to database
  - Link sandbox to conversation
  - Return sandboxInfo, template, error
- [ ] Run tests (GREEN)
- [ ] References: Requirements 1.1-1.6, 4.1-4.10

#### 7.2. Implement Message Management
**References**: Requirements 2.1-2.6, Design Phase 7

- [ ] Create `internal/services/conversation_service/messages.go`
- [ ] Write tests first (RED):
  - Test SaveUserMessage creates message with ROLE_USER
  - Test SaveUserMessage calls OnMessage hook
  - Test SaveAssistantMessage creates message with ROLE_ASSISTANT
  - Test SaveAssistantMessage sets token count
  - Test SaveAssistantMessage calls OnMessage hook
  - Mock DynamoDB and hook execution
- [ ] Define constants: ROLE_USER = 1, ROLE_ASSISTANT = 2, ROLE_TOOL = 3
- [ ] Implement SaveUserMessage(ctx, conversationID, sandboxInfo, template, prompt) function
  - Create Message with ROLE_USER, Body=prompt, Tokens=0
  - Call sandboxService.ExecuteMessageHook (non-critical)
  - Return message
- [ ] Implement SaveAssistantMessage(ctx, conversationID, sandboxInfo, template, response, tokens) function
  - Create Message with ROLE_ASSISTANT, Body=response, Tokens=tokens
  - Call sandboxService.ExecuteMessageHook (non-critical)
  - Return message
- [ ] Run tests (GREEN)
- [ ] References: Requirements 2.1-2.6

#### 7.3. Implement Streaming Coordination
**References**: Requirements 5.1-5.10, Design Phase 7

- [ ] Create `internal/services/conversation_service/streaming.go`
- [ ] Write tests first (RED):
  - Test StreamClaudeWithHooks saves user message
  - Test StreamClaudeWithHooks executes Claude streaming
  - Test StreamClaudeWithHooks calls OnStreamFinish hook
  - Test StreamClaudeWithHooks continues if message save fails
  - Mock streaming and hook execution
- [ ] Implement StreamClaudeWithHooks(ctx, conversationID, sandboxInfo, template, prompt, responseWriter) function
  - Call SaveUserMessage (non-critical if fails)
  - Build ClaudeExecConfig
  - Call sandboxService.ExecuteClaudeStream
  - TODO: Capture response body during streaming (future task)
  - TODO: Extract token usage from ClaudeProcess (requires refactor)
  - Build TokenUsage struct (placeholder for now)
  - Call ExecuteStreamFinishHook (non-critical)
  - Return error only if streaming fails
- [ ] Run tests (GREEN)
- [ ] Add TODO comments for response capture and token extraction
- [ ] References: Requirements 5.1-5.10

---

### Phase 8: Controller Layer - Streaming Endpoint

#### 8.1. Create Conversation Streaming Controller
**References**: Requirements 1.2-1.4, 2.2, Design Phase 8

- [ ] Create `internal/controllers/conversations/streaming.go`
- [ ] Write tests first (RED):
  - Test streamClaude with valid request
  - Test streamClaude creates conversation if not exists
  - Test streamClaude ensures sandbox exists
  - Test streamClaude calls StreamClaudeWithHooks
  - Test streamClaude handles invalid request
  - Test streamClaude handles missing prompt
  - Test streamClaude handles OnColdStart failure
  - Mock HTTP request/response
- [ ] Define StreamRequest struct with Prompt, Provider, AgentID fields
- [ ] Implement streamClaude(w, req) function
  - Get user session and accountID
  - Parse conversationID and sandboxID from URL params
  - Parse StreamRequest from JSON body
  - Validate prompt not empty
  - Default provider to PROVIDER_CLAUDE_CODE if not set
  - Initialize ConversationService
  - Call GetOrCreateConversation
  - Call EnsureSandbox (handles OnColdStart)
  - Call StreamClaudeWithHooks
  - Handle errors appropriately
- [ ] Run tests (GREEN)
- [ ] References: Requirements 1.2-1.4, 2.2

#### 8.2. Add Streaming Route to Conversation Controller
**References**: Requirements 1.2, Design Phase 8

- [ ] Read existing `internal/controllers/conversations/setup.go`
- [ ] Add new route in Setup function:
  - POST `/{conversationId}/sandbox/{sandboxId}` 
  - Role: ROLE_ANY_AUTHORIZED
  - Handler: streamClaude (no wrapper - handles HTTP directly)
- [ ] Write test to verify route is registered
- [ ] Run tests (GREEN)
- [ ] References: Requirements 1.2

---

### Phase 9: Refactoring and Integration

#### 9.1. Refactor ExecuteClaudeStream to Return ClaudeProcess
**References**: Requirements 5.4-5.6, Design Phase 7

- [ ] Read existing `internal/services/sandbox_service/sandbox_service.go`
- [ ] Write test for ExecuteClaudeStream returning ClaudeProcess (RED)
- [ ] Update ExecuteClaudeStream signature to return (*modal.ClaudeProcess, error)
  - Return ClaudeProcess after streaming completes
  - ClaudeProcess contains parsed token usage
- [ ] Update all callers of ExecuteClaudeStream:
  - `internal/controllers/sandboxes/claude.go` (existing endpoint)
  - `internal/services/conversation_service/streaming.go` (new endpoint)
- [ ] Update StreamClaudeWithHooks to extract tokens from returned ClaudeProcess
- [ ] Run tests (GREEN)
- [ ] References: Requirements 5.4-5.6

#### 9.2. Implement Response Capture During Streaming
**References**: Requirements 2.2, 5.2, Design Phase 7

- [ ] Update StreamClaudeOutput in `internal/integrations/modal/claude.go` to capture output
- [ ] Write test for response capture (RED)
- [ ] Add responseBuffer to ClaudeProcess or return as separate value
- [ ] Update StreamClaudeWithHooks to use captured response
- [ ] Call SaveAssistantMessage with captured response and token count
- [ ] Run tests (GREEN)
- [ ] References: Requirements 2.2, 5.2

---

### Phase 10: End-to-End Testing and Documentation

#### 10.1. Create End-to-End Integration Tests
**References**: Requirements 10.3, Design Testing Strategy

- [ ] Create `internal/services/conversation_service/conversation_service_test.go`
- [ ] Write comprehensive integration tests (skip if Modal/DB not configured):
  - Test complete new conversation flow (create → cold start → stream → finish)
  - Test existing conversation resume flow
  - Test multiple messages in same conversation
  - Test token accumulation across messages
  - Test state file updates during cold start and stream finish
  - Test OnColdStart failure prevents sandbox creation
  - Test OnMessage failure doesn't block streaming
  - Test OnStreamFinish failure doesn't affect response
  - Verify messages saved to DynamoDB
  - Verify conversation stats updated correctly
- [ ] Run tests (GREEN)
- [ ] References: Requirements 10.3

#### 10.2. Update Documentation and Examples
**References**: Design Migration Strategy

- [ ] Update `docs/CONTROLLERS.md` with new conversation streaming endpoint
  - Document request/response format
  - Document error handling
  - Provide curl examples
- [ ] Create `internal/services/sandbox_service/README.md` documenting:
  - State files subfolder structure and purpose
  - Lifecycle hooks subfolder structure and purpose
  - How to create custom hooks for new providers
  - Hook error handling patterns (critical vs non-critical)
- [ ] Add migration guide to design.md:
  - Deprecation notice for old `/sandboxes/{id}/claude` endpoint
  - Timeline for removal
  - Frontend migration instructions
- [ ] References: Design Migration Strategy

#### 10.3. Performance and Monitoring
**References**: Requirements 10.4-10.7

- [ ] Add metrics logging for all lifecycle hooks:
  - Log hook execution duration
  - Log sync statistics (files, bytes, duration)
  - Log token usage per message and conversation
  - Log OnColdStart vs incremental sync decision
- [ ] Add error logging with full context:
  - Conversation ID
  - Sandbox ID
  - Hook name
  - Error details
- [ ] Test logging output in integration tests
- [ ] References: Requirements 10.4-10.7

---

## Configuration and Environment

### Environment Variables
```bash
# Add to environment configuration
SANDBOX_SYNC_STALE_THRESHOLD=3600  # 1 hour default
```

### Configuration Updates
- [ ] Add SyncStaleThreshold to environment config if not present
- [ ] Document in environment config documentation

---

## Success Criteria Checklist

Before marking the feature complete, verify:

- [ ] All tests pass with ≥90% coverage for new code
- [ ] `#code_tools lint` reports no issues
- [ ] `#code_tools fmt` applied to all files
- [ ] State files created and maintained correctly in S3 and local
- [ ] Files deleted locally when removed from S3
- [ ] OnColdStart failure prevents sandbox creation
- [ ] OnMessage/OnStreamFinish failures logged but don't block
- [ ] Token usage tracked per message and accumulated in conversation
- [ ] Messages stored in DynamoDB with conversation linkage
- [ ] Conversation stats updated correctly
- [ ] New endpoint `/conversations/{id}/sandbox/{id}` works end-to-end
- [ ] Old endpoint `/sandboxes/{id}/claude` still works (backward compatibility)
- [ ] No files exceed 400 lines (excluding generated code)
- [ ] All Go doc comments are clear and comprehensive
- [ ] Integration tests verify actual S3 operations
- [ ] Documentation updated with examples

---

## Learnings

Use this section to capture important discoveries during implementation that may impact future tasks or design decisions.

**Format**: 
- **What I learned**: [Description of discovery]
- **Impact**: [How this affects remaining tasks or design]
- **Action needed**: [Any updates required to requirements/design]

### Example Entry:
- **What I learned**: DynamoDB requires explicit type hints for map[string]interface{} fields in ToolCall.Input
- **Impact**: Need to use custom marshaler for ToolCall struct
- **Action needed**: Add custom JSON marshaling methods to ToolCall type in task 1.1

---

## Notes

- Tasks are designed to be executed sequentially by sub-agents
- Each task includes TDD workflow: RED → GREEN → REFACTOR → COMMIT
- All tests must pass before moving to next task
- Use `#code_tools` for all testing, linting, and formatting
- Sub-agents should read all context files before starting implementation
- Main agent should track progress by marking completed tasks with ✅
- Update Learnings section if sub-agents discover important information
