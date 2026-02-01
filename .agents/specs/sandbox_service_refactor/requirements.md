# Sandbox Service Refactor - Requirements

## Introduction

This document outlines the requirements for refactoring the `internal/services/sandbox_service` from a stateless service with extensive parameter threading to a stateful service design. The current implementation requires passing `sandboxInfo`, `sandboxModel`, `user`, and other context through nearly every function call, creating verbose signatures and making the code harder to maintain. This refactor will move commonly-passed parameters into the service struct itself, creating a cleaner API where the service is initialized with all necessary context upfront.

The service will implement a universal **GetOrCreate** pattern - when initialized with a sandbox ID, it will intelligently load the existing sandbox if it exists, or create a new one if it doesn't. This eliminates the need for callers to handle conditional logic around sandbox existence.

The refactor aims to maintain all existing functionality while significantly improving code clarity, reducing parameter threading, and making the service easier to use and extend.

---

## Requirements

### 1. Stateful Service Initialization with GetOrCreate

**User Story:** As a developer using the sandbox service, I want to initialize it with all necessary context upfront (sandbox ID, user, account) so that I don't have to pass these parameters to every method call, and the service should intelligently handle whether the sandbox exists or needs to be created.

**Acceptance Criteria:**

1.1. **WHEN** creating a new sandbox service instance, **IF** provided with a sandbox ID and the sandbox exists in the database, **THEN** the service **SHALL** load the existing sandbox model from the database into its internal state.

1.2. **WHEN** creating a new sandbox service instance, **IF** the sandbox ID does not exist in the database, **THEN** the service **SHALL** create a new sandbox using the provided configuration and store it in the database.

1.3. **WHEN** creating a new sandbox service instance, **IF** a sandbox is loaded or created, **THEN** the service **SHALL** reconstruct the SandboxInfo and store it in its internal state.

1.4. **WHEN** the service is initialized, **IF** successful, **THEN** the service **SHALL** contain all necessary state (sandboxInfo, sandboxModel, user, accountID, client) for subsequent operations.

1.5. **WHEN** the service is initialized, **IF** successful, **THEN** all existing service methods **SHALL** no longer require sandboxInfo, sandboxModel, or user as parameters.

---

### 2. Constructor Pattern Simplification

**User Story:** As a developer integrating with the sandbox service, I want a simple, intuitive constructor that clearly communicates what's required to use the service, so that I can quickly understand and use the API correctly.

**Acceptance Criteria:**

2.1. **WHEN** creating a sandbox service, **IF** I have a sandbox ID, user, and account context, **THEN** the constructor **SHALL** be `NewSandboxService(ctx, sandboxID, accountID, user, config) (*SandboxService, error)`.

2.2. **WHEN** the constructor is called, **IF** any required parameter is invalid (zero UUID, nil user), **THEN** it **SHALL** return a clear validation error.

2.3. **WHEN** the constructor completes successfully, **IF** the service is returned, **THEN** it **SHALL** be immediately ready to use without additional setup, regardless of whether the sandbox existed or was newly created.

2.4. **WHEN** looking at the constructor signature, **IF** I'm a new developer, **THEN** it **SHALL** be immediately clear what context is required to use the service.

---

### 3. Method Signature Simplification

**User Story:** As a developer calling sandbox service methods, I want simplified method signatures that don't require passing the same context repeatedly, so that my code is cleaner and less error-prone.

**Acceptance Criteria:**

3.1. **WHEN** calling file operations (ListFiles, GetFileContent), **IF** the service is initialized, **THEN** the methods **SHALL NOT** require sandboxInfo, sandboxModel, or user as parameters.

3.2. **WHEN** calling lifecycle hook methods (ExecuteColdStartHook, ExecuteMessageHook, etc.), **IF** the service is initialized, **THEN** the methods **SHALL NOT** require sandboxInfo as a parameter.

3.3. **WHEN** calling sync operations (SyncToS3, InitFromS3), **IF** the service is initialized, **THEN** the methods **SHALL NOT** require sandboxInfo as a parameter.

3.4. **WHEN** calling TerminateSandbox, **IF** the service is initialized, **THEN** it **SHALL NOT** require sandboxInfo as a parameter.

3.5. **WHEN** calling ExecuteClaudeStream, **IF** the service is initialized, **THEN** it **SHALL NOT** require sandboxInfo as a parameter.

3.6. **WHEN** any method is called, **IF** additional operation-specific parameters are needed (like file paths, options), **THEN** only those specific parameters **SHALL** be required.

---

### 4. GetOrCreate Sandbox Logic

**User Story:** As a developer using the sandbox service, I want the service to intelligently handle whether a sandbox needs to be created or already exists, so that I don't have to write conditional logic in my code.

**Acceptance Criteria:**

