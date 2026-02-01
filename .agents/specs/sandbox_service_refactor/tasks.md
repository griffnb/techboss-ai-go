# Sandbox Service Refactor - Implementation Tasks

## Overview

This document contains the implementation tasks for refactoring `internal/services/sandbox_service` into a stateful service with a GetOrCreate pattern. Each task is designed for autonomous execution by Ralph Wiggum with clear success criteria and automatic verification.

**Requirements Reference:** `.agents/specs/sandbox_service_refactor/requirements.md`
**Design Reference:** `.agents/specs/sandbox_service_refactor/design.md`

---

## Phase 1: Service Struct and Core Infrastructure

### Task 1: Update Service Struct with New Fields

**Requirements:** Requirement 1.4 (service state), 2.1 (constructor parameters)

**Implementation:**
1. Open `internal/services/sandbox_service/sandbox_service.go`
2. Update the `SandboxService` struct to include:
   - `sandboxInfo *modal.SandboxInfo`
   - `sandboxModel *sandbox.Sandbox`
   - `volume *modal.VolumeInfo` (NEW)
   - `account *account.Account` (replaces user)
   - `accountID types.UUID`
   - `config *modal.SandboxConfig`
   - `template *SandboxTemplate`
   - `conversationID types.UUID`
3. Remove any old fields that conflict with the new design
4. Add godoc comment explaining the stateful design and GetOrCreate pattern

**Verification:**
- Run: `make fmt`
- Expected: Code formats successfully with no errors
- Run: `make lint`
- Expected: No linting errors related to struct definition
- Check: Struct has exactly 9 fields as specified in design

**Self-Correction:**
- If compilation errors: Verify all field types are imported correctly
- If lint errors: Fix naming or add missing comments
- If wrong field count: Review design document and add/remove fields

**Completion Criteria:**
- [ ] SandboxService struct matches design document exactly
- [ ] All fields have correct types
- [ ] Struct has godoc comment
- [ ] Code compiles without errors
- [ ] No lint warnings

**Escape Condition:** If stuck after 3 attempts, document the blocker in the struct definition and move to next task.

---

### Task 2: Implement validateParams Helper

**Requirements:** Requirement 9.1, 9.2 (error handling and validation)

**Implementation:**
1. Create private function `validateParams(sandboxID types.UUID, account *account.Account, config *modal.SandboxConfig) error`
2. Validate sandboxID is not zero: `if sandboxID == types.UUID("") || sandboxID == types.NilUUID`
3. Validate account is not nil: `if account == nil`
4. Validate account.ID is not zero: `if account.ID == types.UUID("") || account.ID == types.NilUUID`
5. Validate config is not nil: `if config == nil`
6. Return descriptive errors using `errors.New()` for each validation failure
7. Add godoc comment explaining purpose

**Verification:**
- Write test: `Test_validateParams_validInputs` - should return nil
- Write test: `Test_validateParams_zeroSandboxID` - should return error
- Write test: `Test_validateParams_nilAccount` - should return error
- Write test: `Test_validateParams_zeroAccountID` - should return error
- Write test: `Test_validateParams_nilConfig` - should return error
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service --test=Test_validateParams`
- Expected: All tests pass

**Self-Correction:**
- If tests fail: Check validation logic matches test expectations
- If compilation errors: Verify types.NilUUID constant exists or use appropriate zero check
- If logic errors: Review each validation condition

**Completion Criteria:**
- [ ] validateParams function exists and compiles
- [ ] All 5 validation cases handled
- [ ] All tests pass
- [ ] Function has godoc comment
- [ ] No lint errors

**Escape Condition:** If stuck after 3 attempts, document the validation logic issue and move to next task.

---

### Task 3: Implement reconstructSandboxInfo Private Method

**Requirements:** Requirement 1.3 (reconstruct SandboxInfo from model)

**Implementation:**
1. Create private method: `func (s *SandboxService) reconstructSandboxInfo(ctx context.Context, model *sandbox.Sandbox) (*modal.SandboxInfo, error)`
2. Extract fields from model and build `modal.SandboxInfo`:
   - SandboxID from model.SandboxID
   - AppName from model.AppName
   - VolumeName from model.VolumeName
   - Status from model.Status
   - Config from s.config
   - Other fields as needed
3. Validate reconstructed data has required fields
4. Return constructed SandboxInfo
5. Wrap errors with context using `errors.Wrapf()`
6. Add godoc comment

**Verification:**
- Write test: `Test_reconstructSandboxInfo_success` - valid model returns SandboxInfo
- Write test: `Test_reconstructSandboxInfo_invalidModel` - missing fields returns error
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service --test=Test_reconstructSandboxInfo`
- Expected: All tests pass

**Self-Correction:**
- If tests fail: Verify field mappings from model to SandboxInfo
- If nil pointer errors: Add nil checks for model fields
- If compilation errors: Check modal.SandboxInfo struct definition

**Completion Criteria:**
- [ ] reconstructSandboxInfo method exists
- [ ] All required fields mapped from model
- [ ] Error handling with context wrapping
- [ ] Tests pass
- [ ] Godoc comment added

**Escape Condition:** If stuck after 3 attempts, document field mapping issues and move to next task.

---

### Task 4: Implement ensureSandboxAccessible Private Method

**Requirements:** Requirement 1.2, 1.3 (ensure sandbox accessible)

**Implementation:**
1. Create private method: `func (s *SandboxService) ensureSandboxAccessible(ctx context.Context) error`
2. Check sandbox status via `s.client` (Modal API)
3. Verify sandbox is reachable (not terminated or errored)
4. Validate volume is attached using s.volume
5. Return error with context if not accessible
6. Add godoc comment explaining purpose

