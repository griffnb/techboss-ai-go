package account_service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/model/coremodel"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/models/organization"
	"github.com/griffnb/techboss-ai-go/internal/services/email_sender"
	"github.com/griffnb/techboss-ai-go/internal/services/organization_service"
	"github.com/pkg/errors"
)

func EmailVerified(ctx context.Context, accountObj *account.Account, savingUser coremodel.Model) error {
	accountObj.EmailVerifiedAtTS.Set(time.Now().Unix())
	// Check if organization exists for whitelisted domain
	org, err := IsWhitelistedDomain(accountObj)
	if err != nil {
		return err
	}

	if !tools.Empty(org) {
		// Add to organization
		accountObj.OrganizationID.Set(org.ID())
		accountObj.Status.Set(account.STATUS_ACTIVE)
		accountObj.Role.Set(constants.ROLE_USER)
	} else {
		// They need to setup their organization
		accountObj.Status.Set(account.STATUS_PENDING_ONBOARD)
		org, err := organization_service.CreateDefaultOrganization(ctx, accountObj, savingUser)
		if err != nil {
			return err
		}
		accountObj.OrganizationID.Set(org.ID())
		accountObj.Role.Set(constants.ROLE_ORG_OWNER)
	}

	err = accountObj.Save(savingUser)
	if err != nil {
		return err
	}

	return nil
}

// TODO send email verification
// TODO handle invite links
func Signup(ctx context.Context, accountObj *account.Account, savingUser *account.Account) error {
	signupProps, err := accountObj.SignupProperties.Get()
	if err != nil {
		return err
	}

	if accountObj.Status.Get() == account.STATUS_PENDING_EMAIL_VERIFICATION && signupProps.IsOauth != 1 {
		// Send Verification Email
		err = email_sender.SendVerifyEmail(accountObj)
		if err != nil {
			log.ErrorContext(err, ctx)
		}
	}
	return accountObj.Save(savingUser)
}

// IsWhitelistedDomain : Checks if the account email is from a whitelisted domain for a specific ORG to add them to it
func IsWhitelistedDomain(accountObj *account.Account) (*organization.Organization, error) {
	orgs, err := organization.GetWhitelistDomainOrgs(context.Background())
	if err != nil {
		return nil, err
	}

	for _, org := range orgs {
		emailDomains, err := org.EmailDomains.Get()
		if err != nil {
			return nil, err
		}

		for _, domain := range emailDomains {
			if strings.HasSuffix(accountObj.Email.Get(), fmt.Sprintf("@%s", domain)) {
				return org, nil
			}
		}
	}

	return nil, nil
}

func UpdatePrimaryEmailAddressForUnverified(ctx context.Context, accountObj *account.Account, emailAddress string) error {
	if tools.Empty(accountObj) {
		return errors.New("account object is empty")
	}

	if tools.Empty(emailAddress) {
		return errors.New("email address is empty")
	}

	exists, err := account.Exists(ctx, emailAddress)
	if err != nil {
		return errors.Wrap(err, "failed to check if email address exists")
	}
	if exists {
		return errors.New("email address already exists")
	}

	accountObj.Email.Set(emailAddress)
	err = accountObj.Save(accountObj.ToSavingUser())
	if err != nil {
		return errors.Wrap(err, "failed to save account object with new email address")
	}

	return nil
}
