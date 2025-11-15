package cloudflare

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/CrowdShield/go-core/lib/tools/api_client"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/pkg/errors"
)

var (
	instance *APIClient
	once     sync.Once
)

func Client() *APIClient {
	once.Do(func() {
		if !tools.Empty(environment.GetConfig().Cloudflare) && !tools.Empty(environment.GetConfig().Cloudflare.TurnstileKey) {
			apiConfig := environment.GetConfig().Cloudflare
			instance = NewClient(apiConfig)
		}
	})
	return instance
}

func Configured() bool {
	return !tools.Empty(environment.GetConfig().Cloudflare) && !tools.Empty(environment.GetConfig().Cloudflare.TurnstileKey)
}

type APIClient struct {
	api_client.APIClient
	TurnstileKey string
}

// NewClient creates a new instance of Client
func NewClient(apiConfig *environment.Cloudflare) *APIClient {
	client := &APIClient{
		TurnstileKey: apiConfig.TurnstileKey,
	}

	client.BaseURL = fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s", apiConfig.AccountID)
	client.SuccessCodes = []int{200}
	client.DefaultHeaders = map[string]string{
		"Authorization": "Bearer " + apiConfig.APIKey,
		"Content-Type":  "application/json",
	}
	client.Client = &http.Client{
		Timeout: 10 * time.Second,
	}
	return client
}

// BuildRequestWithOptions creates an API request with a JSON body for browser rendering endpoints
func (this *APIClient) BuildRequestWithOptions(method, path string, options interface{}) (*api_client.Request, error) {
	request := this.NewRequest(method, path)
	if options != nil {
		bodyBytes, err := json.Marshal(options)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		request.WithBody(bodyBytes)
	}
	return request, nil
}