4.1. **WHEN** the service is initialized, **IF** the sandbox ID exists in the database, **THEN** the service **SHALL** load the existing sandbox and its associated SandboxInfo.

4.2. **WHEN** the service is initialized, **IF** the sandbox ID does not exist in the database, **THEN** the service **SHALL** create a new sandbox using the provided configuration.

4.3. **WHEN** a new sandbox is created during initialization, **IF** the creation succeeds, **THEN** the service **SHALL** store the sandbox model in the database and construct the SandboxInfo.

4.4. **WHEN** operations are performed, **IF** the sandbox needs to be running (e.g., terminated sandbox), **THEN** the service **SHALL** ensure the sandbox is in the correct state before proceeding with the operation.

4.5. **WHEN** sandbox state changes are needed, **IF** the service updates the sandbox, **THEN** it **SHALL** update both the database model and internal SandboxInfo atomically.

---

### 5. Template Management

**User Story:** As a developer working with lifecycle hooks, I want the service to manage the sandbox template internally based on the sandbox configuration, so that I don't have to retrieve and pass templates to hook methods.

**Acceptance Criteria:**

5.1. **WHEN** the service is initialized, **IF** the sandbox has a type and optional agent ID, **THEN** the service **SHALL** load the appropriate SandboxTemplate into its state.

5.2. **WHEN** calling lifecycle hook methods, **IF** the service is initialized, **THEN** the methods **SHALL NOT** require a template parameter.

5.3. **WHEN** the template is needed for hooks, **IF** stored in the service, **THEN** the service **SHALL** use its internal template reference.

5.4. **WHEN** the sandbox type or configuration doesn't map to a template, **IF** hooks are not applicable, **THEN** the service **SHALL** handle this gracefully (no-op or skip hooks).

5.5. **WHEN** template configuration changes, **IF** the service needs to reload, **THEN** a method **MAY** be provided to refresh the template.

---

### 6. Conversation Context Integration

**User Story:** As a developer executing lifecycle hooks, I want to optionally associate a conversation ID with the service, so that hook methods automatically include conversation context without requiring it as a parameter.

**Acceptance Criteria:**

6.1. **WHEN** initializing the service, **IF** a conversation ID is relevant, **THEN** it **MAY** be set via a constructor parameter or setter method.

6.2. **WHEN** calling lifecycle hook methods, **IF** a conversation ID is set on the service, **THEN** the methods **SHALL NOT** require conversationID as a parameter.

6.3. **WHEN** calling lifecycle hook methods, **IF** no conversation ID is set on the service, **THEN** the methods **SHALL** use a zero/nil value or skip conversation-specific logic.

6.4. **WHEN** the conversation context changes during service lifetime, **IF** needed, **THEN** a method **MAY** be provided to update the conversation ID.

6.5. **WHEN** conversation context is not applicable, **IF** the service is used without a conversation ID, **THEN** all operations **SHALL** work correctly without it.

---

### 7. Orchestrator Method Integration

**User Story:** As a developer using sync orchestration, I want these to be normal service methods rather than package-level functions, so that the API is consistent and I can use the service's internal client.

**Acceptance Criteria:**

7.1. **WHEN** calling OrchestratePullSync, **IF** converted to a receiver method, **THEN** it **SHALL** use the service's internal client and sandboxInfo.

7.2. **WHEN** calling OrchestratePushSync, **IF** converted to a receiver method, **THEN** it **SHALL** use the service's internal client and sandboxInfo.

7.3. **WHEN** these orchestrator methods are called, **IF** they're receiver methods, **THEN** they **SHALL NOT** require client or sandboxInfo as parameters.

7.4. **WHEN** orchestrator methods are called, **IF** they have operation-specific parameters (like staleThresholdSeconds), **THEN** only those specific parameters **SHALL** be required.

7.5. **WHEN** looking at the service API, **IF** all methods are receiver methods, **THEN** the API **SHALL** be consistent (no mix of receiver methods and package functions).

---

### 8. Backward Compatibility and Migration

**User Story:** As a developer with existing code using the sandbox service, I want clear guidance on migrating to the new API, so that I can update my code efficiently and correctly.

**Acceptance Criteria:**

8.1. **WHEN** the refactor is complete, **IF** the old API is incompatible, **THEN** all calling code **SHALL** be updated to use the new constructor and method signatures.

8.2. **WHEN** updating calling code, **IF** breaking changes exist, **THEN** compiler errors **SHALL** clearly indicate what needs to change.

8.3. **WHEN** the refactor is tested, **IF** complete, **THEN** all existing unit tests **SHALL** pass with updated test code.

8.4. **WHEN** the refactor is tested, **IF** complete, **THEN** all existing integration tests **SHALL** pass with updated test code.

8.5. **WHEN** migration is complete, **IF** any deprecated patterns remain, **THEN** they **SHALL** be documented with migration guidance.

