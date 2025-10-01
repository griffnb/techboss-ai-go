// Package sendpulse integration
// https://sendpulse.com/integrations/api
// https://sendpulse.com/integrations/api/crm
// https://login.sendpulse.com/api/crm-service/v1/openapi/en
package sendpulse

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"

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
		if !tools.Empty(environment.GetConfig().Sendpulse) && !tools.Empty(environment.GetConfig().Sendpulse.ClientSecret) {
			apiConfig := environment.GetConfig().Sendpulse
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
	clientID     string
	clientSecret string
	WebhookKey   string
	token        string
	expires      int64
}

// NewClient creates a new instance of Client
func NewClient(config *environment.Sendpulse) *APIClient {
	client := &APIClient{
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		WebhookKey:   config.WebhookKey,
	}
	client.Client = &http.Client{
		Timeout: 10 * time.Second,
	}
	client.BaseURL = "https://api.sendpulse.com"
	client.SuccessCodes = []int{200, 201, 202}

	return client
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

func (this *APIClient) GetToken() (string, error) {
	if this.token != "" && time.Now().Unix() < this.expires-60 {
		return this.token, nil
	}

	result := &tokenResponse{}

	formData := url.Values{}
	formData.Set("grant_type", "client_credentials")
	formData.Set("client_id", this.clientID)
	formData.Set("client_secret", this.clientSecret)

	bodyBytes := []byte(formData.Encode())

	req := this.NewRequest(http.MethodPost, "/oauth/access_token").
		WithBody(bodyBytes).
		WithSuccessResult(result).WithHeader("Content-Type", "application/x-www-form-urlencoded")

	_, err := this.Call(context.TODO(),
		req,
	)
	if err != nil {
		return "", err
	}

	this.expires = time.Now().Unix() + result.ExpiresIn
	this.token = result.AccessToken

	return this.token, nil
}

type ErrorsResponse struct {
	Errors []string `json:"errors"`
}

type Response[T any] struct {
	Data T `json:"data,omitempty"`
}