**Verification:**
- Write test: `Test_ensureSandboxAccessible_running` - running sandbox returns nil
- Write test: `Test_ensureSandboxAccessible_terminated` - terminated sandbox returns error
- Write test: `Test_ensureSandboxAccessible_noVolume` - missing volume returns error
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service --test=Test_ensureSandboxAccessible`
- Expected: All tests pass

**Self-Correction:**
- If tests fail: Check Modal API client mocking
- If status checks wrong: Review sandbox.Status enum values
- If volume validation missing: Ensure s.volume is checked

**Completion Criteria:**
- [ ] ensureSandboxAccessible method exists
- [ ] Status and volume checks implemented
- [ ] Errors wrapped with context
- [ ] Tests pass
- [ ] Godoc comment added

**Escape Condition:** If stuck after 3 attempts, document accessibility check issues and move to next task.

---

### Task 5: Implement loadOrCreateSandbox Private Method

**Requirements:** Requirement 1.1, 1.2 (GetOrCreate pattern)

**Implementation:**
1. Create private method: `func (s *SandboxService) loadOrCreateSandbox(ctx context.Context, sandboxID types.UUID) error`
2. Query database for sandbox by ID using `sandbox.Get(ctx, sandboxID)`
3. **IF EXISTS**:
   - Load sandbox model into s.sandboxModel
   - Call `s.reconstructSandboxInfo(ctx, model)` to build SandboxInfo
   - Store result in s.sandboxInfo
   - Extract volume info from SandboxInfo and store in s.volume
   - Call `s.ensureSandboxAccessible(ctx)` to verify
4. **IF NOT EXISTS**:
   - Generate names using `s.generateAppName(s.accountID)` and `s.generateVolumeName(s.accountID)`
   - Create sandbox via `s.client.CreateSandbox(ctx, s.config)` with generated names
   - Build new sandbox model with response data
   - Save model to database using `model.Save(ctx)`
   - Store model in s.sandboxModel
   - Construct SandboxInfo from API response and store in s.sandboxInfo
   - Extract volume info and store in s.volume
   - Call `s.ensureSandboxAccessible(ctx)` to verify
5. Wrap all errors with context
6. Add comprehensive godoc comment

**Verification:**
- Write test: `Test_loadOrCreateSandbox_existingSandbox` - loads from DB
- Write test: `Test_loadOrCreateSandbox_newSandbox` - creates and saves
- Write test: `Test_loadOrCreateSandbox_dbError` - handles DB errors
- Write test: `Test_loadOrCreateSandbox_createError` - handles create errors
- Write test: `Test_loadOrCreateSandbox_accessibilityError` - handles accessibility errors
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service --test=Test_loadOrCreateSandbox`
- Expected: All tests pass

**Self-Correction:**
- If tests fail: Check both code paths (exists vs not exists)
- If volume extraction fails: Verify modal.SandboxInfo has volume field
- If create fails: Check config is passed correctly to Modal API
- If save fails: Verify model has all required fields

**Completion Criteria:**
- [ ] loadOrCreateSandbox method exists
- [ ] Both code paths implemented (load existing, create new)
- [ ] Volume extracted and stored
- [ ] Accessibility verified
- [ ] All tests pass
- [ ] Comprehensive godoc comment

**Escape Condition:** If stuck after 3 attempts, document the GetOrCreate logic issue and move to next task.

---

### Task 6: Implement loadTemplate Private Method

**Requirements:** Requirement 5.1 (template management)

**Implementation:**
1. Create private method: `func (s *SandboxService) loadTemplate() error`
2. Get sandbox type from s.sandboxModel.Type or s.config
3. Get agent ID from s.sandboxModel.AgentID if exists
4. Call `GetSandboxTemplate(sandboxType, agentID)` (package-level function)
5. Store result in s.template (may be nil - not an error)
6. Return error only if template loading logic fails, not if no template exists
7. Add godoc comment noting nil template is valid

**Verification:**
- Write test: `Test_loadTemplate_withTemplate` - loads template successfully
- Write test: `Test_loadTemplate_noTemplate` - nil template is OK
- Write test: `Test_loadTemplate_unsupportedType` - handles unknown types
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service --test=Test_loadTemplate`
- Expected: All tests pass

**Self-Correction:**
- If tests fail: Verify GetSandboxTemplate returns nil for unknown types
- If nil panic: Add nil checks for s.sandboxModel before accessing Type
- If wrong template: Check sandbox type enum values

**Completion Criteria:**
- [ ] loadTemplate method exists
- [ ] Handles nil template gracefully
- [ ] Template stored in s.template
- [ ] All tests pass
- [ ] Godoc comment added

**Escape Condition:** If stuck after 3 attempts, document template loading issues and move to next task.

---

### Task 7: Implement NewSandboxService Constructor

**Requirements:** Requirement 2.1, 2.2, 2.3, 2.4 (constructor pattern)

**Implementation:**
1. Update/create `NewSandboxService` function: `func NewSandboxService(ctx context.Context, sandboxID types.UUID, account *account.Account, config *modal.SandboxConfig) (*SandboxService, error)`
2. Call `validateParams(sandboxID, account, config)` and return error if validation fails
3. Initialize service struct:
   - `client: modal.Client()`
   - `account: account`
   - `accountID: account.ID`
   - `config: config`
   - `conversationID: types.NilUUID` (or zero value)
4. Call `s.loadOrCreateSandbox(ctx, sandboxID)` and return error if it fails
5. Call `s.loadTemplate()` and return error if it fails
6. Call `s.executeColdStartHook(ctx)` (will implement later, stub for now)
7. Return initialized service
8. Update godoc comment with GetOrCreate explanation and example usage

**Verification:**
- Write test: `Test_NewSandboxService_existingSandbox` - loads successfully
- Write test: `Test_NewSandboxService_newSandbox` - creates successfully
- Write test: `Test_NewSandboxService_invalidParams` - validation fails
- Write test: `Test_NewSandboxService_loadError` - handles load errors
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service --test=Test_NewSandboxService`
- Expected: All tests pass (skip hook test if not implemented yet)

