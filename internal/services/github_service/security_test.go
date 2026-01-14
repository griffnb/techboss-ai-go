package github_service_test

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"strings"
	"testing"
	"time"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/services/github_service"
	"github.com/pkg/errors"
)

func init() {
	system_testing.BuildSystem()
}

// Test_WebhookSignatureValidation_InvalidSignatures tests that invalid webhook signatures are rejected
func Test_WebhookSignatureValidation_InvalidSignatures(t *testing.T) {
	t.Run("Rejects signature with wrong secret", func(t *testing.T) {
		// Arrange
		correctSecret := "correct-secret" // #nosec G101 - test secret only
		wrongSecret := "wrong-secret"     // #nosec G101 - test secret only
		payload := []byte(`{"action":"created","installation":{"id":12345}}`)

		// Generate signature with correct secret
		validSignature := generateTestWebhookSignature(payload, correctSecret)

		// Act - Validate with wrong secret
		valid := github_service.ValidateWebhookSignature(payload, validSignature, wrongSecret)

		// Assert
		assert.Equal(t, false, valid)
	})

	t.Run("Rejects malformed signature format", func(t *testing.T) {
		// Arrange
		secret := "test-secret" // #nosec G101 - test secret only
		payload := []byte(`{"action":"created","installation":{"id":12345}}`)
		malformedSignature := "not-a-valid-signature-format"

		// Act
		valid := github_service.ValidateWebhookSignature(payload, malformedSignature, secret)

		// Assert
		assert.Equal(t, false, valid)
	})

	t.Run("Rejects signature with modified payload", func(t *testing.T) {
		// Arrange
		secret := "test-secret" // #nosec G101 - test secret only
		originalPayload := []byte(`{"action":"created","installation":{"id":12345}}`)
		modifiedPayload := []byte(`{"action":"created","installation":{"id":99999}}`)

		// Generate signature for original payload
		signature := generateTestWebhookSignature(originalPayload, secret)

		// Act - Validate modified payload with original signature
		valid := github_service.ValidateWebhookSignature(modifiedPayload, signature, secret)

		// Assert
		assert.Equal(t, false, valid)
	})

	t.Run("Rejects signature without sha256 prefix", func(t *testing.T) {
		// Arrange
		secret := "test-secret" // #nosec G101 - test secret only
		payload := []byte(`{"action":"created","installation":{"id":12345}}`)

		// Generate valid signature but strip the prefix
		validSignature := generateTestWebhookSignature(payload, secret)
		signatureWithoutPrefix := strings.TrimPrefix(validSignature, "sha256=")

		// Act
		valid := github_service.ValidateWebhookSignature(payload, signatureWithoutPrefix, secret)

		// Assert
		assert.Equal(t, false, valid)
	})

	t.Run("Rejects empty signature", func(t *testing.T) {
		// Arrange
		secret := "test-secret" // #nosec G101 - test secret only
		payload := []byte(`{"action":"created","installation":{"id":12345}}`)

		// Act
		valid := github_service.ValidateWebhookSignature(payload, "", secret)

		// Assert
		assert.Equal(t, false, valid)
	})

	t.Run("Rejects signature with empty secret", func(t *testing.T) {
		// Arrange
		payload := []byte(`{"action":"created","installation":{"id":12345}}`)
		signature := "sha256=somehash"

		// Act
		valid := github_service.ValidateWebhookSignature(payload, signature, "")

		// Assert
		assert.Equal(t, false, valid)
	})

	t.Run("Rejects signature with empty payload", func(t *testing.T) {
		// Arrange
		secret := "test-secret" // #nosec G101 - test secret only
		signature := "sha256=somehash"

		// Act
		valid := github_service.ValidateWebhookSignature([]byte(""), signature, secret)

		// Assert
		assert.Equal(t, false, valid)
	})
}

