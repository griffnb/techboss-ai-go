package subscriptions

import (
	"net/http"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/models/billing_plan"
	"github.com/griffnb/techboss-ai-go/internal/models/billing_plan_price"
	"github.com/griffnb/techboss-ai-go/internal/models/subscription"
)

type CurrentSubscriptionResponse struct {
	Subscription     *subscription.Subscription           `json:"subscription"`
	BillingPlan      *billing_plan.BillingPlan            `json:"billing_plan"`
	BillingPlanPrice *billing_plan_price.BillingPlanPrice `json:"billing_plan_price"`
}

// @link {models}/src/models/subscription/services/_subscription.ts:currentSubscription
func authCurrent(_ http.ResponseWriter, req *http.Request) (*CurrentSubscriptionResponse, int, error) {
	accountObj := helpers.GetLoadedUser(req)

	subObj, err := subscription.GetActiveByOrganizationID(req.Context(), accountObj.OrganizationID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*CurrentSubscriptionResponse]()

	}

	billingPlanPrice, err := billing_plan_price.Get(req.Context(), subObj.BillingPlanPriceID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*CurrentSubscriptionResponse]()
	}

	billingPlan, err := billing_plan.Get(req.Context(), billingPlanPrice.BillingPlanID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*CurrentSubscriptionResponse]()
	}

	return response.Success(&CurrentSubscriptionResponse{
		Subscription:     subObj,
		BillingPlan:      billingPlan,
		BillingPlanPrice: billingPlanPrice,
	})
}
