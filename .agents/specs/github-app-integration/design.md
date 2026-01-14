# GitHub App Integration - Design Document

## Overview

This design implements GitHub App authentication and Modal sandbox integration for automated Git operations with Claude assistance. The system enables users to install a GitHub App on their repositories, provides secure credential management through JWT and installation tokens, and supports launching Modal sandboxes pre-configured with Git access.

### Key Components

1. **GitHub App Registration & Config** - Central configuration for GitHub App credentials
2. **GithubInstallation Model** - Stores installation data per user/organization  
3. **GitHub Authentication Service** - Generates JWTs and exchanges for installation tokens
4. **GitHub API Service** - Wraps GitHub REST API operations
5. **Modal GitHub Template** - Pre-configured sandbox template with Git setup
6. **Sandbox Service Integration** - Extends sandbox service to support GitHub credentials
7. **Controller Endpoints** - REST API for installation management

## Architecture

### High-Level Flow

```
User Installs GitHub App
        ↓
GitHub Webhook → Store Installation
        ↓
User Launches Sandbox
        ↓
Sandbox Service → Auth Service (Get Token)
        ↓
Modal Template → Clone Repo → Create Branch → Create PR
        ↓
Claude Works in Sandbox
        ↓
Claude Commits & Pushes
```

### Component Relationships

```
┌─────────────────────────────────────────────────────────────┐
│              Controllers (Go Chi Router)                    │
│  /api/github_installations/*                                │
└────────────────────┬──────────────────────────────────────┘
                     │
     ┌───────────────┴─────────────────┐
     │                                   │
┌────▼──────────────────┐   ┌───────────▼────────────────┐
│  GithubInstallation    │   │  Sandbox Service           │
│  Model & Queries       │   │  - Create GitHub sandbox   │
└────────────────────────┘   │  - Pass credentials        │
                             └───────────┬─────────────────┘
                                         │
                             ┌───────────▼────────────────┐
                             │  GitHub Auth Service       │
                             │  - Generate JWT            │
                             │  - Get installation token  │
                             │  - Cache tokens            │
                             └───────────┬─────────────────┘
                                         │
                             ┌───────────▼────────────────┐
                             │  GitHub API Service        │
                             │  - Repository operations   │
                             │  - Branch operations       │
                             │  - PR operations           │
                             └────────────────────────────┘
```

### Data Flow

1. **Installation Flow**
   - User clicks "Install GitHub App" in UI
   - Redirects to GitHub with App ID
   - User authorizes repositories
   - GitHub redirects to callback endpoint with `installation_id`
   - System stores installation in database

2. **Sandbox Creation Flow**
   - User/System requests GitHub-enabled sandbox
   - Sandbox Service calls GitHub Auth Service for token
   - Auth Service generates JWT → exchanges for installation token
   - Sandbox Service passes token + template to Modal
   - Modal creates sandbox with Git configured
   - Template clones repo, creates branch, creates PR
   - Claude operates in sandbox with Git access

## Components and Interfaces

### 1. GitHub App Configuration

**Location:** AWS Secrets Manager

**Secret Name:** `github-app-credentials`

**Secret Structure:**

```json
{
  "app_id": "123456",
  "private_key": "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
  "client_id": "Iv1.abc123def456",
  "client_secret": "secret123",
  "webhook_secret": "webhook_secret_123",
  "app_slug": "techboss-ai-dev"
}
```


**Generation:** Use `#code_tools make_object` to generate the model scaffold, then customize.

```go
//go:generate core_gen model GithubInstallation

package github_installation

import (
    "github.com/griffnb/core/lib/model"
    "github.com/griffnb/core/lib/model/fields"
    "github.com/griffnb/techboss-ai-go/internal/models/base"
)

const (
    TABLE        = "github_installations"
    CHANGE_LOGS  = true
    CLIENT       = environment.CLIENT_DEFAULT
    IS_VERSIONED = false
)

// GithubAccountType represents the type of GitHub account
type GithubAccountType int

const (
    ACCOUNT_TYPE_USER GithubAccountType = iota
    ACCOUNT_TYPE_ORGANIZATION
)

type Structure struct {
    DBColumns
    JoinData
}

type DBColumns struct {
    base.Structure
    AccountID           *fields.UUIDField                                `column:"account_id"           type:"uuid"     default:"null"               index:"true" null:"true"`
    InstallationID      *fields.StringField                              `column:"installation_id"      type:"text"     index:"true" unique:"true"`
    GithubAccountID     *fields.StringField                              `column:"github_account_id"    type:"text"     index:"true"`
    GithubAccountType   *fields.IntConstantField[GithubAccountType]      `column:"github_account_type"  type:"smallint" default:"0"`
    GithubAccountName   *fields.StringField                              `column:"github_account_name"  type:"text"`
    RepositoryAccess    *fields.StringField                              `column:"repository_access"    type:"text"     default:"all"`
    Permissions         *fields.StructField[map[string]any]              `column:"permissions"          type:"jsonb"    default:"{}"`
    Suspended           *fields.IntField                                 `column:"suspended"            type:"smallint" default:"0"  index:"true"`
    AppSlug             *fields.StringField                              `column:"app_slug"             type:"text"`
}

type JoinData struct{}

type GithubInstallation struct {
    model.BaseModel
    DBColumns
}

type GithubInstallationJoined struct {
    GithubInstallation
    JoinData
}
```

