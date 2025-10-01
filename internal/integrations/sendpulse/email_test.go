package sendpulse_test

import (
	"context"
	"testing"

	"github.com/griffnb/techboss-ai-go/internal/integrations/sendpulse"
)

func TestAPIClient_AddContactEmail(t *testing.T) {
	if !sendpulse.Configured() {
		t.Skip("Sendpulse API Key not set")
	}

	client := sendpulse.Client()

	// First create a contact to add email to
	createRequest := &sendpulse.CreateContactRequest{
		ResponsibleID: 1,
		FirstName:     "Test",
		LastName:      "Contact",
	}

	createResp, err := client.CreateContact(context.Background(), createRequest)
	if err != nil {
		t.Fatal(err)
	}

	if createResp == nil || createResp.Data == nil {
		t.Fatal("Failed to create contact for email test")
	}

	contactID := int64(createResp.Data.ID)

	// Now add email to the contact
	emailRequest := &sendpulse.AddContactEmailRequest{
		Emails: []sendpulse.EmailRequest{
			{
				Email:  "test@example.com",
				IsMain: true,
			},
		},
	}

	resp, err := client.AddContactEmail(context.Background(), contactID, emailRequest)
	if err != nil {
		t.Fatal(err)
	}

	if resp == nil || resp.Data == nil || len(resp.Data) == 0 {
		t.Fatal("No email data returned")
	}

	if resp.Data[0].Email != emailRequest.Emails[0].Email {
		t.Errorf("Expected email %s, got %s", emailRequest.Emails[0].Email, resp.Data[0].Email)
	}
}

func TestAPIClient_UpdateContactEmail(t *testing.T) {
	if !sendpulse.Configured() {
		t.Skip("Sendpulse API Key not set")
	}

	client := sendpulse.Client()

	// First create a contact to update email
	createRequest := &sendpulse.CreateContactRequest{
		ResponsibleID: 1,
		FirstName:     "Test",
		LastName:      "Contact",
	}

	createResp, err := client.CreateContact(context.Background(), createRequest)
	if err != nil {
		t.Fatal(err)
	}

	if createResp == nil || createResp.Data == nil {
		t.Fatal("Failed to create contact for email test")
	}

	contactID := int64(createResp.Data.ID)

	// Add an email first
	emailRequest := &sendpulse.AddContactEmailRequest{
		Emails: []sendpulse.EmailRequest{
			{
				Email:  "test@example.com",
				IsMain: true,
			},
		},
	}

	addResp, err := client.AddContactEmail(context.Background(), contactID, emailRequest)
	if err != nil {
		t.Fatal(err)
	}

	if addResp == nil || addResp.Data == nil || len(addResp.Data) == 0 {
		t.Fatal("Failed to add email for update test")
	}

	emailID := int64(addResp.Data[0].ID)

	// Now update the email
	updateRequest := &sendpulse.UpdateContactEmailRequest{
		Email: "updated@example.com",
	}

	resp, err := client.UpdateContactEmail(context.Background(), contactID, emailID, updateRequest)
	if err != nil {
		t.Fatal(err)
	}

	if resp == nil || resp.Data == nil {
		t.Fatal("No response data")
	}

	if resp.Data.Email != updateRequest.Email {
		t.Errorf("Expected email %s, got %s", updateRequest.Email, resp.Data.Email)
	}
}

func TestAPIClient_DeleteContactEmail(t *testing.T) {
	if !sendpulse.Configured() {
		t.Skip("Sendpulse API Key not set")
	}

	client := sendpulse.Client()

	// First create a contact to delete email from
	createRequest := &sendpulse.CreateContactRequest{
		ResponsibleID: 1,
		FirstName:     "Test",
		LastName:      "Contact",
	}

	createResp, err := client.CreateContact(context.Background(), createRequest)
	if err != nil {
		t.Fatal(err)
	}

	if createResp == nil || createResp.Data == nil {
		t.Fatal("Failed to create contact for email test")
	}

	contactID := int64(createResp.Data.ID)

	// Add an email first
	emailRequest := &sendpulse.AddContactEmailRequest{
		Emails: []sendpulse.EmailRequest{
			{
				Email:  "test@example.com",
				IsMain: true,
			},
		},
	}

	addResp, err := client.AddContactEmail(context.Background(), contactID, emailRequest)
	if err != nil {
		t.Fatal(err)
	}

	if addResp == nil || addResp.Data == nil || len(addResp.Data) == 0 {
		t.Fatal("Failed to add email for delete test")
	}

	emailID := int64(addResp.Data[0].ID)

	// Now delete the email
	err = client.DeleteContactEmail(context.Background(), contactID, emailID)
	if err != nil {
		t.Fatal(err)
	}

	// Success if no error returned
}
