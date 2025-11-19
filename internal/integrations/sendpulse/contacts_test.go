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

func TestAPIClient_CreateContact(t *testing.T) {
	if !sendpulse.Configured() {
		t.Skip("Sendpulse API Key not set")
	}

	client := sendpulse.Client()

	request := &sendpulse.CreateContactRequest{
		ResponsibleID: 1, // This should be a valid user ID in your SendPulse account
		FirstName:     "Test",
		LastName:      "Contact",
	}

	resp, err := client.CreateContact(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}

	if resp == nil || resp.Data == nil {
		t.Fatal("No response data")
	}

	if resp.Data.ID == 0 {
		t.Fatal("Contact ID should not be 0")
	}

	if resp.Data.FirstName != request.FirstName {
		t.Errorf("Expected first name %s, got %s", request.FirstName, resp.Data.FirstName)
	}

	if resp.Data.LastName != request.LastName {
		t.Errorf("Expected last name %s, got %s", request.LastName, resp.Data.LastName)
	}
}

func TestAPIClient_UpdateContact(t *testing.T) {
	if !sendpulse.Configured() {
		t.Skip("Sendpulse API Key not set")
	}

	client := sendpulse.Client()

	// First create a contact to update
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
		t.Fatal("Failed to create contact for update test")
	}

	contactID := int64(createResp.Data.ID)

	// Now update the contact
	updateRequest := &sendpulse.UpdateContactRequest{
		ResponsibleID: 1,
		FirstName:     "Updated",
		LastName:      "Contact",
	}

	resp, err := client.UpdateContact(context.Background(), contactID, updateRequest)
	if err != nil {
		t.Fatal(err)
	}

	if resp == nil || resp.Data == nil {
		t.Fatal("No response data")
	}

	if resp.Data.FirstName != updateRequest.FirstName {
		t.Errorf("Expected first name %s, got %s", updateRequest.FirstName, resp.Data.FirstName)
	}

	if resp.Data.LastName != updateRequest.LastName {
		t.Errorf("Expected last name %s, got %s", updateRequest.LastName, resp.Data.LastName)
	}
}

func TestAPIClient_GetContact(t *testing.T) {
	if !sendpulse.Configured() {
		t.Skip("Sendpulse API Key not set")
	}

	client := sendpulse.Client()

	// First create a contact to get
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
		t.Fatal("Failed to create contact for get test")
	}

	contactID := int64(createResp.Data.ID)

	// Now get the contact
	resp, err := client.GetContact(context.Background(), contactID)
	if err != nil {
		t.Fatal(err)
	}

	if resp == nil || resp.Data == nil {
		t.Fatal("No response data")
	}

	if resp.Data.ID != createResp.Data.ID {
		t.Errorf("Expected contact ID %d, got %d", createResp.Data.ID, resp.Data.ID)
	}

	if resp.Data.FirstName != createRequest.FirstName {
		t.Errorf("Expected first name %s, got %s", createRequest.FirstName, resp.Data.FirstName)
	}
}

func TestAPIClient_DeleteContact(t *testing.T) {
	if !sendpulse.Configured() {
		t.Skip("Sendpulse API Key not set")
	}

	client := sendpulse.Client()

	// First create a contact to delete
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
		t.Fatal("Failed to create contact for delete test")
	}

	contactID := int64(createResp.Data.ID)

	// Now delete the contact
	err = client.DeleteContact(context.Background(), contactID)
	if err != nil {
		t.Fatal(err)
	}

	// Verify contact is deleted by trying to get it (should fail)
	_, err = client.GetContact(context.Background(), contactID)
	if err == nil {
		t.Error("Expected error when getting deleted contact, but got none")
	}
}
