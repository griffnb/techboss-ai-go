# GitHub App Integration - Implementation Tasks

This document contains implementation tasks optimized for Ralph Wiggum autonomous execution. Each task includes clear completion criteria, automatic verification steps, and self-correction loops.

## Task 1: Add GitHub Configuration to Environment

**Requirements:** Requirement 1.3 - Store GitHub App credentials

**Implementation:**
1. Open `internal/environment/config.go`
2. Add `Github` field to `Config` struct
3. Create `Github` struct with fields: `AppID`, `PrivateKey`, `ClientID`, `ClientSecret`, `WebhookSecret`
4. Ensure all fields are exported (capitalized)

**Verification:**
- Run linter: `#code_tools lint`
- Expected: No lint errors in config.go

**Self-Correction:**
- If compilation errors: Check field names are exported, fix syntax, re-compile
- If tests fail: Review error output, ensure struct tags are correct, re-run tests
- If lint errors: Fix formatting issues, re-run linter

**Completion Criteria:**
- [x] Github struct added to Config with all required fields
- [x] All tests passing
- [x] No lint errors

**Status:** ✅ COMPLETED

**Escape Condition:** If stuck after 3 iterations, document the blocker and move to next task.

---

## Task 2: Generate GithubInstallation Model Scaffold

**Requirements:** Requirement 4.1 - GithubInstallation model structure

**Implementation:**
1. Run: `#code_tools make_object github_installation github_installations`
2. This will generate:
   - `internal/models/github_installation/github_installation.go`
   - `internal/controllers/github_installations/` directory with CRUD endpoints
3. Verify generated files exist

**Verification:**
- Check file exists: `internal/models/github_installation/github_installation.go`
- Check controller exists: `internal/controllers/github_installations/setup.go`
- Run: `#code_tools run_tests ./internal/models/github_installation`
- Expected: Generated tests compile and pass

**Self-Correction:**
- If make_object fails: Check command syntax, verify tool is available, retry
- If files not created: Review error messages, check permissions, retry
- If compilation errors: The generated scaffold should compile - investigate error, may need to proceed to Task 3 to fix

**Completion Criteria:**
- [x] Model scaffold generated at correct path
- [x] Controller scaffold generated
- [x] Generated code compiles

**Escape Condition:** If make_object tool unavailable, document blocker and skip to manual implementation in Task 3.

---

## Task 3: Customize GithubInstallation Model

**Requirements:** Requirements 2.3, 4.1 - Installation data storage

**Implementation:**
1. Open `internal/models/github_installation/github_installation.go`
2. Define `GithubAccountType` constant type:
   ```go
   type GithubAccountType int
   const (
       ACCOUNT_TYPE_USER GithubAccountType = iota
       ACCOUNT_TYPE_ORGANIZATION
   )
   ```
3. Update `DBColumns` struct with these fields:
   - `AccountID *fields.UUIDField` with tags: `column:"account_id" type:"uuid" default:"null" index:"true" null:"true"`
   - `InstallationID *fields.StringField` with tags: `column:"installation_id" type:"text" index:"true" unique:"true"`
   - `GithubAccountID *fields.StringField` with tags: `column:"github_account_id" type:"text" index:"true"`
   - `GithubAccountType *fields.IntConstantField[GithubAccountType]` with tags: `column:"github_account_type" type:"smallint" default:"0"`
   - `GithubAccountName *fields.StringField` with tags: `column:"github_account_name" type:"text"`
   - `RepositoryAccess *fields.StringField` with tags: `column:"repository_access" type:"text" default:"all"`
   - `Permissions *fields.StructField[map[string]any]` with tags: `column:"permissions" type:"jsonb" default:"{}"`
   - `Suspended *fields.IntField` with tags: `column:"suspended" type:"smallint" default:"0" index:"true"`
   - `AppSlug *fields.StringField` with tags: `column:"app_slug" type:"text"`
