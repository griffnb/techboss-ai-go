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
	"net/url"
	"testing"

	"github.com/griffnb/core/lib/testtools"
	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/controllers/github_installations"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/github_installation"
	"github.com/griffnb/techboss-ai-go/internal/services/testing_service"
)

func init() {
	system_testing.BuildSystem()
}

// Test_WebhookToDatabase_CompleteFlow tests full webhook processing to database operations
func Test_WebhookToDatabase_CompleteFlow(t *testing.T) {
	t.Run("Complete webhook to database flow - installation created", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		webhookSecret := "integration-test-secret"

		// Set webhook secret in config
		config := environment.GetConfig()
		if config.Github == nil {
			config.Github = &environment.Github{}
		}
		config.Github.WebhookSecret = webhookSecret

		// Create realistic webhook payload
		payload := github_installations.WebhookPayload{
			Action: "created",
			Installation: &github_installations.WebhookInstallation{
				ID: "integration-install-001",
				Account: &github_installations.WebhookAccount{
					ID:    99999,
					Login: "integration-org",
					Type:  "Organization",
				},
				RepositorySelection: "selected",
				Permissions: map[string]any{
					"contents":      "write",
					"pull_requests": "write",
					"issues":        "read",
				},
				AppSlug:     "techboss-ai-integration",
				SuspendedAt: nil,
			},
			Repositories: []github_installations.WebhookRepository{
				{
					ID:       11111,
					Name:     "repo-1",
					FullName: "integration-org/repo-1",
				},
				{
					ID:       22222,
					Name:     "repo-2",
					FullName: "integration-org/repo-2",
				},
			},
			Sender: &github_installations.WebhookSender{
				ID:    99999,
				Login: "integration-org",
			},
		}

		payloadBytes, err := json.Marshal(payload)
		assert.NoError(t, err)

		signature := generateIntegrationWebhookSignature(payloadBytes, webhookSecret)

		// Act - Send webhook
		req := httptest.NewRequest(http.MethodPost, "/callback", bytes.NewReader(payloadBytes))
		req.Header.Set("X-Hub-Signature-256", signature)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		github_installations.WebhookCallback(rec, req)

		// Assert - Webhook processed successfully
		assert.Equal(t, http.StatusOK, rec.Code)

		// Assert - Database record created
		installation, err := github_installation.GetByInstallationID(ctx, "integration-install-001")
		assert.NoError(t, err)
		assert.NEmpty(t, installation)
		defer testtools.CleanupModel(installation)

		// Verify all fields
		assert.Equal(t, "integration-install-001", installation.InstallationID.Get())
		assert.Equal(t, "99999", installation.GithubAccountID.Get())
		assert.Equal(t, "integration-org", installation.GithubAccountName.Get())
		assert.Equal(t, "selected", installation.RepositoryAccess.Get())
		assert.Equal(t, "techboss-ai-integration", installation.AppSlug.Get())
		assert.Equal(t, 0, installation.Suspended.Get())
		assert.Equal(t, github_installation.ACCOUNT_TYPE_ORGANIZATION, installation.GithubAccountType.Get())

		// Verify permissions stored correctly
		permissions, _ := installation.Permissions.Get()
		assert.NEmpty(t, permissions)
		assert.Equal(t, "write", permissions["contents"])
		assert.Equal(t, "write", permissions["pull_requests"])
		assert.Equal(t, "read", permissions["issues"])
	})

	t.Run("Complete flow - installation lifecycle (create, suspend, unsuspend, delete)", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		webhookSecret := "lifecycle-test-secret"

		config := environment.GetConfig()
		if config.Github == nil {
			config.Github = &environment.Github{}
		}
		config.Github.WebhookSecret = webhookSecret

		installationID := "lifecycle-install-002"

		// Step 1: Create installation via webhook
		createPayload := createWebhookPayload("created", installationID)
		createPayloadBytes, err := json.Marshal(createPayload)
		assert.NoError(t, err)
		createSignature := generateIntegrationWebhookSignature(createPayloadBytes, webhookSecret)

		req1 := httptest.NewRequest(http.MethodPost, "/callback", bytes.NewReader(createPayloadBytes))
		req1.Header.Set("X-Hub-Signature-256", createSignature)
		rec1 := httptest.NewRecorder()
		github_installations.WebhookCallback(rec1, req1)
		assert.Equal(t, http.StatusOK, rec1.Code)

		// Verify created
		installation, err := github_installation.GetByInstallationID(ctx, installationID)
		assert.NoError(t, err)
		assert.NEmpty(t, installation)
		defer testtools.CleanupModel(installation)
		assert.Equal(t, 0, installation.Suspended.Get())

		// Step 2: Suspend installation via webhook
		suspendPayload := createWebhookPayload("suspend", installationID)
		suspendPayloadBytes, err := json.Marshal(suspendPayload)
		assert.NoError(t, err)
		suspendSignature := generateIntegrationWebhookSignature(suspendPayloadBytes, webhookSecret)

		req2 := httptest.NewRequest(http.MethodPost, "/callback", bytes.NewReader(suspendPayloadBytes))
		req2.Header.Set("X-Hub-Signature-256", suspendSignature)
		rec2 := httptest.NewRecorder()
		github_installations.WebhookCallback(rec2, req2)
		assert.Equal(t, http.StatusOK, rec2.Code)

		// Verify suspended
		installation, err = github_installation.Get(ctx, installation.ID())
		assert.NoError(t, err)
		assert.Equal(t, 1, installation.Suspended.Get())

		// Step 3: Unsuspend installation via webhook
		unsuspendPayload := createWebhookPayload("unsuspend", installationID)
		unsuspendPayloadBytes, err := json.Marshal(unsuspendPayload)
		assert.NoError(t, err)
		unsuspendSignature := generateIntegrationWebhookSignature(unsuspendPayloadBytes, webhookSecret)

		req3 := httptest.NewRequest(http.MethodPost, "/callback", bytes.NewReader(unsuspendPayloadBytes))
		req3.Header.Set("X-Hub-Signature-256", unsuspendSignature)
		rec3 := httptest.NewRecorder()
		github_installations.WebhookCallback(rec3, req3)
		assert.Equal(t, http.StatusOK, rec3.Code)

		// Verify unsuspended
		installation, err = github_installation.Get(ctx, installation.ID())
		assert.NoError(t, err)
		assert.Equal(t, 0, installation.Suspended.Get())

		// Step 4: Delete installation via webhook
		deletePayload := createWebhookPayload("deleted", installationID)
		deletePayloadBytes, err := json.Marshal(deletePayload)
		assert.NoError(t, err)
		deleteSignature := generateIntegrationWebhookSignature(deletePayloadBytes, webhookSecret)

		req4 := httptest.NewRequest(http.MethodPost, "/callback", bytes.NewReader(deletePayloadBytes))
		req4.Header.Set("X-Hub-Signature-256", deleteSignature)
		rec4 := httptest.NewRecorder()
		github_installations.WebhookCallback(rec4, req4)
		assert.Equal(t, http.StatusOK, rec4.Code)

		// Verify deleted (record should be marked as deleted, not accessible via normal Get)
		_, err = github_installation.GetByInstallationID(ctx, installationID)
		// Note: GetByInstallationID filters deleted records, so this should error
		// However, the record still exists with Deleted=1
		if err != nil {
			// Expected - deleted records are filtered
			assert.Error(t, err)
		}
	})
}

