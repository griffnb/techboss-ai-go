package accounts

import (
	"net/http"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/session"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
)

// authMe retrieves the current authenticated user's account details
//
//	@Public
//	@Summary		Get current user
//	@Description	Retrieves the authenticated user's account details with joined data
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.SuccessResponse{data=account.AccountWithFeatures}
//	@Failure		400	{object}	response.ErrorResponse
//	@Router			/me [get]
func authMe(_ http.ResponseWriter, req *http.Request) (*account.AccountWithFeatures, int, error) {
	userSession := req.Context().Value(router.SessionContextKey("session")).(*session.Session)
	accountObj, err := account.GetAccountWithFeatures(req.Context(), userSession.User.ID())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*account.AccountWithFeatures]()
	}
	/*
		adminSession := helpers.GetAdminSession(req)
		if adminSession != nil && !tools.Empty(adminSession.User) {
			accountObj.IsSuperUserSession.Set(1)
		}
	*/

	return response.Success(accountObj)
}