4. Add query functions:
   - `GetByInstallationID(ctx context.Context, installationID string) (*GithubInstallation, error)`
   - `GetByAccountID(ctx context.Context, accountID types.UUID) ([]*GithubInstallation, error)`
   - `GetActiveByAccountID(ctx context.Context, accountID types.UUID) ([]*GithubInstallation, error)`

**Verification:**
- Run tests: `#code_tools run_tests ./internal/models/github_installation`
- Expected: All tests pass
- Run type check: `#code_tools typecheck ./internal/models/github_installation`
- Expected: No type errors
- Verify struct tags are correctly formatted (no missing quotes or commas)

**Self-Correction:**
- If compilation errors: Check field types match imports, fix syntax, re-compile
- If tests fail: Review test output, ensure query functions have correct signatures, fix and re-run
- If type errors: Ensure `IntConstantField` generic type is correct, fix imports

**Completion Criteria:**
- [x] All fields added to DBColumns with correct types and tags
- [x] GithubAccountType constants defined
- [x] Query functions implemented
- [x] All tests passing
- [x] No type errors

**Escape Condition:** If stuck after 3 iterations, document specific compilation errors and continue.

---

## Task 4: Add GithubInstallation to Init Migration

**Requirements:** Requirement 4.1 - Database schema

**Implementation:**
1. Locate the init migration file (likely in `internal/migrations/`)
2. Create struct `GithubInstallationV1` with same fields as model's `DBColumns`
3. Add to migration's model list
4. Ensure struct is registered for table creation

**Verification:**
- Run migration test: `#code_tools run_tests ./internal/migrations`
- Expected: Migration compiles and can generate SQL
- Check that struct matches model (field names, types, tags)
- Run: `#code_tools lint ./internal/migrations`
- Expected: No lint errors

**Self-Correction:**
- If migration fails: Check struct tags match model exactly, fix discrepancies, re-run
- If tests fail: Review error message, ensure struct is properly registered, fix and retry
- If lint errors: Fix formatting, re-run linter

**Completion Criteria:**
- [x] GithubInstallationV1 added to init migration
- [x] Struct properly registered
- [x] Migration tests pass
- [x] No lint errors

**Escape Condition:** If migration pattern unclear after 3 attempts, document structure and continue.

---

## Task 5: Implement GitHub Auth Service - JWT Generation

**Requirements:** Requirement 3.1, 3.2 - JWT token generation

**Implementation:**
1. Create file: `internal/services/github_service/auth.go`
2. Create `AuthService` struct with fields:
   - `appID string`
   - `privateKey *rsa.PrivateKey`
   - `tokenCache *TokenCache`
3. Create `TokenCache` struct with:
   - `mu sync.RWMutex`
   - `tokens map[string]*InstallationToken`
4. Create `InstallationToken` struct with `Token string` and `ExpiresAt time.Time`
5. Implement `NewAuthService(appID string, privateKeyPEM string) (*AuthService, error)`:
   - Parse PEM private key using `x509.ParsePKCS1PrivateKey`
   - Initialize token cache
   - Return AuthService instance
6. Implement `GenerateJWT() (string, error)`:
   - Create JWT claims with `iss` (app ID), `iat` (now), `exp` (now + 10 min)
   - Sign with RS256 using private key
   - Return JWT string

**Verification:**
- Write test in `auth_test.go` that verifies JWT can be generated
- Run tests: `#code_tools run_tests ./internal/services/github_service`
- Expected: JWT generation test passes, token is valid format
- Verify JWT contains correct claims (decode and check)

**Self-Correction:**
- If key parsing fails: Check PEM format expectations, handle errors properly, retry
- If JWT signing fails: Verify RS256 algorithm, check private key type, fix and retry
- If tests fail: Review JWT structure, ensure claims are correct, fix and re-run

**Completion Criteria:**
- [x] AuthService struct implemented
- [x] NewAuthService constructor works
- [x] GenerateJWT produces valid JWT
- [x] Tests pass
- [x] No compilation errors

**Status:** ✅ COMPLETED

**Escape Condition:** If JWT library issues after 3 attempts, document the problem and continue to next task.

---

## Task 6: Implement GitHub Auth Service - Token Exchange

