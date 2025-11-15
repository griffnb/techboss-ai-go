package email_sender_test

import (
	"strings"
	"testing"

	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/services/email_sender"
)

func TestBuildVerifyEmail(t *testing.T) {
	accountObj := account.TESTCreateAccount()
	template, err := email_sender.BuildVerifyEmail(accountObj, "abcd")
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(template, "{{") || strings.Contains(template, "}}") {
		t.Fatal("Template not parsed properly")
	}
}

func TestSendVerifyEmail(t *testing.T) {
	if tools.Empty(environment.GetConfig().Email) || tools.Empty(environment.GetConfig().Email.SMTP) ||
		tools.Empty(environment.GetConfig().Email.SMTP.UserName) {
		t.Skip("Email configuration is not set")
	}

	accountObj := account.TESTCreateAccount()
	accountObj.Email.Set("nick+test@gmail.com")

	err := email_sender.SendVerifyEmail(accountObj)
	if err != nil {
		t.Fatal(err)
	}
}
