package billing_plans

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/models/billing_plan"
	"github.com/griffnb/techboss-ai-go/internal/services/billing"
	"github.com/pkg/errors"
)

func adminCreate(_ http.ResponseWriter, req *http.Request) (*billing_plan.BillingPlan, int, error) {
	userSession := request.GetReqSession(req)
	data := request.GetModelPostData(req)
	billingPlanObj := billing_plan.New()
	billingPlanObj.MergeData(data)

	err := billing.CreateStripeProduct(req.Context(), billingPlanObj, userSession.User)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*billing_plan.BillingPlan](err)
	}

	return response.Success(billingPlanObj)
}

func adminUpdate(_ http.ResponseWriter, req *http.Request) (*billing_plan.BillingPlanJoined, int, error) {
	userSession := request.GetReqSession(req)
	data := request.GetModelPostData(req)
	id := chi.URLParam(req, "id")
	billingPlanObj, err := billing_plan.GetJoined(req.Context(), types.UUID(id))
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*billing_plan.BillingPlanJoined](err)
	}

	if tools.Empty(billingPlanObj) {
		return response.AdminBadRequestError[*billing_plan.BillingPlanJoined](errors.Errorf("Object not found with ID: %s", id))
	}

	billingPlanObj.MergeData(data)
	if billingPlanObj.HasStripeChanges() {
		err = billing.UpdateStripeProduct(req.Context(), &billingPlanObj.BillingPlan, userSession.User)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.AdminBadRequestError[*billing_plan.BillingPlanJoined](err)
		}
	} else {
		err = billingPlanObj.Save(userSession.User)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.AdminBadRequestError[*billing_plan.BillingPlanJoined](err)
		}
	}

	return response.Success(billingPlanObj)
}