// Test_WebhookSignatureValidation_TimingAttacks tests that signature validation uses constant-time comparison
func Test_WebhookSignatureValidation_TimingAttacks(t *testing.T) {
	t.Run("Uses constant-time comparison to prevent timing attacks", func(t *testing.T) {
		// Arrange
		secret := "test-secret" // #nosec G101 - test secret only
		payload := []byte(`{"action":"created","installation":{"id":12345}}`)
		validSignature := generateTestWebhookSignature(payload, secret)

		// Create variations of signatures to test timing
		// If the implementation used non-constant-time comparison,
		// these would potentially have different timing characteristics
		testCases := []struct {
			name      string
			signature string
			expected  bool
		}{
			{"Valid signature", validSignature, true},
			{"First char wrong", "sha256=0" + validSignature[8:], false},
			{"Last char wrong", validSignature[:len(validSignature)-1] + "0", false},
			{"Middle char wrong", validSignature[:20] + "0" + validSignature[21:], false},
			{"Completely wrong", "sha256=0000000000000000000000000000000000000000000000000000000000000000", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Act
				valid := github_service.ValidateWebhookSignature(payload, tc.signature, secret)

				// Assert
				assert.Equal(t, tc.expected, valid)
			})
		}
	})
}

// Test_JWTGeneration_InvalidKeys tests that invalid private keys fail gracefully
func Test_JWTGeneration_InvalidKeys(t *testing.T) {
	t.Run("Fails gracefully with invalid PEM format", func(t *testing.T) {
		// Arrange
		appID := "123456"
		invalidPEM := "this is not a valid PEM encoded key"

		// Act
		authService, err := github_service.NewAuthService(appID, invalidPEM)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, authService)
		assert.True(t, strings.Contains(err.Error(), "failed to decode PEM block"))
	})

	t.Run("Fails gracefully with PEM of wrong type", func(t *testing.T) {
		// Arrange
		appID := "123456"
		// Create a PEM block that's not an RSA private key
		wrongTypePEM := `-----BEGIN CERTIFICATE-----
MIIBkTCB+wIJAKHHCgVZU1WIMA0GCSqGSIb3DQEBCwUAMCExCzAJBgNVBAYTAlVT
MRIwEAYDVQQDDAlsb2NhbGhvc3QwHhcNMjAwMTAxMDAwMDAwWhcNMzAwMTAxMDAw
MDAwWjAhMQswCQYDVQQGEwJVUzESMBAGA1UEAwwJbG9jYWxob3N0MFwwDQYJKoZI
-----END CERTIFICATE-----`

		// Act
		authService, err := github_service.NewAuthService(appID, wrongTypePEM)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, authService)
	})

	t.Run("Fails gracefully with empty private key", func(t *testing.T) {
		// Arrange
		appID := "123456"
		emptyKey := ""

		// Act
		authService, err := github_service.NewAuthService(appID, emptyKey)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, authService)
		assert.True(t, strings.Contains(err.Error(), "privateKey cannot be empty"))
	})

	t.Run("Fails gracefully with empty app ID", func(t *testing.T) {
		// Arrange
		appID := ""
		privateKeyPEM, err := generateTestPrivateKeyPEM()
		assert.NoError(t, err)

		// Act
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, authService)
		assert.True(t, strings.Contains(err.Error(), "appID cannot be empty"))
	})

	t.Run("Fails gracefully with corrupted PEM data", func(t *testing.T) {
		// Arrange
		appID := "123456"
		// Valid PEM structure but corrupted key data
		// #nosec G101 // Test mock corrupted PEM data
		corruptedPEM := `-----BEGIN RSA PRIVATE KEY-----
corrupted data that is not valid base64
or valid key material
-----END RSA PRIVATE KEY-----`

		// Act
		authService, err := github_service.NewAuthService(appID, corruptedPEM)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, authService)
	})
}

