package organization_service

import (
	"context"

	"github.com/griffnb/core/lib/model/coremodel"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/models/organization"

	"github.com/pkg/errors"
)

// CompleteSetup finalizes the setup process for an organization and updates the user's status accordingly.
func CompleteSetup(ctx context.Context, org *organization.Organization, user *account.Account) error {
	if user.Status.Get() == account.STATUS_PENDING_ONBOARD {
		user.Status.Set(account.STATUS_ACTIVE)
		err := user.Save(user)
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO default plan
func CreateDefaultOrganization(
	ctx context.Context,
	accountObj *account.Account,
	savingUser coremodel.Model,
) (*organization.Organization, error) {
	org := organization.New()
	org.Name.Set(accountObj.GetName())
	org.Properties.Set(&organization.Properties{
		BillingEmail: accountObj.Email.Get(),
	})

	err := org.Save(savingUser)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return org, nil
}
