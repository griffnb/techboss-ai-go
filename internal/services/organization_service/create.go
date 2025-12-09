package organization_service

import (
	"context"

	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/models/organization"

	"github.com/pkg/errors"
)

func CreateOrganization(_ context.Context, accountObj *account.Account) (*organization.Organization, error) {
	org := organization.New()
	org.Name.Set(accountObj.GetName())
	org.Properties.Set(&organization.Properties{
		BillingEmail: accountObj.Email.Get(),
	})
	err := org.Save(nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return org, nil
}
