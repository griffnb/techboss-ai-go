package openai

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
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

func TestProxyRequest_MockOpenAI(t *testing.T) {
	// Create a mock OpenAI server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.URL.Path != "/responses" {
			t.Errorf("Expected /responses path, got %s", r.URL.Path)
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-key" {
			t.Errorf("Expected Authorization header 'Bearer test-key', got %s", authHeader)
		}

		// Send a mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"id":"test","object":"chat.completion","choices":[{"message":{"content":"Hello"}}]}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer mockServer.Close()

	// Create client with mock server URL
	client := NewClient("test-key").WithBaseURL(mockServer.URL)

	// Test data
	requestBody := []byte(`{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"Hello"}]}`)

	// Create response recorder
	responseRecorder := httptest.NewRecorder()

	// Make the request
	err := client.ProxyRequest(context.Background(), requestBody, responseRecorder)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check response
	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	expectedBody := `{"id":"test","object":"chat.completion","choices":[{"message":{"content":"Hello"}}]}`
	if responseRecorder.Body.String() != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, responseRecorder.Body.String())
	}
}

func TestProxyStreamRequest_MockOpenAI(t *testing.T) {
	// Create a mock OpenAI server for streaming
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Send a mock streaming response
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("ResponseWriter doesn't support flushing")
		}

		// Send some mock SSE data
		if _, err := w.Write([]byte("data: {\"id\":\"test\",\"object\":\"chat.completion.chunk\",\"choices\":[{\"delta\":{\"content\":\"Hello\"}}]}\n")); err != nil {
			t.Errorf("Failed to write SSE data: %v", err)
		}
		flusher.Flush()

		if _, err := w.Write([]byte("data: [DONE]\n")); err != nil {
			t.Errorf("Failed to write DONE marker: %v", err)
		}
		flusher.Flush()
	}))
	defer mockServer.Close()

	// Create client with mock server URL
	client := NewClient("test-key").WithBaseURL(mockServer.URL)

	// Test data (stream will be added automatically)
	requestBody := []byte(`{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"Hello"}]}`)

	// Create response recorder
	responseRecorder := httptest.NewRecorder()

	// Make the request
	err := client.ProxyStreamRequest(context.Background(), requestBody, responseRecorder)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check response
	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, responseRecorder.Code)
	}

	contentType := responseRecorder.Header().Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("Expected Content-Type 'text/event-stream', got %s", contentType)
	}

	// Check that the response contains the expected streaming data
	responseBody := responseRecorder.Body.String()
	if !bytes.Contains(responseRecorder.Body.Bytes(), []byte("data: {")) {
		t.Errorf("Expected streaming response to contain SSE data, got: %s", responseBody)
	}
}
