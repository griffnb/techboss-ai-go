# Modal Sandbox System - Implementation Tasks

## CRITICAL: Task Execution Instructions for Main Agent

**DO NOT IMPLEMENT TASKS YOURSELF.** Your role is to delegate each task to a specialized sub-agent, one task at a time.

### Task Delegation Process

1. **Work Sequentially**: Execute tasks in order (1, 2, 3, etc.)
2. **One Task Per Sub-Agent**: Launch a new sub-agent for each individual task
3. **Complete Context**: Provide the sub-agent with ALL necessary context (see template below)
4. **Wait for Completion**: Do not move to next task until current sub-agent completes
5. **Track Progress**: Mark tasks as complete (✅) after sub-agent finishes
6. **Update Learnings**: After each task, check if sub-agent added learnings and review them

### Sub-Agent Prompt Template

When launching a sub-agent, use this template (adapt for the specific task):

```
You are a specialized coding agent implementing Task {N} for the Modal Sandbox System.

## YOUR SPECIFIC TASK
{Copy the exact task description from tasks.md, including all sub-bullets}

## REQUIREMENTS THIS TASK SATISFIES
{List the specific requirements from requirements.md that this task addresses}

## CRITICAL: READ THESE FILES FIRST
Before writing ANY code, you MUST read:
1. /Users/griffnb/projects/techboss/techboss-ai-go/AGENTS.md - Core development patterns and TDD requirements
2. /Users/griffnb/projects/techboss/techboss-ai-go/.github/instructions/go.instructions.md - Go coding standards
3. /Users/griffnb/projects/techboss/techboss-ai-go/.github/instructions/test.instructions.md - Testing patterns
4. /Users/griffnb/projects/techboss/techboss-ai-go/.agents/specs/modal-sandbox-system/requirements.md - Feature requirements
5. /Users/griffnb/projects/techboss/techboss-ai-go/.agents/specs/modal-sandbox-system/design.md - Detailed design
6. /Users/griffnb/projects/techboss/techboss-ai-go/internal/integrations/modal/client.go - Existing client pattern
7. {Any other relevant files for this specific task}

## EXISTING PATTERNS TO FOLLOW
Look at these examples for reference:
- **Integration client pattern**: /Users/griffnb/projects/techboss/techboss-ai-go/internal/integrations/cloudflare/client.go
- **Streaming SSE pattern**: /Users/griffnb/projects/techboss/techboss-ai-go/internal/services/ai_proxies/openai/client.go
- **Controller pattern**: /Users/griffnb/projects/techboss/techboss-ai-go/internal/controllers/ai/setup.go
- **Service layer pattern**: /Users/griffnb/projects/techboss/techboss-ai-go/internal/services/ai_proxies/openai/service.go
- **Integration tests**: /Users/griffnb/projects/techboss/techboss-ai-go/internal/integrations/cloudflare/*_test.go

## MANDATORY TDD WORKFLOW
You MUST follow this exact workflow:

1. **RED**: Write the test FIRST - make it fail
2. **GREEN**: Write minimal code to make test pass
3. **REFACTOR**: Clean up code while keeping tests green
4. **COMMIT**: Only commit when tests are passing

DO NOT write implementation code before writing tests!

## GO-SPECIFIC PATTERNS (CRITICAL)
- Use pointer receivers for all methods: `func (c *APIClient) Method()`
- Wrap ALL errors with context: `errors.Wrapf(err, "context about what failed")`
- First parameter is ALWAYS `ctx context.Context`
- Error is ALWAYS the last return value
- Use `types.UUID` for account IDs
- Never use `interface{}`, use `any`
- Use `tools.Empty()` to check for empty strings/values
- All exported functions/types MUST have godoc comments
- Use `#code_tools` for running tests, formatting, linting
- Use modal SDK types from `github.com/modal-labs/libmodal/modal-go`

