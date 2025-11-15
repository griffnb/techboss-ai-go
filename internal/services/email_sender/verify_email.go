package email_sender

import (
	"context"
	"fmt"
	"html/template"
	"net/url"
	"strings"
	"time"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/session"

	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/account"

	"github.com/pkg/errors"
)

type VerifyEmailTemplate struct {
	Logo  string
	Name  string
	Link  string
	Email string
}

func BuildVerifyEmail(requestAccount *account.Account, verifyToken string) (string, error) {
	baseURL := environment.GetConfig().Server.AppURL
	params := url.Values{}
	params.Add("verify", verifyToken)
	params.Add("email", requestAccount.Email.Get())
	link := fmt.Sprintf("%s/verify?%s", baseURL, params.Encode())

	data := VerifyEmailTemplate{
		Name:  fmt.Sprintf("%s %s", requestAccount.FirstName.Get(), requestAccount.LastName.Get()),
		Link:  link,
		Email: requestAccount.Email.Get(),
		Logo:  "https://assettradingdesk.com/img/logo.png",
	}

	emailTemplate, err := GetEmailTemplate("EmailVerification.html")
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

	log.Debug("Verify Email Link: ", link)

	return bufBody.String(), nil
}

func sendVerifyEmail(emailAddress string, requestAccount *account.Account, verifyToken string) error {
	message, err := BuildVerifyEmail(requestAccount, verifyToken)
	if err != nil {
		return err
	}
	subject := "Verify Email Address"

	return Send(context.Background(), "verify_email", []string{emailAddress}, subject, message)
}

func SendVerifyEmail(accountObj *account.Account) error {
	emailAddress := accountObj.Email.Get()

	verifySession := session.New("").WithUser(accountObj)
	err := verifySession.SaveWithExpiration(time.Now().Add(48 * time.Hour).Unix())
	if err != nil {
		return errors.Wrap(err, "failed to save verify session")
	}

	err = sendVerifyEmail(emailAddress, accountObj, verifySession.Key)
	if err != nil {
		return errors.Wrap(err, "failed to send verify email")
	}

	props, err := accountObj.Properties.Get()
	if err != nil {
		return errors.Wrap(err, "failed to get account properties")
	}

	props.VerifyEmailKey = verifySession.Key
	accountObj.Properties.Set(props)
	err = accountObj.Save(nil)
	if err != nil {
		return errors.Wrap(err, "failed to save account properties with verify email key")
	}

	return nil
}
