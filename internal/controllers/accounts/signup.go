package accounts

import (
	"net/http"

	"github.com/griffnb/techboss-ai-go/internal/common"
	"github.com/griffnb/techboss-ai-go/internal/services/account_service"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/session"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/controllers/login"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/integrations/cloudflare"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
)

type SignupResponse struct {
	Token             string `public:"view" json:"token"`
	RedirectURL       string `public:"view" json:"redirect_url,omitempty"`
	VerificationToken string `public:"view" json:"verification_token,omitempty"`
}

// oauthSignup handles OAuth-based signup flow
//
//	@Summary		OAuth signup
//	@Description	Completes signup process for OAuth authenticated users
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Param			body	body		object	true	"Signup data with OAuth token"
//	@Success		200		{object}	response.SuccessResponse{data=SignupResponse}
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		404		{object}	response.ErrorResponse
//	@Failure		409		{object}	response.ErrorResponse
//	@Router			/signup/oauth [post]
func oauthSignup(res http.ResponseWriter, req *http.Request) (*SignupResponse, int, error) {
	data := request.GetJSONPostMap(req)
	token := tools.ParseStringI(data["token"])
	guestSession := session.Load(token)
	if tools.Empty(guestSession) {
		return response.PublicNotFoundError[*SignupResponse]()
	}

	profile := guestSession.GetData()

	if !tools.IsEqual(profile["email"], data["email"]) {
		return response.PublicBadRequestError[*SignupResponse]()
	}

	// Invalid session and start signup process
	err := guestSession.Invalidate()
	if err != nil {
		log.ErrorContext(err, req.Context())
	}

	// Dont let them set password if its oauth
	delete(data, "password")
	accountObj := account.New()
	account.MergeSignup(accountObj, data, nil)
	// Don't want frontend setting is oauth flag
	signupProps, err := accountObj.SignupProperties.Get()
	if err != nil {
		return response.PublicBadRequestError[*SignupResponse]()
	}
	signupProps.IsOauth = 1
	accountObj.SignupProperties.Set(signupProps)

	// Set random password
	accountObj.Set("password", tools.GUID())
	common.GenerateURN(accountObj) // setup IDs

	if !tools.Empty(cloudflare.Client()) && !tools.Empty(cloudflare.Client().TurnstileKey) {
		turnstileToken := data["cf_token"]
		if tools.Empty(turnstileToken) {
			return response.PublicBadRequestError[*SignupResponse]()
		}
		resp, err := cloudflare.Client().
			ValidateTurnstileResponse(tools.ParseStringI(turnstileToken), req.RemoteAddr)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[*SignupResponse]()
		}

		if !resp {
			return response.PublicBadRequestError[*SignupResponse]()
		}
	}

	// Check if email already exists (including disabled accounts)
	// This uses GetExistingByEmail which checks Deleted = 0, matching the /account/check endpoint
	existingAccount, err := account.Exists(req.Context(), accountObj.Email.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*SignupResponse]()
	}

	if !tools.Empty(existingAccount) {
		return response.PublicCustomError[*SignupResponse]("Email already exists", http.StatusConflict)
	}

	savingUser := accountObj.ToSavingUser()
	err = account_service.Signup(req.Context(), accountObj, savingUser)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*SignupResponse]()
	}

	// Create a new session with the user
	sessionObj := session.New(tools.ParseStringI(req.Context().Value("ip"))).
		WithUser(accountObj)
	err = sessionObj.Save()
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*SignupResponse]()
	}

	// Set accout as email verified and move to onboarding
	err = account_service.EmailVerified(req.Context(), accountObj, savingUser)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*SignupResponse]()
	}
	successResponse := &SignupResponse{}
	successResponse.Token = sessionObj.Key
	// No verification token for OAuth signup - email is already verified

	// attach session cookies
	login.SendSessionCookie(res, environment.GetConfig().Server.SessionKey, sessionObj.Key)
	login.SendOrgCookie(res, accountObj.OrganizationID.Get().String())

	return response.Success(successResponse)
}

