package clerk

import (
	"context"
	"encoding/json"

	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/models/admin"
	"github.com/griffnb/techboss-ai-go/internal/models/organization"
	"github.com/pkg/errors"
)

func SyncAdmin(ctx context.Context, claims *clerk.SessionClaims) (*admin.Admin, error) {
	if claims == nil || claims.Subject == "" {
		return nil, errors.New("invalid claims")
	}

	customClaims, ok := claims.Custom.(*CustomSessionClaims)
	if !ok {
		return nil, errors.New("invalid custom claims")
	}

	adminObj, err := admin.GetByExternalID(ctx, customClaims.ExternalID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if tools.Empty(adminObj) {
		return nil, nil
	}

	clerkUser, err := user.Get(ctx, claims.Subject)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if clerkUser == nil {
		return nil, errors.New("could not find user in clerk")
	}

	_, err = UpdateClerkAdmin(ctx, adminObj)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return adminObj, nil
}

func CreateAccount(ctx context.Context, claims *clerk.SessionClaims) (*account.Account, error) {
	if claims == nil || claims.Subject == "" {
		return nil, errors.New("invalid claims")
	}

	customClaims, ok := claims.Custom.(*CustomSessionClaims)
	if !ok {
		return nil, errors.New("invalid custom claims")
	}

	accountObj, err := account.GetByExternalID(ctx, customClaims.ExternalID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !tools.Empty(accountObj) {
		return accountObj, nil
	}

	clerkUser, err := user.Get(ctx, claims.Subject)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if clerkUser == nil {
		return nil, errors.New("could not find user in clerk")
	}

	accountObj = account.New()
	accountObj.ExternalID.Set(customClaims.ExternalID)
	accountObj.Email.Set(customClaims.Email)
	accountObj.FirstName.Set(*clerkUser.FirstName)
	accountObj.LastName.Set(*clerkUser.LastName)
	accountObj.Role.Set(constants.ROLE_USER)
	org := organization.New()
	org.Name.Set(accountObj.GetName())
	err = org.Save(nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	accountObj.OrganizationID.Set(org.ID())
	err = accountObj.Save(nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	_, err = UpdateClerkUser(ctx, accountObj)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return accountObj, nil
}

func UpdateClerkAdmin(ctx context.Context, adminObj *admin.Admin) (*clerk.User, error) {
	metadata := map[string]any{
		"role":        adminObj.Role.Get(),
		"external_id": adminObj.ID(),
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	resp, err := user.Update(ctx, adminObj.ExternalID.Get(), &user.UpdateParams{
		ExternalID:     clerk.String(adminObj.ExternalID.Get()),
		PublicMetadata: clerk.JSONRawMessage(metadataBytes),
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return resp, nil
}

func UpdateClerkUser(ctx context.Context, accountObj *account.Account) (*clerk.User, error) {
	metadata := map[string]any{
		"role":            accountObj.Role.Get(),
		"external_id":     accountObj.ID(),
		"organization_id": accountObj.OrganizationID.Get(),
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	resp, err := user.Update(ctx, accountObj.ExternalID.Get(), &user.UpdateParams{
		ExternalID:     clerk.String(accountObj.ExternalID.Get()),
		PublicMetadata: clerk.JSONRawMessage(metadataBytes),
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return resp, nil
}
