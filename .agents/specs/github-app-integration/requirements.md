# GitHub App Integration - Requirements

## Introduction

This feature enables GitHub App authentication and integration with Modal sandboxes to perform automated Git operations (pull, push, PR creation) and Claude-assisted development workflows. The system will allow users to install a GitHub App on their organization or repository, provide secure credential management, and execute automated branch operations with Claude integration through Modal sandboxes.

## Requirements

### 1. GitHub App Registration and Configuration

**User Story:** As a system administrator, I want to register and configure a GitHub App, so that users can install it on their repositories and organizations for automated Git operations.

**Acceptance Criteria:**

1.1. **WHEN** the system is deployed, **THE** GitHub App **SHALL** be registered with the following scopes:
   - Repository contents (read/write)
   - Pull requests (read/write)
   - Repository metadata (read)
   - Commit statuses (read/write)

1.2. **WHEN** registering the GitHub App, **THE** system **SHALL** configure webhook endpoints for:
   - Installation events
   - Installation repository events
   - Repository events

1.3. **WHEN** the GitHub App is registered, **THE** system **SHALL** securely store:
   - App ID
   - Private key (PEM format)
   - Client ID
   - Client secret
   - Webhook secret

1.4. **IF** the GitHub App configuration is invalid, **THEN** the system **SHALL** prevent GitHub operations and display configuration error messages.

### 2. GitHub App Installation Flow

**User Story:** As a user, I want to install the GitHub App on my organization or repository, so that the system can access and manage my code through automated workflows.

**Acceptance Criteria:**

2.1. **WHEN** a user initiates GitHub App installation, **THE** system **SHALL** redirect them to GitHub's installation page with proper authorization scopes.

2.2. **WHEN** the user completes installation, **THE** system **SHALL** receive and process the installation webhook event.

2.3. **WHEN** processing an installation event, **THE** system **SHALL** store:
   - Installation ID
   - Account ID (user or organization)
   - Repository access permissions
   - Installation timestamp

2.4. **WHEN** a user has multiple installations, **THE** system **SHALL** allow them to select which installation to use for operations.

2.5. **IF** an installation is suspended or uninstalled, **THEN** the system **SHALL** update the installation status and prevent operations using that installation.

### 3. GitHub App Authentication Service

**User Story:** As a developer, I want a service that handles GitHub App authentication, so that I can obtain access tokens for API operations securely.

**Acceptance Criteria:**

3.1. **WHEN** the system needs to perform GitHub operations, **THE** authentication service **SHALL** generate a JWT token signed with the App's private key.

3.2. **WHEN** generating installation access tokens, **THE** service **SHALL**:
   - Verify the installation is active
   - Generate JWT with appropriate claims (iss, iat, exp)
   - Exchange JWT for installation access token
   - Cache tokens until 10 minutes before expiration

3.3. **WHEN** an installation access token expires, **THE** service **SHALL** automatically refresh it before making API calls.

3.4. **IF** authentication fails, **THEN** the service **SHALL** return specific error codes indicating:
   - Invalid credentials
   - Expired token
   - Insufficient permissions
   - Installation suspended

3.5. **WHEN** authenticating, **THE** service **SHALL** support both organization-level and repository-level installations.

### 4. GitHub Operations Model

**User Story:** As a system, I need a data model for tracking GitHub installations and operations, so that I can maintain state and audit Git activities.

**Acceptance Criteria:**

4.1. **WHEN** storing GitHub installations, **THE** system **SHALL** maintain a GithubInstallation model with:
   - Installation ID (unique identifier)
   - Account ID (organization or user)
   - Account type (organization/user)
   - Repository access (all or selected)
   - Installation status (active/suspended/deleted)
   - Permissions granted
   - Associated account/user ID
   - Created/updated timestamps

### 5. Modal Sandbox GitHub Template

**User Story:** As a user, I want a Modal sandbox template for GitHub operations, so that I can automate branch creation, changes, and PR workflows with Claude assistance.

**Acceptance Criteria:**