**Requirements:** Requirement 3.2 - Installation token exchange and caching

**Implementation:**
1. In `internal/services/github_service/auth.go`, implement `GetInstallationToken(ctx context.Context, installationID string) (string, error)`:
   - Check cache for valid token (not expired within 10 minutes)
   - If cached and valid, return cached token
   - If not cached or expired:
     - Generate JWT using `GenerateJWT()`
     - Create GitHub client with JWT
     - Call GitHub API: `POST /app/installations/{installationID}/access_tokens`
     - Cache the token with expiration time
     - Return token
2. Implement `RefreshInstallationToken(ctx context.Context, installationID string) (string, error)`:
   - Force token refresh (skip cache)
   - Generate new token via API
   - Update cache
   - Return new token
3. Add thread-safe cache operations using RWMutex

**Verification:**
- Write test that verifies token caching works
- Write test that verifies token refresh works
- Mock GitHub API responses for testing
- Run tests: `#code_tools run_tests ./internal/services/github_service`
- Expected: All token exchange tests pass
- Verify cache prevents unnecessary API calls

**Self-Correction:**
- If API call fails: Check GitHub client setup, verify JWT is valid, mock properly in tests
- If caching broken: Review mutex usage, ensure thread safety, fix race conditions
- If tests fail: Check mock expectations, verify token structure, fix and retry

**Completion Criteria:**
- [x] GetInstallationToken implemented with caching
- [x] RefreshInstallationToken implemented
- [x] Thread-safe cache operations
- [x] All tests pass
- [x] Token caching verified

**Escape Condition:** If GitHub API integration complex after 3 attempts, implement basic version and document limitations.

---

## Task 7: Implement GitHub API Service

**Requirements:** Requirement 6.2 - GitHub API operations

**Implementation:**
1. Create file: `internal/services/github_service/api.go`
2. Create `APIService` struct with `authService *AuthService`
3. Implement `NewAPIService(authService *AuthService) *APIService`
4. Implement `getClient(ctx context.Context, installationID string) (*github.Client, error)`:
   - Get token from auth service
   - Create OAuth2 token source
   - Return authenticated GitHub client
5. Implement repository operations:
   - `GetRepository(ctx, installationID, owner, repo string) (*github.Repository, error)`
   - `ListRepositories(ctx, installationID string) ([]*github.Repository, error)`
6. Implement branch operations:
   - `GetBranch(ctx, installationID, owner, repo, branch string) (*github.Branch, error)`
   - `CreateBranch(ctx, installationID, owner, repo, branch, sha string) error`
7. Implement PR operations:
   - `CreatePullRequest(ctx, installationID, owner, repo string, pr *github.NewPullRequest) (*github.PullRequest, error)`

**Verification:**
- Write tests with mocked GitHub API
- Run tests: `#code_tools run_tests ./internal/services/github_service`
- Expected: All API operations tests pass
- Verify each operation calls correct GitHub API endpoint

**Self-Correction:**
- If GitHub client creation fails: Check token format, verify OAuth2 setup, fix and retry
- If API calls fail: Review GitHub SDK usage, check parameters, fix and retry
- If tests fail: Verify mock setup, check return types, fix and re-run

**Completion Criteria:**
- [x] APIService implemented
- [x] All repository operations work
- [x] All branch operations work
- [x] All PR operations work
- [x] Tests pass

**Escape Condition:** If GitHub SDK integration problematic after 3 attempts, implement core operations only and document missing features.

---

## Task 8: Add Webhook Validation to Auth Service

**Requirements:** Requirement 7.1 - Webhook signature validation

**Implementation:**
1. In `internal/services/github_service/auth.go`, implement `ValidateWebhookSignature(payload []byte, signature string, secret string) bool`:
   - Extract signature from header (format: `sha256=...`)
   - Compute HMAC-SHA256 of payload with secret
   - Compare computed signature with provided signature
   - Return true if valid, false otherwise
2. Use constant-time comparison to prevent timing attacks

