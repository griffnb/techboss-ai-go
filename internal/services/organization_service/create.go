package organization_service

import (
	"context"

	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/models/organization"

	"github.com/pkg/errors"
)

func CreateOrganization(ctx context.Context, accountObj *account.Account) (*organization.Organization, error) {
	org := organization.New()
	org.Name.Set(accountObj.GetName())
	err := org.Save(nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return org, nil
}