## MODAL SDK PATTERNS (CRITICAL)
- Apps are scoped per account: `fmt.Sprintf("app-%s", accountID)`
- Volumes are created with `VolumeFromNameParams{CreateIfMissing: true}`
- Secrets retrieved with `Secrets.FromName(ctx, name, nil)`
- CloudBucketMounts require Secret parameter
- Sandboxes.Exec requires PTY for Claude: `PTY: true`
- S3 paths use timestamp versioning: `{account}/{timestamp}/files/...`

## SUCCESS CRITERIA
- ✅ Tests written FIRST and initially fail (RED)
- ✅ Minimal implementation makes tests pass (GREEN)
- ✅ Code is refactored and clean (REFACTOR)
- ✅ All tests pass when run with `#code_tools run_tests`
- ✅ Code coverage ≥90% for new code
- ✅ No linting errors (`#code_tools lint`)
- ✅ Code properly formatted (`#code_tools format`)
- ✅ All errors properly wrapped with context
- ✅ Godoc comments on all exported symbols
- ✅ Integration tests use real Modal API (skip if not configured)

## DELIVERABLES
1. Test file with comprehensive test cases
2. Implementation file with clean, idiomatic Go code
3. Confirmation that all tests pass
4. Any learnings or issues discovered (add to Learnings section in tasks.md)

