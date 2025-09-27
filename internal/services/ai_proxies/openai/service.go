package openai

import (
	"context"
	"io"
	"net/http"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/pkg/errors"
)

// Service represents a service layer wrapper for the OpenAI client
type Service struct {
	client *Client
}

// NewService creates a new OpenAI service
func NewService(apiKey string) *Service {
	client := NewClient(apiKey)
	return &Service{
		client: client,
	}
}

// NewServiceFromEnv creates a new OpenAI service using the OPENAI_API_KEY environment variable
func NewServiceFromEnv() (*Service, error) {
	apiKey := environment.GetConfig().AIKeys.OpenAI.APIKey
	if apiKey == "" {
		return nil, errors.New("openai key is required")
	}

	return NewService(apiKey), nil
}

// ProxyNonStreaming proxies a non-streaming request to OpenAI
// This method reads the request body, forwards it to OpenAI, and pipes the response back
func (s *Service) ProxyNonStreaming(ctx context.Context, request *http.Request, responseWriter http.ResponseWriter) error {
	// Read the request body
	requestBody, err := io.ReadAll(request.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read request body")
	}
	defer func() {
		if closeErr := request.Body.Close(); closeErr != nil {
			log.ErrorContext(closeErr, ctx)
		}
	}()

	// Proxy the request
	return s.client.ProxyRequest(ctx, requestBody, responseWriter)
}

// ProxyStreaming proxies a streaming request to OpenAI
// This method reads the request body, forwards it to OpenAI with stream=true, and pipes the SSE response back
func (s *Service) ProxyStreaming(ctx context.Context, request *http.Request, responseWriter http.ResponseWriter) error {
	// Read the request body
	requestBody, err := io.ReadAll(request.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read request body")
	}
	defer func() {
		if closeErr := request.Body.Close(); closeErr != nil {
			log.ErrorContext(closeErr, ctx)
		}
	}()

	// Proxy the streaming request
	return s.client.ProxyStreamRequest(ctx, requestBody, responseWriter)
}

// GetClient returns the underlying client for advanced use cases
func (s *Service) GetClient() *Client {
	return s.client
}