**Self-Correction:**
- If tests fail: Check initialization order
- If nil pointers: Ensure all required fields set before method calls
- If validation bypassed: Verify validateParams called first
- If hook errors: Stub executeColdStartHook to return nil for now

**Completion Criteria:**
- [ ] NewSandboxService constructor exists with new signature
- [ ] Validation called first
- [ ] All initialization steps completed
- [ ] Service returned fully initialized
- [ ] All tests pass
- [ ] Godoc with example usage

**Escape Condition:** If stuck after 3 attempts, document constructor initialization issues and move to next task.

---

## Phase 2: State Management Helpers

### Task 8: Implement ensureSandboxRunning Private Method

**Requirements:** Requirement 4.4 (ensure sandbox running)

**Implementation:**
1. Create private method: `func (s *SandboxService) ensureSandboxRunning(ctx context.Context, operation string) error`
2. Check `s.sandboxInfo.Status`:
   - If running: return nil (no-op)
   - If terminated: create new sandbox
3. If terminated:
   - Create new sandbox via `s.client.CreateSandbox(ctx, s.config)` using same volume
   - Call `s.updateSandboxState(ctx, newSandboxInfo)` to update state atomically
4. Wrap errors with operation context
5. Add godoc comment explaining auto-restart behavior

**Verification:**
- Write test: `Test_ensureSandboxRunning_alreadyRunning` - no-op
- Write test: `Test_ensureSandboxRunning_terminated` - creates new sandbox
- Write test: `Test_ensureSandboxRunning_createFails` - handles create error
- Write test: `Test_ensureSandboxRunning_updateFails` - handles update error
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service --test=Test_ensureSandboxRunning`
- Expected: All tests pass

**Self-Correction:**
- If tests fail: Check status enum values
- If create fails: Verify config includes volume reference
- If state not updated: Ensure updateSandboxState called after create
- If errors not wrapped: Add operation name to error context

**Completion Criteria:**
- [ ] ensureSandboxRunning method exists
- [ ] Running check is no-op
- [ ] Terminated triggers recreation with volume
- [ ] State updated atomically
- [ ] All tests pass

**Escape Condition:** If stuck after 3 attempts, document auto-restart logic issues and move to next task.

---

### Task 9: Implement updateSandboxState Private Method

**Requirements:** Requirement 4.5 (atomic state updates)

**Implementation:**
1. Create private method: `func (s *SandboxService) updateSandboxState(ctx context.Context, newSandboxInfo *modal.SandboxInfo) error`
2. Update database model fields from newSandboxInfo
3. Save model to database using `s.sandboxModel.Save(ctx)`
4. If save succeeds:
   - Update `s.sandboxInfo = newSandboxInfo`
   - Update `s.sandboxModel` with saved model
   - Extract and update `s.volume` from newSandboxInfo
5. If save fails: return error WITHOUT updating memory (atomicity)
6. Add godoc comment emphasizing atomicity

**Verification:**
- Write test: `Test_updateSandboxState_success` - updates DB and memory
- Write test: `Test_updateSandboxState_dbError` - leaves memory unchanged
- Write test: `Test_updateSandboxState_atomicity` - verify rollback on error
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service --test=Test_updateSandboxState`
- Expected: All tests pass

**Self-Correction:**
- If tests fail: Check update order (DB first, then memory)
- If atomicity broken: Ensure memory not updated if DB save fails
- If volume not updated: Extract volume from newSandboxInfo

**Completion Criteria:**
- [ ] updateSandboxState method exists
- [ ] Database updated before memory
- [ ] Atomicity enforced (no partial updates)
- [ ] Volume extracted and updated
- [ ] All tests pass

**Escape Condition:** If stuck after 3 attempts, document atomicity implementation issues and move to next task.

---

## Phase 3: Public Method Refactoring

### Task 10: Refactor TerminateSandbox Method

**Requirements:** Requirement 3.4 (simplified signature)

**Implementation:**
1. Update method signature: `func (s *SandboxService) TerminateSandbox(ctx context.Context, syncToS3 bool) error`
2. Remove sandboxInfo, sandboxModel, user parameters (now use s.sandboxInfo, s.sandboxModel, s.account)
3. If syncToS3 is true: call `s.SyncFiles(ctx)` (will implement later, stub for now)
4. Call Modal API to terminate sandbox using s.sandboxInfo
5. Update s.sandboxModel.Status to terminated
6. Save model using `s.sandboxModel.Save(ctx)`
7. Call `s.executeTerminateHook(ctx)` (will implement later, stub for now)
8. Update godoc comment to reflect new signature

