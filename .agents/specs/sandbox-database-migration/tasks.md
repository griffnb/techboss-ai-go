# Sandbox Database Migration - Implementation Tasks

## CRITICAL: Task Execution Instructions for Main Agent

**DO NOT IMPLEMENT TASKS YOURSELF.** Your role is to delegate each task to a specialized sub-agent, one task at a time.

### Task Delegation Process

1. **Work Sequentially**: Execute tasks in the order listed below
2. **One Task Per Sub-Agent**: Launch a new sub-agent for each individual task
3. **Complete Context**: Provide the sub-agent with ALL necessary context including:
   - The specific task description from this file
   - Requirements this task satisfies (from requirements.md)
   - All necessary file paths to read (AGENTS.md, design.md, requirements.md, existing code)
   - Existing patterns to follow (reference sandbox.go, other model files)
   - The mandatory TDD workflow (RED → GREEN → REFACTOR → COMMIT)
   - Success criteria (tests pass, code quality, coverage, documentation)
4. **Wait for Completion**: Do not move to the next task until the current sub-agent completes
5. **Track Progress**: Mark tasks as complete (✓) after the sub-agent finishes
6. **Review Output**: Verify tests pass and code quality before proceeding

### Sub-Agent Prompt Template

When launching a sub-agent, use this template structure:

```
You are implementing task [X] for the Sandbox Database Migration feature.

**Task**: [Copy exact task description]

**Requirements**: This task satisfies:
- [List specific requirement numbers from requirements.md]

**Context Files to Read**:
1. /Users/griffnb/projects/techboss/techboss-ai-go/AGENTS.md - Core Go development patterns
2. /Users/griffnb/projects/techboss/techboss-ai-go/.claude/skills/testing/SKILL.md - TDD patterns and testing requirements
3. /Users/griffnb/projects/techboss/techboss-ai-go/.agents/specs/sandbox-database-migration/design.md - Technical design
4. /Users/griffnb/projects/techboss/techboss-ai-go/.agents/specs/sandbox-database-migration/requirements.md - Feature requirements
5. /Users/griffnb/projects/techboss/techboss-ai-go/internal/models/sandbox/sandbox.go - Existing sandbox model
6. [Additional relevant files based on task]

**Existing Patterns to Follow**:
- Use pointer fields for all model fields (e.g., `AccountID *fields.UUIDField`)
- Use `errors.Wrapf()` for error wrapping with context
- Use `log.ErrorContext(err, ctx)` before returning errors
- Use `response.AdminBadRequestError`, `response.AdminNotFoundError`, etc.
- Controller methods return `(*ModelType, int, error)`
- Use `testing_service.BuildSystem()` in test `init()`
- Use `testing_service.NewGETRequest/NewPOSTRequest` for controller tests
- Use `req.WithAccount()` for auth endpoint tests (not WithAdmin)
- Use `defer testtools.CleanupModel(obj)` to clean up test data
- Extend `testing_service.Builder` if new test fixtures needed

**TDD Workflow (MANDATORY)**:
1. 🔴 RED: Write test FIRST (see it fail)
2. 🟢 GREEN: Write minimal code to make it pass
3. 🔵 REFACTOR: Clean up while keeping tests green
4. 📝 COMMIT: Commit only when tests pass

**Critical Patterns**:
- MetaData uses `*int64` for LastS3Sync (unix timestamp)
- ExternalID column stores Modal sandbox ID (not in metadata)
- Status column uses base.Structure.Status (constants.STATUS_ACTIVE, etc.)
- Provider field uses PROVIDER_CLAUDE_CODE enum
- All database queries must include AccountID for multi-tenancy
- Never use `sandboxCache` - always query database
- Use `#code_tools` for running tests: `#code_tools run_tests <package> <test_name>`

**Success Criteria**:
- All tests pass (≥90% coverage)
- Tests written BEFORE implementation
- Code follows Go idioms from AGENTS.md
- Error handling uses errors.Wrapf with context
- Proper logging on all error paths
- No linter errors
- Documentation updated if adding new patterns

