package accounts

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/models/organization"
)

type APIResponse struct {
	Account      *account.Account           `json:"account"`
	Organization *organization.Organization `json:"organization"`
}

// internalAPIAccount retrieves account and organization data for internal API use
//
//	@Title			Get account with organization
//	@Summary		Get account with organization
//	@Description	Retrieves account and associated organization data by account ID
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Account ID"
//	@Success		200	{object}	response.SuccessResponse{data=APIResponse}
//	@Failure		400	{object}	response.ErrorResponse
//	@Router			/api/account/{id} [get]
func internalAPIAccount(_ http.ResponseWriter, req *http.Request) (*APIResponse, int, error) {
	id := chi.URLParam(req, "id")

	accountObj, err := account.Get(req.Context(), types.UUID(id))
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*APIResponse](err)
	}

	organizationObj, err := organization.Get(req.Context(), accountObj.OrganizationID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*APIResponse](err)
	}

	responseObj := &APIResponse{
		Account:      accountObj,
		Organization: organizationObj,
	}

	return response.Success(responseObj)
}
