package email_sender_test

import (
	"context"
	"testing"

	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	email_sender "github.com/griffnb/techboss-ai-go/internal/services/email_sender"
)

func init() {
	system_testing.BuildSystem()
}

func TestSend(t *testing.T) {
	if tools.Empty(environment.GetConfig().Email) || tools.Empty(environment.GetConfig().Email.SMTP) ||
		tools.Empty(environment.GetConfig().Email.SMTP.UserName) {
		t.Skip("Email configuration is not set")
	}

	err := email_sender.Send(context.Background(), "verify_email", []string{"nick@atlas.net"}, "Test SMTPEmail", "Test <h1>HTML</h1>")
	if err != nil {
		t.Error(err)
	}
}