**Verification:**
- Write test with known payload and signature
- Run tests: `#code_tools run_tests ./internal/services/github_service`
- Expected: Signature validation test passes
- Verify both valid and invalid signatures are handled correctly

**Self-Correction:**
- If HMAC computation fails: Check crypto library usage, verify algorithm, fix and retry
- If comparison fails: Ensure proper string format (remove `sha256=` prefix), fix and retry
- If tests fail: Check test vectors, verify hex encoding, fix and re-run

**Completion Criteria:**
- [x] ValidateWebhookSignature implemented
- [x] Uses constant-time comparison
- [x] Tests pass for valid signatures
- [x] Tests pass for invalid signatures

**Escape Condition:** If HMAC validation complex after 3 attempts, implement basic version and document security considerations.

---

## Task 9: Add Custom Webhook Endpoint to Controller

**Requirements:** Requirements 2.2, 2.3, 2.5 - Process installation webhooks

**Implementation:**
1. Create file: `internal/controllers/github_installations/webhook.go`
2. Define webhook payload structs:
   - `WebhookPayload` with `Action`, `Installation`, `Repositories`, `Sender`
   - `WebhookInstallation` with installation fields matching GitHub webhook format
3. Implement `webhookCallback(res http.ResponseWriter, req *http.Request)`:
   - Read request body
   - Validate webhook signature using auth service
   - Parse JSON payload into WebhookPayload struct
   - Handle events:
     - `installation.created`: Create GithubInstallation record
     - `installation.deleted`: Mark installation as deleted
     - `installation.suspend`: Set suspended flag
     - `installation.unsuspend`: Clear suspended flag
   - Return 200 on success, 400/401 on errors
4. In `internal/controllers/github_installations/setup.go`, add route:
   - `r.Post("/callback", webhookCallback)`

**Verification:**
- Write test that simulates webhook delivery
- Mock GitHub webhook payload
- Run tests: `#code_tools run_tests ./internal/controllers/github_installations`
- Expected: Webhook tests pass
- Verify database records are created/updated correctly

**Self-Correction:**
- If signature validation fails: Check secret configuration, verify HMAC implementation, fix and retry
- If JSON parsing fails: Check struct tags, verify payload structure, fix and retry
- If database operations fail: Check model methods, verify transaction handling, fix and retry
- If tests fail: Review mock payload format, check assertions, fix and re-run

**Completion Criteria:**
- [x] webhookCallback implemented
- [x] All webhook events handled
- [x] Signature validation integrated
- [x] Route added to setup
- [x] Tests pass

**Escape Condition:** If webhook handling complex after 3 attempts, implement basic event handling and document incomplete events.

---

## Task 10: Implement Modal GitHub Template

**Requirements:** Requirements 5.1, 5.2, 5.3 - GitHub sandbox template

**Implementation:**
1. Create file: `internal/services/sandbox_service/github_template.go`
2. Define `GitHubTemplateConfig` struct with:
   - `InstallationID string`
   - `Repository string` (format: "owner/repo")
   - `SourceBranch string`
   - `TargetBranch string`
   - `PRTargetBranch string`
   - `PRTitle string`
   - `PRBody string`
   - `GitUserName string`
   - `GitUserEmail string`
3. Implement `GetGitHubImage() *modal.ImageConfig`:
   - Base image: `python:3.11-slim`
   - Install Git and GitHub CLI via apt
   - Return ImageConfig
4. Implement `GetGitHubTemplate(config *GitHubTemplateConfig) *SandboxTemplate`:
   - Set Provider to Modal
   - Set ImageConfig from GetGitHubImage()
   - Define OnColdStart hook that:
     - Sets up Git credentials from GITHUB_TOKEN env var
     - Configures git user.name and user.email
     - Clones repository
     - Checks out source branch
     - Creates target branch
     - Creates draft PR
   - Define OnTerminate hook that:
     - Commits uncommitted changes
     - Pushes to remote
   - Return SandboxTemplate

**Verification:**
- Run tests: `#code_tools run_tests ./internal/services/sandbox_service`
- Expected: Template generation tests pass
- Verify lifecycle hooks are correctly defined
- Check image config includes Git and GitHub CLI

