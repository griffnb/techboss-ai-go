package billing_plans

import (
	"net/http"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/models/billing_plan"
	"github.com/griffnb/techboss-ai-go/internal/models/billing_plan_price"
)

func authIndex(_ http.ResponseWriter, req *http.Request) ([]*billing_plan.BillingPlanJoined, int, error) {
	userSession := request.GetReqSession(req)

	user := userSession.User

	parameters := request.BuildIndexParams(req.Context(), req.URL.Query(), TABLE_NAME)

	if tools.Empty(parameters.Limit) {
		parameters.Limit = constants.SYSTEM_LIMIT
	}

	if !tools.Empty(req.URL.Query().Get("q")) {
		addSearch(parameters, req.URL.Query().Get("q"))
	}

	billingPlanObjs, err := billing_plan.FindAllRestrictedJoined(req.Context(), parameters, user)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[[]*billing_plan.BillingPlanJoined]()

	}

	return response.Success(billingPlanObjs)
}

type PlanResponse struct {
	Plans []*PlanWithPrices `json:"plans"`
}
type PlanWithPrices struct {
	Plan   *billing_plan.BillingPlanJoined        `json:"plan"`
	Prices []*billing_plan_price.BillingPlanPrice `json:"prices"`
}

// @link {models}/src/models/billing_plan/services/_checkout.ts:getPlans
func authPlans(_ http.ResponseWriter, req *http.Request) (*PlanResponse, int, error) {
	userSession := request.GetReqSession(req)

	user := userSession.User
	parameters := model.NewOptions().WithCondition("%s = 0")
	billingPlanObjs, err := billing_plan.GetAllActivePlans(req.Context())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*PlanResponse]()

	}

	prices, err := billing_plan_price.GetAllActivePrices(req.Context())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*PlanResponse]()

	}

	planPriceMap := make(map[types.UUID][]*billing_plan_price.BillingPlanPrice)
	for _, price := range prices {
		planID := price.BillingPlanID.Get()
		planPriceMap[planID] = append(planPriceMap[planID], price)
	}

	var responsePlans []*PlanWithPrices
	for _, plan := range billingPlanObjs {
		planID := plan.ID()
		responsePlans = append(responsePlans, &PlanWithPrices{
			Plan:   plan,
			Prices: planPriceMap[planID],
		})
	}

	return response.Success(&PlanResponse{Plans: responsePlans})
}