---

### 9. Error Handling and Validation

**User Story:** As a developer using the sandbox service, I want clear, actionable errors when initialization fails or state becomes invalid, so that I can quickly diagnose and fix issues.

**Acceptance Criteria:**

9.1. **WHEN** initializing the service, **IF** sandbox creation is required but fails, **THEN** the error **SHALL** clearly indicate what went wrong during creation and include relevant context.

9.2. **WHEN** initializing the service, **IF** database queries fail, **THEN** the error **SHALL** be wrapped with context about what failed.

9.3. **WHEN** reconstructing SandboxInfo fails, **IF** due to invalid data, **THEN** the error **SHALL** indicate what data was invalid.

9.4. **WHEN** any method is called, **IF** the service is in an invalid state, **THEN** the method **SHALL** return a clear error indicating the state issue.

9.5. **WHEN** ensuring sandbox is running fails, **IF** the sandbox cannot be started or created, **THEN** the error **SHALL** include details about why the operation failed and what the caller should do.

---

### 10. Testing and Code Quality

**User Story:** As a developer maintaining the sandbox service, I want comprehensive tests for the new stateful design, so that I can confidently refactor and extend the service.

**Acceptance Criteria:**

10.1. **WHEN** the refactor is complete, **IF** tests are written, **THEN** they **SHALL** cover the new constructor with various scenarios (existing sandbox, new sandbox creation, invalid parameters).

10.2. **WHEN** testing method calls, **IF** the service is stateful, **THEN** tests **SHALL** verify methods use internal state correctly.

10.3. **WHEN** testing GetOrCreate behavior, **IF** state is loaded or created, **THEN** tests **SHALL** verify internal state is correctly initialized for both scenarios.

10.4. **WHEN** the refactor is complete, **IF** code coverage is measured, **THEN** it **SHALL** be ≥90% for the refactored code.

10.5. **WHEN** running the full test suite, **IF** the refactor is complete, **THEN** all tests **SHALL** pass without modification to test assertions (only test setup should change).

10.6. **WHEN** the refactor is complete, **IF** linting is run, **THEN** there **SHALL** be no new linting errors or warnings.

---

### 11. Documentation and Code Comments

**User Story:** As a developer learning the new sandbox service API, I want clear documentation and examples, so that I can quickly understand how to use the stateful service pattern.

**Acceptance Criteria:**

11.1. **WHEN** the refactor is complete, **IF** godoc is generated, **THEN** the SandboxService struct **SHALL** have a package comment explaining the stateful design.

11.2. **WHEN** reading the constructor documentation, **IF** looking at godoc, **THEN** it **SHALL** explain what the service loads and what state it maintains.

11.3. **WHEN** reading method documentation, **IF** updated, **THEN** it **SHALL** reflect the simplified signatures (no mention of removed parameters).

11.4. **WHEN** examples are needed, **IF** provided in documentation, **THEN** they **SHALL** show the typical usage pattern (initialize once, call multiple methods).

11.5. **WHEN** complex behavior exists (GetOrCreate, state management), **IF** documented, **THEN** it **SHALL** clearly explain when and how the service initializes or creates sandboxes.

---

## Non-Functional Requirements

### Performance

- Service initialization should not add significant overhead compared to the current pattern
- Caching of template and sandbox info should improve performance for repeated operations
- No performance regression in existing operations

### Maintainability

- Code should be easier to understand with reduced parameter threading
- Adding new methods should be simpler (fewer parameters to thread through)
- Service state should be clearly defined in the struct

### Testability

- Tests should be simpler to write with less mocking of repeated parameters
- State-based testing should be straightforward
- Existing test patterns should translate cleanly to the new design

---

## Success Criteria

This refactor will be considered successful when:

1. ✅ All method signatures are simplified (no more sandboxInfo, sandboxModel, user threading)
2. ✅ Service initialization implements GetOrCreate pattern (loads existing or creates new sandbox)
3. ✅ Service handles all state management internally
4. ✅ All existing functionality works identically to before
5. ✅ All tests pass with ≥90% coverage
6. ✅ Calling code is cleaner and easier to understand
7. ✅ No performance regressions
8. ✅ Documentation clearly explains the new stateful GetOrCreate pattern

---

## Out of Scope

The following are explicitly **NOT** part of this refactor:

- ❌ Changing the Modal API integration or client behavior
- ❌ Modifying lifecycle hook execution logic (only parameter simplification)
- ❌ Changing S3 sync orchestration behavior (only making them receiver methods)
- ❌ Adding new features or capabilities to the sandbox service
- ❌ Modifying the database schema or sandbox model
- ❌ Changing external APIs or contracts with calling code (beyond constructor/method signatures)
- ❌ Performance optimization beyond what naturally comes from reduced allocations