5.1. **WHEN** creating a GitHub Modal sandbox, **THE** template **SHALL** accept the following inputs:
   - Installation ID
   - Repository full name (owner/repo)
   - Source branch name (branch to clone from)
   - Target branch name (branch to create)
   - PR target branch (branch to open PR against)
   - PR title (draft PR title)
   - PR body (draft PR description)
   - Modal configuration (Claude model, instructions, etc.)

5.2. **WHEN** the Modal sandbox starts, **IT** **SHALL**:
   - Authenticate using the GitHub App installation
   - Validate repository access permissions
   - Clone the repository at the specified source branch
   - Create and checkout the new target branch
   - Configure git user information

5.3. **WHEN** git operations fail, **THE** sandbox **SHALL**:
   - Capture detailed error messages
   - Return specific error codes (auth failure, permission denied, branch exists, etc.)
   - Clean up any partial work

5.4. **WHEN** the target branch is ready, **THE** sandbox **SHALL**:
   - Create a draft pull request against the PR target branch
   - Set PR title and description from inputs

5.5. **WHEN** the PR is created, **THE** sandbox **SHALL**:
   - Install Claude with specified configuration
   - Configure Claude to work on the checked-out branch
   - Provide Claude with context about the PR and branch
   - Enable Claude to make commits and push changes

5.6. **WHEN** the sandbox terminates, **IT** **SHALL**:
   - Clean up temporary files
   - Preserve git history and PR state

### 6. GitHub API Integration Service

**User Story:** As a developer, I need a service that wraps GitHub API operations, so that I can perform Git operations programmatically with proper error handling.

**Acceptance Criteria:**

6.1. **WHEN** performing repository operations, **THE** service **SHALL** support:
   - Getting repository information
   - Listing branches
   - Getting branch details
   - Creating branches
   - Comparing branches

6.2. **WHEN** performing pull request operations, **THE** service **SHALL** support:
   - Creating pull requests (draft and ready)
   - Updating pull request status
   - Getting pull request details
   - Adding pull request comments
   - Requesting pull request reviews

6.3. **WHEN** performing commit operations, **THE** service **SHALL** support:
   - Getting commit history
   - Creating commits
   - Getting commit details
   - Comparing commits

6.4. **WHEN** API calls fail, **THE** service **SHALL**:
   - Retry transient failures (rate limits, network errors)
   - Return structured error responses
   - Log errors for debugging
   - Implement exponential backoff for retries

6.5. **WHEN** encountering rate limits, **THE** service **SHALL**:
   - Respect GitHub's rate limit headers
   - Wait until rate limit reset when necessary
   - Use conditional requests to save rate limit quota
   - Log rate limit status

### 7. Controller Endpoints for GitHub Operations

**User Story:** As a frontend developer, I need API endpoints to manage GitHub installations and trigger operations, so that users can interact with GitHub functionality through the UI.

**Acceptance Criteria:**

7.1. **WHEN** accessing GitHub endpoints, **THE** API **SHALL** provide:
   - `GET /api/github/installations` - List user's GitHub installations
   - `GET /api/github/installations/:id` - Get installation details
   - `POST /api/github/installations/callback` - Handle GitHub installation callback
   - `DELETE /api/github/installations/:id` - Remove installation record

7.2. **WHEN** validating requests, **THE** endpoints **SHALL**:
   - Verify user authentication
   - Validate installation ownership
   - Validate repository access permissions
   - Check for required parameters
   - Return appropriate error codes (400, 401, 403, 404)

7.3. **WHEN** returning responses, **THE** endpoints **SHALL**:
   - Return structured error messages
   - Include relevant resource links (installation URLs, repository URLs)
   - Follow consistent response format

### 8. Sandbox Service GitHub Integration

**User Story:** As a sandbox service, I need to launch Modal sandboxes with GitHub credentials and template, so that automated Git workflows can be executed securely.

**Acceptance Criteria:**

8.1. **WHEN** the sandbox service needs to create a GitHub sandbox, **IT** **SHALL** accept:
   - Installation ID
   - GitHub template identifier
   - Template-specific parameters (repo, branches, etc.)

8.2. **WHEN** launching a GitHub sandbox, **THE** service **SHALL**:
   - Retrieve installation access token from authentication service
   - Validate installation is active and accessible
   - Pass credentials securely to the Modal sandbox
   - Track sandbox creation in system

