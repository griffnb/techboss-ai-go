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

	"github.com/pkg/errors"
)

type VerifyInput struct {
	Verify string `json:"verify"`
	Email  string `json:"email,omitempty"`
}

type VerifyResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url,omitempty"`
}

// openVerifyInvite verifies an account invitation and creates a session
//
//	@Public
//	@Title			Verify account invite
//	@Summary		Verify account invite
//	@Description	Verifies an account invitation using the verification key and creates a user session
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Param			body	body		VerifyInput	true	"Verification details"
//	@Success		200		{object}	response.SuccessResponse{data=VerifyResponse}
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		404		{object}	response.ErrorResponse
//	@Router			/account/verify/invite [post]
func openVerifyInvite(res http.ResponseWriter, req *http.Request) (*VerifyResponse, int, error) {
	data, err := request.GetJSONPostAs[*VerifyInput](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*VerifyResponse]()
	}
	verifySessionObj := session.LoadExpired(data.Verify)

	if tools.Empty(verifySessionObj) && tools.Empty(data.Email) {
		log.ErrorContext(errors.New("no verify session found"), req.Context())
		return response.PublicNotFoundError[*VerifyResponse]()
	}

	var accountObj *account.Account

	// Use the verify session to get the account
	if !tools.Empty(verifySessionObj) {
		accountObj, err = account.Get(req.Context(), verifySessionObj.User.ID())
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[*VerifyResponse]()
		}

		if tools.Empty(accountObj) {
			log.ErrorContext(errors.Errorf("account not found %s", verifySessionObj.User.ID()), req.Context())
			return response.PublicNotFoundError[*VerifyResponse]()
		}
	} else {
		log.Debugf("No verify session, looking up by email %s", data.Email)
		// No session, so need to lookup by email and verify key
		accountObj, err = account.GetByEmail(req.Context(), data.Email)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[*VerifyResponse]()
		}

		if tools.Empty(accountObj) {
			log.ErrorContext(errors.Errorf("account not found %s", data.Email), req.Context())
			return response.PublicNotFoundError[*VerifyResponse]()
		}

		props, err := accountObj.Properties.Get()
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[*VerifyResponse]()
		}

		if tools.Empty(props.InviteKey) {
			log.ErrorContext(errors.Errorf("account not invited %s", accountObj.ID()), req.Context())
			return response.PublicNotFoundError[*VerifyResponse]()
		}

		if props.InviteKey != data.Verify {
			log.ErrorContext(
				errors.Errorf("invite key does not match Key:%s vs %s | Account:%s", props.InviteKey, data.Verify, accountObj.ID()),
				req.Context(),
			)
			return response.PublicNotFoundError[*VerifyResponse]()
		}

	}

	if accountObj.Status.Get() != account.STATUS_PENDING_EMAIL_VERIFICATION {
		err := errors.Errorf("account already accepted invite %s", accountObj.ID())
		log.ErrorContext(err, req.Context())
		return response.PublicCustomError[*VerifyResponse]("Account is not pending email verification", http.StatusBadRequest)
	}

	savingUser := accountObj.ToSavingUser()

	// Verfies email and gets the flow they need to do
	err = account_service.EmailVerified(req.Context(), accountObj, savingUser)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*VerifyResponse]()
	}

	// Create the user a session
	userSession, ok := request.CheckReqSession(req)
	if !ok {
		// create a real session for the user
		userSession = session.New(tools.ParseStringI(req.Context().Value("ip"))).
			WithUser(accountObj)
		err = userSession.Save()
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[*VerifyResponse]()
		}

		login.SendSessionCookie(res, environment.GetConfig().Server.SessionKey, userSession.Key)
		login.SendOrgCookie(res, accountObj.OrganizationID.Get().String())

	}

	successResponse := &VerifyResponse{}
	successResponse.Token = userSession.Key

	return response.Success(successResponse)
}

// openVerifyEmail verifies a user's email address and creates a session
//
//	@Public
//	@Title			Verify email address
//	@Summary		Verify email address
//	@Description	Verifies a user's email address using the verification key and creates a user session
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Param			body	body		VerifyInput	true	"Verification details"
//	@Success		200		{object}	response.SuccessResponse{data=VerifyResponse}
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		404		{object}	response.ErrorResponse
//	@Router			/account/verify/email [post]
func openVerifyEmail(res http.ResponseWriter, req *http.Request) (*VerifyResponse, int, error) {
	data, err := request.GetJSONPostAs[*VerifyInput](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*VerifyResponse]()
	}
	verifySessionObj := session.LoadExpired(data.Verify)

	if tools.Empty(verifySessionObj) {
		log.ErrorContext(errors.New("no verify session found"), req.Context())
		return response.PublicNotFoundError[*VerifyResponse]()
	}

	accountObj, err := account.Get(req.Context(), verifySessionObj.User.ID())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*VerifyResponse]()
	}

	if tools.Empty(accountObj) {
		log.ErrorContext(errors.Errorf("account not found %s", verifySessionObj.User.ID()), req.Context())
		return response.PublicNotFoundError[*VerifyResponse]()
	}

	// Verfies email and gets the flow they need to do
	err = account_service.EmailVerified(req.Context(), accountObj, verifySessionObj.User)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*VerifyResponse]()
	}

	// Create the user a session
	userSession, ok := request.CheckReqSession(req)
	if !ok {
		// create a real session for the user
		userSession = session.New(tools.ParseStringI(req.Context().Value("ip"))).
			WithUser(accountObj)
		err = userSession.Save()
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[*VerifyResponse]()
		}

		login.SendSessionCookie(res, environment.GetConfig().Server.SessionKey, userSession.Key)
		login.SendOrgCookie(res, accountObj.OrganizationID.Get().String())

	}

	successResponse := &VerifyResponse{}
	successResponse.Token = userSession.Key

	// redirect if on waitlist
	if accountObj.Status.Get() == account.STATUS_PENDING_ONBOARD {
		successResponse.RedirectURL = "/signup/organization"
		return response.Success(successResponse)
	}

	// redirect if already fully active
	if accountObj.Status.Get() == account.STATUS_ACTIVE {
		successResponse.RedirectURL = "/dashboard"
		return response.Success(successResponse)
	}

	return response.Success(successResponse)
}
