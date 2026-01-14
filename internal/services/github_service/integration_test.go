package github_service_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/services/github_service"
)

func init() {
	system_testing.BuildSystem()
}

// Test_FullAuthFlow_JWTGenerationToTokenExchange tests the complete authentication flow
func Test_FullAuthFlow_JWTGenerationToTokenExchange(t *testing.T) {
	t.Run("Complete JWT generation and token exchange flow", func(t *testing.T) {
		// Arrange - Set up auth service with test credentials
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)

		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)
		assert.NEmpty(t, authService)

		ctx := context.Background()
		installationID := "test-installation-full-flow"

		// Mock GitHub API response
		mockToken := "ghs_integration_test_token"
		mockExpiry := time.Now().Add(60 * time.Minute)
		authService.SetMockTokenResponse(mockToken, mockExpiry)

		// Act - Step 1: Generate JWT
		jwt, err := authService.GenerateJWT()
		assert.NoError(t, err)
		assert.NotEmpty(t, jwt)

		// Act - Step 2: Exchange JWT for installation token (first call)
		token1, err := authService.GetInstallationToken(ctx, installationID)
		assert.NoError(t, err)
		assert.Equal(t, mockToken, token1)

		// Act - Step 3: Verify token is cached (second call should return cached token)
		token2, err := authService.GetInstallationToken(ctx, installationID)
		assert.NoError(t, err)
		assert.Equal(t, token1, token2)

		// Act - Step 4: Verify cache works
		cachedToken, found := authService.GetCachedToken(installationID)
		assert.True(t, found)
		assert.Equal(t, mockToken, cachedToken)
	})

	t.Run("Token refresh when cache expires", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)

		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		ctx := context.Background()
		installationID := "test-installation-expiry"

		// Set up initial expired token in cache
		expiredToken := "ghs_expired"
		expiresAt := time.Now().Add(-5 * time.Minute)
		authService.SetCachedToken(installationID, expiredToken, expiresAt)

		// Mock new token response
		newToken := "ghs_refreshed"
		newExpiry := time.Now().Add(60 * time.Minute)
		authService.SetMockTokenResponse(newToken, newExpiry)

		// Act - GetInstallationToken should detect expiry and refresh
		token, err := authService.GetInstallationToken(ctx, installationID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, newToken, token)
		assert.NotEqual(t, expiredToken, token)

		// Verify new token is cached
		cachedToken, found := authService.GetCachedToken(installationID)
		assert.True(t, found)
		assert.Equal(t, newToken, cachedToken)
	})

	t.Run("Forced token refresh bypasses cache", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)

		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		ctx := context.Background()
		installationID := "test-installation-force-refresh"

		// Set up valid token in cache
		oldToken := "ghs_old_but_valid"
		expiresAt := time.Now().Add(30 * time.Minute)
		authService.SetCachedToken(installationID, oldToken, expiresAt)

		// Mock new token response
		newToken := "ghs_force_refreshed"
		newExpiry := time.Now().Add(60 * time.Minute)
		authService.SetMockTokenResponse(newToken, newExpiry)

		// Act - RefreshInstallationToken should bypass cache
		token, err := authService.RefreshInstallationToken(ctx, installationID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, newToken, token)
		assert.NotEqual(t, oldToken, token)
	})
}

