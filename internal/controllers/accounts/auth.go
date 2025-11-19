package accounts

import (
	"net/http"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/session"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
)

func authMe(_ http.ResponseWriter, req *http.Request) (*account.AccountJoined, int, error) {
	userSession := req.Context().Value(router.SessionContextKey("session")).(*session.Session)
	accountObj, err := account.GetJoined(req.Context(), userSession.User.ID())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*account.AccountJoined]()
	}
	/*
		adminSession := helpers.GetAdminSession(req)
		if adminSession != nil && !tools.Empty(adminSession.User) {
			accountObj.IsSuperUserSession.Set(1)
		}
	*/

	return response.Success(accountObj)
}