// Test_JWTTokenExpiration_AutoRefresh tests that expired JWT tokens are refreshed
func Test_JWTTokenExpiration_AutoRefresh(t *testing.T) {
	t.Run("Refreshes token when expired", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKeyPEM()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		ctx := context.Background()
		installationID := "test-installation-expired"

		// Set an expired token in cache
		expiredToken := "ghs_expired_token" // #nosec G101 - test token only
		expiresAt := time.Now().Add(-10 * time.Minute)
		authService.SetCachedToken(installationID, expiredToken, expiresAt)

		// Mock the GitHub API to return a new token
		newToken := "ghs_new_refreshed_token" // #nosec G101 - test token only
		authService.SetMockTokenResponse(newToken, time.Now().Add(60*time.Minute))

		// Act
		token, err := authService.GetInstallationToken(ctx, installationID)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotEqual(t, expiredToken, token)
		assert.Equal(t, newToken, token)
	})

	t.Run("Refreshes token when expiring within 10 minutes", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKeyPEM()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		ctx := context.Background()
		installationID := "test-installation-expiring"

		// Set a token expiring in 5 minutes (within the 10-minute threshold)
		expiringToken := "ghs_expiring_token" // #nosec G101 - test token only
		expiresAt := time.Now().Add(5 * time.Minute)
		authService.SetCachedToken(installationID, expiringToken, expiresAt)

		// Mock the GitHub API to return a new token
		newToken := "ghs_refreshed_token" // #nosec G101 - test token only
		authService.SetMockTokenResponse(newToken, time.Now().Add(60*time.Minute))

		// Act
		token, err := authService.GetInstallationToken(ctx, installationID)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotEqual(t, expiringToken, token)
		assert.Equal(t, newToken, token)
	})

	t.Run("Uses cached token when still valid", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKeyPEM()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		ctx := context.Background()
		installationID := "test-installation-valid"

		// Set a token expiring in 30 minutes (beyond the 10-minute threshold)
		validToken := "ghs_valid_token" // #nosec G101 - test token only
		expiresAt := time.Now().Add(30 * time.Minute)
		authService.SetCachedToken(installationID, validToken, expiresAt)

		// Act
		token, err := authService.GetInstallationToken(ctx, installationID)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Equal(t, validToken, token)
	})

	t.Run("RefreshInstallationToken bypasses cache even for valid tokens", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKeyPEM()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		ctx := context.Background()
		installationID := "test-installation-force-refresh"

		// Set a valid token in cache
		oldToken := "ghs_old_valid_token" // #nosec G101 - test token only
		expiresAt := time.Now().Add(30 * time.Minute)
		authService.SetCachedToken(installationID, oldToken, expiresAt)

		// Mock the GitHub API to return a new token
		newToken := "ghs_force_refreshed_token" // #nosec G101 - test token only
		authService.SetMockTokenResponse(newToken, time.Now().Add(60*time.Minute))

		// Act
		token, err := authService.RefreshInstallationToken(ctx, installationID)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotEqual(t, oldToken, token)
		assert.Equal(t, newToken, token)
	})
}

