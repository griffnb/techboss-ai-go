package accounts

import (
	"net/http"

	"github.com/griffnb/techboss-ai-go/internal/common"
	"github.com/griffnb/techboss-ai-go/internal/services/account_service"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/router/request"
	"github.com/CrowdShield/go-core/lib/router/response"
	"github.com/CrowdShield/go-core/lib/session"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/CrowdShield/go-core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/controllers/login"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/integrations/cloudflare"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
)

type SignupResponse struct {
	Token             string `json:"token"`
	RedirectURL       string `json:"redirect_url,omitempty"`
	VerificationToken string `json:"verification_token,omitempty"`
}

func oauthSignup(res http.ResponseWriter, req *http.Request) (*SignupResponse, int, error) {
	body := request.GetJSONPostData(req)
	token := tools.ParseStringI(body["token"])
	guestSession := session.Load(token)
	if tools.Empty(guestSession) {
		return response.PublicNotFoundError[*SignupResponse]()
	}
	data := request.ConvertPost(body)

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
	turnstileToken := data["cf_token"]

	if tools.Empty(turnstileToken) && !tools.Empty(cloudflare.Client()) && !tools.Empty(cloudflare.Client().TurnstileKey) {
		return response.PublicBadRequestError[*SignupResponse]()
	}

	if !tools.Empty(cloudflare.Client()) && !tools.Empty(cloudflare.Client().TurnstileKey) {
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

	savingUser := account.New()
	dataCopy := accountObj.GetDataCopy()
	savingUser.SetData(dataCopy)
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

func openSignup(res http.ResponseWriter, req *http.Request) (*SignupResponse, int, error) {
	rawdata := request.GetJSONPostData(req)
	data := request.ConvertPost(rawdata)

	log.Debugf("in signup")

	var accountObj *account.Account
	var err error

	// New user from signup
	if tools.Empty(data["id"]) {
		accountObj = account.New()
		account.MergeSignup(accountObj, data, nil)
		common.GenerateURN(accountObj) // setup IDs
	} else {
		// User is from invite
		accountObj, err = account.Get(req.Context(), types.UUID(tools.ParseStringI(data["id"])))
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[*SignupResponse]()
		}
		account.MergeSignup(accountObj, data, nil)
	}

	turnstileToken := data["cf_token"]

	if tools.Empty(turnstileToken) && !tools.Empty(cloudflare.Client()) && !tools.Empty(cloudflare.Client().TurnstileKey) {
		return response.PublicBadRequestError[*SignupResponse]()
	}

	if !tools.Empty(cloudflare.Client()) && !tools.Empty(cloudflare.Client().TurnstileKey) {
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

	savingUser := account.New()
	dataCopy := accountObj.GetDataCopy()
	savingUser.SetData(dataCopy)

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

	successResponse := &SignupResponse{
		Token: sessionObj.Key,
	}

	// Include verification token in response when in dev mode
	if environment.IsLocalDev() {
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

// @link {models}/src/models/account/services/_existing.ts:checkExisting
func openCheckExisting(_ http.ResponseWriter, req *http.Request) (*ExistingResponse, int, error) {
	data := &ExistingCheck{}
	err := request.GetJSONPostDataStruct(req, data)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*ExistingResponse]()
	}

	if tools.Empty(data.CFToken) && !tools.Empty(cloudflare.Client()) && !tools.Empty(cloudflare.Client().TurnstileKey) {
		return response.PublicBadRequestError[*ExistingResponse]()
	}

	if !tools.Empty(cloudflare.Client()) && !tools.Empty(cloudflare.Client().TurnstileKey) {
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
