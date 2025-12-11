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

func Client() *APIClient {
	once.Do(func() {
		if !tools.Empty(environment.GetConfig().Modal) && !tools.Empty(environment.GetConfig().Modal.TokenSecret) {
			apiConfig := environment.GetConfig().Modal
			instance = NewClient(apiConfig)
		}
	})
	return instance
}

func Configured() bool {
	return !tools.Empty(environment.GetConfig().Sendpulse) && !tools.Empty(environment.GetConfig().Sendpulse.ClientSecret)
}

type APIClient struct {
	tokenID     string
	tokenSecret string
	client      *modal.Client
}

// NewClient creates a new instance of Client
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
