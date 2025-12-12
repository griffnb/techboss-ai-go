package billing

import (
	"context"
	"fmt"
	"net/http"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/stripe_wrapper"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/services/billing"
)

// Test by activating the local cli to reroute, be sure to set the webhook key from the client
func openStripeHook(res http.ResponseWriter, req *http.Request) {
	// store event with event ID as the PK

	webHookKey := environment.GetConfig().Stripe.WebhookKey

	event, err := stripe_wrapper.ParseWebhookRequest(res, req, webHookKey)
	if err != nil {
		log.Error(err)
		response.ErrorWrapper(res, req, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("\n\n--------Received stripe webhook event %s------------\n\n", event.Type)

	bgContext := context.WithoutCancel(req.Context())
	go func() {
		err := billing.ProcessStripeEvent(bgContext, event)
		if err != nil {
			log.ErrorContext(err, bgContext)
			return
		}
	}()

	response.JSONDataResponseWrapper(res, req, "success")
}
