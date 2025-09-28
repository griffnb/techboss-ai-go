// Package clerk integration
// https://pkg.go.dev/github.com/clerk/clerk-sdk-go/v2
package clerk

import (
	"sync"

	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/CrowdShield/go-core/lib/tools/api_client"
	"github.com/griffnb/techboss-ai-go/internal/environment"
)

var (
	instance *APIClient
	once     sync.Once
)

func Client() *APIClient {
	once.Do(func() {
		if !tools.Empty(environment.GetConfig().Clerk) && !tools.Empty(environment.GetConfig().Clerk.APIKey) {
			apiConfig := environment.GetConfig().Clerk
			instance = NewClient(apiConfig)
		}
	})
	return instance
}

func Configured() bool {
	return !tools.Empty(environment.GetConfig().Sendpulse) && !tools.Empty(environment.GetConfig().Sendpulse.ClientSecret)
}

type APIClient struct {
	api_client.APIClient
	apiKey string
}

// NewClient creates a new instance of Client
func NewClient(config *environment.Clerk) *APIClient {
	client := &APIClient{
		apiKey: config.APIKey,
	}

	return client
}