**Self-Correction:**
- If image config invalid: Review Modal API docs, fix Dockerfile commands, retry
- If hooks fail: Check bash script syntax, verify environment variables, fix and retry
- If tests fail: Review template structure, check hook logic, fix and re-run

**Completion Criteria:**
- [x] GitHubTemplateConfig struct defined
- [x] GetGitHubImage implemented
- [x] GetGitHubTemplate implemented with lifecycle hooks
- [x] Tests pass

**Escape Condition:** If Modal template complex after 3 attempts, implement basic version with documented limitations.

---

## Task 11: Extend Sandbox Service with GitHub Support

**Requirements:** Requirement 5.5 - Integration with sandbox service

**Implementation:**
1. Create file: `internal/services/sandbox_service/github.go`
2. Implement `CreateGitHubSandbox(ctx context.Context, accountID types.UUID, installationID string, config *GitHubTemplateConfig) (*modal.SandboxInfo, error)`:
   - Get installation token from github auth service
   - Build sandbox config from GetGitHubTemplate(config)
   - Add environment variables:
     - `GITHUB_TOKEN`: installation token
     - `REPOSITORY`: config.Repository
     - `SOURCE_BRANCH`: config.SourceBranch
     - `TARGET_BRANCH`: config.TargetBranch
     - `PR_TARGET_BRANCH`: config.PRTargetBranch
     - `PR_TITLE`: config.PRTitle
     - `PR_BODY`: config.PRBody
     - `GIT_USER_NAME`: config.GitUserName
     - `GIT_USER_EMAIL`: config.GitUserEmail
   - Call existing CreateSandbox method with modified config
   - Return sandbox info

**Verification:**
- Write test with mocked auth service and Modal client
- Run tests: `#code_tools run_tests ./internal/services/sandbox_service`
- Expected: GitHub sandbox creation test passes
- Verify all environment variables are set correctly

**Self-Correction:**
- If token retrieval fails: Check auth service integration, verify installation ID, fix and retry
- If sandbox creation fails: Review config structure, check Modal client, fix and retry
- If tests fail: Verify mocks, check environment setup, fix and re-run

**Completion Criteria:**
- [x] CreateGitHubSandbox implemented
- [x] Token retrieval integrated
- [x] Environment variables set correctly
- [x] Tests pass

**Escape Condition:** If sandbox integration complex after 3 attempts, implement basic version and document integration points.

---

## Task 12: Add Controller Documentation (Swagger)

**Requirements:** Requirement 8.3 - API documentation

**Implementation:**
1. Open generated CRUD handlers in `internal/controllers/github_installations/`
2. Add Swagger annotations to each handler:
   - `@Public` tag
   - `@Summary` with brief description
   - `@Description` with detailed explanation
   - `@Tags GitHub`
   - `@Accept json`
   - `@Produce json`
   - `@Param` annotations for path/query parameters
   - `@Success` with response models
   - `@Failure` with error responses
   - `@Router` with path and method
3. Add Swagger doc to webhook handler in `webhook.go`

**Verification:**
- Run: `#code_tools lint ./internal/controllers/github_installations`
- Expected: No lint errors
- Generate Swagger docs (if tool available)
- Verify documentation is correctly formatted

**Self-Correction:**
- If Swagger syntax errors: Review swaggo documentation, fix annotations, retry
- If lint errors: Fix formatting, re-run linter
- If documentation incomplete: Add missing annotations, regenerate

**Completion Criteria:**
- [x] All handlers have Swagger documentation
- [x] No lint errors
- [x] Documentation is complete and accurate

**Status:** ✅ COMPLETED

**Escape Condition:** If Swagger documentation unclear after 3 attempts, add basic comments and document manual documentation needs.

---

## Task 13: Write Integration Tests

**Requirements:** Requirement 11.2 - Integration testing

**Implementation:**
1. Create file: `internal/services/github_service/integration_test.go`
2. Write test that:
   - Configures auth service with test credentials
   - Tests full JWT generation → token exchange flow
   - Mocks GitHub API responses
   - Verifies token caching works
