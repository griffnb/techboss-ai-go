package github_service

import (
	"context"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

// AuthService handles GitHub App authentication including JWT generation and token caching
type AuthService struct {
	appID      string
	privateKey *rsa.PrivateKey
	tokenCache *TokenCache

	// For testing - allows mocking GitHub API responses
	mockTokenResponse *MockTokenResponse
}

// TokenCache stores installation access tokens with expiration times
type TokenCache struct {
	mu     sync.RWMutex
	tokens map[string]*InstallationToken
}

// InstallationToken represents a cached installation access token
type InstallationToken struct {
	Token     string
	ExpiresAt time.Time
}

// MockTokenResponse allows testing without hitting GitHub API
type MockTokenResponse struct {
	Token     string
	ExpiresAt time.Time
}

// NewAuthService creates a new GitHub authentication service
func NewAuthService(appID string, privateKeyPEM string) (*AuthService, error) {
	if appID == "" {
		return nil, errors.New("appID cannot be empty")
	}

	if privateKeyPEM == "" {
		return nil, errors.New("privateKey cannot be empty")
	}

	// Parse PEM private key
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	// Parse RSA private key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse RSA private key")
	}

	return &AuthService{
		appID:      appID,
		privateKey: privateKey,
		tokenCache: &TokenCache{
			tokens: make(map[string]*InstallationToken),
		},
	}, nil
}

// GenerateJWT generates a JWT token for authenticating as the GitHub App
func (s *AuthService) GenerateJWT() (string, error) {
	now := time.Now()

	// Create JWT claims
	claims := jwt.MapClaims{
		"iss": s.appID,
		"iat": now.Unix(),
		"exp": now.Add(10 * time.Minute).Unix(),
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Sign token with private key
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", errors.Wrapf(err, "failed to sign JWT")
	}

	return tokenString, nil
}

// GetInstallationToken retrieves an installation access token, using cache when possible
func (s *AuthService) GetInstallationToken(ctx context.Context, installationID string) (string, error) {
	if installationID == "" {
		return "", errors.New("installationID cannot be empty")
	}

	// Check cache for valid token
	s.tokenCache.mu.RLock()
	cachedToken, exists := s.tokenCache.tokens[installationID]
	s.tokenCache.mu.RUnlock()

	if exists {
		// Check if token is still valid (not expired within 10 minutes)
		if time.Now().Add(10 * time.Minute).Before(cachedToken.ExpiresAt) {
			return cachedToken.Token, nil
		}
	}

	// Token not in cache or expired - fetch new token
	return s.fetchInstallationToken(ctx, installationID)
}

// RefreshInstallationToken forces a token refresh, bypassing the cache
func (s *AuthService) RefreshInstallationToken(ctx context.Context, installationID string) (string, error) {
	if installationID == "" {
		return "", errors.New("installationID cannot be empty")
	}

	return s.fetchInstallationToken(ctx, installationID)
}

// fetchInstallationToken fetches a new installation token from GitHub API (or mock)
func (s *AuthService) fetchInstallationToken(_ context.Context, installationID string) (string, error) {
	// For testing - use mock response if available
	if s.mockTokenResponse != nil {
		token := s.mockTokenResponse.Token
		expiresAt := s.mockTokenResponse.ExpiresAt

		// Update cache
		s.setCachedToken(installationID, token, expiresAt)

		return token, nil
	}

	// TODO: In production, this would call GitHub API:
	// 1. Generate JWT
	// 2. Create GitHub client with JWT
	// 3. Call POST /app/installations/{installationID}/access_tokens
	// 4. Cache and return the token

	return "", errors.New("GitHub API integration not yet implemented - use SetMockTokenResponse for testing")
}

// setCachedToken updates the token cache (thread-safe internal method)
func (s *AuthService) setCachedToken(installationID string, token string, expiresAt time.Time) {
	s.tokenCache.mu.Lock()
	defer s.tokenCache.mu.Unlock()

	s.tokenCache.tokens[installationID] = &InstallationToken{
		Token:     token,
		ExpiresAt: expiresAt,
	}
}

// SetCachedToken manually sets a cached token (for testing)
func (s *AuthService) SetCachedToken(installationID string, token string, expiresAt time.Time) {
	s.setCachedToken(installationID, token, expiresAt)
}

// GetCachedToken retrieves a token from cache (for testing)
func (s *AuthService) GetCachedToken(installationID string) (string, bool) {
	s.tokenCache.mu.RLock()
	defer s.tokenCache.mu.RUnlock()

	cachedToken, exists := s.tokenCache.tokens[installationID]
	if !exists {
		return "", false
	}

	return cachedToken.Token, true
}

// SetMockTokenResponse sets a mock token response (for testing)
func (s *AuthService) SetMockTokenResponse(token string, expiresAt time.Time) {
	s.mockTokenResponse = &MockTokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}
}

// ValidateWebhookSignature validates a GitHub webhook signature using HMAC-SHA256
func ValidateWebhookSignature(payload []byte, signature string, secret string) bool {
	if len(payload) == 0 || signature == "" || secret == "" {
		return false
	}

	// GitHub signature format: "sha256=<hex-encoded-hash>"
	if len(signature) < 7 || signature[:7] != "sha256=" {
		return false
	}

	// Extract hex signature (remove "sha256=" prefix)
	hexSignature := signature[7:]

	// Compute HMAC-SHA256 of payload
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	expectedMAC := h.Sum(nil)
	expectedHex := hex.EncodeToString(expectedMAC)

	// Use constant-time comparison to prevent timing attacks
	// Note: hmac.Equal internally uses crypto/subtle.ConstantTimeCompare
	return hmac.Equal([]byte(hexSignature), []byte(expectedHex))
}