**Verification:**
- Write test: `Test_TerminateSandbox_withSync` - syncs before terminating
- Write test: `Test_TerminateSandbox_withoutSync` - skips sync
- Write test: `Test_TerminateSandbox_apiError` - handles API errors
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service --test=Test_TerminateSandbox`
- Expected: All tests pass (stub sync and hook if needed)

**Self-Correction:**
- If tests fail: Check internal state access (s.sandboxInfo, etc.)
- If sync errors: Stub SyncFiles to return nil for now
- If hook errors: Stub executeTerminateHook to return nil for now

**Completion Criteria:**
- [ ] Signature simplified (no sandboxInfo, sandboxModel, user params)
- [ ] Uses internal state
- [ ] Sync and hook integration points added
- [ ] All tests pass
- [ ] Godoc updated

**Escape Condition:** If stuck after 3 attempts, document termination logic issues and move to next task.

---

### Task 11: Refactor ListFiles Method

**Requirements:** Requirement 3.1 (simplified file operations)

**Implementation:**
1. Update method signature: `func (s *SandboxService) ListFiles(ctx context.Context, opts *ListFilesOptions) ([]FileInfo, error)`
2. Remove sandboxInfo, sandboxModel, user parameters
3. Call `s.ensureSandboxRunning(ctx, "list files")` at start
4. Call `s.buildListFilesCommand(opts)` to generate command
5. Execute command using `s.client` with `s.sandboxInfo`
6. Parse and return file list
7. Update godoc comment

**Verification:**
- Write test: `Test_ListFiles_success` - lists files successfully
- Write test: `Test_ListFiles_sandboxNotRunning` - auto-restarts
- Write test: `Test_ListFiles_apiError` - handles errors
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service --test=Test_ListFiles`
- Expected: All tests pass

**Self-Correction:**
- If tests fail: Check ensureSandboxRunning called first
- If auto-restart fails: Verify ensureSandboxRunning implementation
- If command build fails: Check buildListFilesCommand exists

**Completion Criteria:**
- [ ] Signature simplified
- [ ] Uses internal state
- [ ] Ensures sandbox running
- [ ] All tests pass
- [ ] Godoc updated

**Escape Condition:** If stuck after 3 attempts, document file listing issues and move to next task.

---

### Task 12: Refactor GetFileContent Method

**Requirements:** Requirement 3.1 (simplified file operations)

**Implementation:**
1. Update method signature: `func (s *SandboxService) GetFileContent(ctx context.Context, source string, filePath string) (string, error)`
2. Remove sandboxInfo, sandboxModel, user parameters
3. Call `s.ensureSandboxRunning(ctx, "get file content")` at start
4. Call `s.buildReadFileCommand(source, filePath, maxSize)` to generate command
5. Execute command using `s.client` with `s.sandboxInfo`
6. Return file content
7. Update godoc comment

**Verification:**
- Write test: `Test_GetFileContent_success` - reads file successfully
- Write test: `Test_GetFileContent_sandboxNotRunning` - auto-restarts
- Write test: `Test_GetFileContent_fileNotFound` - handles errors
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service --test=Test_GetFileContent`
- Expected: All tests pass

**Self-Correction:**
- If tests fail: Check ensureSandboxRunning called first
- If command build fails: Check buildReadFileCommand exists
- If maxSize undefined: Use default value or constant

**Completion Criteria:**
- [ ] Signature simplified
- [ ] Uses internal state
- [ ] Ensures sandbox running
- [ ] All tests pass
- [ ] Godoc updated

**Escape Condition:** If stuck after 3 attempts, document file reading issues and move to next task.

---

### Task 13: Refactor ExecuteClaudeStream Method

**Requirements:** Requirement 3.5 (simplified execution)

**Implementation:**
1. Update method signature: `func (s *SandboxService) ExecuteClaudeStream(ctx context.Context, config *claude.ExecutionConfig, responseWriter io.Writer) error`
2. Remove sandboxInfo parameter
3. Call `s.ensureSandboxRunning(ctx, "execute claude stream")` at start
4. Execute command via Modal API using `s.client` with `s.sandboxInfo`
5. Stream response to responseWriter
6. Update godoc comment

**Verification:**
- Write test: `Test_ExecuteClaudeStream_success` - executes and streams
- Write test: `Test_ExecuteClaudeStream_sandboxNotRunning` - auto-restarts
- Write test: `Test_ExecuteClaudeStream_streamError` - handles stream errors
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service --test=Test_ExecuteClaudeStream`
- Expected: All tests pass

**Self-Correction:**
- If tests fail: Check ensureSandboxRunning called first
- If streaming fails: Verify responseWriter handling
- If config issues: Check claude.ExecutionConfig structure

**Completion Criteria:**
- [ ] Signature simplified
- [ ] Uses internal state
- [ ] Ensures sandbox running
- [ ] All tests pass
- [ ] Godoc updated

**Escape Condition:** If stuck after 3 attempts, document streaming execution issues and move to next task.

---

### Task 14: Implement SyncFiles Public Method

**Requirements:** Requirement 3.3 (intelligent sync)

**Implementation:**
1. Create new method: `func (s *SandboxService) SyncFiles(ctx context.Context) error`
2. Call `s.ensureSandboxRunning(ctx, "sync files")` at start
3. Check if S3 configured: `s.config.S3Config != nil`
4. If not configured: return nil (no-op)
5. Analyze current state vs S3 state (use state files):
   - Call `s.orchestratePullSync(ctx, staleThresholdSeconds)` if local is stale
   - Call `s.orchestratePushSync(ctx)` if local has changes
6. Handle bidirectional sync if needed
7. Add comprehensive godoc explaining intelligent sync behavior

