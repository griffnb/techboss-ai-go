package github_installations_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/testtools"
	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/controllers/github_installations"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/github_installation"
)

func init() {
	system_testing.BuildSystem()
}

// generateWebhookSignature generates a valid GitHub webhook signature
func generateWebhookSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	signature := hex.EncodeToString(h.Sum(nil))
	return "sha256=" + signature
}

// createWebhookPayload creates a test webhook payload
func createWebhookPayload(action string, installationID string) github_installations.WebhookPayload {
	return github_installations.WebhookPayload{
		Action: action,
		Installation: &github_installations.WebhookInstallation{
			ID: installationID,
			Account: &github_installations.WebhookAccount{
				ID:    12345,
				Login: "test-user",
				Type:  "User",
			},
			RepositorySelection: "all",
			Permissions: map[string]any{
				"contents":      "read",
				"pull_requests": "write",
			},
			AppSlug:     "test-app",
			SuspendedAt: nil,
		},
		Repositories: []github_installations.WebhookRepository{
			{
				ID:       67890,
				Name:     "test-repo",
				FullName: "test-user/test-repo",
			},
		},
		Sender: &github_installations.WebhookSender{
			ID:    12345,
			Login: "test-user",
		},
	}
}

func Test_webhookCallback_installationCreated(t *testing.T) {
	t.Run("Creates installation record successfully", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		webhookSecret := "test-webhook-secret"

		// Set webhook secret in config
		config := environment.GetConfig()
		if config.Github == nil {
			config.Github = &environment.Github{}
		}
		config.Github.WebhookSecret = webhookSecret

		payload := createWebhookPayload("created", "12345678")
		payloadBytes, err := json.Marshal(payload)
		assert.NoError(t, err)

		signature := generateWebhookSignature(payloadBytes, webhookSecret)

		req := httptest.NewRequest(http.MethodPost, "/callback", bytes.NewReader(payloadBytes))
		req.Header.Set("X-Hub-Signature-256", signature)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		// Act
		github_installations.WebhookCallback(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)

		// Verify installation was created
		installation, err := github_installation.GetByInstallationID(ctx, "12345678")
		assert.NoError(t, err)
		assert.NEmpty(t, installation)
		defer testtools.CleanupModel(installation)

		assert.Equal(t, "12345678", installation.InstallationID.Get())
		assert.Equal(t, "12345", installation.GithubAccountID.Get())
		assert.Equal(t, "test-user", installation.GithubAccountName.Get())
		assert.Equal(t, "all", installation.RepositoryAccess.Get())
		assert.Equal(t, 0, installation.Suspended.Get())
	})
}

func Test_webhookCallback_installationDeleted(t *testing.T) {
	t.Run("Marks installation as deleted", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		webhookSecret := "test-webhook-secret"

		// Set webhook secret in config
		config := environment.GetConfig()
		if config.Github == nil {
			config.Github = &environment.Github{}
		}
		config.Github.WebhookSecret = webhookSecret

		// Create an existing installation
		existingInstallation := github_installation.New()
		existingInstallation.InstallationID.Set("87654321")
		existingInstallation.GithubAccountID.Set("54321")
		existingInstallation.GithubAccountName.Set("existing-user")
		existingInstallation.RepositoryAccess.Set("all")
		existingInstallation.AppSlug.Set("test-app")
		err := existingInstallation.SaveWithContext(ctx, nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(existingInstallation)

		payload := createWebhookPayload("deleted", "87654321")
		payloadBytes, err := json.Marshal(payload)
		assert.NoError(t, err)

		signature := generateWebhookSignature(payloadBytes, webhookSecret)

		req := httptest.NewRequest(http.MethodPost, "/callback", bytes.NewReader(payloadBytes))
		req.Header.Set("X-Hub-Signature-256", signature)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		// Act
		github_installations.WebhookCallback(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)

		// Verify installation was marked as deleted
		// Note: Use FindFirst instead of Get because Get filters out deleted records
		installation, err := github_installation.FindFirst(ctx, model.NewOptions().
			WithCondition("id = :id:").
			WithParam(":id:", existingInstallation.ID()))
		assert.NoError(t, err)
		assert.NEmpty(t, installation)
		assert.Equal(t, installation.Deleted.Get(), 1)
	})
}

