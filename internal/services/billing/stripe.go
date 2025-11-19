package billing

import (
	"sync"

	"github.com/griffnb/core/lib/stripe_wrapper"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/environment"
)

var (
	instance *stripe_wrapper.APIClient
	once     sync.Once
)

func Client() *stripe_wrapper.APIClient {
	once.Do(func() {
		if !tools.Empty(environment.GetConfig().Stripe) && !tools.Empty(environment.GetConfig().Stripe.SecretKey) {
			apiConfig := environment.GetConfig().Stripe
			instance = stripe_wrapper.NewClient(apiConfig.SecretKey)
		}
	})
	return instance
}

func Configured() bool {
	return !tools.Empty(environment.GetConfig().Stripe) && !tools.Empty(environment.GetConfig().Stripe.SecretKey)
}
