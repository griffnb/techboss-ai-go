package github_service_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/services/github_service"
	"github.com/pkg/errors"
)

func init() {
	system_testing.BuildSystem()
}

// generateTestPrivateKey generates an RSA private key for testing
func generateTestPrivateKey() (string, error) {
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

func Test_NewAuthService(t *testing.T) {
	t.Run("Creates auth service successfully with valid key", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)

		// Act
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, authService)
	})

	t.Run("Returns error with invalid PEM", func(t *testing.T) {
		// Arrange
		appID := "123456"
		invalidPEM := "not a valid pem"

		// Act
		authService, err := github_service.NewAuthService(appID, invalidPEM)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, authService)
	})

	t.Run("Returns error with empty appID", func(t *testing.T) {
		// Arrange
		appID := ""
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)

		// Act
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, authService)
	})
}

func Test_GenerateJWT(t *testing.T) {
	t.Run("Generates valid JWT token", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		// Act
		token, err := authService.GenerateJWT()

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Verify token structure
		parts := strings.Split(token, ".")
		assert.Equal(t, 3, len(parts)) // JWT has 3 parts: header.payload.signature
	})

	t.Run("JWT contains correct claims", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		// Act
		tokenString, err := authService.GenerateJWT()
		assert.NoError(t, err)

		// Parse token without verification (we just want to check claims)
		parser := jwt.NewParser(jwt.WithoutClaimsValidation())
		token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
		assert.NoError(t, err)

		claims, ok := token.Claims.(jwt.MapClaims)
		assert.True(t, ok)

		// Assert
		iss, ok := claims["iss"].(string)
		assert.True(t, ok)
		assert.Equal(t, appID, iss)
		assert.NEmpty(t, claims["iat"])
		assert.NEmpty(t, claims["exp"])

		// Check expiration is ~10 minutes from now
		exp := int64(claims["exp"].(float64))
		iat := int64(claims["iat"].(float64))
		duration := exp - iat
		assert.True(t, duration >= 590 && duration <= 610) // 10 minutes ±10 seconds
	})
}

func Test_GetInstallationToken(t *testing.T) {
	t.Run("Returns cached token if valid", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		ctx := context.Background()
		installationID := "test-installation-123"

		// Pre-populate cache with a valid token
		testToken := "ghs_testtoken123"
		expiresAt := time.Now().Add(20 * time.Minute)
		authService.SetCachedToken(installationID, testToken, expiresAt)

		// Act
		token, err := authService.GetInstallationToken(ctx, installationID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, testToken, token)
	})

	t.Run("Refreshes token if expired", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		ctx := context.Background()
		installationID := "test-installation-456"

		// Pre-populate cache with an expired token
		expiredToken := "ghs_expiredtoken"
		expiresAt := time.Now().Add(-5 * time.Minute) // Already expired
		authService.SetCachedToken(installationID, expiredToken, expiresAt)

		// Mock the GitHub API call
		newToken := "ghs_newtoken789" // #nosec G101 - test token only
		authService.SetMockTokenResponse(newToken, time.Now().Add(60*time.Minute))

		// Act
		token, err := authService.GetInstallationToken(ctx, installationID)

		// Assert
		assert.NoError(t, err)
		assert.NotEqual(t, expiredToken, token)
		assert.Equal(t, newToken, token)
	})

	t.Run("Refreshes token if expiring within 10 minutes", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		ctx := context.Background()
		installationID := "test-installation-789"

		// Pre-populate cache with a token expiring in 5 minutes
		oldToken := "ghs_expiringtoken" // #nosec G101 - test token only
		expiresAt := time.Now().Add(5 * time.Minute)
		authService.SetCachedToken(installationID, oldToken, expiresAt)

		// Mock the GitHub API call
		newToken := "ghs_refreshedtoken" // #nosec G101 - test token only
		authService.SetMockTokenResponse(newToken, time.Now().Add(60*time.Minute))

		// Act
		token, err := authService.GetInstallationToken(ctx, installationID)

		// Assert
		assert.NoError(t, err)
		assert.NotEqual(t, oldToken, token)
		assert.Equal(t, newToken, token)
	})

	t.Run("Returns error with empty installation ID", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		ctx := context.Background()
		installationID := ""

		// Act
		token, err := authService.GetInstallationToken(ctx, installationID)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, token)
	})
}

func Test_RefreshInstallationToken(t *testing.T) {
	t.Run("Forces token refresh and updates cache", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		ctx := context.Background()
		installationID := "test-installation-refresh"

		// Pre-populate cache with a valid token
		oldToken := "ghs_oldtoken" // #nosec G101 - test token only
		expiresAt := time.Now().Add(30 * time.Minute)
		authService.SetCachedToken(installationID, oldToken, expiresAt)

		// Mock the GitHub API call
		newToken := "ghs_forcedrefreshtoken" // #nosec G101 - test token only
		authService.SetMockTokenResponse(newToken, time.Now().Add(60*time.Minute))

		// Act
		token, err := authService.RefreshInstallationToken(ctx, installationID)

		// Assert
		assert.NoError(t, err)
		assert.NotEqual(t, oldToken, token)
		assert.Equal(t, newToken, token)

		// Verify cache was updated
		cachedToken, _ := authService.GetCachedToken(installationID)
		assert.Equal(t, newToken, cachedToken)
	})

	t.Run("Returns error with empty installation ID", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		ctx := context.Background()
		installationID := ""

		// Act
		token, err := authService.RefreshInstallationToken(ctx, installationID)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, token)
	})
}

