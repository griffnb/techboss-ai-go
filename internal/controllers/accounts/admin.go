package accounts

import (
	"net/http"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/router/request"
	"github.com/CrowdShield/go-core/lib/router/response"
	"github.com/CrowdShield/go-core/lib/session"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/controllers/login"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/services/account_service"
)

func adminTestCreate(res http.ResponseWriter, req *http.Request) (*account.Account, int, error) {
	userSession := request.GetReqSession(req)

	createInput, err := request.GetJSONPostAs[*account_service.TestUserInput](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*account.Account](err)
	}

	accountObj, err := account_service.CreateTestUser(req.Context(), createInput, userSession.User)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*account.Account](err)
	}

	if !tools.Empty(req.URL.Query().Get("login_as")) {
		accountSession := session.New(tools.ParseStringI(req.Context().Value("ip"))).WithUser(accountObj)
		err = accountSession.Save()
		if err != nil {
			log.ErrorContext(err, req.Context())
			response.ErrorWrapper(res, req, err.Error(), http.StatusBadRequest)
		}

		login.SendSessionCookie(res, environment.GetConfig().Server.SessionKey, accountSession.Key)
		log.Debugf("Sending login cookie for test user %s", accountObj.ID().String())
	}

	return response.Success(accountObj)
}
