package sendpulse_test

import (
	"context"
	"testing"

	"github.com/griffnb/techboss-ai-go/internal/integrations/sendpulse"
)

func TestAPIClient_AddContactPhone(t *testing.T) {
	if !sendpulse.Configured() {
		t.Skip("Sendpulse API Key not set")
	}

	client := sendpulse.Client()

	// First create a contact to add phone to
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
		t.Fatal("Failed to create contact for phone test")
	}

	contactID := int64(createResp.Data.ID)

	// Now add phone to the contact
	phoneRequest := &sendpulse.AddContactPhoneRequest{
		Phone: "+1234567890",
	}

	resp, err := client.AddContactPhone(context.Background(), contactID, phoneRequest)
	if err != nil {
		t.Fatal(err)
	}

	if resp == nil || resp.Data == nil {
		t.Fatal("No phone data returned")
	}

	if resp.Data.Phone != phoneRequest.Phone {
		t.Errorf("Expected phone %s, got %s", phoneRequest.Phone, resp.Data.Phone)
	}
}

func TestAPIClient_UpdateContactPhone(t *testing.T) {
	if !sendpulse.Configured() {
		t.Skip("Sendpulse API Key not set")
	}

	client := sendpulse.Client()

	// First create a contact to update phone
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
		t.Fatal("Failed to create contact for phone test")
	}

	contactID := int64(createResp.Data.ID)

	// Add a phone first
	phoneRequest := &sendpulse.AddContactPhoneRequest{
		Phone: "+1234567890",
	}

	addResp, err := client.AddContactPhone(context.Background(), contactID, phoneRequest)
	if err != nil {
		t.Fatal(err)
	}

	if addResp == nil || addResp.Data == nil {
		t.Fatal("Failed to add phone for update test")
	}

	phoneID := int64(addResp.Data.ID)

	// Now update the phone
	updateRequest := &sendpulse.UpdateContactPhoneRequest{
		Phone: "+9876543210",
	}

	resp, err := client.UpdateContactPhone(context.Background(), contactID, phoneID, updateRequest)
	if err != nil {
		t.Fatal(err)
	}

	if resp == nil || resp.Data == nil {
		t.Fatal("No response data")
	}

	if resp.Data.Phone != updateRequest.Phone {
		t.Errorf("Expected phone %s, got %s", updateRequest.Phone, resp.Data.Phone)
	}
}

func TestAPIClient_DeleteContactPhone(t *testing.T) {
	if !sendpulse.Configured() {
		t.Skip("Sendpulse API Key not set")
	}

	client := sendpulse.Client()

	// First create a contact to delete phone from
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
		t.Fatal("Failed to create contact for phone test")
	}

	contactID := int64(createResp.Data.ID)

	// Add a phone first
	phoneRequest := &sendpulse.AddContactPhoneRequest{
		Phone: "+1234567890",
	}

	addResp, err := client.AddContactPhone(context.Background(), contactID, phoneRequest)
	if err != nil {
		t.Fatal(err)
	}

	if addResp == nil || addResp.Data == nil {
		t.Fatal("Failed to add phone for delete test")
	}

	phoneID := int64(addResp.Data.ID)

	// Now delete the phone
	err = client.DeleteContactPhone(context.Background(), contactID, phoneID)
	if err != nil {
		t.Fatal(err)
	}

	// Success if no error returned
}