// openSignup handles standard email/password signup flow
//
//	@Summary		Standard signup
//	@Description	Creates a new account with email and password or completes invited user signup
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Param			body	body		object	true	"Signup data"
//	@Success		200		{object}	response.SuccessResponse{data=SignupResponse}
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		409		{object}	response.ErrorResponse
//	@Router			/signup [post]
func openSignup(res http.ResponseWriter, req *http.Request) (*SignupResponse, int, error) {
	data := request.GetJSONPostMap(req)
	var accountObj *account.Account
	var err error

	// New user from signup
	if tools.Empty(data["id"]) {
		accountObj = account.New()
		account.MergeSignup(accountObj, data, nil)
		common.GenerateURN(accountObj) // setup IDs
		accountObj.Status.Set(account.STATUS_PENDING_EMAIL_VERIFICATION)
	} else {
		// User is from invite
		accountObj, err = account.Get(req.Context(), types.UUID(tools.ParseStringI(data["id"])))
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[*SignupResponse]()
		}
		account.MergeSignup(accountObj, data, nil)
	}

	// Cloudflare

	if !tools.Empty(cloudflare.Client()) && !tools.Empty(cloudflare.Client().TurnstileKey) {
		turnstileToken := data["cf_token"]
		if tools.Empty(turnstileToken) {
			return response.PublicBadRequestError[*SignupResponse]()
		}
		resp, err := cloudflare.Client().
			ValidateTurnstileResponse(tools.ParseStringI(turnstileToken), req.RemoteAddr)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[*SignupResponse]()
		}
		if !resp {
			return response.PublicBadRequestError[*SignupResponse]()
		}
	}

	// Check if email already exists (including disabled accounts)
	// This uses GetExistingByEmail which checks Deleted = 0, matching the /account/check endpoint
	existingAccount, err := account.GetExistingByEmail(req.Context(), accountObj.Email.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*SignupResponse]()
	}

	if !tools.Empty(existingAccount) {
		return response.PublicCustomError[*SignupResponse]("Email already exists", http.StatusConflict)
	}

	err = account_service.Signup(req.Context(), accountObj, accountObj.ToSavingUser())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*SignupResponse]()
	}

	// Create a new session with the user
	sessionObj := session.New(tools.ParseStringI(req.Context().Value("ip"))).
		WithUser(accountObj)
	err = sessionObj.Save()
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*SignupResponse]()
	}

	successResponse := &SignupResponse{
		Token: sessionObj.Key,
	}

	// Include verification token in response when in dev mode
	if !environment.IsProduction() {
		props := accountObj.Properties.GetI()
		if !tools.Empty(props.VerifyEmailKey) {
			successResponse.VerificationToken = props.VerifyEmailKey
		}
	}

	// attach session cookies
	login.SendSessionCookie(res, environment.GetConfig().Server.SessionKey, sessionObj.Key)
	return response.Success(successResponse)
}

type ExistingCheck struct {
	Email   string `json:"email"`
	CFToken string `json:"cf_token"`
}

type ExistingResponse struct {
	Exists bool `json:"exists"`
}

// openCheckExisting checks if an email already exists in the system
//
//	@Summary		Check existing email
//	@Description	Checks whether an email address is already registered
//	@Tags			Account
//	@Accept			json
//	@Produce		json
//	@Param			body	body		ExistingCheck	true	"Email to check"
//	@Success		200		{object}	response.SuccessResponse{data=ExistingResponse}
//	@Failure		400		{object}	response.ErrorResponse
//	@Router			/account/check [post]
func openCheckExisting(_ http.ResponseWriter, req *http.Request) (*ExistingResponse, int, error) {
	data, err := request.GetJSONPostAs[*ExistingCheck](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*ExistingResponse]()
	}

	if !tools.Empty(cloudflare.Client()) && !tools.Empty(cloudflare.Client().TurnstileKey) {
		if tools.Empty(data.CFToken) {
			return response.PublicBadRequestError[*ExistingResponse]()
		}
		resp, err := cloudflare.Client().
			ValidateTurnstileResponse(tools.ParseStringI(data.CFToken), req.RemoteAddr)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[*ExistingResponse]()
		}

		if !resp {
			return response.PublicBadRequestError[*ExistingResponse]()
		}
	}

	existingAccount, err := account.GetExistingByEmail(req.Context(), data.Email)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*ExistingResponse]()
	}

	return response.Success(&ExistingResponse{
		Exists: !tools.Empty(existingAccount),
	})
}
