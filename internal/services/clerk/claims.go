package clerk

import (
	"context"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/pkg/errors"
)

type CustomSessionClaims struct {
	Email      string         `json:"email"`
	ClerkID    string         `json:"clerk_id"`
	ExternalID string         `json:"external_id"`
	Role       constants.Role `json:"role"`
}

func CustomClaims(claims *clerk.SessionClaims) (*CustomSessionClaims, error) {
	customClaims, ok := claims.Custom.(*CustomSessionClaims)
	if !ok {
		return nil, errors.New("invalid custom claims")
	}
	return customClaims, nil
}

func customClaimsConstructor(_ context.Context) any {
	return &CustomSessionClaims{}
}

func WithCustomClaimsConstructor(params *clerkhttp.AuthorizationParams) error {
	// nolint: staticcheck
	params.VerifyParams.CustomClaimsConstructor = customClaimsConstructor
	return nil
}