**Deliverables**:
[Specify exact files to create/modify and what should be in them]

Read ALL context files listed above BEFORE starting implementation. If you discover new information that impacts future tasks, add it to the Learnings section at the bottom of tasks.md.

Do you understand the task? Begin by reading the context files, then proceed with TDD.
```

### Example Sub-Agent Invocation

```typescript
await runSubagent({
  description: "Update MetaData structure",
  prompt: `You are implementing task 1.1 for the Sandbox Database Migration feature.

**Task**: Update meta_data.go with simplified MetaData structure using unix timestamp for LastS3Sync

**Requirements**: This task satisfies requirement 7.1 (MetaData structure design)

**Context Files to Read**:
1. /Users/griffnb/projects/techboss/techboss-ai-go/AGENTS.md
2. /Users/griffnb/projects/techboss/techboss-ai-go/.claude/skills/testing/SKILL.md
3. /Users/griffnb/projects/techboss/techboss-ai-go/.agents/specs/sandbox-database-migration/design.md
4. /Users/griffnb/projects/techboss/techboss-ai-go/internal/models/sandbox/meta_data.go

**Existing Patterns**: [list patterns as in template]
**TDD Workflow**: [as in template]
**Success Criteria**: [as in template]

Read the context files and implement the new MetaData structure with tests.`
});
```

---

## Implementation Tasks

### Phase 1: Model Layer Updates

- [ ] 1.1. Update `internal/models/sandbox/meta_data.go` with simplified MetaData structure
  - Change LastS3Sync to `*int64` (unix timestamp)
  - Add SyncStats struct with FilesProcessed, BytesTransferred, DurationMs
  - Add UpdateLastSync() helper method
  - Write unit tests for UpdateLastSync()
  - Satisfies: Requirements 1.2, 7.1

- [ ] 1.2. Add database query functions to `internal/models/sandbox/queries.go`
  - Implement FindByExternalID(ctx, externalID, accountID) with account scoping
  - Implement FindAllByAccount(ctx, accountID) with deleted/disabled filtering
  - Implement CountByAccount(ctx, accountID)
  - Write table-driven tests for all query functions
  - Test ownership verification (wrong accountID returns empty)
  - Satisfies: Requirements 2.1, 2.4, 5.1

- [ ] 1.3. Write comprehensive model tests in `internal/models/sandbox/sandbox_test.go`
  - Test SaveWithMetaData creates record with minimal metadata
  - Test FindByExternalID with correct/incorrect accountID
  - Test FindAllByAccount excludes deleted sandboxes
  - Test MetaData UpdateLastSync updates timestamp and stats
  - Use testing_service.Builder if needed for test fixtures
  - Satisfies: Requirements 1.1, 1.3, 2.2, 2.4

### Phase 2: Service Layer Updates

- [ ] 2.1. Create `internal/services/sandbox_service/templates.go` with premade configurations
  - Define SandboxTemplate struct (Provider, ImageConfig, S3Config)
  - Implement GetSandboxTemplate(provider, agentID) for PROVIDER_CLAUDE_CODE
  - Implement BuildSandboxConfig(accountID) method on template
  - Write table-driven tests for GetSandboxTemplate (valid/unsupported providers)
  - Test that BuildSandboxConfig creates proper modal.SandboxConfig
  - Satisfies: Requirements 6.2 (templates enable frontend to create without config details)

- [ ] 2.2. Add `reconstructSandboxInfo()` helper function in `sandbox_service`
  - Create function that builds modal.SandboxInfo from Sandbox model fields
  - Use ExternalID, Provider, AccountID from model
  - Build SandboxConfig from template based on Provider
  - Write unit tests for reconstruction with different providers
  - Satisfies: Requirements 2.5

### Phase 3: Controller Layer Updates

- [ ] 3.1. Update `internal/controllers/sandboxes/sandbox.go` - createSandbox function
  - Add CreateSandboxRequest struct (Provider, AgentID fields only)
  - Call GetSandboxTemplate(provider, agentID) to get premade config
  - Call BuildSandboxConfig(accountID) from template
  - Create Modal sandbox via service.CreateSandbox()
  - Save to database with ExternalID, Provider, AgentID, Status, empty MetaData
  - Write controller test using testing_service.NewPOSTRequest
  - Use req.WithAccount() for auth endpoint
  - Test successful creation and DB save failure scenarios
  - Cleanup Modal sandbox in defer/cleanup
  - Satisfies: Requirements 1.1, 1.4, 1.5, 6.2, 6.3

- [ ] 3.2. Update `internal/controllers/sandboxes/sandbox.go` - authDelete function
  - Query database by ID (auth framework verifies ownership)
  - Call reconstructSandboxInfo() to build SandboxInfo
  - Call service.TerminateSandbox(sandboxInfo, syncToS3=true)
  - Update Status to STATUS_TERMINATED
  - Set Deleted=1 for soft delete
  - Save model
  - Write controller test using testing_service.NewDELETERequest
  - Test Modal termination failure continues with soft delete
  - Satisfies: Requirements 4.1, 4.2, 4.3, 4.4, 4.5

- [ ] 3.3. Update `internal/controllers/sandboxes/sandbox.go` - syncSandbox function
  - Add SyncSandboxRequest struct (no params needed)
  - Query database by ID (auth framework verifies ownership)
  - Call reconstructSandboxInfo() to build SandboxInfo
  - Call service.SyncToS3() with SandboxInfo
  - Update MetaData using UpdateLastSync(filesProcessed, bytes, duration)
  - Save model with updated metadata
  - Write controller test verifying metadata updates
  - Satisfies: Requirements 3.4

- [ ] 3.4. Update `internal/controllers/sandboxes/claude.go` - streamClaude function
  - Query database using Get(ctx, sandboxID) with ownership verification
  - Call reconstructSandboxInfo() to build SandboxInfo
  - Pass SandboxInfo to service.ExecuteClaudeStream()
  - Remove all sandboxCache.Load() calls
  - Write controller test for streaming with owned/unowned sandbox
  - Test 404 when sandbox doesn't exist
  - Satisfies: Requirements 2.1, 2.4, 2.5

### Phase 4: Cache Removal

- [ ] 4.1. Remove sandboxCache from `internal/controllers/sandboxes/sandbox.go`
  - Delete `var sandboxCache sync.Map` declaration
  - Remove all sandboxCache.Store() calls (should be replaced by DB saves)
  - Remove all sandboxCache.Load() calls (should be replaced by DB queries)
  - Remove all sandboxCache.Delete() calls (should be replaced by soft delete)
  - Grep for "sandboxCache" to ensure no references remain
  - Run all tests to verify nothing broke
  - Satisfies: Core migration goal

- [ ] 4.2. Remove Phase 2 TODO comments from codebase
  - Search for "Phase 2" or "TODO.*database" comments in sandbox controllers
  - Remove or update comments that reference cache migration
  - Add comments explaining database-backed implementation where helpful
  - Satisfies: Requirements cleanup

### Phase 5: Frontend Updates

- [ ] 5.1. Update `static/modal-sandbox-ui.html` - Add sandbox list loading
  - Add loadSandboxList() function that calls GET /sandbox
  - Parse response and store sandboxes array
  - Display sandbox list in UI with ID, Status, CreatedAt
  - Handle empty list case with "No sandboxes available" message
  - Add error handling for failed API calls
  - Satisfies: Requirements 5.1, 5.2, 5.4

- [ ] 5.2. Update `static/modal-sandbox-ui.html` - Add sandbox selection handlers
  - Add click handlers for sandbox list items
  - Verify sandbox status is active before allowing chat
  - Update currentSandboxID with selected sandbox's database ID (not external_id)
  - Open chat interface when valid sandbox selected
  - Show warning if sandbox is terminated
  - Satisfies: Requirements 5.3, 5.5

- [ ] 5.3. Update `static/modal-sandbox-ui.html` - Add create new sandbox button
  - Add "Create New Sandbox" button to UI
  - Add click handler that calls POST /sandbox with provider type
  - Show loading indicator during creation
  - Auto-select new sandbox and open chat on success
  - Display error message on failure
  - Satisfies: Requirements 6.1, 6.2, 6.3, 6.4, 6.5

### Phase 6: Integration Testing & Validation

- [ ] 6.1. Write end-to-end integration tests
  - Test full flow: create → save to DB → retrieve → chat → delete
  - Test multi-user isolation (user1 cannot access user2's sandboxes)
  - Test server restart scenario (mock by clearing cache, verify DB retrieval)
  - Test concurrent sandbox creation by same user
  - Satisfies: Requirements 1.1, 2.4, 4.1-4.5

- [ ] 6.2. Run full test suite and verify coverage
  - Run `#code_tools run_tests ./internal/models/sandbox` 
  - Run `#code_tools run_tests ./internal/controllers/sandboxes`
  - Run `#code_tools run_tests ./internal/services/sandbox_service`
  - Verify ≥90% coverage for new code
  - Fix any failing tests
  - Satisfies: TDD requirements

