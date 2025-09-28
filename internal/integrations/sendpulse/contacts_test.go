package sendpulse_test

import (
	"context"
	"testing"

	"github.com/griffnb/techboss-ai-go/internal/integrations/sendpulse"
)

func TestAPIClient_GetContactList(t *testing.T) {
	if !sendpulse.Configured() {
		t.Skip("Sendpulse API Key not set")
	}

	client := sendpulse.Client()

	resp, err := client.GetContactList(context.Background(), &sendpulse.ContactListBody{})
	if err != nil {
		t.Fatal(err)
	}

	if resp == nil || resp.Data == nil {
		t.Fatal("No response")
	}
}
