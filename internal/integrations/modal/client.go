// Package modal provides integration with Modal Labs for sandboxed code execution.
// It enables creating isolated sandbox environments with Docker images, persistent volumes,
// S3 storage mounts, and Claude Code CLI execution. The package supports multi-tenant
// architecture with account-scoped resources and timestamp-based versioning for S3 storage.
//
//go:generate core_gen mock APIClient
package modal

import (
	"log"
	"sync"

	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/modal-labs/libmodal/modal-go"
)

var (
	instance *APIClient
	once     sync.Once
)

// Client returns the singleton Modal API client instance.
// It initializes the client on first call if Modal credentials are configured.
// Returns nil if Modal is not configured.
func Client() *APIClient {
	once.Do(func() {
		if !tools.Empty(environment.GetConfig().Modal) && !tools.Empty(environment.GetConfig().Modal.TokenSecret) {
			apiConfig := environment.GetConfig().Modal
			instance = NewClient(apiConfig)
		}
	})
	return instance
}

// Configured checks if Modal is properly configured.
// Returns true if both TokenID and TokenSecret are present in environment config.
func Configured() bool {
	return !tools.Empty(environment.GetConfig().Modal) && !tools.Empty(environment.GetConfig().Modal.TokenSecret)
}

// APIClient provides methods for interacting with Modal Labs API.
// It encapsulates sandbox creation, Claude Code execution, and S3 storage operations.
// The client is thread-safe and should be accessed via the Client() singleton.
type APIClient struct {
	tokenID     string
	tokenSecret string
	client      *modal.Client
}

// NewClient creates a new Modal API client with the given configuration.
// This function is typically called by the Client() singleton and should not be
// called directly unless you need multiple client instances.
func NewClient(config *environment.Modal) *APIClient {
	client := &APIClient{
		tokenID:     config.TokenID,
		tokenSecret: config.TokenSecret,
	}

	mc, err := modal.NewClientWithOptions(&modal.ClientParams{
		TokenID:     client.tokenID,
		TokenSecret: client.tokenSecret,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	client.client = mc

	return client
}