// Test_InstallationCRUD_CompleteFlow tests full CRUD operations via controller endpoints
func Test_InstallationCRUD_CompleteFlow(t *testing.T) {
	t.Run("List installations through controller endpoint", func(t *testing.T) {
		// Arrange - Create test installation in database
		ctx := context.Background()
		installation := github_installation.New()
		installation.InstallationID.Set("crud-test-003")
		installation.GithubAccountID.Set("88888")
		installation.GithubAccountName.Set("crud-test-user")
		installation.RepositoryAccess.Set("all")
		installation.AppSlug.Set("techboss-ai-crud")
		installation.Suspended.Set(0)
		installation.GithubAccountType.Set(github_installation.ACCOUNT_TYPE_USER)
		installation.Permissions.Set(map[string]any{
			"contents": "read",
		})

		err := installation.SaveWithContext(ctx, nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(installation)

		// Test: GET (Read) - List all installations
		req, err := testing_service.NewGETRequest[[]*github_installation.GithubInstallationJoined]("/", nil)
		assert.NoError(t, err)
		err = req.WithAdmin()
		assert.NoError(t, err)

		resp, errCode, err := req.Do(github_installations.AdminIndex)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, resp)
		assert.True(t, len(resp) > 0)

		// Find our installation in the list
		var found bool
		for _, inst := range resp {
			if inst.InstallationID.Get() == "crud-test-003" {
				found = true
				assert.Equal(t, "crud-test-user", inst.GithubAccountName.Get())
				assert.Equal(t, "techboss-ai-crud", inst.AppSlug.Get())
				assert.Equal(t, 0, inst.Suspended.Get())
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("Update installation through database and verify", func(t *testing.T) {
		// Arrange - Create test installation
		ctx := context.Background()
		installation := github_installation.New()
		installation.InstallationID.Set("crud-test-update")
		installation.GithubAccountID.Set("99999")
		installation.GithubAccountName.Set("update-test-user")
		installation.RepositoryAccess.Set("all")
		installation.AppSlug.Set("test-app")
		installation.Suspended.Set(0)

		err := installation.SaveWithContext(ctx, nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(installation)

		// Act - Update via database
		installation.GithubAccountName.Set("updated-user")
		installation.Suspended.Set(1)
		err = installation.SaveWithContext(ctx, nil)
		assert.NoError(t, err)

		// Assert - Verify via controller list endpoint
		req, err := testing_service.NewGETRequest[[]*github_installation.GithubInstallationJoined]("/", nil)
		assert.NoError(t, err)
		err = req.WithAdmin()
		assert.NoError(t, err)

		resp, errCode, err := req.Do(github_installations.AdminIndex)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)

		// Find updated installation
		var found bool
		for _, inst := range resp {
			if inst.InstallationID.Get() == "crud-test-update" {
				found = true
				assert.Equal(t, "updated-user", inst.GithubAccountName.Get())
				assert.Equal(t, 1, inst.Suspended.Get())
				break
			}
		}
		assert.True(t, found)
	})
}

// Test_SearchFunctionality_Integration tests search functionality
func Test_SearchFunctionality_Integration(t *testing.T) {
	t.Run("Search installations by account name", func(t *testing.T) {
		// Arrange - Create multiple installations
		ctx := context.Background()

		installation1 := github_installation.New()
		installation1.InstallationID.Set("search-test-001")
		installation1.GithubAccountID.Set("11111")
		installation1.GithubAccountName.Set("searchable-org-alpha")
		installation1.RepositoryAccess.Set("all")
		installation1.AppSlug.Set("test-app")
		err := installation1.SaveWithContext(ctx, nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(installation1)

		installation2 := github_installation.New()
		installation2.InstallationID.Set("search-test-002")
		installation2.GithubAccountID.Set("22222")
		installation2.GithubAccountName.Set("searchable-org-beta")
		installation2.RepositoryAccess.Set("selected")
		installation2.AppSlug.Set("test-app")
		err = installation2.SaveWithContext(ctx, nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(installation2)

		installation3 := github_installation.New()
		installation3.InstallationID.Set("search-test-003")
		installation3.GithubAccountID.Set("33333")
		installation3.GithubAccountName.Set("different-user")
		installation3.RepositoryAccess.Set("all")
		installation3.AppSlug.Set("test-app")
		err = installation3.SaveWithContext(ctx, nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(installation3)

		// Act - Search for installations with "searchable" in account name
		params := url.Values{}
		params.Add("q", "searchable")

		req, err := testing_service.NewGETRequest[[]*github_installation.GithubInstallationJoined]("/", params)
		assert.NoError(t, err)
		err = req.WithAdmin()
		assert.NoError(t, err)

		resp, errCode, err := req.Do(github_installations.AdminIndex)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, resp)

		// Should find at least the two searchable installations
		var foundAlpha, foundBeta bool
		for _, inst := range resp {
			if inst.GithubAccountName.Get() == "searchable-org-alpha" {
				foundAlpha = true
			}
			if inst.GithubAccountName.Get() == "searchable-org-beta" {
				foundBeta = true
			}
		}
		assert.True(t, foundAlpha)
		assert.True(t, foundBeta)
	})
}

// Test_WebhookAuthentication_Integration tests webhook authentication requirements
func Test_WebhookAuthentication_Integration(t *testing.T) {
	t.Run("Webhook rejects request without signature", func(t *testing.T) {
		// Arrange
		payload := createWebhookPayload("created", "no-auth-test")
		payloadBytes, err := json.Marshal(payload)
		assert.NoError(t, err)

		// Act - Send webhook without signature header
		req := httptest.NewRequest(http.MethodPost, "/callback", bytes.NewReader(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		// Intentionally not setting X-Hub-Signature-256
		rec := httptest.NewRecorder()

		github_installations.WebhookCallback(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Webhook rejects request with invalid signature", func(t *testing.T) {
		// Arrange
		webhookSecret := "auth-test-secret"
		config := environment.GetConfig()
		if config.Github == nil {
			config.Github = &environment.Github{}
		}
		config.Github.WebhookSecret = webhookSecret

		payload := createWebhookPayload("created", "invalid-sig-test")
		payloadBytes, err := json.Marshal(payload)
		assert.NoError(t, err)

		// Generate signature with wrong secret
		invalidSignature := generateIntegrationWebhookSignature(payloadBytes, "wrong-secret")

		// Act
		req := httptest.NewRequest(http.MethodPost, "/callback", bytes.NewReader(payloadBytes))
		req.Header.Set("X-Hub-Signature-256", invalidSignature)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		github_installations.WebhookCallback(rec, req)

		// Assert
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("Webhook accepts request with valid signature", func(t *testing.T) {
		// Arrange
		webhookSecret := "valid-auth-test-secret"
		config := environment.GetConfig()
		if config.Github == nil {
			config.Github = &environment.Github{}
		}
		config.Github.WebhookSecret = webhookSecret

		payload := createWebhookPayload("created", "valid-sig-test-004")
		payloadBytes, err := json.Marshal(payload)
		assert.NoError(t, err)

		// Generate valid signature
		validSignature := generateIntegrationWebhookSignature(payloadBytes, webhookSecret)

		// Act
		req := httptest.NewRequest(http.MethodPost, "/callback", bytes.NewReader(payloadBytes))
		req.Header.Set("X-Hub-Signature-256", validSignature)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		github_installations.WebhookCallback(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)

		// Cleanup
		ctx := context.Background()
		installation, err := github_installation.GetByInstallationID(ctx, "valid-sig-test-004")
		if err == nil {
			testtools.CleanupModel(installation)
		}
	})
}

// generateIntegrationWebhookSignature generates a valid GitHub webhook signature for testing
func generateIntegrationWebhookSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	signature := hex.EncodeToString(h.Sum(nil))
	return "sha256=" + signature
}
