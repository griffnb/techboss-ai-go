package email_sender

import (
	"context"
	"fmt"
	"html/template"
	"strings"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/services/magic_link"

	"github.com/pkg/errors"
)

type LoginLinkEmailTemplate struct {
	Name       string
	Link       string
	ResendLink string
	Email      string
	Logo       string
}

func BuildLoginLinkEmail(requestAccount *account.Account, loginToken string, redirectURL string) (string, error) {
	magicLink := magic_link.GetLoginLink(loginToken, redirectURL)
	resendLink := magic_link.GetResendLink(requestAccount.Email.Get(), loginToken, redirectURL)

	data := LoginLinkEmailTemplate{
		Name:       fmt.Sprintf("%s %s", requestAccount.FirstName.Get(), requestAccount.LastName.Get()),
		Link:       magicLink,
		ResendLink: resendLink,
		Email:      requestAccount.Email.Get(),
		Logo:       "https://app.atlas.net/img/logo.png",
	}

	emailTemplate, err := GetEmailTemplate("LoginLink.html")
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

	log.Debug("Login Link: ", magicLink)

	return bufBody.String(), nil
}

func SendLoginLinkEmail(requestAccount *account.Account, token string, redirectURL string) error {
	message, err := BuildLoginLinkEmail(requestAccount, token, redirectURL)
	if err != nil {
		return err
	}
	subject := "Your Techboss.ai Login Link"

	return Send(context.Background(), "login_link", []string{requestAccount.Email.Get()}, subject, message)
}
