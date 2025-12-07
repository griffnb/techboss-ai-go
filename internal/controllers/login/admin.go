package login

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/router/route_helpers"
	"github.com/griffnb/core/lib/session"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/models/admin"
)

func adminLogInAs(res http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	accountObj, err := account.Get(req.Context(), types.UUID(id))
	if err != nil {
		log.ErrorContext(err, req.Context())
		response.ErrorWrapper(res, req, err.Error(), http.StatusBadRequest)
		return
	}

	userSession := session.New(tools.ParseStringI(req.Context().Value("ip"))).WithUser(accountObj)
	err = userSession.Save()
	if err != nil {
		log.ErrorContext(err, req.Context())
		response.ErrorWrapper(res, req, err.Error(), http.StatusBadRequest)
	}

	SendSessionCookie(res, environment.GetConfig().Server.SessionKey, userSession.Key)
	SendOrgCookie(res, accountObj.OrganizationID.Get().String())

	response.JSONDataResponseWrapper(res, req, &TokenResponse{
		Token: userSession.Key,
	})
}

// This is for loging in on the frontend with a token
func adminTokenLogin(res http.ResponseWriter, req *http.Request) (*TokenResponse, int, error) {
	input, err := request.GetJSONPostAs[*TokenInput](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*TokenResponse](err)
	}

	profile, token, err := route_helpers.HandleTokenLogin(req.Context(), input.Token, environment.GetOauth())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*TokenResponse](err)
	}

	adminObj, err := admin.GetByEmail(req.Context(), profile.Email)
	// Query error occured
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*TokenResponse](err)
	}

	// create a session and set its value as the same as the token
	userSession := session.New(tools.ParseStringI(req.Context().Value("ip"))).WithUser(adminObj)
	userSession.Key = helpers.CreateCustomAdminKey(token)
	err = userSession.Save()
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[*TokenResponse](err)
	}

	SendSessionCookie(res, environment.GetConfig().Server.AdminSessionKey, userSession.Key)

	return response.Success(&TokenResponse{
		Token: userSession.Key,
	})
}
