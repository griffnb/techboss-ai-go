package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const (
	// OpenAI API base URL
	BaseURL = "https://api.openai.com/v1"

	// Default timeout for HTTP requests
	DefaultTimeout = 30 * time.Second

	// Content types
	ContentTypeJSON = "application/json"
	ContentTypeSSE  = "text/event-stream"
)

// Client represents an OpenAI API proxy client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new OpenAI proxy client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: BaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// WithBaseURL allows setting a custom base URL (useful for testing)
func (c *Client) WithBaseURL(baseURL string) *Client {
	c.baseURL = baseURL
	return c
}

// WithTimeout allows setting a custom timeout
func (c *Client) WithTimeout(timeout time.Duration) *Client {
	c.httpClient.Timeout = timeout
	return c
}

// ProxyRequest proxies a request to OpenAI API without streaming
func (c *Client) ProxyRequest(ctx context.Context, requestBody []byte, responseWriter http.ResponseWriter) error {
	// Create the request to OpenAI
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(requestBody))
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	// Set headers
	req.Header.Set("Content-Type", ContentTypeJSON)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to make request to OpenAI")
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			err = errors.Wrap(closeErr, "failed to close response body")
		}
	}()

	// Copy status code
	responseWriter.WriteHeader(resp.StatusCode)

	// Copy headers
	for key, values := range resp.Header {
		for _, value := range values {
			responseWriter.Header().Add(key, value)
		}
	}

	// Copy response body
	_, err = io.Copy(responseWriter, resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to copy response body")
	}

	return nil
}

// ProxyStreamRequest proxies a streaming request to OpenAI API
func (c *Client) ProxyStreamRequest(ctx context.Context, requestBody []byte, responseWriter http.ResponseWriter) error {
	// Parse the request body to add stream: true
	var requestData map[string]interface{}
	if err := json.Unmarshal(requestBody, &requestData); err != nil {
		return errors.Wrap(err, "failed to parse request body")
	}

	// Ensure stream is set to true
	requestData["stream"] = true

	// Re-marshal the request body
	modifiedRequestBody, err := json.Marshal(requestData)
	if err != nil {
		return errors.Wrap(err, "failed to marshal modified request body")
	}

	// Create the request to OpenAI
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(modifiedRequestBody))
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	// Set headers
	req.Header.Set("Content-Type", ContentTypeJSON)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to make request to OpenAI")
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			err = errors.Wrap(closeErr, "failed to close response body")
		}
	}()

	// Set headers for SSE response
	responseWriter.Header().Set("Content-Type", ContentTypeSSE)
	responseWriter.Header().Set("Cache-Control", "no-cache")
	responseWriter.Header().Set("Connection", "keep-alive")
	responseWriter.Header().Set("Access-Control-Allow-Origin", "*")
	responseWriter.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// If OpenAI returned an error, just return the status code
	if resp.StatusCode != http.StatusOK {
		responseWriter.WriteHeader(resp.StatusCode)
		_, err = io.Copy(responseWriter, resp.Body)
		return errors.Wrap(err, "failed to copy error response")
	}

	// Stream the response
	flusher, ok := responseWriter.(http.Flusher)
	if !ok {
		return errors.New("response writer does not support flushing")
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// Write the line to the response
		_, err := fmt.Fprintf(responseWriter, "%s\n", line)
		if err != nil {
			return errors.Wrap(err, "failed to write streaming response")
		}

		// Flush the response
		flusher.Flush()

		// Check if the stream is done
		if line == "data: [DONE]" {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return errors.Wrap(err, "error reading streaming response")
	}

	return nil
}
