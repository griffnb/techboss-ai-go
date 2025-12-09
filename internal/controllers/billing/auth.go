package billing

import (
	"net/http"

	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/models/organization"
	"github.com/griffnb/techboss-ai-go/internal/models/subscription"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/techboss-ai-go/internal/services/billing"
	"github.com/pkg/errors"
)

func authCancel(_ http.ResponseWriter, req *http.Request) (any, int, error) {
	session := request.GetReqSession(req)
	accountObj := helpers.GetLoadedUser(req)

	subscriptionInfo, err := subscription.GetActiveByOrganizationID(req.Context(), accountObj.OrganizationID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[any]()
	}

	org, err := organization.Get(req.Context(), accountObj.OrganizationID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[any]()
	}

	if tools.Empty(subscriptionInfo) {
		log.ErrorContext(errors.Errorf("failed to get active subscription for organization %s", accountObj.OrganizationID.Get()), req.Context())
		return response.PublicBadRequestError[any]()
	}

	err = billing.ProcessStripeCancel(req.Context(), org, subscriptionInfo, session.User)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[any]()
	}

	return response.Success(subscriptionInfo)
}

func authResume(_ http.ResponseWriter, req *http.Request) (any, int, error) {
	userSession := request.GetReqSession(req)

	accountObj := helpers.GetLoadedUser(req)

	subscriptionInfo, err := subscription.GetActiveByOrganizationID(req.Context(), accountObj.OrganizationID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[any]()
	}

	if tools.Empty(subscriptionInfo) {
		log.ErrorContext(errors.Errorf("failed to get active subscription for organization %s", accountObj.OrganizationID.Get()), req.Context())
		return response.PublicBadRequestError[any]()
	}

	organization, err := organization.Get(req.Context(), accountObj.OrganizationID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[any]()
	}

	err = billing.ProcessStripeResume(req.Context(), organization, subscriptionInfo, userSession.User)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[any]()
	}

	return response.Success(subscriptionInfo)
}