func Test_webhookCallback_installationSuspend(t *testing.T) {
	t.Run("Sets suspended flag", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		webhookSecret := "test-webhook-secret"

		// Set webhook secret in config
		config := environment.GetConfig()
		if config.Github == nil {
			config.Github = &environment.Github{}
		}
		config.Github.WebhookSecret = webhookSecret

		// Create an existing installation
		existingInstallation := github_installation.New()
		existingInstallation.InstallationID.Set("11111111")
		existingInstallation.GithubAccountID.Set("11111")
		existingInstallation.GithubAccountName.Set("suspend-user")
		existingInstallation.RepositoryAccess.Set("all")
		existingInstallation.AppSlug.Set("test-app")
		existingInstallation.Suspended.Set(0)
		err := existingInstallation.SaveWithContext(ctx, nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(existingInstallation)

		payload := createWebhookPayload("suspend", "11111111")
		payloadBytes, err := json.Marshal(payload)
		assert.NoError(t, err)

		signature := generateWebhookSignature(payloadBytes, webhookSecret)

		req := httptest.NewRequest(http.MethodPost, "/callback", bytes.NewReader(payloadBytes))
		req.Header.Set("X-Hub-Signature-256", signature)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		// Act
		github_installations.WebhookCallback(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)

		// Verify installation was suspended
		installation, err := github_installation.Get(ctx, existingInstallation.ID())
		assert.NoError(t, err)
		assert.NEmpty(t, installation)
		assert.Equal(t, 1, installation.Suspended.Get())
	})
}

func Test_webhookCallback_installationUnsuspend(t *testing.T) {
	t.Run("Clears suspended flag", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		webhookSecret := "test-webhook-secret"

		// Set webhook secret in config
		config := environment.GetConfig()
		if config.Github == nil {
			config.Github = &environment.Github{}
		}
		config.Github.WebhookSecret = webhookSecret

		// Create an existing suspended installation
		existingInstallation := github_installation.New()
		existingInstallation.InstallationID.Set("22222222")
		existingInstallation.GithubAccountID.Set("22222")
		existingInstallation.GithubAccountName.Set("unsuspend-user")
		existingInstallation.RepositoryAccess.Set("all")
		existingInstallation.AppSlug.Set("test-app")
		existingInstallation.Suspended.Set(1)
		err := existingInstallation.SaveWithContext(ctx, nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(existingInstallation)

		payload := createWebhookPayload("unsuspend", "22222222")
		payloadBytes, err := json.Marshal(payload)
		assert.NoError(t, err)

		signature := generateWebhookSignature(payloadBytes, webhookSecret)

		req := httptest.NewRequest(http.MethodPost, "/callback", bytes.NewReader(payloadBytes))
		req.Header.Set("X-Hub-Signature-256", signature)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		// Act
		github_installations.WebhookCallback(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)

		// Verify installation was unsuspended
		installation, err := github_installation.Get(ctx, existingInstallation.ID())
		assert.NoError(t, err)
		assert.NEmpty(t, installation)
		assert.Equal(t, 0, installation.Suspended.Get())
	})
}

func Test_webhookCallback_invalidSignature(t *testing.T) {
	t.Run("Returns 401 for invalid signature", func(t *testing.T) {
		// Arrange
		webhookSecret := "test-webhook-secret"

		// Set webhook secret in config
		config := environment.GetConfig()
		if config.Github == nil {
			config.Github = &environment.Github{}
		}
		config.Github.WebhookSecret = webhookSecret

		payload := createWebhookPayload("created", "99999999")
		payloadBytes, err := json.Marshal(payload)
		assert.NoError(t, err)

		// Use wrong secret to generate invalid signature
		invalidSignature := generateWebhookSignature(payloadBytes, "wrong-secret")

		req := httptest.NewRequest(http.MethodPost, "/callback", bytes.NewReader(payloadBytes))
		req.Header.Set("X-Hub-Signature-256", invalidSignature)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		// Act
		github_installations.WebhookCallback(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func Test_webhookCallback_missingSignature(t *testing.T) {
	t.Run("Returns 401 for missing signature header", func(t *testing.T) {
		// Arrange
		payload := createWebhookPayload("created", "99999999")
		payloadBytes, err := json.Marshal(payload)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/callback", bytes.NewReader(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		// Intentionally not setting X-Hub-Signature-256 header
		rec := httptest.NewRecorder()

		// Act
		github_installations.WebhookCallback(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func Test_webhookCallback_invalidJSON(t *testing.T) {
	t.Run("Returns 400 for invalid JSON payload", func(t *testing.T) {
		// Arrange
		webhookSecret := "test-webhook-secret"

		// Set webhook secret in config
		config := environment.GetConfig()
		if config.Github == nil {
			config.Github = &environment.Github{}
		}
		config.Github.WebhookSecret = webhookSecret

		invalidPayload := []byte("not valid json")
		signature := generateWebhookSignature(invalidPayload, webhookSecret)

		req := httptest.NewRequest(http.MethodPost, "/callback", bytes.NewReader(invalidPayload))
		req.Header.Set("X-Hub-Signature-256", signature)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		// Act
		github_installations.WebhookCallback(rec, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
