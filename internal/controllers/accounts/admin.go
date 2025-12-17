package accounts

import (
	"net/http"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/session"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/controllers/login"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/services/account_service"
)

// adminTestCreate creates a test account and optionally logs in as that account
//
//	@Summary		Create test account
//	@Description	Creates a test account with the provided details and optionally logs in as that account
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Param			body		body		account_service.TestUserInput	true	"Test user details"
//	@Param			login_as	query		string							false	"Set to login as the created user"
//	@Success		200			{object}	response.SuccessResponse{data=account.Account}
//	@Failure		400			{object}	response.ErrorResponse
//	@Failure		500			{object}	response.ErrorResponse
//	@Router			/admin/testUser [post]
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