**Migration Structure (goes in init-migration):**

```go
type GithubInstallationV1 struct {
    base.Structure
    AccountID           *fields.UUIDField                                `column:"account_id"           type:"uuid"     default:"null"               index:"true" null:"true"`
    InstallationID      *fields.StringField                              `column:"installation_id"      type:"text"     index:"true" unique:"true"`
    GithubAccountID     *fields.StringField                              `column:"github_account_id"    type:"text"     index:"true"`
    GithubAccountType   *fields.IntConstantField[GithubAccountType]      `column:"github_account_type"  type:"smallint" default:"0"`
    GithubAccountName   *fields.StringField                              `column:"github_account_name"  type:"text"`
    RepositoryAccess    *fields.StringField                              `column:"repository_access"    type:"text"     default:"all"`
    Permissions         *fields.StructField[map[string]any]              `column:"permissions"          type:"jsonb"    default:"{}"`
    Suspended           *fields.IntField                                 `column:"suspended"            type:"smallint" default:"0"  index:"true"`
    AppSlug             *fields.StringField                              `column:"app_slug"             type:"text"`
}
```

**Queries:**

```go
// GetByInstallationID retrieves an installation by GitHub installation ID
func GetByInstallationID(ctx context.Context, installationID string) (*GithubInstallation, error)

// GetByAccountID retrieves all installations for an account
func GetByAccountID(ctx context.Context, accountID types.UUID) ([]*GithubInstallation, error)

// GetActiveByAccountID retrieves non-suspended installations for an account
func GetActiveByAccountID(ctx context.Context, accountID types.UUID) ([]*GithubInstallation, error)
```

### 3. GitHub Authentication Service

**Location:** `internal/services/github_service/auth.go`

```go
package github_service

import (
    "context"
    "crypto/rsa"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
    "github.com/google/go-github/v66/github"
    "golang.org/x/oauth2"
)

type AuthService struct {
    appID      int64
    privateKey *rsa.PrivateKey
    tokenCache string
    privateKey *rsa.PrivateKey
    tokenCache *TokenCache
}

type InstallationToken struct {
    Token     string
    ExpiresAt time.Time
}

type TokenCache struct {
    mu     sync.RWMutex
    tokens map[string]*InstallationToken // installationID -> token
}

// NewAuthService creates a new GitHub authentication service
func NewAuthService(config *environment.Github) (*AuthService, error)

// GenerateJWT generates a JWT token for the GitHub App
// JWT is valid for 10 minutes and used to authenticate as the app
func (s *AuthService) GenerateJWT() (string, error)

// GetInstallationToken retrieves or generates an installation access token
// Tokens are cached until 10 minutes before expiration
func (s *AuthService) GetInstallationToken(ctx context.Context, installationID string) (string, error)

// RefreshInstallationToken forces a refresh of an installation token
func (s *AuthService) RefreshInstallationToken(ctx context.Context, installationID string
func (s *AuthService) ValidateWebhookSignature(payload []byte, signature string, secret string) bool
```

**Implementation Details:**

- JWT generation using RS256 with GitHub App private key
- JWT claims: `iss` (app ID), `iat` (issued at), `exp` (10 min expiration)
- Installation token exchange via GitHub API `POST /app/installations/{installation_id}/access_tokens`
- Token caching with 10-minute buffer before expiration (tokens expire in 1 hour)
- Thread-safe token cache with RWMutex

### 4. GitHub API Service

**Location:** `internal/services/github_service/api.go`