**Verification:**
- Write test: `Test_SyncFiles_noS3Config` - no-op
- Write test: `Test_SyncFiles_pullNeeded` - pulls from S3
- Write test: `Test_SyncFiles_pushNeeded` - pushes to S3
- Write test: `Test_SyncFiles_bidirectional` - handles both
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service --test=Test_SyncFiles`
- Expected: All tests pass

**Self-Correction:**
- If tests fail: Check orchestrator methods exist (stub if needed)
- If state analysis fails: Verify state file reading logic
- If sync direction wrong: Review staleness and change detection logic

**Completion Criteria:**
- [ ] SyncFiles method exists
- [ ] Intelligent sync logic implemented
- [ ] No-op if no S3 config
- [ ] All tests pass
- [ ] Comprehensive godoc

**Escape Condition:** If stuck after 3 attempts, document sync logic issues and move to next task.

---

### Task 15: Implement Getter Methods

**Requirements:** Testing and debugging support

**Implementation:**
1. Create method: `func (s *SandboxService) GetSandboxInfo() *modal.SandboxInfo { return s.sandboxInfo }`
2. Create method: `func (s *SandboxService) GetSandboxModel() *sandbox.Sandbox { return s.sandboxModel }`
3. Create method: `func (s *SandboxService) GetVolume() *modal.VolumeInfo { return s.volume }`
4. Add godoc comments noting these are for testing/debugging

**Verification:**
- Write test: `Test_GetSandboxInfo` - returns correct info
- Write test: `Test_GetSandboxModel` - returns correct model
- Write test: `Test_GetVolume` - returns correct volume
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service --test=Test_Get`
- Expected: All tests pass

**Self-Correction:**
- If tests fail: Verify fields are set during initialization
- If nil returned: Check NewSandboxService sets all fields

**Completion Criteria:**
- [ ] All three getter methods exist
- [ ] Methods return correct fields
- [ ] All tests pass
- [ ] Godoc comments added

**Escape Condition:** If stuck after 3 attempts, document getter implementation issues and move to next task.

---

### Task 16: Implement SetConversationID Method

**Requirements:** Requirement 6.1, 6.4 (conversation context)

**Implementation:**
1. Create method: `func (s *SandboxService) SetConversationID(conversationID types.UUID)`
2. Set `s.conversationID = conversationID`
3. Add godoc comment noting this is optional for lifecycle hooks

**Verification:**
- Write test: `Test_SetConversationID` - sets and retrieves correctly
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service --test=Test_SetConversationID`
- Expected: Test passes

**Self-Correction:**
- If test fails: Verify field is exported and accessible
- If getter needed: Add getter for testing

**Completion Criteria:**
- [ ] SetConversationID method exists
- [ ] Updates internal field
- [ ] Test passes
- [ ] Godoc comment added

**Escape Condition:** If stuck after 2 attempts, document setter issues and move to next task.

---

## Phase 4: Private Method Implementation

### Task 17: Implement Command Builder Methods

**Requirements:** Helper methods for operations

**Implementation:**
1. Move/keep `buildListFilesCommand` as private method: `func (s *SandboxService) buildListFilesCommand(opts *ListFilesOptions) string`
2. Move/keep `buildReadFileCommand` as private method: `func (s *SandboxService) buildReadFileCommand(source string, filePath string, maxSize int64) string`
3. Ensure both are private (lowercase first letter)
4. Update any references to use `s.buildListFilesCommand()` instead of package function
5. Add godoc comments

**Verification:**
- Run: `make fmt && make lint`
- Expected: No errors
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service`
- Expected: All existing tests still pass

**Self-Correction:**
- If tests fail: Update test code to use service methods
- If references broken: Find and update all call sites

**Completion Criteria:**
- [ ] Both methods are private receiver methods
- [ ] Existing logic preserved
- [ ] All tests pass
- [ ] Godoc comments added

**Escape Condition:** If stuck after 2 attempts, document command builder issues and move to next task.

---

### Task 18: Implement Naming Utility Methods

**Requirements:** Private naming helpers

**Implementation:**
1. Move `GenerateAppName` to private method: `func (s *SandboxService) generateAppName(accountID types.UUID) string`
2. Move `GenerateVolumeName` to private method: `func (s *SandboxService) generateVolumeName(accountID types.UUID) string`
3. Move `BuildPermissionFixCommand` to private method: `func (s *SandboxService) buildPermissionFixCommand(workdir, username string) string`
4. Update loadOrCreateSandbox to use `s.generateAppName()` and `s.generateVolumeName()`
5. Remove old package-level versions
6. Add godoc comments

**Verification:**
- Run: `make fmt && make lint`
- Expected: No errors
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service`
- Expected: All tests still pass

**Self-Correction:**
- If tests fail: Update test code to initialize service first
- If references broken: Update all call sites to use service methods

**Completion Criteria:**
- [ ] All three methods are private receiver methods
- [ ] Package-level functions removed
- [ ] loadOrCreateSandbox updated
- [ ] All tests pass
- [ ] Godoc comments added

**Escape Condition:** If stuck after 2 attempts, document naming utility issues and move to next task.

---

### Task 19: Implement Private Sync Operations

**Requirements:** Requirement 3.3 (sync operations)

**Implementation:**
1. Move `OrchestratePullSync` to private method: `func (s *SandboxService) orchestratePullSync(ctx context.Context, staleThresholdSeconds int) error`
2. Move `OrchestratePushSync` to private method: `func (s *SandboxService) orchestratePushSync(ctx context.Context) error`
3. Create private method: `func (s *SandboxService) initFromS3(ctx context.Context) error` (calls lifecycle action)
4. Update methods to use internal state (s.client, s.sandboxInfo) instead of parameters
5. Remove client parameter from signatures
6. Update SyncFiles to call these private methods
7. Add godoc comments

**Verification:**
- Write test: `Test_orchestratePullSync` - pulls with state files
- Write test: `Test_orchestratePushSync` - pushes with state files
- Write test: `Test_initFromS3` - initializes from S3
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service --test=Test_orchestrate`
- Expected: All tests pass