## IMPORTANT NOTES
- If Modal credentials not configured, tests should skip gracefully
- Always cleanup resources in defer blocks
- Use `assert` package from lib/testtools/assert
- DO NOT ask clarifying questions - all context is provided above
- If you discover something that impacts future tasks, ADD IT TO LEARNINGS
```

### Example Sub-Agent Invocation

```typescript
await runSubagent({
  description: "Implement Task 1: Sandbox Config Types",
  prompt: `{Use the template above with specific task details}`
});
```

---

## Phase 1: Core Integration Service

### 1. Create Sandbox Configuration Types and SandboxInfo Structure
**References**: Requirements 1.1-1.9, Design Section "Sandbox Management"

- [ ] 1.1. Create `SandboxConfig` struct in `internal/integrations/modal/sandbox.go`
  - Fields: `AccountID`, `Image`, `VolumeName`, `VolumeMountPath`, `S3Config`, `Workdir`, `Secrets`, `EnvironmentVars`
  - Add godoc comments explaining each field
  - Use pointer fields where appropriate (following design spec)
  
- [ ] 1.2. Create `ImageConfig` struct
  - Fields: `BaseImage`, `DockerfileCommands`
  - Support both registry images and custom build commands
  
- [ ] 1.3. Create `S3MountConfig` struct with timestamp versioning
  - Fields: `BucketName`, `SecretName`, `KeyPrefix`, `MountPath`, `ReadOnly`, `Timestamp`
  - KeyPrefix format: `{account}/{timestamp}/folder/structure`
  
- [ ] 1.4. Create `SandboxInfo` struct
  - Fields: `SandboxID`, `Sandbox`, `Config`, `CreatedAt`, `Status`
  - Use pointer to `modal.Sandbox` from SDK
  
- [ ] 1.5. Create `SandboxStatus` type and constants
  - Constants: `SandboxStatusRunning`, `SandboxStatusTerminated`, `SandboxStatusError`

### 2. Implement CreateSandbox with Tests (TDD)
**References**: Requirements 1.1-1.9, Design Section "Sandbox Management"

- [ ] 2.1. Write test: "Basic sandbox with volume only" (RED)
  - Test in `sandbox_test.go`
  - Use `system_testing.BuildSystem()` in `init()`
  - Skip if `modal.Configured()` returns false
  - Create minimal config (alpine image, volume only)
  - Assert sandbox created successfully
  - Defer cleanup with `TerminateSandbox`
  
- [ ] 2.2. Implement `CreateSandbox` method to pass test (GREEN)
  - Create or get Modal app scoped by accountID: `app-{accountID}`
  - Build image from `ImageConfig` (registry + dockerfile commands)
  - Create or get volume with `VolumeFromNameParams{CreateIfMissing: true}`
  - Create sandbox with `Sandboxes.Create`
  - Return `SandboxInfo` with metadata
  
- [ ] 2.3. Write test: "Sandbox with custom Docker commands" (RED)
  - Test custom dockerfile commands for Claude CLI setup
  - Use DockerfileCommands: `RUN apk add --no-cache bash curl git libgcc libstdc++ ripgrep aws-cli`
  - Verify sandbox accepts and processes commands
  
- [ ] 2.4. Refactor CreateSandbox for custom images (GREEN)
  - Add conditional Dockerfile command application
  - Ensure image building works with custom commands
  
- [ ] 2.5. Write test: "Sandbox with S3 bucket mount" (RED)
  - Test S3Config with bucket, secret, and key prefix
  - Use timestamp in key prefix: `docs/{account}/{timestamp}/`
  - Verify CloudBucketMount created correctly
  
- [ ] 2.6. Implement S3 mount logic in CreateSandbox (GREEN)
  - Retrieve S3 secret with `Secrets.FromName`
  - Create CloudBucketMount with timestamp key prefix
  - Add to sandbox creation params
  
- [ ] 2.7. Write test: "Error handling for invalid configs" (RED)
  - Test missing required fields
  - Test invalid secret names
  - Verify errors properly wrapped with context
  
- [ ] 2.8. Add validation and error handling (GREEN + REFACTOR)
  - Validate required config fields
  - Wrap all errors with `errors.Wrapf` including context
  - Ensure cleanup on partial failure

### 3. Implement TerminateSandbox and GetSandboxStatus with Tests
**References**: Requirements 5.1-5.7, Design Section "Sandbox Management"

- [ ] 3.1. Write test: "Terminate sandbox successfully" (RED)
  - Create sandbox, then terminate it
  - Verify sandbox.Terminate() called
  - Check status changes to terminated
  
- [ ] 3.2. Implement `TerminateSandbox` method (GREEN)
  - Call `sb.Terminate(ctx)`
  - Handle `syncToS3` parameter (call SyncVolumeToS3 if true)
  - Wrap errors with context
  
- [ ] 3.3. Write test: "Get sandbox status" (RED)
  - Create sandbox and check status is "running"
  - Terminate and check status is "terminated"
  
- [ ] 3.4. Implement `GetSandboxStatus` method (GREEN)
  - Query sandbox state from Modal SDK
  - Return appropriate `SandboxStatus` constant
  
- [ ] 3.5. Write test: "Terminate with sync to S3" (RED)
  - Will test integration with storage layer (implement basic version now)
  
- [ ] 3.6. Refactor and add edge case tests
  - Test terminating already terminated sandbox
  - Test timeout scenarios

### 4. Implement Storage Operations (S3 Sync) with Tests
**References**: Requirements 2.1-2.8, 3.1-3.6, Design Section "Storage Operations"

- [ ] 4.1. Create `SyncStats` struct in `storage.go`
  - Fields: `FilesProcessed`, `BytesTransferred`, `Duration`, `Errors`
  
- [ ] 4.2. Write test: "InitVolumeFromS3 copies files from S3" (RED)
  - Create sandbox with S3 mount
  - Pre-populate test S3 bucket with files
  - Call InitVolumeFromS3
  - Verify files exist in volume path
  - Check SyncStats returned correctly
  
- [ ] 4.3. Implement `InitVolumeFromS3` method (GREEN)
  - Use shell command: `cp -r {s3_mount}/* {volume_mount}/`
  - Execute with `sb.Exec(ctx, cmd, &SandboxExecParams{})`
  - Parse output for stats (files count, duration)
  - Return SyncStats
  
- [ ] 4.4. Write test: "SyncVolumeToS3 creates timestamped version" (RED)
  - Create sandbox and write files to volume
  - Call SyncVolumeToS3
  - Verify new timestamped folder created in S3: `{account}/{timestamp}/`
  - Verify files synced correctly
  - Check SyncStats
  
- [ ] 4.5. Implement `SyncVolumeToS3` method (GREEN)
  - Generate unix timestamp: `time.Now().Unix()`
  - Build S3 path: `{accountID}/{timestamp}/`
  - Use AWS CLI command: `aws s3 sync {volume} s3://{bucket}/{path}/`
  - Execute with AWS secrets from `Secrets.FromName`
  - Parse output for stats
  
- [ ] 4.6. Write test: "GetLatestVersion returns most recent timestamp" (RED)
  - Create multiple timestamped versions in S3
  - Call GetLatestVersion
  - Verify returns highest timestamp
  
- [ ] 4.7. Implement `GetLatestVersion` method (GREEN)
  - Use AWS CLI: `aws s3 ls s3://{bucket}/{account}/ | sort | tail -1`
  - Parse timestamp from directory name
  - Return as int64
  
- [ ] 4.8. Refactor storage operations
  - Extract common exec patterns
  - Add timeout protection
  - Handle empty S3 buckets gracefully
  - Improve error messages

### 5. Implement Claude Execution with Tests
**References**: Requirements 4.1-4.10, Design Section "Claude Execution"

- [ ] 5.1. Create Claude types in `claude.go`
  - `ClaudeExecConfig` struct with fields: `Prompt`, `Workdir`, `OutputFormat`, `SkipPermissions`, `Verbose`, `AdditionalFlags`
  - `ClaudeProcess` struct with fields: `Process`, `Config`, `StartedAt`
  
- [ ] 5.2. Write test: "ExecClaude starts Claude process" (RED)
  - Create sandbox with Claude CLI installed (use docker commands)
  - Create ClaudeExecConfig with simple prompt
  - Call ExecClaude
  - Verify ClaudeProcess returned
  - Verify process has stdout reader
  
- [ ] 5.3. Implement `ExecClaude` method (GREEN)
  - Build Claude command array: `["claude", "-c", "-p", prompt, ...]`
  - Add flags based on config (--dangerously-skip-permissions, --verbose, --output-format)
  - Retrieve Anthropic/AWS secrets from environment config
  - Create secrets map with `Secrets.FromMap`
  - Execute with `sb.Exec(ctx, cmd, &SandboxExecParams{PTY: true, Secrets: secrets, Workdir: workdir})`
  - **CRITICAL**: `PTY: true` is required for Claude CLI
  - Return ClaudeProcess
  
- [ ] 5.4. Write test: "Claude executes with correct flags" (RED)
  - Test different config combinations
  - Verify command built correctly
  - Check PTY enabled
  
- [ ] 5.5. Write test: "WaitForClaude returns exit code" (RED)
  - Execute simple Claude command
  - Call WaitForClaude
  - Verify returns exit code 0 for success
  
- [ ] 5.6. Implement `WaitForClaude` method (GREEN)
  - Call `claudeProcess.Process.Wait(ctx)`
  - Return exit code
  - Wrap errors with context
  
- [ ] 5.7. Refactor Claude execution
  - Validate config fields
  - Add default workdir (volume mount path)
  - Improve error messages

### 6. Implement Claude Streaming Output with Tests
**References**: Requirements 4.1-4.10, Design Section "Claude Execution"

- [ ] 6.1. Write test: "StreamClaudeOutput streams to ResponseWriter" (RED)
  - Use `httptest.NewRecorder()` as ResponseWriter
  - Execute Claude with simple prompt
  - Call StreamClaudeOutput
  - Verify SSE headers set correctly
  - Verify output lines streamed
  - Verify [DONE] event sent
  
- [ ] 6.2. Implement `StreamClaudeOutput` method (GREEN)
  - Set SSE headers: `Content-Type: text/event-stream`, `Cache-Control: no-cache`, `Connection: keep-alive`
  - Get flusher: `responseWriter.(http.Flusher)`
  - Create scanner: `bufio.NewScanner(claudeProcess.Process.Stdout)`
  - Loop: scan line, write `data: {line}\n\n`, flush
  - Send completion: `data: [DONE]\n\n`
  - Handle scanner errors
  
- [ ] 6.3. Write test: "Streaming handles Claude errors" (RED)
  - Execute Claude with invalid prompt/config
  - Verify error output streamed
  - Verify proper error returned
  
- [ ] 6.4. Write test: "Streaming handles context cancellation" (RED)
  - Start streaming Claude output
  - Cancel context mid-stream
  - Verify cleanup happens gracefully
  
- [ ] 6.5. Refactor streaming implementation
  - Add timeout protection
  - Ensure cleanup in defer blocks
  - Handle connection drops gracefully
  - Follow SSE pattern from openai/client.go

### 7. Integration Tests: End-to-End Sandbox Workflow
**References**: All Phase 1 requirements

- [ ] 7.1. Write test: "Complete sandbox lifecycle with Claude" (RED)
  - Create sandbox with S3 mount
  - Initialize volume from S3 (if files exist)
  - Execute Claude with prompt
  - Stream output and verify response
  - Sync volume back to S3 with new timestamp
  - Verify new version created in S3
  - Terminate sandbox
  - Verify cleanup complete
  
- [ ] 7.2. Write test: "Multiple sandboxes for same account" (RED)
  - Create 2 sandboxes for same account ID
  - Verify both use same app and volume
  - Verify both operate independently
  - Cleanup both
  
- [ ] 7.3. Write test: "Sandbox with all configuration options" (RED)
  - Test with custom image, volume, S3, workdir, secrets, env vars
  - Verify all options applied correctly
  
- [ ] 7.4. Ensure all integration tests pass
  - Run with `#code_tools run_tests internal/integrations/modal`
  - Verify ≥90% code coverage
  - Fix any failing tests
  - Add missing edge case tests

### 8. Code Quality and Documentation
**References**: AGENTS.md requirements

- [ ] 8.1. Run linter and fix all issues
  - Use `#code_tools lint internal/integrations/modal`
  - Fix any reported issues
  
- [ ] 8.2. Run formatter
  - Use `#code_tools format internal/integrations/modal`
  - Ensure consistent formatting
  
- [ ] 8.3. Verify test coverage
  - Run tests with coverage: `#code_tools run_tests internal/integrations/modal`
  - Ensure ≥90% coverage
  - Add tests for any uncovered paths
  
- [ ] 8.4. Add/improve godoc comments
  - All exported types have clear descriptions
  - All exported functions explain parameters and return values
  - Examples where helpful
  
- [ ] 8.5. Review error messages
  - Ensure all errors have helpful context
  - Check error wrapping is consistent
  - Verify no sensitive data in errors

---

## Phase 2: Service Layer, HTTP Endpoints & Web UI ✅ COMPLETE

### 9. Implement Service Layer ✅
**References**: Design Section "Service Layer"

- [x] 9.1. Create `SandboxService` struct in `internal/services/modal/sandbox_service.go`
  - Field: `client *modal.APIClient`
  - Constructor: `NewSandboxService() *SandboxService`
  
- [ ] 9.2. Implement service methods (pass-through to integration)
  - `CreateSandbox(ctx, accountID, config) (*SandboxInfo, error)` - adds accountID to config
  - `TerminateSandbox(ctx, sandboxInfo, syncToS3) error`
  - `ExecuteClaudeStream(ctx, sandboxInfo, config, responseWriter) error` - combines Exec + Stream
  - `InitFromS3(ctx, sandboxInfo) (*SyncStats, error)`
  - `SyncToS3(ctx, sandboxInfo) (*SyncStats, error)`
  
- [ ] 9.3. Add validation logic in service layer
  - Validate config fields before passing to integration
  - Add business logic checks (quotas, permissions, etc.) - leave as TODO for now
  
- [ ] 9.4. Write unit tests for service layer (optional for Phase 2)
  - Mock Modal client
  - Test service logic

### 10. Implement Sandbox Controller Routes
**References**: Requirements 6.1-6.10, Design Section "HTTP Controllers"

- [ ] 10.1. Create route setup in `internal/controllers/sandbox/setup.go`
  - Const: `ROUTE = "sandbox"`
  - Routes: `POST /sandbox`, `GET /sandbox/{sandboxID}`, `DELETE /sandbox/{sandboxID}`, `POST /sandbox/{sandboxID}/claude`
  - Use `helpers.RoleHandler` with `ROLE_ANY_AUTHORIZED`
  - Use `StandardRequestWrapper` for CRUD, `NoTimeoutStreamingMiddleware` for Claude
  
- [ ] 10.2. Create request/response types in `sandbox.go`
  - `CreateSandboxRequest` struct
  - `CreateSandboxResponse` struct
  - Add JSON tags to all fields
  
- [ ] 10.3. Implement `createSandbox` handler
  - Get user session with `request.GetReqSession(req)`
  - Parse request body with `request.GetJSONPostData` and `request.ConvertPost`
  - Build `SandboxConfig` from request
  - Call `sandboxService.CreateSandbox`
  - Add TODO: Store sandboxInfo in database/cache
  - Return `CreateSandboxResponse` with `response.Success`
  
- [ ] 10.4. Implement `getSandbox` handler (stub)
  - Get sandboxID from URL params: `chi.URLParam(req, "sandboxID")`
  - Return error with TODO message for now
  - Phase 2 enhancement: retrieve from database
  
- [ ] 10.5. Implement `deleteSandbox` handler (stub)
  - Get sandboxID from URL params
  - Return error with TODO message for now
  - Phase 2 enhancement: retrieve from database and terminate

### 11. Implement Claude Streaming Endpoint
**References**: Requirements 6.6-6.7, 4.1-4.10, Design Section "HTTP Controllers"

- [ ] 11.1. Create request type in `claude.go`
  - `ClaudeRequest` struct with `Prompt` field
  
- [ ] 11.2. Implement `streamClaude` handler
  - Get sandboxID from URL params
  - Parse request body for prompt
  - Validate prompt not empty
  - TODO: Retrieve sandboxInfo from database/cache (stub for now)
  - Build `ClaudeExecConfig` with prompt
  - Call `sandboxService.ExecuteClaudeStream` with responseWriter
  - Handle errors appropriately (log, don't expose internal details)
  
- [ ] 11.3. Test streaming endpoint with curl or httptest
  - Verify SSE headers set
  - Verify streaming works
  - Verify [DONE] event sent

### 12. Create Simple Web UI
**References**: Requirements 7.1-7.10, Design Section "Web UI"

- [ ] 12.1. Create `static/modal-sandbox-ui.html`
  - Clean HTML5 structure
  - CSS for chat interface (embedded in `<style>`)
  - Vanilla JavaScript (no build process)
  
- [ ] 12.2. Implement sandbox creation UI
  - Button to create sandbox
  - Loading state while creating
  - Display sandbox ID when ready
  - Hide creation section after success
  
- [ ] 12.3. Implement chat interface
  - Message history display
  - Input field and send button
  - Disable input while streaming
  - Auto-scroll to latest message
  
- [ ] 12.4. Implement SSE streaming in JavaScript
  - Use `fetch()` with streaming response
  - Read response body with reader
  - Parse SSE format: `data: {content}\n\n`
  - Handle [DONE] event
  - Display streamed content in real-time
  
- [ ] 12.5. Add error handling to UI
  - Display network errors
  - Display API errors
  - Handle disconnections gracefully
  
- [ ] 12.6. Style and polish UI
  - Responsive design (mobile-friendly)
  - Clean, modern styling
  - Loading indicators
  - User feedback for all actions

### 13. End-to-End Testing and Integration
**References**: All Phase 2 requirements

- [ ] 13.1. Test complete flow manually
  - Open UI in browser
  - Create sandbox
  - Send message to Claude
  - Verify real-time streaming
  - Verify Claude responds correctly
  - Test error scenarios
  
- [ ] 13.2. Wire up controller routes in main router
  - Add `sandbox.Setup(coreRouter)` to router initialization
  - Verify routes accessible
  
- [ ] 13.3. Test with real Modal infrastructure
  - Use real S3 bucket
  - Use real Claude API
  - Verify end-to-end flow works
  
- [ ] 13.4. Load testing (optional)
  - Test multiple concurrent sandboxes
  - Test streaming under load
  - Verify cleanup works correctly

### 14. Documentation and Cleanup
**References**: All requirements

- [ ] 14.1. Update README with usage instructions
  - How to configure Modal credentials
  - How to run the application
  - How to access the UI
  
- [ ] 14.2. Add inline code comments where needed
  - Complex logic explained
  - TODO markers for future enhancements
  
- [ ] 14.3. Final code review
  - Check all requirements satisfied
  - Verify TDD followed throughout
  - Ensure code quality standards met
  
- [ ] 14.4. Create summary of what was built
  - List all features implemented
  - Note any limitations or future work
  - Document any deviations from design

---

## Learnings

### What Was Discovered During Implementation

*Sub-agents should add new learnings here as they discover information that impacts future tasks or provides valuable context for the next developer.*

**Format**:
```
**Task {N} - {Title}**: {What was learned and why it matters for future tasks}
```

**Example**:
```
**Task 2.6 - S3 Mount Implementation**: Modal's CloudBucketMount API requires the secret to have specific AWS credential keys: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, and optionally AWS_SESSION_TOKEN. This should be documented in the configuration setup.
```

### Actual Learnings

**Task 9 - Service Layer Implementation**: The Modal integration has a generated mock implementation in `internal/integrations/modal/x_gen_mock.go` that provides `MockAPIClient` and `APIClientInterface`. The service layer currently uses the concrete `*modal.APIClient` type, but should be refactored to use `modal.APIClientInterface` for better testability. This would allow unit tests to mock the Modal client without requiring real Modal infrastructure. This is a future enhancement that would improve test isolation and speed.

**Task 10 - Controller Routes**: The controller successfully follows the established patterns from `docs/CONTROLLERS.md`. The `testing_service.New()` helper provides clean test fixtures for controller testing. Stub implementations for GET and DELETE endpoints work well with TODO comments for Phase 2 database integration.

**Task 13 - End-to-End Testing**: Discovered that end-to-end testing was blocked by lack of database persistence for sandbox retrieval (Claude streaming endpoint needs to retrieve sandbox by ID). Implemented in-memory cache (`sync.Map`) as a temporary solution to enable Phase 1 testing. This allows full end-to-end testing without implementing the complete Phase 2 database layer. The cache is thread-safe and persists for the session lifetime. All CRUD operations (create, get, delete) and Claude streaming now work with the cache. This is documented as a temporary solution with TODO comments for Phase 2 database implementation. Modal integration tests pass successfully, confirming Modal is properly configured. Router integration (Task 13.2) was already complete from previous tasks. System is now ready for manual testing with the UI.

---

## Future Enhancements (Not in Current Scope)

- Database persistence for sandbox metadata
- Sandbox lifecycle management (automatic timeouts)
- Usage tracking and billing integration
- Multi-user collaboration (shared sandboxes)
- WebSocket for bidirectional communication
- File upload/download endpoints
- Sandbox snapshot/restore functionality
- GPU-enabled sandboxes
- Custom Docker image registry integration