// Test_FullAPIServiceFlow_AuthToClientCreation tests API service with auth integration
func Test_FullAPIServiceFlow_AuthToClientCreation(t *testing.T) {
	t.Run("Complete flow from auth service to GitHub client creation", func(t *testing.T) {
		// Arrange - Set up auth service
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)

		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		// Mock token response
		mockToken := "ghs_api_service_token"
		mockExpiry := time.Now().Add(60 * time.Minute)
		authService.SetMockTokenResponse(mockToken, mockExpiry)

		// Create API service with auth service
		apiService := github_service.NewAPIService(authService)
		assert.NEmpty(t, apiService)

		ctx := context.Background()
		installationID := "test-installation-api"

		// Act - Create GitHub client (this tests the full flow)
		client, err := apiService.GetClient(ctx, installationID)

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, client)

		// Verify token was cached by auth service
		cachedToken, found := authService.GetCachedToken(installationID)
		assert.True(t, found)
		assert.Equal(t, mockToken, cachedToken)
	})

	t.Run("Multiple API operations reuse cached token", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)

		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		mockToken := "ghs_reuse_token"
		mockExpiry := time.Now().Add(60 * time.Minute)
		authService.SetMockTokenResponse(mockToken, mockExpiry)

		apiService := github_service.NewAPIService(authService)
		ctx := context.Background()
		installationID := "test-installation-reuse"

		// Act - Create multiple clients
		client1, err := apiService.GetClient(ctx, installationID)
		assert.NoError(t, err)
		assert.NEmpty(t, client1)

		client2, err := apiService.GetClient(ctx, installationID)
		assert.NoError(t, err)
		assert.NEmpty(t, client2)

		client3, err := apiService.GetClient(ctx, installationID)
		assert.NoError(t, err)
		assert.NEmpty(t, client3)

		// Assert - All clients should be created with same cached token
		cachedToken, found := authService.GetCachedToken(installationID)
		assert.True(t, found)
		assert.Equal(t, mockToken, cachedToken)
	})
}

// Test_WebhookSignatureValidation_Integration tests webhook validation in realistic scenario
func Test_WebhookSignatureValidation_Integration(t *testing.T) {
	t.Run("Validates webhook signature with realistic payload", func(t *testing.T) {
		// Arrange - Simulate real GitHub webhook payload
		secret := "production-webhook-secret"
		payload := []byte(`{
			"action": "created",
			"installation": {
				"id": 12345678,
				"account": {
					"login": "test-org",
					"id": 87654321,
					"type": "Organization"
				},
				"repository_selection": "selected",
				"permissions": {
					"contents": "write",
					"pull_requests": "write"
				},
				"app_slug": "techboss-ai"
			},
			"repositories": [
				{
					"id": 123,
					"name": "test-repo",
					"full_name": "test-org/test-repo"
				}
			],
			"sender": {
				"login": "test-user",
				"id": 999
			}
		}`)

		// Generate valid signature
		validSignature := generateWebhookSignature(payload, secret)

		// Act & Assert - Valid signature
		valid := github_service.ValidateWebhookSignature(payload, validSignature, secret)
		assert.True(t, valid)

		// Act & Assert - Invalid signature (wrong secret)
		invalid := github_service.ValidateWebhookSignature(payload, validSignature, "wrong-secret")
		assert.Equal(t, false, invalid)

		// Act & Assert - Tampered payload
		tamperedPayload := []byte(`{"action": "deleted"}`)
		invalid = github_service.ValidateWebhookSignature(tamperedPayload, validSignature, secret)
		assert.Equal(t, false, invalid)
	})
}

// Test_ConcurrentTokenAccess_Integration tests thread-safety in realistic concurrent scenario
func Test_ConcurrentTokenAccess_Integration(t *testing.T) {
	t.Run("Handles concurrent token requests safely", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)

		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		mockToken := "ghs_concurrent_token"
		mockExpiry := time.Now().Add(60 * time.Minute)
		authService.SetMockTokenResponse(mockToken, mockExpiry)

		ctx := context.Background()
		installationID := "test-installation-concurrent"

		// Act - Simulate concurrent requests from multiple goroutines
		const numGoroutines = 20
		done := make(chan string, numGoroutines)
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				token, err := authService.GetInstallationToken(ctx, installationID)
				if err != nil {
					errors <- err
					return
				}
				done <- token
			}()
		}

		// Assert - All goroutines should succeed with same token
		receivedTokens := make([]string, 0, numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			select {
			case token := <-done:
				receivedTokens = append(receivedTokens, token)
			case err := <-errors:
				t.Fatalf("Concurrent token request failed: %v", err)
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for concurrent requests")
			}
		}

		// All tokens should be the same
		for _, token := range receivedTokens {
			assert.Equal(t, mockToken, token)
		}

		// Verify cache consistency
		cachedToken, found := authService.GetCachedToken(installationID)
		assert.True(t, found)
		assert.Equal(t, mockToken, cachedToken)
	})
}

// generateWebhookSignature generates a valid GitHub webhook signature for testing
func generateWebhookSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	signature := hex.EncodeToString(h.Sum(nil))
	return "sha256=" + signature
}