func Test_TokenCacheThreadSafety(t *testing.T) {
	t.Run("Cache operations are thread-safe", func(t *testing.T) {
		// Arrange
		appID := "123456"
		privateKeyPEM, err := generateTestPrivateKey()
		assert.NoError(t, err)
		authService, err := github_service.NewAuthService(appID, privateKeyPEM)
		assert.NoError(t, err)

		ctx := context.Background()
		installationID := "test-installation-concurrent"

		// Mock the GitHub API call
		newToken := "ghs_concurrenttoken"
		authService.SetMockTokenResponse(newToken, time.Now().Add(60*time.Minute))

		// Act - Simulate concurrent access
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				_, _ = authService.GetInstallationToken(ctx, installationID)
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		// Assert - No panic should occur, and we should have a cached token
		token, found := authService.GetCachedToken(installationID)
		assert.True(t, found)
		assert.NotEmpty(t, token)
	})
}

func Test_ValidateWebhookSignature(t *testing.T) {
	t.Run("Returns true for valid signature", func(t *testing.T) {
		// Arrange
		secret := "test-webhook-secret" // #nosec G101 - test secret only
		payload := []byte(`{"action":"created","installation":{"id":12345}}`)
		// Pre-computed HMAC-SHA256 signature for the above payload with secret
		// Computed using: echo -n '{"action":"created","installation":{"id":12345}}' | openssl dgst -sha256 -hmac 'test-webhook-secret'
		signature := "sha256=8c84fe36fcf57db1f29851c8e521303a01b74b6f0183bdc246f666e5cd50660c"

		// Act
		valid := github_service.ValidateWebhookSignature(payload, signature, secret)

		// Assert
		assert.True(t, valid)
	})

	t.Run("Returns false for invalid signature", func(t *testing.T) {
		// Arrange
		secret := "test-webhook-secret" // #nosec G101 - test secret only
		payload := []byte(`{"action":"created","installation":{"id":12345}}`)
		invalidSignature := "sha256=invalid_signature_here"

		// Act
		valid := github_service.ValidateWebhookSignature(payload, invalidSignature, secret)

		// Assert
		assert.Equal(t, false, valid)
	})

	t.Run("Returns false for signature with wrong secret", func(t *testing.T) {
		// Arrange
		wrongSecret := "wrong-secret"
		payload := []byte(`{"action":"created","installation":{"id":12345}}`)
		// Signature computed with correct secret (test-webhook-secret)
		signature := "sha256=8c84fe36fcf57db1f29851c8e521303a01b74b6f0183bdc246f666e5cd50660c"

		// Act
		valid := github_service.ValidateWebhookSignature(payload, signature, wrongSecret)

		// Assert
		assert.Equal(t, false, valid)
	})

	t.Run("Returns false for signature without sha256 prefix", func(t *testing.T) {
		// Arrange
		secret := "test-webhook-secret" // #nosec G101 - test secret only
		payload := []byte(`{"action":"created","installation":{"id":12345}}`)
		signatureWithoutPrefix := "8c84fe36fcf57db1f29851c8e521303a01b74b6f0183bdc246f666e5cd50660c"

		// Act
		valid := github_service.ValidateWebhookSignature(payload, signatureWithoutPrefix, secret)

		// Assert
		assert.Equal(t, false, valid)
	})

	t.Run("Returns false for empty signature", func(t *testing.T) {
		// Arrange
		secret := "test-webhook-secret" // #nosec G101 - test secret only
		payload := []byte(`{"action":"created","installation":{"id":12345}}`)
		emptySignature := ""

		// Act
		valid := github_service.ValidateWebhookSignature(payload, emptySignature, secret)

		// Assert
		assert.Equal(t, false, valid)
	})

	t.Run("Returns false for empty secret", func(t *testing.T) {
		// Arrange
		emptySecret := ""
		payload := []byte(`{"action":"created","installation":{"id":12345}}`)
		signature := "sha256=8c84fe36fcf57db1f29851c8e521303a01b74b6f0183bdc246f666e5cd50660c"

		// Act
		valid := github_service.ValidateWebhookSignature(payload, signature, emptySecret)

		// Assert
		assert.Equal(t, false, valid)
	})

	t.Run("Returns false for empty payload", func(t *testing.T) {
		// Arrange
		secret := "test-webhook-secret" // #nosec G101 - test secret only
		emptyPayload := []byte("")
		signature := "sha256=8c84fe36fcf57db1f29851c8e521303a01b74b6f0183bdc246f666e5cd50660c"

		// Act
		valid := github_service.ValidateWebhookSignature(emptyPayload, signature, secret)

		// Assert
		assert.Equal(t, false, valid)
	})
}
