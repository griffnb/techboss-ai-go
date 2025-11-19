package sendpulse_test

import (
	"context"
	"testing"

	"github.com/griffnb/techboss-ai-go/internal/integrations/sendpulse"
)

func TestAPIClient_AddContactMessenger(t *testing.T) {
	if !sendpulse.Configured() {
		t.Skip("SendPulse not configured")
	}

	ctx := context.Background()
	client := sendpulse.Client()

	// Test with valid messenger data
	request := &sendpulse.AddContactMessengerRequest{
		TypeID:        1, // Telegram
		Login:         "@testuser",
		IsMainChatbot: true,
	}

	response, err := client.AddContactMessenger(ctx, 1234567, request)
	if err != nil {
		t.Errorf("AddContactMessenger failed: %v", err)
	}
	if response == nil {
		t.Error("AddContactMessenger returned nil response")
	}

	// Test with invalid contact ID
	_, err = client.AddContactMessenger(ctx, -1, request)
	if err == nil {
		t.Error("AddContactMessenger should fail with invalid contact ID")
	}
}

func TestAPIClient_UpdateContactMessenger(t *testing.T) {
	if !sendpulse.Configured() {
		t.Skip("SendPulse not configured")
	}

	ctx := context.Background()
	client := sendpulse.Client()

	// Test with valid messenger data
	request := &sendpulse.UpdateContactMessengerRequest{
		TypeID:        1, // Telegram
		Login:         "@updateduser",
		IsMainChatbot: false,
	}

	response, err := client.UpdateContactMessenger(ctx, 1234567, 9876543, request)
	if err != nil {
		t.Errorf("UpdateContactMessenger failed: %v", err)
	}
	if response == nil {
		t.Error("UpdateContactMessenger returned nil response")
	}

	// Test with invalid contact ID
	_, err = client.UpdateContactMessenger(ctx, -1, 9876543, request)
	if err == nil {
		t.Error("UpdateContactMessenger should fail with invalid contact ID")
	}

	// Test with invalid messenger ID
	_, err = client.UpdateContactMessenger(ctx, 1234567, -1, request)
	if err == nil {
		t.Error("UpdateContactMessenger should fail with invalid messenger ID")
	}
}

func TestAPIClient_DeleteContactMessengers(t *testing.T) {
	if !sendpulse.Configured() {
		t.Skip("SendPulse not configured")
	}

	ctx := context.Background()
	client := sendpulse.Client()

	// Test with valid IDs
	err := client.DeleteContactMessengers(ctx, 1234567, 9876543)
	if err != nil {
		t.Errorf("DeleteContactMessengers failed: %v", err)
	}

	// Test with invalid contact ID
	err = client.DeleteContactMessengers(ctx, -1, 9876543)
	if err == nil {
		t.Error("DeleteContactMessengers should fail with invalid contact ID")
	}

	// Test with invalid messenger ID
	err = client.DeleteContactMessengers(ctx, 1234567, -1)
	if err == nil {
		t.Error("DeleteContactMessengers should fail with invalid messenger ID")
	}
}
