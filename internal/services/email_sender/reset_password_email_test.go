package email_sender_test

import (
	"strings"
	"testing"

	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/services/email_sender"
)

func TestBuildResetPasswordEmail(t *testing.T) {
	accountObj := account.TESTCreateAccount()
	template, err := email_sender.BuildResetPasswordEmail(accountObj, "abcd")
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(template, "{{") || strings.Contains(template, "}}") {
		t.Fatal("Template not parsed properly")
	}
}

func TestSendResetPasswordEmail(t *testing.T) {
	if tools.Empty(environment.GetConfig().Email) || tools.Empty(environment.GetConfig().Email.SMTP) ||
		tools.Empty(environment.GetConfig().Email.SMTP.UserName) {
		t.Skip("Email configuration is not set")
	}

	accountObj := account.TESTCreateAccount()

	err := email_sender.SendResetPasswordEmail(accountObj, "reset-token")
	if err != nil {
		t.Fatal(err)
	}
}
