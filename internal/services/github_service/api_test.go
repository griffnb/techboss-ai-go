package github_service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-github/v66/github"
	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/services/github_service"
)

func init() {
	system_testing.BuildSystem()
}

func Test_NewAPIService_success(t *testing.T) {
	t.Run("creates API service with auth service", func(t *testing.T) {
		// Arrange
		authService := &github_service.AuthService{}

		// Act
		apiService := github_service.NewAPIService(authService)

		// Assert
		assert.NEmpty(t, apiService)
	})
}

func Test_getClient_success(t *testing.T) {
	t.Run("creates authenticated GitHub client", func(t *testing.T) {
		// Arrange
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService("test-app-id", privateKeyPEM)
		assert.NoError(t, err)
		authService.SetMockTokenResponse("test-token", time.Now().Add(1*time.Hour))

		apiService := github_service.NewAPIService(authService)
		ctx := context.Background()
		installationID := "12345"

		// Act
		client, err := apiService.GetClient(ctx, installationID)

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, client)
	})

	t.Run("returns error when token fetch fails", func(t *testing.T) {
		// Arrange
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService("test-app-id", privateKeyPEM)
		assert.NoError(t, err)
		// No mock token response set - will fail

		apiService := github_service.NewAPIService(authService)
		ctx := context.Background()
		installationID := "12345"

		// Act
		client, err := apiService.GetClient(ctx, installationID)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, client)
	})

	t.Run("returns error for empty installation ID", func(t *testing.T) {
		// Arrange
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService("test-app-id", privateKeyPEM)
		assert.NoError(t, err)

		apiService := github_service.NewAPIService(authService)
		ctx := context.Background()

		// Act
		client, err := apiService.GetClient(ctx, "")

		// Assert
		assert.Error(t, err)
		assert.Empty(t, client)
	})
}

func Test_GetRepository_success(t *testing.T) {
	t.Run("returns error for empty installation ID", func(t *testing.T) {
		// Arrange
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService("test-app-id", privateKeyPEM)
		assert.NoError(t, err)

		apiService := github_service.NewAPIService(authService)
		ctx := context.Background()

		// Act
		repo, err := apiService.GetRepository(ctx, "", "owner", "repo")

		// Assert
		assert.Error(t, err)
		assert.Empty(t, repo)
	})
}

func Test_ListRepositories_success(t *testing.T) {
	t.Run("returns error for empty installation ID", func(t *testing.T) {
		// Arrange
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService("test-app-id", privateKeyPEM)
		assert.NoError(t, err)

		apiService := github_service.NewAPIService(authService)
		ctx := context.Background()

		// Act
		repos, err := apiService.ListRepositories(ctx, "")

		// Assert
		assert.Error(t, err)
		assert.Empty(t, repos)
	})
}

func Test_GetBranch_success(t *testing.T) {
	t.Run("returns error for empty installation ID", func(t *testing.T) {
		// Arrange
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService("test-app-id", privateKeyPEM)
		assert.NoError(t, err)

		apiService := github_service.NewAPIService(authService)
		ctx := context.Background()

		// Act
		branch, err := apiService.GetBranch(ctx, "", "owner", "repo", "main")

		// Assert
		assert.Error(t, err)
		assert.Empty(t, branch)
	})
}

func Test_CreateBranch_success(t *testing.T) {
	t.Run("returns error for empty installation ID", func(t *testing.T) {
		// Arrange
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService("test-app-id", privateKeyPEM)
		assert.NoError(t, err)

		apiService := github_service.NewAPIService(authService)
		ctx := context.Background()

		// Act
		err = apiService.CreateBranch(ctx, "", "owner", "repo", "new-branch", "abc123")

		// Assert
		assert.Error(t, err)
	})
}

func Test_CreatePullRequest_success(t *testing.T) {
	t.Run("returns error for empty installation ID", func(t *testing.T) {
		// Arrange
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService("test-app-id", privateKeyPEM)
		assert.NoError(t, err)

		apiService := github_service.NewAPIService(authService)
		ctx := context.Background()

		title := "Test PR"
		head := "feature-branch"
		base := "main"

		pr := &github.NewPullRequest{
			Title: &title,
			Head:  &head,
			Base:  &base,
		}

		// Act
		result, err := apiService.CreatePullRequest(ctx, "", "owner", "repo", pr)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, result)
	})
}
