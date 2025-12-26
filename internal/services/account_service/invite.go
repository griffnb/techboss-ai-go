package account_service

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/griffnb/core/lib/model/coremodel"
	"github.com/griffnb/core/lib/session"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/services/email_sender"
	"github.com/pkg/errors"
)

func GenerateInviteSession(_ context.Context, accountObj *account.Account, savingUser coremodel.Model) (string, error) {
	verifySession := session.New("").WithUser(accountObj)
	err := verifySession.SaveWithExpiration(time.Now().Add(24 * 7 * time.Hour).Unix())
	if err != nil {
		return "", err
	}

	props, err := accountObj.Properties.Get()
	if err != nil {
		return "", err
	}

	props.InviteKey = verifySession.Key
	props.InviteTS = time.Now().Unix()
	accountObj.Properties.Set(props)
	err = accountObj.Save(savingUser)
	if err != nil {
		return "", err
	}

	return verifySession.Key, nil
}

func SendInviteEmail(
	ctx context.Context,
	newAccount, primaryAccount *account.Account,
	_ coremodel.Model,
) error {
	link, err := CreateInviteLink(ctx, newAccount, primaryAccount, false)
	if err != nil {
		return err
	}

	// if environment.IsProduction() {
	err = email_sender.SendInviteEmail(newAccount, primaryAccount, link)
	if err != nil {
		return err
	}
	/*
		} else {
			err = email_sender.SendCustomerio(ctx, email_sender.MESSAGE_ID_ACCEPT_INVITE, []string{newAccount.Email.Get()}, map[string]any{
				"invite_name": fmt.Sprintf("%s %s", primaryAccount.FirstName.Get(), primaryAccount.LastName.Get()),
				"first_name":  newAccount.FirstName.Get(),
				"link":        link,
			})
			if err != nil {
				return err
			}

			log.Debug("Invite Email Link: ", link)
		}
	*/

	return nil
}

func CreateInviteLink(ctx context.Context, invitedAccount, primaryAccount *account.Account, emailLess bool) (string, error) {
	if invitedAccount.Status.Get() > account.STATUS_PENDING_EMAIL_VERIFICATION {
		return "", errors.Errorf(
			"account %s is not in a valid state to be invited",
			invitedAccount.ID(),
		)
	} else if invitedAccount.OrganizationID.Get() != primaryAccount.OrganizationID.Get() {
		return "", errors.Errorf(
			"account %s is not in the same org as the primary account %s",
			invitedAccount.ID(),
			primaryAccount.ID())
	}

	sessionKey, err := GenerateInviteSession(ctx, invitedAccount, primaryAccount)
	if err != nil {
		return "", err
	}

	baseURL := environment.GetConfig().Server.AppURL

	params := url.Values{}
	params.Set("verify", sessionKey)
	params.Set("email", invitedAccount.Email.Get())
	if emailLess {
		params.Set("emailless", "true")
	}
	link := fmt.Sprintf("%s/verify/invite?%s", baseURL, params.Encode())
	return link, nil
}