```go
package github_service

import (
    "context"
    
    "github.com/google/go-github/v66/github"
    "golang.org/x/oauth2"
)

type APIService struct {
    authService *AuthService
}

// NewAPIService creates a new GitHub API service
func NewAPIService(authService *AuthService) *APIService

// getClient creates an authenticated GitHub client for an installastring) (*github.Client, error)

// Repository Operations
func (s *APIService) GetRepository(ctx context.Context, installationID string, owner, repo string) (*github.Repository, error)
func (s *APIService) ListRepositories(ctx context.Context, installationID string) ([]*github.Repository, error)

// Branch Operations  
func (s *APIService) GetBranch(ctx context.Context, installationID string, owner, repo, branch string) (*github.Branch, error)
func (s *APIService) ListBranches(ctx context.Context, installationID string, owner, repo string) ([]*github.Branch, error)
func (s *APIService) CreateBranch(ctx context.Context, installationID string, owner, repo, branch, sha string) error

// Pull Request Operations
func (s *APIService) CreatePullRequest(ctx context.Context, installationID string, owner, repo string, pr *github.NewPullRequest) (*github.PullRequest, error)
func (s *APIService) GetPullRequest(ctx context.Context, installationID string, owner, repo string, number int) (*github.PullRequest, error)
func (s *APIService) UpdatePullRequest(ctx context.Context, installationID string, owner, repo string, number int, pr *github.PullRequest) (*github.PullRequest, error)

// Commit Operations
func (s *APIService) GetCommit(ctx context.Context, installationID string, owner, repo, sha string) (*github.Commit, error)
func (s *APIService) CompareCommits(ctx context.Context, installationID stringer, repo, sha string) (*github.Commit, error)
func (s *APIService) CompareCommits(ctx context.Context, installationID int64, owner, repo, base, head string) (*github.CommitsComparison, error)
```

**Rate Limiting:**
- Implement automatic retry with exponential backoff for rate limit errors
- Respect `X-RateLimit-*` headers
- Use conditional requests (`If-None-Match`) where applicable

### 5. Modal GitHub Template

**Location:** `internal/services/sandbox_service/github_template.go`

```go
package sandbox_service

import (
    "github.com/griffnb/techboss-ai-go/internal/integrations/modal"
    "github.com/griffnb/techboss-ai-go/internal/services/sandbox_service/lifecycle"
)string // UUID string

type GitHubTemplateConfig struct {
    InstallationID  int64
    Repository      string // "owner/repo"
    SourceBranch    string
    TargetBranch    string
    PRTargetBranch  string
    PRTitle         string
    PRBody          string
    GitUserName     string // Optional, defaults to app name
    GitUserEmail    string // Optional, defaults to app email
}

// GetGitHubTemplate returns a Modal sandbox template configured for GitHub operations
func GetGitHubTemplate(config *GitHubTemplateConfig) *SandboxTemplate

// GetGitHubImage returns image config with Git + GitHub CLI preinstalled
func GetGitHubImage() *modal.ImageConfig
```

**Image Configuration:**

```go
&modal.ImageConfig{
    BaseImage: "python:3.11-slim",
    DockerfileCommands: []string{
        "RUN apt-get update && apt-get install -y git curl",
        "RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg",
        "RUN echo \"deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main\" | tee /etc/apt/sources.list.d/github-cli.list > /dev/null",
        "RUN apt-get update && apt-get install -y gh",
        "RUN apt-get clean && rm -rf /var/lib/apt/lists/*",
    },
}
```

**Lifecycle Hooks:**

```go
OnColdStart: func(ctx context.Context, sandbox *modal.SandboxInfo) error {
    // 1. Set up Git credentials from installation token
    // 2. Configure git user.name and user.email  
    // 3. Clone repository at source branch
    // 4. Create and checkout target branch
    // 5. Create draft pull request
    // 6. Store PR URL in sandbox metadata
    return nil
}

OnTerminate: func(ctx context.Context, sandbox *modal.SandboxInfo) error {
    // 1. Commit any uncommitted changes
    // 2. Push final changes to target branch
    // 3. Optionally mark PR as ready for review
    return nil
}
```

**Git Operations Script (executed in OnColdStart):**

