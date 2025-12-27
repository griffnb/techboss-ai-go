# Agent Learnings

This file tracks important lessons learned during implementation that should inform future work.

## Date: 2025-12-27

### Issue: Sandbox Database Migration - Architectural Violations

**What I learned:**

The previous agent implementing the sandbox-database-migration spec created duplicate `reconstructSandboxInfo()` logic in two places:

1. **Service Layer** (✅ Correct): `internal/services/sandbox_service/reconstruct.go` - Created `ReconstructSandboxInfo()` as a proper service method
2. **Controller Layer** (❌ Wrong): `internal/controllers/sandboxes/sandbox.go` - Created a local helper `reconstructSandboxInfo()`

The controller used its own local helper instead of calling the service layer method, violating the core architectural principle that **controllers should be simplistic and delegate business logic to services**.

**Why this happened:**

The spec said "Add `reconstructSandboxInfo()` helper function in `sandbox_service`" but the agent:
- ✅ Created it correctly in the service layer
- ❌ Also created a duplicate in the controller
- ❌ Used the controller version instead of the service version

This suggests the agent may have:
1. Misunderstood "helper function" to mean a local utility
2. Not recognized that helper functions with business logic belong in services, not controllers
3. Created the controller version first, then the service version, but forgot to refactor

**Impact:**

- Helper functions in controllers = ❌ Wrong architecture
- Business logic in controllers = ❌ Violates separation of concerns
- Duplicate code = ❌ Maintenance burden
- Controllers not calling services = ❌ Defeats purpose of service layer

**How to improve instructions:**

1. **Be explicit about architecture layers:**
   - ❌ "Add a helper function" (ambiguous)
   - ✅ "Add a service method in `sandbox_service` and call it from controllers"

2. **Emphasize the principle:**
   - "Controllers must be thin - they orchestrate, not implement"
   - "All business logic must be in the service layer"
   - "Controllers should only: parse requests, call services, return responses"

3. **Add verification steps:**
   - "Verify no business logic exists in controller files"
   - "Grep for helper functions in controllers - there should be none"
   - "Controllers should only have request/response handling code"

4. **Reference examples:**
   - "Follow the pattern in `other_controller.go` where `service.Method()` is called"
   - "Look at existing controllers - they should never have helper functions"

**The fix:**

1. Deleted the controller helper `reconstructSandboxInfo()` from `sandbox.go` (lines 98-143)
2. Updated all controller references to use `sandbox_service.ReconstructSandboxInfo()`
3. Updated all test references to use the service layer method
4. Removed unused imports (`time`, `modal`)
5. Verified code compiles and no local helpers remain

**Files changed:**
- `internal/controllers/sandboxes/sandbox.go` - Removed helper, updated call
- `internal/controllers/sandboxes/auth.go` - Updated call
- `internal/controllers/sandboxes/claude.go` - Updated call
- `internal/controllers/sandboxes/sandbox_test.go` - Updated all test calls
- `internal/controllers/sandboxes/claude_test.go` - Updated all test calls

**Key takeaway for future specs:**

When writing specs or delegating to agents, be crystal clear about:
1. **WHERE** code should live (which layer, which file)
2. **WHO** should call it (service calls integration, controller calls service)
3. **WHAT** each layer is responsible for (thin controllers, fat services)
4. **WHY** this matters (separation of concerns, testability, maintainability)

Don't assume agents will infer architectural patterns - be explicit and reference existing good examples.
