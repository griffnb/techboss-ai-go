package clerk

import (
	"context"

	"github.com/CrowdShield/go-core/lib/types"
	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/pkg/errors"
)

type CustomSessionClaims struct {
	Email          string         `json:"email,omitempty"`
	AccountID      types.UUID     `json:"account_id,omitempty"`
	Role           constants.Role `json:"role,omitempty"`
	AdminRole      constants.Role `json:"admin_role,omitempty"`
	AdminID        types.UUID     `json:"admin_id,omitempty"`
	OrganizationID types.UUID     `json:"organization_id,omitempty"`
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
