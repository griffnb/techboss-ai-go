package email_sender

import (
	"context"
	"fmt"
	"html/template"
	"strings"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/account"

	"github.com/pkg/errors"
)

type ResetPasswordEmailTemplate struct {
	Name  string
	Link  string
	Email string
	Logo  string
}

func BuildResetPasswordEmail(requestAccount *account.Account, resetToken string) (string, error) {
	baseURL := environment.GetConfig().Server.AppURL
	link := fmt.Sprintf("%s/password/reset?resetKey=%s", baseURL, resetToken)

	data := ResetPasswordEmailTemplate{
		Name:  fmt.Sprintf("%s %s", requestAccount.FirstName.Get(), requestAccount.LastName.Get()),
		Link:  link,
		Email: requestAccount.Email.Get(),
		Logo:  "https://app.atlas.net/img/logo.png",
	}

	emailTemplate, err := GetEmailTemplate("ResetPassword.html")
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

	log.Debug("Reset Password Link: ", link)

	return bufBody.String(), nil
}

func SendResetPasswordEmail(requestAccount *account.Account, resetToken string) error {
	message, err := BuildResetPasswordEmail(requestAccount, resetToken)
	if err != nil {
		return err
	}
	subject := "Reset Password"

	return Send(context.Background(), "reset_password", []string{requestAccount.Email.Get()}, subject, message)
}
