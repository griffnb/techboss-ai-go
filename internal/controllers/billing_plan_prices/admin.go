package billing_plan_prices

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/models/billing_plan"
	"github.com/griffnb/techboss-ai-go/internal/models/billing_plan_price"
	"github.com/griffnb/techboss-ai-go/internal/services/billing"
	"github.com/pkg/errors"
)

func adminCreate(_ http.ResponseWriter, req *http.Request) (*billing_plan_price.BillingPlanPrice, int, error) {
	userSession := request.GetReqSession(req)
	data := request.GetModelPostData(req)
	billingPlanPriceObj := billing_plan_price.New()
	billingPlanPriceObj.MergeData(data)

	plan, err := billing_plan.Get(req.Context(), billingPlanPriceObj.BillingPlanID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*billing_plan_price.BillingPlanPrice](err)

	}
	err = billing.CreateStripePrice(req.Context(), plan.StripeProductID.Get(), billingPlanPriceObj, userSession.User)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*billing_plan_price.BillingPlanPrice](err)

	}

	return response.Success(billingPlanPriceObj)
}

func adminUpdate(_ http.ResponseWriter, req *http.Request) (*billing_plan_price.BillingPlanPriceJoined, int, error) {
	userSession := request.GetReqSession(req)
	data := request.GetModelPostData(req)
	id := chi.URLParam(req, "id")
	billingPlanPriceObj, err := billing_plan_price.GetJoined(req.Context(), types.UUID(id))
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*billing_plan_price.BillingPlanPriceJoined](err)
	}

	if tools.Empty(billingPlanPriceObj) {
		return response.AdminBadRequestError[*billing_plan_price.BillingPlanPriceJoined](errors.Errorf("Object not found with ID: %s", id))
	}

	billingPlanPriceObj.MergeData(data)
	if billingPlanPriceObj.HasStripeChanges() {
		err := billing.UpdateStripePrice(req.Context(), &billingPlanPriceObj.BillingPlanPrice, userSession.User)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.AdminBadRequestError[*billing_plan_price.BillingPlanPriceJoined](err)
		}
	} else {
		err = billingPlanPriceObj.Save(userSession.User)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.AdminBadRequestError[*billing_plan_price.BillingPlanPriceJoined](err)
		}
	}

	return response.Success(billingPlanPriceObj)
}
