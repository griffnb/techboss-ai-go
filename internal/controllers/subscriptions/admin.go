package subscriptions

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/models/subscription"
	"github.com/griffnb/techboss-ai-go/internal/services/billing"
	"github.com/stripe/stripe-go/v83"
)

func adminDetails(_ http.ResponseWriter, req *http.Request) (*stripe.Subscription, int, error) {
	id := chi.URLParam(req, "id")

	subObj, err := subscription.Get(req.Context(), types.UUID(id))
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*stripe.Subscription](err)
	}

	sub, err := billing.Client().GetSubscriptionByID(req.Context(), subObj.StripeSubscriptionID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*stripe.Subscription](err)
	}
	return response.Success(sub)
}