```bash
#!/bin/bash
set -euo pipefail

# Set up authentication
echo "$GITHUB_TOKEN" | gh auth login --with-token

# Configure Git
git config --global user.name "$GIT_USER_NAME"
git config --global user.email "$GIT_USER_EMAIL"

# Clone repository
gh repo clone "$REPOSITORY" /mnt/workspace
cd /mnt/workspace

# Checkout source branch
git checkout "$SOURCE_BRANCH"
git pull origin "$SOURCE_BRANCH"

# Create new branch
git checkout -b "$TARGET_BRANCH"

# Create draft PR
gh pr create \
  --title "$PR_TITLE" \
  --body "$PR_BODY" \
  --base "$PR_TARGET_BRANCH" \
  --head "$TARGET_BRANCH" \
  --draft

echo "GitHub setup complete"
```

### 6. Sandbox Service Extensions

**Location:** `internal/services/sandbox_service/github.go`

```go
package sandbox_service

import (
    "context"
    
    "github.com/griffnb/core/lib/types"
    "github.com/griffnb/techboss-ai-go/internal/integrations/modal"
    "github.com/griffnb/techboss-ai-go/internal/services/github_service"
)

// CreateGitHubSandbox creates a Modal sandbox with GitHub integration
func (s *SandboxSerstring CreateGitHubSandbox(
    ctx context.Context,
    accountID types.UUID,
    installationID int64,
    config *GitHubTemplateConfig,
) (*modal.SandboxInfo, error) {
    // 1. Get installation token from auth service
    // 2. Build sandbox config from GitHub template
    // 3. Add GitHub token as secret
    // 4. Add Git config as environment variables
    // 5. Call CreateSandbox with modified config
    return sandboxInfo, nil
}
```

**Environment Variables Passed to Sandbox:**

```go
env := map[string]string{
    "GITHUB_TOKEN":       installationToken,
    "REPOSITORY":         config.Repository,
    "SOURCE_BRANCH":      config.SourceBranch,
    "TARGET_BRANCH":      config.TargetBranch,
    "PR_TARGET_BRANCH":   config.PRTargetBranch,
    "PR_TITLE":           config.PRTitle,
    "PR_BODY":            config.PRBody,
    "GIT_USER_NAME":      gitUserName,
    "GIT_USER_EMAIL":     gitUserEmail,
}
```

### 7. Controller Endpoints

**Generation:** Use `#code_tools make_object` to generate controller scaffold for `github_installation` model.

**Location:** `internal/controllers/github_installations/`

The code gen will create the standard CRUD endpoints. We need to add a custom webhook endpoint.

**Additional Endpoint (`webhook.go`):**

```go
package github_installations

import (
    "net/http"
    "github.com/griffnb/core/lib/log"
)

// webhookCallback handles GitHub App installation webhooks
// This is called by GitHub when installation events occur
func webhookCallback(res http.ResponseWriter, req *http.Request) {
    // 1. Validate webhook signature
    // 2. Parse webhook payload
    // 3. Handle installation events
    // 4. Create/update/delete installation records
}
```

**Webhook Route Setup (add to `setup.go`):**

```go
// Add to the Setup function after code gen creates it
r.Post("/callback", webhookCallback)
```

**Webhook Event Handling:**

```go
type WebhookPayload struct {
    Action       string                    `json:"action"`
    Installation *WebhookInstallation      `json:"installation"`
    Repositories []WebhookRepository       `json:"repositories"`
    Sender       *WebhookSender            `json:"sender"`
}

type WebhookInstallation struct {
    ID              string                  `json:"id"`  // GitHub returns string IDs
    Account         *WebhookAccount         `json:"account"`
    RepositorySelection string              `json:"repository_selection"`
    Permissions     map[string]string       `json:"permissions"`
    AppSlug         string                  `json:"app_slug"`
    SuspendedAt     *time.Time              `json:"suspended_at"`
}

// Handle events:
// - installation.created
// - installation.deleted
// - installation.suspend
// - installation.unsuspend
// - installation_repositories.added
// - installation_repositories.removed
```

### 8. Error Handling

**Error Types:**

```go
type GitHubError struct {
    Code    string
    Message string
    Details map[string]any
}

const (
    ErrCodeInvalidCredentials    = "GITHUB_INVALID_CREDENTIALS"
    ErrCodeInstallationNotFound  = "GITHUB_INSTALLATION_NOT_FOUND"
    ErrCodeInstallationSuspended = "GITHUB_INSTALLATION_SUSPENDED"
    ErrCodeRateLimited           = "GITHUB_RATE_LIMITED"
    ErrCodeInsufficientPermissions = "GITHUB_INSUFFICIENT_PERMISSIONS"
    ErrCodeRepositoryNotFound    = "GITHUB_REPOSITORY_NOT_FOUND"
    ErrCodeBranchAlreadyExists   = "GITHUB_BRANCH_EXISTS"
    ErrCodeInvalidWebhookSignature = "GITHUB_INVALID_SIGNATURE"
)
```

