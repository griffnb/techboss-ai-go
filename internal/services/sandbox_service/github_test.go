package sandbox_service_test

import (
	"context"
	"testing"
	"time"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/services/github_service"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service"
)

func init() {
	system_testing.BuildSystem()
}

func setupMockGitHubAuth(t *testing.T) *github_service.AuthService { //nolint:unused
	// Use the mock token response functionality
	// #nosec G101 -- This is a test key, not a real credential
	privateKey := `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA0Z3VS5JJcds3xfn/ygWyF8qvJjHLqgvT3xHs8M3L3E0FhYGl
L6wY/XCXdMb5LrKMUMDaFDsqPJWEjQpJ5d+QM6pLNvOdBkYdFwC3Zr7L9bCmNBhs
xq6Y5VJ+GNx4N3K4P5xL3F2RRQZ9kLzDhW4Y1A0j3Q3xL7N9sKE3fFYAGUK5vP3Y
fJQ7B3L4Y1F6WQH8K3Y7F9L4N1K3Q9Z5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3
F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9
J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1QwIDAQABAoIBAFJxT1B3Y4L7
K3Q9Z5B3F4Y1K3Q9J5B3L7F4K3Y1Q9J5B3L7F4K3Y1Q9J5B3L7F4K3Y1Q9J5B3L7
F4K3Y1Q9J5B3L7F4K3Y1Q9J5B3L7F4K3Y1Q9J5B3L7F4K3Y1Q9J5B3L7F4K3Y1Q9
J5B3L7F4K3Y1Q9J5B3L7F4K3Y1Q9J5B3L7F4K3Y1Q9J5B3L7F4K3Y1Q9J5B3L7F4
K3Y1Q9J5B3L7F4K3Y1Q9J5B3L7F4K3Y1Q9J5B3L7F4K3Y1Q9J5B3L7F4K3Y1Q9J5
B3L7F4K3Y1Q9J5B3L7F4K3Y1Q9J5B3L7F4K3Y1Q9J5B3L7F4K3Y1Q9J5B3L7F4K3
Y1Q9J5ECgYEA8F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5
L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3
Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7
B3F4K3Y1QwKBgQDR5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4
K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5
L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3
Y1Q9J5QQKBgQCF4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5
L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3
Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7
B3F4K3YwKBgQC3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9
J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4
K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5
L7B3QQKBgQDY1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4
K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5
L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3Y1Q9J5L7B3F4K3
Y1Q9J5L7B3
-----END RSA PRIVATE KEY-----`

	authService, err := github_service.NewAuthService("test-app-id", privateKey)
	assert.NoError(t, err)

	// Set mock token response to avoid real GitHub API calls
	authService.SetMockTokenResponse("mock-token-12345", time.Now().Add(1*time.Hour))

	return authService
}

func skipIfGitHubNotConfigured(t *testing.T) {
	config := environment.GetConfig()
	if config == nil || config.Github == nil || config.Github.AppID == "" {
		t.Skip("GitHub is not configured, skipping test")
	}
}

func skipIfModalNotConfigured(t *testing.T) {
	if !modal.Configured() {
		t.Skip("Modal client is not configured, skipping test")
	}
}

func Test_CreateGitHubSandbox(t *testing.T) {
	t.Run("validates required fields", func(t *testing.T) {
		// Arrange
		service := sandbox_service.NewSandboxService()
		ctx := context.Background()
		accountID := types.UUID("test-account-123")

		t.Run("returns error when config is nil", func(t *testing.T) {
			// Act
			sandboxInfo, err := service.CreateGitHubSandbox(ctx, accountID, "12345", nil)

			// Assert
			assert.Error(t, err)
			assert.Empty(t, sandboxInfo)
		})

		t.Run("returns error when installationID is empty", func(t *testing.T) {
			// Arrange
			config := &sandbox_service.GitHubTemplateConfig{
				Repository: "owner/repo",
			}

			// Act
			sandboxInfo, err := service.CreateGitHubSandbox(ctx, accountID, "", config)

			// Assert
			assert.Error(t, err)
			assert.Empty(t, sandboxInfo)
		})

		t.Run("returns error when repository is empty", func(t *testing.T) {
			// Arrange
			config := &sandbox_service.GitHubTemplateConfig{
				Repository: "",
			}

			// Act
			sandboxInfo, err := service.CreateGitHubSandbox(ctx, accountID, "12345", config)

			// Assert
			assert.Error(t, err)
			assert.Empty(t, sandboxInfo)
		})
	})

	t.Run("applies default values for optional fields", func(t *testing.T) {
		skipIfGitHubNotConfigured(t)
		skipIfModalNotConfigured(t)

		// This test would need to verify the defaults are applied
		// Since we can't easily inspect the internal config without creating the sandbox,
		// we skip this for now or mock the Modal client
		t.Skip("Requires mocked Modal client to test without creating real sandbox")
	})

	t.Run("sets environment variables correctly", func(t *testing.T) {
		skipIfGitHubNotConfigured(t)
		skipIfModalNotConfigured(t)

		// This test would verify environment variables are set
		// Would need mocked Modal client to inspect the config
		t.Skip("Requires mocked Modal client to test without creating real sandbox")
	})

	t.Run("returns error when GitHub is not configured", func(t *testing.T) {
		// This test verifies that CreateGitHubSandbox checks for GitHub config
		// Skip if GitHub IS configured, since we want to test the error case
		if environment.GetConfig() != nil && environment.GetConfig().Github != nil {
			t.Skip("GitHub is configured - cannot test unconfigured case")
		}

		// Arrange
		service := sandbox_service.NewSandboxService()
		ctx := context.Background()
		accountID := types.UUID("test-account-123")
		config := &sandbox_service.GitHubTemplateConfig{
			InstallationID: "12345",
			Repository:     "owner/repo",
		}

		// Act
		sandboxInfo, err := service.CreateGitHubSandbox(ctx, accountID, "12345", config)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, sandboxInfo)
	})
}

func Test_CreateGitHubSandbox_Integration(t *testing.T) {
	skipIfGitHubNotConfigured(t)
	skipIfModalNotConfigured(t)

	t.Run("creates sandbox with GitHub configuration", func(t *testing.T) {
		// Skip in CI/CD as this would create a real sandbox
		t.Skip("Integration test - requires real GitHub App and Modal configuration")

		// This test would:
		// 1. Create a real GitHub sandbox
		// 2. Verify it has the correct environment variables
		// 3. Verify the lifecycle hooks are set
		// 4. Clean up by terminating the sandbox
	})
}

func Test_GitHubTemplateConfig_Structure(t *testing.T) {
	t.Run("config struct has all required fields", func(t *testing.T) {
		// Arrange & Act
		config := &sandbox_service.GitHubTemplateConfig{
			InstallationID: "12345",
			Repository:     "owner/repo",
			SourceBranch:   "main",
			TargetBranch:   "feature/test",
			PRTargetBranch: "main",
			PRTitle:        "Test PR",
			PRBody:         "Test body",
			GitUserName:    "Test User",
			GitUserEmail:   "test@example.com",
		}

		// Assert
		assert.Equal(t, "12345", config.InstallationID)
		assert.Equal(t, "owner/repo", config.Repository)
		assert.Equal(t, "main", config.SourceBranch)
		assert.Equal(t, "feature/test", config.TargetBranch)
		assert.Equal(t, "main", config.PRTargetBranch)
		assert.Equal(t, "Test PR", config.PRTitle)
		assert.Equal(t, "Test body", config.PRBody)
		assert.Equal(t, "Test User", config.GitUserName)
		assert.Equal(t, "test@example.com", config.GitUserEmail)
	})
}