8.3. **WHEN** credentials are passed to sandboxes, **THE** service **SHALL**:
   - Use secure environment variables
   - Never log or expose credentials
   - Ensure credentials are scoped to minimum permissions
   - Clean up credentials after sandbox terminates

8.4. **IF** credential retrieval fails, **THEN** the service **SHALL**:
   - Return specific error indicating auth failure
   - Not create the sandbox
   - Log failure for debugging

### 9. Security and Permissions

**User Story:** As a security-conscious user, I want GitHub operations to be secure and properly authorized, so that only authorized users can perform Git operations on my repositories.

**Acceptance Criteria:**

9.1. **WHEN** storing GitHub App credentials, **THE** system **SHALL**:
   - Encrypt private keys at rest
   - Store webhook secrets securely
   - Never expose credentials in logs or API responses
   - Use environment variables for sensitive configuration

9.2. **WHEN** verifying webhook signatures, **THE** system **SHALL**:
   - Validate HMAC signature on all webhook payloads
   - Reject webhooks with invalid signatures
   - Log suspicious webhook attempts

9.3. **WHEN** authorizing operations, **THE** system **SHALL**:
   - Verify user owns or has access to the installation
   - Check repository permissions before operations
   - Validate installation is active
   - Prevent cross-account access

9.4. **WHEN** performing Git operations in sandboxes, **THE** system **SHALL**:
   - Use installation tokens (not personal tokens)
   - Scope tokens to minimum required permissions
   - Rotate tokens appropriately
   - Clean up credentials after operations complete

### 10. Error Handling and Recovery

**User Story:** As a user, I want clear error messages and automatic recovery when GitHub operations fail, so that I can understand and resolve issues quickly.

**Acceptance Criteria:**

10.1. **WHEN** errors occur, **THE** system **SHALL** provide error messages that indicate:
   - The specific operation that failed
   - The reason for failure
   - Suggested remediation steps
   - Whether retry is possible

9.2. **WHEN** transient failures occur, **THE** system **SHALL**:
   - Automatically retry with exponential backoff
   - Track retry attempts
   - Give up after maximum retry limit
   - Log failure context for debugging

10.3. **IF** authentication expires mid-operation, **THEN** the system **SHALL**:
   - Attempt to refresh tokens
   - Resume operation if possible
   - Fail gracefully if refresh fails
   - Record authentication failure details

### 11. Frontend Integration

**User Story:** As a user, I want to manage GitHub installations and trigger operations through the UI, so that I can easily integrate GitHub into my workflows without using API calls directly.

**Acceptance Criteria:**

11.1. **WHEN** viewing GitHub settings, **THE** UI **SHALL** display:
   - List of installed GitHub Apps
   - Installation status (active/suspended)
   - Repository access scope
   - Installation date
   - Uninstall/reinstall options

11.2. **IF** installation is required, **THEN** the UI **SHALL**:
   - Show "Install GitHub App" button
   - Redirect to GitHub installation flow
   - Handle installation callback
   - Refresh installation list after completion

### 12. Testing and Validation

**User Story:** As a developer, I need comprehensive tests for GitHub integration, so that I can ensure reliability and catch regressions.

**Acceptance Criteria:**

11.1. **WHEN** testing authentication, **THE** test suite **SHALL** include:
   - JWT generation and validation
   - Installation token generation
   - Token caching and refresh
   - Authentication failure scenarios

11.2. **WHEN** testing GitHub API integration, **THE** test suite **SHALL** include:
   - Mocked API responses for all operations
   - Rate limit handling
   - Error response handling
   - Retry logic validation

11.3. **WHEN** testing Modal sandbox operations, **THE** test suite **SHALL** include:
   - Git operation simulation
   - Branch creation and switching
   - PR creation workflows
   - Failure and cleanup scenarios

12.4. **WHEN** testing controllers, **THE** test suite **SHALL** include:
   - Authorization checks
   - Input validation
   - Error responses
   - Successful operation flows

12.5. **WHEN** testing security, **THE** test suite **SHALL** include:
   - Webhook signature verification
   - Cross-account access prevention
   - Credential encryption/decryption
   - Permission validation