**Retry Strategy:**

```go
type RetryConfig struct {
    MaxAttempts    int
    InitialBackoff time.Duration
    MaxBackoff     time.Duration
    Multiplier     float64
}

// Default: 3 attempts, 1s initial, 10s max, 2x multiplier
```

**Error Handling in Controllers:**

```go
func handleGitHubError(err error) (any, int, error) {
    if ghErr, ok := err.(*GitHubError); ok {
        switch ghErr.Code {
        case ErrCodeInstallationNotFound:
            return response.PublicCustomError[any]("Installation not found", http.StatusNotFound)
        case ErrCodeInstallationSuspended:
            return response.PublicCustomError[any]("Installation is suspended", http.StatusForbidden)
        case ErrCodeRateLimited:
            return response.PublicCustomError[any]("Rate limit exceeded, try again later", http.StatusTooManyRequests)
        default:
            return response.PublicBadRequestError[any]()
        }
    }
    return response.PublicBadRequestError[any]()
}
```

## Data Models

### Database Schema

**github_installations table:**

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | UUID | PK | Primary key |
| urn | TEXT | UNIQUE | Resource identifier |
| account_id | INTEGER | INDEX | User's account ID |
| installation_id | INTEGER | UNIQUE, INDEX | GitHub installation ID |
| github_account_id | INTEGER | INDEX | GitHub account/org ID |
| github_account_type | TEXT | | "User" or "Organization" |
| github_account_name | TEXT | | GitHub username/org name |
| repository_access | TEXT | | "all" or "selected" |
| permissions | JSONB | | JSON of granted permissions |
| suspended | SMALLINT | INDEX, DEFAULT 0 | 1 if suspended |
| app_slug | TEXT | | GitHub App slug |
| status | SMALLINT | DEFAULT 0 | Record status |
| created_by_urn | TEXT | | Creator URN |
| updated_by_urn | TEXT | | Last updater URN |
| created_at | TIMESTAMP | DEFAULT NOW() | Creation time |
| updated_at | TIMESTAMP | DEFAULT NOW() | Last update time |
| disabled | SMALLINT | DEFAULT 0 | Soft delete flag |
| deleted | SMALLINT | DEFAULT 0 | Hard delete flag |

**Indexes:**
- `idx_github_installations_account_id` on `account_id`
- `idx_github_installations_installation_id` on `installation_id`
- `idx_github_installations_suspended` on `suspended`

## Testing Strategy

### Unit Tests

**Authentication Service Tests:**
```go
// TestGenerateJWT - Verify JWT structure and claims
// TestGetInstallationToken - Test token retrieval and caching
// TestRefreshInstallationToken - Test force refresh
// TestTokenExpiration - Verify cache expiration logic
// TestValidateWebhookSignature - Test HMAC validation
```

**API Service Tests:**
```go
// Mock GitHub API responses using httptest
// TestGetRepository - Test repository retrieval
// TestCreateBranch - Test branch creation
// TestCreatePullRequest - Test PR creation
// TestRateLimitHandling - Test retry logic
// TestErrorHandling - Test various API errors
```

**Model Tests:**
```go
// TestGithubInstallationCreate - Test installation creation
// TestGithubInstallationQuery - Test query operations
// TestGithubInstallationUpdate - Test updates (suspension, etc.)
```

### Integration Tests