**Self-Correction:**
- If tests fail: Check internal state access
- If client issues: Verify s.client is used
- If state files fail: Check state_files package integration

**Completion Criteria:**
- [ ] All three methods are private receiver methods
- [ ] Use internal state
- [ ] Old package functions removed
- [ ] All tests pass
- [ ] Godoc comments added

**Escape Condition:** If stuck after 3 attempts, document sync operation issues and move to next task.

---

### Task 20: Implement executeColdStartHook Private Method

**Requirements:** Requirement 5.1, 5.2 (lifecycle hooks)

**Implementation:**
1. Create private method: `func (s *SandboxService) executeColdStartHook(ctx context.Context) error`
2. Check if `s.template == nil || s.template.Hooks == nil || s.template.Hooks.OnColdStart == nil`
3. If nil: return nil (no-op)
4. Build `lifecycle.HookData`:
   ```go
   hookData := &lifecycle.HookData{
       ConversationID: s.conversationID,
       SandboxInfo:    s.sandboxInfo,
   }
   ```
5. Call `return lifecycle.ExecuteHook(ctx, "OnColdStart", s.template.Hooks.OnColdStart, hookData)`
6. Add godoc comment explaining CRITICAL hook behavior

**Verification:**
- Write test: `Test_executeColdStartHook_noHook` - nil template is no-op
- Write test: `Test_executeColdStartHook_success` - executes hook
- Write test: `Test_executeColdStartHook_error` - propagates errors
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service --test=Test_executeColdStartHook`
- Expected: All tests pass

**Self-Correction:**
- If tests fail: Check nil checks for template
- If hook not called: Verify lifecycle.ExecuteHook import
- If error not propagated: Ensure error returned from ExecuteHook

**Completion Criteria:**
- [ ] executeColdStartHook method exists
- [ ] Nil template handled
- [ ] HookData built from internal state
- [ ] Errors propagated (CRITICAL)
- [ ] All tests pass

**Escape Condition:** If stuck after 3 attempts, document hook execution issues and move to next task.

---

### Task 21: Implement executeTerminateHook Private Method

**Requirements:** Requirement 5.2 (terminate hook)

**Implementation:**
1. Create private method: `func (s *SandboxService) executeTerminateHook(ctx context.Context) error`
2. Check if `s.template == nil || s.template.Hooks == nil || s.template.Hooks.OnTerminate == nil`
3. If nil: return nil (no-op)
4. Build `lifecycle.HookData`:
   ```go
   hookData := &lifecycle.HookData{
       ConversationID: s.conversationID,
       SandboxInfo:    s.sandboxInfo,
   }
   ```
5. Call `_ = lifecycle.ExecuteHook(ctx, "OnTerminate", s.template.Hooks.OnTerminate, hookData)`
6. Return nil (NON-CRITICAL - errors swallowed)
7. Add godoc comment explaining NON-CRITICAL behavior

**Verification:**
- Write test: `Test_executeTerminateHook_noHook` - nil template is no-op
- Write test: `Test_executeTerminateHook_success` - executes hook
- Write test: `Test_executeTerminateHook_errorSwallowed` - errors don't propagate
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service --test=Test_executeTerminateHook`
- Expected: All tests pass

**Self-Correction:**
- If tests fail: Check nil checks for template
- If errors propagate: Ensure return nil, not return err
- If hook not called: Verify lifecycle.ExecuteHook import

**Completion Criteria:**
- [ ] executeTerminateHook method exists
- [ ] Nil template handled
- [ ] HookData built from internal state
- [ ] Always returns nil (NON-CRITICAL)
- [ ] All tests pass

**Escape Condition:** If stuck after 2 attempts, document hook execution issues and move to next task.

---

### Task 22: Implement Future Hook Stubs

**Requirements:** Future extensibility

**Implementation:**
1. Create private method stub: `func (s *SandboxService) executeMessageHook(ctx context.Context, msg *message.Message) error { return nil }`
2. Create private method stub: `func (s *SandboxService) executeStreamFinishHook(ctx context.Context, tokenUsage *lifecycle.TokenUsage) error { return nil }`
3. Add godoc comments noting these are for future use when conversation_service is implemented
4. Add TODO comments explaining when these should be called

**Verification:**
- Run: `make fmt && make lint`
- Expected: No errors
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service`
- Expected: All tests still pass

**Self-Correction:**
- If lint errors: Add proper godoc comments
- If compilation errors: Check types are imported

**Completion Criteria:**
- [ ] Both stub methods exist
- [ ] Return nil (no implementation needed yet)
- [ ] Godoc with "future use" note
- [ ] Code compiles

**Escape Condition:** If stuck after 1 attempt, skip this task and note in documentation.

---

## Phase 5: Integration and Migration

### Task 23: Update Calling Code - Controllers

**Requirements:** Requirement 8.1 (update calling code)

**Implementation:**
1. Find all usages of `sandbox_service.NewSandboxService()` in controllers
2. Update each to new signature: `NewSandboxService(ctx, sandboxID, account, config)`
3. Replace `user` parameter with `account`
4. Ensure all method calls remove sandboxInfo, sandboxModel, user parameters
5. Update error handling if needed

**Verification:**
- Run: `make fmt && make lint`
- Expected: No compilation errors
- Run: `#code_tools run_tests --package=./internal/controllers/...`
- Expected: Controller tests still pass