3. Create file: `internal/controllers/github_installations/integration_test.go`
4. Write test that:
   - Sends webhook payload to callback endpoint
   - Verifies database record creation
   - Tests installation CRUD operations
   - Verifies auth requirements

**Verification:**
- Run tests: `#code_tools run_tests ./internal/services/github_service`
- Run tests: `#code_tools run_tests ./internal/controllers/github_installations`
- Expected: All integration tests pass
- Verify tests cover main user flows

**Self-Correction:**
- If tests fail: Review error output, check mocking setup, fix and retry
- If database operations fail: Check test fixtures, verify cleanup, fix and retry
- If auth fails: Review test credentials, check mock setup, fix and retry

**Completion Criteria:**
- [x] GitHub service integration tests pass
- [x] Controller integration tests pass
- [x] Test coverage ≥90% for new code

**Escape Condition:** If integration tests complex after 3 attempts, implement basic tests and document coverage gaps.

---

## Task 14: Write Security Tests

**Requirements:** Requirement 10.1, 10.2, 10.3 - Security validation

**Implementation:**
1. Create file: `internal/services/github_service/security_test.go`
2. Write tests for:
   - Invalid webhook signatures are rejected
   - Expired JWT tokens are refreshed
   - Invalid private keys fail gracefully
   - Token cache prevents timing attacks
3. Create file: `internal/controllers/github_installations/security_test.go`
4. Write tests for:
   - Unauthorized access is blocked
   - Installation ID validation works
   - Cross-account access is prevented

**Verification:**
- Run tests: `#code_tools run_tests ./internal/services/github_service`
- Run tests: `#code_tools run_tests ./internal/controllers/github_installations`
- Expected: All security tests pass
- Verify negative test cases work correctly

**Self-Correction:**
- If security tests fail: Review attack scenarios, strengthen validation, fix and retry
- If false positives: Adjust test expectations, verify logic, fix and retry

**Completion Criteria:**
- [x] All security tests pass
- [x] Webhook signature validation tested
- [x] Auth failures handled correctly

**Escape Condition:** If security testing complex after 3 attempts, implement basic tests and document security review needs.

---

## Final Checklist When Complete

### Requirements Coverage
- [ ] Requirement 1.3: GitHub App configuration stored
- [ ] Requirement 2.2, 2.3: Installation webhooks processed
- [ ] Requirement 2.5: Installation suspension handled
- [ ] Requirement 3.1, 3.2: JWT and token exchange implemented
- [ ] Requirement 4.1: GithubInstallation model created
- [ ] Requirement 5.1, 5.2, 5.3: Modal template with lifecycle hooks
- [ ] Requirement 6.2: GitHub API operations implemented
- [ ] Requirement 7.1: Webhook signature validation
- [ ] Requirement 8.3: API documentation added
- [ ] Requirement 10.1, 10.2: Security measures implemented
- [ ] Requirement 11.2: Integration tests written

### Design Implementation
- [ ] GitHub config in environment matches design.md section 1
- [ ] GithubInstallation model matches design.md section 2
- [ ] Auth service matches design.md section 3
- [ ] API service matches design.md section 4
- [ ] Modal template matches design.md section 5
- [ ] Sandbox service extensions match design.md section 6
- [ ] Controllers match design.md section 7

### Code Quality
- [ ] All tests pass: `#code_tools run_tests ./internal/models/github_installation ./internal/services/github_service ./internal/controllers/github_installations ./internal/services/sandbox_service`
- [ ] No lint errors: `#code_tools lint`
- [ ] Test coverage ≥90%
- [ ] All public functions documented
- [ ] No compilation errors
- [ ] Security tests pass

### Verification
- [ ] JWT generation works
- [ ] Token caching works
- [ ] Webhook processing creates database records
- [ ] GitHub API calls succeed (with mocks)
- [ ] Modal template creates valid sandbox config
- [ ] Controller endpoints are accessible
- [ ] Swagger documentation generated

**Output:** <promise>COMPLETE</promise>