**Sandbox Integration Tests:**
```goUUID | INDEX | User's account ID |
| installation_id | TEXT | UNIQUE, INDEX | GitHub installation ID (string) |
| github_account_id | TEXT | INDEX | GitHub account/org ID (string) |
| github_account_type | SMALLINT | DEFAULT 0 | 0=User, 1=Organization |
| github_account_name | TEXT | | GitHub username/org name |
| repository_access | TEXT | DEFAULT 'all' | "all" or "selected" |
| permissions | JSONB | DEFAULT '{}'
**Controller Tests:**
```go
// TestInstallationWebhook - Test webhook handling
// TestListInstallations - Test list endpoint with auth
// TestGetInstallation - Test get endpoint with ownership check
// TestDeleteInstallation - Test delete endpoint
```

### Security Tests

```go (Generated by code gen based on field tags)posed
// TestJWTExpiration - Verify JWTs expire appropriately
```

### Manual Testing Checklist

- [ ] Install GitHub App on test repository
- [ ] Verify webhook creates installation record
- [ ] List installations in UI
- [ ] Create sandbox with GitHub template
- [ ] Verify repo cloned correctly in sandbox
- [ ] Verify branch created
- [ ] Verify PR created as draft
- [ ] Make changes and commit in sandbox
- [ ] Verify changes pushed to GitHub
- [ ] Suspend installation via GitHub
- [ ] Verify suspended installations can't create tokens
- [ ] Uninstall app
- [ ] Verify deletion webhook processed

## Security Considerations

### Credential Management

1. **Private Key Storage**
   - Store GitHub App private key in environment variable
   - Base64 encode for easier storage
   - Never log or expose in API responses
   - Rotate keys periodically

2. **Token Security**
   - Installation tokens valid for 1 hour
   - Cache tokens but refresh 10 minutes before expiry
   - Pass tokens to sandboxes via secure environment variables
   - Clean up tokens when sandbox terminates

3. **Webhook Validation**
   - Always validate HMAC signature using webhook secret
   - Reject webhooks with invalid signatures
   - Log suspicious webhook attempts
   - Use constant-time comparison for signatures

### Authorization

1. **Installation Ownership**
   - Verify user owns installation before operations
   - Check `account_id` matches authenticated user
   - Prevent cross-account access

2. **Repository Permissions**
   - Check installation has access to repository
   - Verify installation not suspended
   - Check specific permissions (read, write, etc.)

3. **Sandbox Isolation**
   - Each sandbox gets unique installation token
   - Tokens scoped to specific installation
   - Sandbox can only access permitted repositories

### Audit Logging

1. **Track Operations**
   - Log installation creation/deletion
   - Log token generation
   - Log sandbox creation with GitHub config
   - Log failed authentication attempts

2. **Change Logs**
   - Enable change logs on GithubInstallation model
   - Track suspension/unsuspension
   - Track permission changes

## Implementation Notes

### Phase 1: Core Infrastructure
1. Add GitHub config to environment
2. Use `#code_tools make_object GithubInstallation` to generate model scaffold
3. Customize model with proper field types (UUID, IntConstant, etc.)
4. Add init-migration struct
5. Implement Auth Service (JWT, token exchange, caching)
6. Write unit tests for Auth Service

### Phase 2: API Integration
1. Implement GitHub API Service
2. Use `#code_tools make_object` to generate controller scaffold
3. Add custom webhook endpoint to controller
4. Implement webhook handler with signature validation
5. Write controller tests

### Phase 3: Modal Template
1. Create GitHub template with lifecycle hooks
2. Implement Git operations script
3. Extend Sandbox Service with GitHub support
4. Write integration tests

### Dependencies

**Go Packages:**
```
github.com/google/go-github/v66 - GitHub API client
github.com/golang-jwt/jwt/v5 - JWT generation
golang.org/x/oauth2 - OAuth2 token handling
```

**Modal Requirements:**
- Docker image with Git + GitHub CLI
- Environment variable support for credentials
- Lifecycle hooks for cold start and termination

### Configuration Example

```go
// Config values are auto-injected from secrets management
config := environment.GetConfig()
appID := config.Github.AppID
privateKey := config.Github.PrivateKey
```

## Open Questions

1. **Token Rotation**: Should we implement automatic token rotation or rely on GitHub's 1-hour expiration?
   - **Decision**: Rely on GitHub's expiration with 10-minute refresh buffer

2. **Repository Selection**: Should we store selected repositories when access is limited?
   - **Decision**: Store in permissions JSONB, query via GitHub API when needed

3. **PR Management**: Should system automatically convert draft PRs to ready when Claude completes work?
   Store GitHub credentials in AWS Secrets Manager
2. Use `#code_tools make_object GithubInstallation` to generate model scaffold
3. Customize model with proper field types (UUID, IntConstant, etc.)
4. Add init-migration struct
5. Implement Auth Service (JWT, token exchange, caching)
6. Write unit tests for Auth Service

### Phase 2: API Integration
1. Implement GitHub API Service
2. Use `#code_tools make_object` to generate controller scaffold
3. Add custom webhook endpoint to controller
4. Implement webhook handler with signature validation
5. Write controller tests

### Phase 3: Modal Template
1. Create GitHub template with lifecycle hooks
2. Implement Git operations script
3. Extend Sandbox Service with GitHub support
4. Write integration tests

### Dependencies