**Self-Correction:**
- If compilation errors: Find remaining old signatures
- If tests fail: Update test setup to use new signature
- If account missing: Load account model where user was used

**Completion Criteria:**
- [ ] All controller code updated
- [ ] No compilation errors
- [ ] All controller tests pass

**Escape Condition:** If stuck after 3 attempts, document remaining issues and move to next task.

---

### Task 24: Update Calling Code - Other Services

**Requirements:** Requirement 8.1 (update calling code)

**Implementation:**
1. Find all usages of `sandbox_service.NewSandboxService()` in other services
2. Update each to new signature
3. Update method calls to remove threaded parameters
4. Replace any `ReconstructSandboxInfo` package calls with service initialization

**Verification:**
- Run: `make fmt && make lint`
- Expected: No compilation errors
- Run: `#code_tools run_tests --package=./internal/services/...`
- Expected: Service tests still pass

**Self-Correction:**
- If compilation errors: Search for remaining old usages
- If tests fail: Update service test setup
- If ReconstructSandboxInfo calls remain: Remove and use NewSandboxService

**Completion Criteria:**
- [ ] All service code updated
- [ ] No compilation errors
- [ ] All service tests pass

**Escape Condition:** If stuck after 3 attempts, document remaining issues and move to next task.

---

### Task 25: Update Unit Tests

**Requirements:** Requirement 10.1, 10.2, 10.3 (testing)

**Implementation:**
1. Update all sandbox_service tests to use new constructor
2. Create test helpers for building service with mocked dependencies
3. Update test assertions to use getter methods where needed
4. Ensure test coverage ≥90% for refactored code
5. Update mocks for new signatures

**Verification:**
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service`
- Expected: All tests pass
- Check coverage: Should be ≥90%
- Run: `make lint`
- Expected: No lint errors

**Self-Correction:**
- If tests fail: Update test setup to match new patterns
- If coverage low: Add missing test cases
- If mocks broken: Regenerate or update mock interfaces

**Completion Criteria:**
- [ ] All unit tests updated
- [ ] Test coverage ≥90%
- [ ] All tests pass
- [ ] Test helpers created

**Escape Condition:** If stuck after 3 attempts, document test issues and move to next task.

---

### Task 26: Update Integration Tests

**Requirements:** Requirement 10.5 (integration testing)

**Implementation:**
1. Find integration tests that use sandbox_service
2. Update to use new constructor and signatures
3. Verify full service lifecycle tests work (create → operate → terminate)
4. Verify auto-restart behavior tested
5. Verify sync operations tested

**Verification:**
- Run integration tests: `#code_tools run_tests --package=./...` (or integration test command)
- Expected: All integration tests pass

**Self-Correction:**
- If tests fail: Check test database setup
- If lifecycle tests broken: Verify all hooks execute correctly
- If auto-restart fails: Check ensureSandboxRunning implementation

**Completion Criteria:**
- [ ] All integration tests updated
- [ ] Lifecycle tests pass
- [ ] Auto-restart tests pass
- [ ] All integration tests pass

**Escape Condition:** If stuck after 3 attempts, document integration test issues and move to next task.

---

## Phase 6: Cleanup and Documentation

### Task 27: Remove Deprecated Code

**Requirements:** Requirement 8.5 (cleanup)

**Implementation:**
1. Remove old package-level `ReconstructSandboxInfo` function (now private method)
2. Remove old package-level naming utilities (now private methods)
3. Remove any old orchestrator package functions (now private methods)
4. Search for any remaining unused old signatures
5. Remove any deprecated constants or types

**Verification:**
- Run: `make fmt && make lint`
- Expected: No errors
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service`
- Expected: All tests still pass
- Search codebase: No references to removed functions

**Self-Correction:**
- If compilation errors: Some code still references old functions
- If tests fail: Tests still using old functions
- If grep finds usages: Update remaining call sites

**Completion Criteria:**
- [ ] All deprecated code removed
- [ ] No compilation errors
- [ ] All tests pass
- [ ] No references to old functions

**Escape Condition:** If stuck after 2 attempts, document remaining deprecated code and move to next task.

---

### Task 28: Aggressive Unused Code Removal

**Requirements:** Code cleanliness and maintainability

**Implementation:**
1. Analyze all functions/methods in `internal/services/sandbox_service` package
2. For each function/method, check if it's used:
   - Search codebase for references: `grep -r "FunctionName" --include="*.go"`
   - Check if it's called from controllers in the standard flow
3. **IF COMPLETELY UNUSED:**
   - Delete the function/method entirely
   - Delete associated tests
   - Remove from any exports or documentation
4. **IF USED OUTSIDE STANDARD CONTROLLER FLOW:**
   - Add comment: `// @deprecated not sure if unused - called from [location]`
   - Keep the function for manual review
5. Standard controller flow is defined as:
   - Calls from `internal/controllers/*sandbox*` controllers
   - Calls from services used by sandbox controllers
   - Direct path from HTTP endpoint → controller → sandbox_service
6. Create a list of marked functions for manual review

**What to Check:**
- All public exported functions (capitalized names)
- All public methods on SandboxService
- Helper functions in the package
- Types and structs that might be unused
- Constants that are no longer referenced