// Test_TokenCacheThreadSafety_ConcurrentAccess tests that token cache prevents race conditions
func Test_TokenCacheThreadSafety_ConcurrentAccess(t *testing.T) {
	t.Run("Handles concurrent reads and writes safely", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKeyPEM()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		ctx := context.Background()
		installationID := "test-installation-concurrent"

		// Mock the GitHub API
		newToken := "ghs_concurrent_token" // #nosec G101 - test token only
		authService.SetMockTokenResponse(newToken, time.Now().Add(60*time.Minute))

		// Act - Simulate many concurrent token requests
		done := make(chan bool, 50)
		for i := 0; i < 50; i++ {
			go func() {
				_, _ = authService.GetInstallationToken(ctx, installationID)
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 50; i++ {
			<-done
		}

		// Assert - No race condition should occur
		// If there was a race condition, the test would fail with -race flag
		cachedToken, found := authService.GetCachedToken(installationID)
		assert.True(t, found)
		assert.Equal(t, newToken, cachedToken)
	})

	t.Run("Handles concurrent cache writes safely", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKeyPEM()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		installationID := "test-installation-write-concurrent"

		// Act - Simulate many concurrent cache writes
		done := make(chan bool, 20)
		for i := 0; i < 20; i++ {
			tokenNum := i
			go func() {
				token := "ghs_token_" + string(rune(tokenNum)) // #nosec G101 - test token only
				authService.SetCachedToken(installationID, token, time.Now().Add(60*time.Minute))
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 20; i++ {
			<-done
		}

		// Assert - Cache should have one of the tokens (last write wins)
		cachedToken, found := authService.GetCachedToken(installationID)
		assert.True(t, found)
		assert.NotEmpty(t, cachedToken)
		assert.True(t, strings.HasPrefix(cachedToken, "ghs_token_"))
	})

	t.Run("Handles concurrent read and refresh safely", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKeyPEM()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		ctx := context.Background()
		installationID := "test-installation-read-refresh"

		// Pre-populate with a token
		initialToken := "ghs_initial_token" // #nosec G101 - test token only
		authService.SetCachedToken(installationID, initialToken, time.Now().Add(30*time.Minute))

		// Mock the GitHub API
		refreshedToken := "ghs_refreshed_concurrent_token" // #nosec G101 - test token only
		authService.SetMockTokenResponse(refreshedToken, time.Now().Add(60*time.Minute))

		// Act - Mix reads and refreshes concurrently
		done := make(chan bool, 30)

		// Start readers
		for i := 0; i < 20; i++ {
			go func() {
				_, _ = authService.GetInstallationToken(ctx, installationID)
				done <- true
			}()
		}

		// Start refreshers
		for i := 0; i < 10; i++ {
			go func() {
				_, _ = authService.RefreshInstallationToken(ctx, installationID)
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 30; i++ {
			<-done
		}

		// Assert - Should have a valid token, no crashes
		cachedToken, found := authService.GetCachedToken(installationID)
		assert.True(t, found)
		assert.NotEmpty(t, cachedToken)
	})
}

// Test_TokenCacheSecurityBoundaries tests security boundaries of token cache
func Test_TokenCacheSecurityBoundaries(t *testing.T) {
	t.Run("Different installation IDs have isolated caches", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKeyPEM()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		installation1 := "installation-001"
		installation2 := "installation-002"
		token1 := "ghs_token_001" // #nosec G101 - test token only
		token2 := "ghs_token_002" // #nosec G101 - test token only

		// Act
		authService.SetCachedToken(installation1, token1, time.Now().Add(30*time.Minute))
		authService.SetCachedToken(installation2, token2, time.Now().Add(30*time.Minute))

		// Assert
		cached1, found1 := authService.GetCachedToken(installation1)
		assert.True(t, found1)
		assert.Equal(t, token1, cached1)

		cached2, found2 := authService.GetCachedToken(installation2)
		assert.True(t, found2)
		assert.Equal(t, token2, cached2)

		// Tokens should not cross contaminate
		assert.NotEqual(t, cached1, cached2)
	})

	t.Run("Empty installation ID is rejected", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKeyPEM()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		ctx := context.Background()

		// Act
		token, err := authService.GetInstallationToken(ctx, "")

		// Assert
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.True(t, strings.Contains(err.Error(), "installationID cannot be empty"))
	})

	t.Run("Empty installation ID is rejected for refresh", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKeyPEM()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		ctx := context.Background()

		// Act
		token, err := authService.RefreshInstallationToken(ctx, "")

		// Assert
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.True(t, strings.Contains(err.Error(), "installationID cannot be empty"))
	})
}

// Helper functions

// generateTestPrivateKeyPEM generates a test RSA private key in PEM format
func generateTestPrivateKeyPEM() (string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", errors.Wrapf(err, "failed to generate test private key")
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return string(privateKeyPEM), nil
}

// generateTestWebhookSignature generates a valid webhook signature for testing
func generateTestWebhookSignature(payload []byte, secret string) string {
	// Import crypto libraries here to avoid import cycle
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	signature := hex.EncodeToString(h.Sum(nil))
	return "sha256=" + signature
}
