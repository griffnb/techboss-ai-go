package billing_plans

import (
	"net/http"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/models/billing_plan"
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
