package email_sender

import (
	"context"
	"embed"
	"strings"

	"github.com/griffnb/core/lib/email"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/environment"

	"github.com/pkg/errors"
)

const REPLY_TO = "support@techboss.ai"

//go:embed raw_templates/*
var emailTemplates embed.FS

func GetEmailTemplate(filename string) (string, error) {
	content, err := emailTemplates.ReadFile("raw_templates/" + filename)
	if err != nil {
		return "", errors.WithMessagef(err, "failed to load template %s", filename)
	}
	return string(content), nil
}

type MessageID string

func Send(ctx context.Context, messageID MessageID, to []string, subject, message string) error {
	// Email isnt Setup
	if tools.Empty(environment.GetConfig().Email) {
		if environment.IsUnitTest() || environment.IsLocalDev() {
			return nil
		}
		return errors.Errorf("email not setup")
	}

	from := environment.GetConfig().Email.From

	req := email.EmailRequest{
		To:        to,
		From:      from,
		ReplyTo:   REPLY_TO,
		Subject:   subject,
		Body:      message,
		MessageID: string(messageID),
	}

	if !tools.Empty(environment.GetConfig().Email.DevEmail) {
		req.To = []string{environment.GetConfig().Email.DevEmail}
		req.Subject = "[DEV] " + req.Subject
	}

	if !environment.IsProduction() {
		if !strings.HasSuffix(to[0], "@techboss.ai") {
			return errors.Errorf("cant send emails to non techboss.ai addresses in non-production environments")
		}
	}
	/*
		// If the generic emailer isnt set, use the CustomerIO client
		if environment.GetEmailer() == nil {
			return customerio.Client().SendEmail(ctx, req)
		}
	*/

	return environment.GetEmailer().SendEmail(ctx, req)
}

/*
func SendCustomerio(ctx context.Context, messageID MessageID, to []string, messageData map[string]any) error {
	// Email isnt Setup
	if tools.Empty(environment.GetConfig().Email) {
		if environment.IsUnitTest() {
			return nil
		}
		return errors.Errorf("email not setup")
	}

	from := environment.GetConfig().Email.From

	if !environment.IsProduction() {
		if !strings.HasSuffix(to[0], "@techboss.ai") {
			return errors.Errorf("cant send emails to non techboss.ai addresses in non-production environments")
		}
	}

	req := email.EmailRequest{
		To:        to,
		From:      from,
		ReplyTo:   REPLY_TO,
		MessageID: string(messageID),
	}

	if !tools.Empty(environment.GetConfig().Email.DevEmail) {
		req.To = []string{environment.GetConfig().Email.DevEmail}
	}

	return customerio.Client().SendTransactionEmailWithTemplate(ctx, req, messageData)
}
*/