- [ ] 6.3. Manual testing of UI flows
  - Test: Load page → see sandbox list
  - Test: Click sandbox → chat opens
  - Test: Create new → auto-selects → chat opens
  - Test: Delete sandbox → removed from list
  - Test: Refresh page → sandboxes persist
  - Satisfies: Requirements 5.1-5.5, 6.1-6.5

- [ ] 6.4. Code quality checks
  - Run `#code_tools lint ./internal/models/sandbox`
  - Run `#code_tools lint ./internal/controllers/sandboxes`
  - Run `#code_tools fmt` on all modified files
  - Fix any linter warnings
  - Satisfies: Code quality standards

### Phase 7: Documentation & Cleanup

- [ ] 7.1. Update model documentation
  - Add godoc comments to new query functions
  - Document MetaData structure and unix timestamp usage
  - Document UpdateLastSync() behavior
  - Satisfies: Documentation standards

- [ ] 7.2. Update controller documentation
  - Document new createSandbox request structure
  - Document template-based configuration approach
  - Add examples of database queries in controller comments
  - Satisfies: Documentation standards

- [ ] 7.3. Final verification
  - Verify no sandboxCache references remain
  - Verify all Phase 2 TODOs removed
  - Verify all tests pass
  - Verify UI works end-to-end
  - Ready for PR/merge

---

## Progress Tracking

After each sub-agent completes a task:

1. Mark the task as complete: `- [x] Task description`
2. Note any issues or deviations in comments below the task
3. Update the Learnings section if new information was discovered
4. Verify tests pass before moving to next task

---

## Learnings

**Purpose**: Capture new information discovered during implementation that may impact future tasks.

**Format**: 
```
### Task X.Y - [Task Name]
**What was learned**: [Description]
**Impact on future tasks**: [Which tasks affected]
**Action taken**: [How design/implementation was adjusted]
```

### Example:
```
### Task 1.2 - Database Query Functions
**What was learned**: The base.Structure already provides a Status column, so we don't need to duplicate it in MetaData
**Impact on future tasks**: Tasks 3.1-3.4 should use model.Status.Set() instead of MetaData.Status
**Action taken**: Updated design.md to clarify Status column usage
```

---

## Notes

- All tasks must follow TDD: write tests FIRST, then implementation
- Use `#code_tools` for all test execution and linting
- Each sub-agent should be given complete context to work independently
- Main agent should verify tests pass after each task before proceeding
- Sub-agents should add discoveries to Learnings section above
