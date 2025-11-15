package email_sender

import (
	"context"
	"fmt"
	"html/template"
	"strings"

	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/pkg/errors"
)

type PasswordWasResetEmailTemplate struct {
	Name  string
	Email string
	Logo  string
}

func BuildPasswordWasResetEmail(requestAccount *account.Account) (string, error) {
	data := PasswordWasResetEmailTemplate{
		Name:  fmt.Sprintf("%s %s", requestAccount.FirstName.Get(), requestAccount.LastName.Get()),
		Email: requestAccount.Email.Get(),
		Logo:  "https://app.atlas.net/img/logo.png",
	}

	emailTemplate, err := GetEmailTemplate("PasswordWasReset.html")
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

	return bufBody.String(), nil
}

func SendPasswordWasResetEmail(requestAccount *account.Account) error {
	message, err := BuildPasswordWasResetEmail(requestAccount)
	if err != nil {
		return err
	}
	subject := "Your Techboss.ai Password Was Reset"

	return Send(context.Background(), "reset_password", []string{requestAccount.Email.Get()}, subject, message)
}
