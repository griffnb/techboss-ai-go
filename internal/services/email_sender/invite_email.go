package email_sender

import (
	"context"
	"fmt"
	"html/template"
	"strings"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/techboss-ai-go/internal/models/account"

	"github.com/pkg/errors"
)

type InviteEmailTemplate struct {
	Logo     string
	Name     string
	FromName string
	Link     string
	Email    string
}

func BuildInviteEmail(requestAccount *account.Account, fromAccount *account.Account, inviteLink string) (string, error) {
	data := InviteEmailTemplate{
		Name:     fmt.Sprintf("%s %s", requestAccount.FirstName.Get(), requestAccount.LastName.Get()),
		FromName: fmt.Sprintf("%s %s", fromAccount.FirstName.Get(), fromAccount.LastName.Get()),
		Link:     inviteLink,
		Email:    requestAccount.Email.Get(),
		Logo:     "https://app.atlas.net/img/logo.png",
	}

	emailTemplate, err := GetEmailTemplate("AcceptInvite.html")
	if err != nil {
		return "", err
	}

	// Parse the template file.
	tmpl, err := template.New("emailTemplate").Parse(emailTemplate)
	if err != nil {
		return "", errors.WithStack(err)
	}
	var bufBody strings.Builder
	// Execute the template, passing in the data to fill in the placeholders.
	err = tmpl.Execute(&bufBody, data) // Use os.Stdout for demonstration, but you can use any io.Writer like a file or a buffer.
	if err != nil {
		return "", errors.WithStack(err)
	}

	log.Debug("Invite Email Link: ", inviteLink)

	return bufBody.String(), nil
}

func SendInviteEmail(requestAccount *account.Account, primaryAccount *account.Account, inviteLink string) error {
	message, err := BuildInviteEmail(requestAccount, primaryAccount, inviteLink)
	if err != nil {
		return err
	}
	subject := "You've Been Invited to Join TechBoss"

	return Send(context.Background(), "accept_invite", []string{requestAccount.Email.Get()}, subject, message)
}
