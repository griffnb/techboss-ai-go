package openai

import (
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	apiKey := "test-api-key"
	client := NewClient(apiKey)

	if client.apiKey != apiKey {
		t.Errorf("Expected API key %s, got %s", apiKey, client.apiKey)
	}

	if client.baseURL != BaseURL {
		t.Errorf("Expected base URL %s, got %s", BaseURL, client.baseURL)
	}

	if client.httpClient.Timeout != DefaultTimeout {
		t.Errorf("Expected timeout %v, got %v", DefaultTimeout, client.httpClient.Timeout)
	}
}

func TestClientWithBaseURL(t *testing.T) {
	client := NewClient("test-key")
	customURL := "https://custom.example.com"

	client = client.WithBaseURL(customURL)

	if client.baseURL != customURL {
		t.Errorf("Expected base URL %s, got %s", customURL, client.baseURL)
	}
}

func TestClientWithTimeout(t *testing.T) {
	client := NewClient("test-key")
	customTimeout := 10 * time.Second

	client = client.WithTimeout(customTimeout)

	if client.httpClient.Timeout != customTimeout {
		t.Errorf("Expected timeout %v, got %v", customTimeout, client.httpClient.Timeout)
	}
}

func TestNewService(t *testing.T) {
	apiKey := "test-api-key"
	service := NewService(apiKey)

	if service.client.apiKey != apiKey {
		t.Errorf("Expected API key %s, got %s", apiKey, service.client.apiKey)
	}
}