**What NOT to Remove:**
- Public API methods from the design document (even if low usage)
- Lifecycle hooks (even if only called internally)
- Template functions (GetSandboxTemplate, etc.)
- Package-level utilities still in use (ConvertTreePathsToUserFacing)

**Verification:**
- Run: `make fmt && make lint`
- Expected: No errors, code still compiles
- Run: `#code_tools run_tests --package=./internal/services/sandbox_service`
- Expected: All tests pass
- Check: No unused imports remain
- Review: List of @deprecated marked functions created

**Self-Correction:**
- If compilation errors: Something still uses deleted code, restore it
- If tests fail: Deleted something that tests needed, restore it
- If uncertain: Mark with @deprecated instead of deleting
- If used in tests only: Delete both function and test

**Completion Criteria:**
- [ ] All completely unused code removed
- [ ] Functions used outside standard flow marked with @deprecated
- [ ] No compilation errors
- [ ] All tests pass
- [ ] List of @deprecated functions documented for manual review
- [ ] No unused imports

**Example @deprecated Comment:**
```go
// @deprecated not sure if unused - called from internal/services/github_service/setup.go
// Used during GitHub integration setup, may be legacy code
func (s *SandboxService) SomeOldMethod() {}
```

**Escape Condition:** If stuck after 3 attempts analyzing usage, mark remaining uncertain functions with @deprecated and document them for manual review.

---

### Task 29: Update Package Godoc

**Requirements:** Requirement 11.1, 11.2 (documentation)

**Note:** This task was previously Task 28, renumbered after adding aggressive cleanup task.

**Implementation:**
1. Update package-level godoc comment in `sandbox_service.go`
2. Explain the stateful design and GetOrCreate pattern
3. Provide example usage showing:
   - Constructor with parameters
   - Multiple method calls without threading
   - Clean API usage
4. Document the lifecycle hook system
5. Note volume persistence across reboots

**Verification:**
- Run: `go doc internal/services/sandbox_service`
- Expected: Shows updated package documentation
- Review: Documentation is clear and includes example

**Self-Correction:**
- If doc not showing: Check godoc comment format
- If example wrong: Test example code compiles
- If unclear: Simplify language and add more context

**Completion Criteria:**
- [ ] Package godoc updated
- [ ] GetOrCreate pattern explained
- [ ] Example usage provided
- [ ] Hook system documented
- [ ] Volume persistence noted

**Escape Condition:** If stuck after 2 attempts, document documentation issues and move to next task.

---

### Task 30: Update Method Godocs

**Requirements:** Requirement 11.3 (method documentation)

**Note:** This task was previously Task 29, renumbered after adding aggressive cleanup task.

**Implementation:**
1. Review all public method godoc comments
2. Update to reflect simplified signatures (no mention of removed parameters)
3. Add notes about auto-restart where applicable
4. Document GetOrCreate behavior in constructor
5. Ensure private methods have appropriate comments

**Verification:**
- Run: `make lint`
- Expected: No missing comment warnings
- Review each public method: Has clear godoc
- Run: `go doc internal/services/sandbox_service.NewSandboxService`
- Expected: Shows complete documentation

**Self-Correction:**
- If lint warnings: Add missing comments
- If unclear: Simplify and clarify language
- If examples needed: Add inline examples

**Completion Criteria:**
- [ ] All public methods have updated godocs
- [ ] No mention of removed parameters
- [ ] Auto-restart documented where relevant
- [ ] No lint warnings

**Escape Condition:** If stuck after 2 attempts, note documentation issues and proceed.

---

### Task 31: Final Verification and Cleanup

**Requirements:** All requirements verified

**Note:** This task was previously Task 30, renumbered after adding aggressive cleanup task.

**Implementation:**
1. Run full test suite: `#code_tools run_tests --package=./...`
2. Check test coverage: Should be ≥90% for sandbox_service
3. Run linter: `make lint`
4. Run formatter: `make fmt`
5. Verify all TODO comments addressed or documented
6. Check for any remaining old patterns or deprecated code
7. Verify no breaking changes to external APIs (besides intended refactor)

**Verification:**
- All tests pass: `#code_tools run_tests --package=./...`
- Coverage ≥90%: Check coverage report
- No lint errors: `make lint` returns 0
- Code formatted: `make fmt` makes no changes
- No TODOs: Search for TODO in refactored files

**Self-Correction:**
- If tests fail: Identify and fix failing tests
- If coverage low: Add missing test cases
- If lint errors: Fix linting issues
- If TODOs remain: Address or document

**Completion Criteria:**
- [ ] All tests pass
- [ ] Test coverage ≥90%
- [ ] No lint errors
- [ ] Code properly formatted
- [ ] All TODOs addressed
- [ ] No deprecated code remains
- [ ] Breaking changes documented

**Escape Condition:** Document any remaining issues in a summary document.

---

## Summary

This refactor transforms the sandbox_service from a stateless service with extensive parameter threading to a clean, stateful service with a universal GetOrCreate pattern. The refactor maintains all existing functionality while significantly simplifying the API and making the code easier to maintain and extend.

**Total Tasks:** 31 tasks across 6 phases
**Estimated Complexity:** Large refactor requiring careful migration
**Key Success Metrics:**
- All tests passing
- Test coverage ≥90%
- No compilation errors
- Clean API with no parameter threading
- GetOrCreate pattern working correctly
- Auto-restart functionality preserved
- Lifecycle hooks executing correctly

**Next Steps After Completion:**
1. Update any external documentation referencing the old API
2. Communicate breaking changes to team
3. Monitor for any issues in staging/production
4. Consider adding performance benchmarks
