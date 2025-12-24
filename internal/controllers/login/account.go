package login

import (
	"net/http"
	"strings"
	"time"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/router/route_helpers"
	"github.com/griffnb/core/lib/session"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
)

// standard login

type LoginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token    string `json:"token"     public:"view"`
	IsSignup bool   `json:"is_signup" public:"view"`
}

type TokenInput struct {
	Token string `json:"token"`
}

type TokenResponse struct {
	Token    string `json:"token"    public:"view"`
	Redirect string `json:"redirect" public:"view"`
}

func login(res http.ResponseWriter, req *http.Request) (*LoginResponse, int, error) {
	loginData, err := request.GetJSONPostAs[*LoginData](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*LoginResponse]()
	}

	accountObj, err := account.GetByEmail(req.Context(), loginData.Email)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicCustomError[*LoginResponse]("Invalid User/Password", http.StatusBadRequest)
	}

	// User not found - invalid login detected
	if tools.Empty(accountObj) {
		return response.PublicCustomError[*LoginResponse]("Invalid User/Password", http.StatusBadRequest)
	}

	if !account.VerifyPassword(accountObj.HashedPassword.Get(), loginData.Password, accountObj.ID()) {
		return response.PublicCustomError[*LoginResponse]("Invalid User/Password", http.StatusBadRequest)
	}

	accountObj.LastLoginTS.Set(time.Now().Unix())
	err = accountObj.Save(nil)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*LoginResponse]()
	}

	userSession := session.New(tools.ParseStringI(req.Context().Value("ip"))).WithUser(accountObj)
	err = userSession.Save()
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*LoginResponse]()
	}

	/* TODO slack notification
	bgCtx := context.WithoutCancel(req.Context())
	go func(accountObj *account.Account) {
		fullJoined, err := account.GetJoined(bgCtx, accountObj.ID())
		if err != nil {
			log.ErrorContext(err, bgCtx)
			return
		}
		err = slacknotifications.Login(bgCtx, fullJoined, slacknotifications.LOGIN_TYPE_USERNAME_PW)
		if err != nil {
			log.ErrorContext(err, bgCtx)
		}
	}(accountObj)
	*/

	successResponse := &LoginResponse{
		Token: userSession.Key,
	}
	SendSessionCookie(res, environment.GetConfig().Server.SessionKey, userSession.Key)
	SendOrgCookie(res, accountObj.OrganizationID.Get().String())
	return response.Success(successResponse)
}

// Gets oauth profile for signup
func getProfile(_ http.ResponseWriter, req *http.Request) (any, int, error) {
	profileData, err := request.GetJSONPostAs[*TokenInput](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*account.AccountJoined]()
	}

	token := profileData.Token
	guestSession := session.Load(token)

	if tools.Empty(guestSession) {
		return response.PublicCustomError[any]("Invalid token", http.StatusBadRequest)
	}

	profileInfo := guestSession.GetData()
	cleanProfile := map[string]any{}
	for k, v := range profileInfo {
		if !strings.EqualFold(k, "oa_key") {
			cleanProfile[k] = v
		}
	}
	return response.Success(cleanProfile)
}

// This is for loging in on the frontend with an oauth token
func tokenLogin(res http.ResponseWriter, req *http.Request) (*LoginResponse, int, error) {
	input, err := request.GetJSONPostAs[*TokenInput](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*LoginResponse]()
	}
	profile, token, err := route_helpers.HandleTokenLogin(req.Context(), input.Token, environment.GetOauth())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*LoginResponse]()
	}

	accountObj, err := account.GetByEmail(req.Context(), profile.Email)
	// Query error occured
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*LoginResponse]()
	}

	if tools.Empty(accountObj) {
		guestSession := session.New(tools.ParseStringI(req.Context().Value("ip")))
		guestSession.Key = tools.SessionKey()
		data := tools.StructToMap(profile)

		data["oa_key"] = token
		guestSession.SetData(data)
		err = guestSession.Save()
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[*LoginResponse]()
		}
		successResponse := &LoginResponse{
			Token:    guestSession.Key,
			IsSignup: true,
		}
		return response.Success(successResponse)
	}

	if accountObj.Disabled.Bool() {
		return response.PublicCustomError[*LoginResponse]("Invalid User/Password", http.StatusBadRequest)
	}

	accountObj.LastLoginTS.Set(time.Now().Unix())
	err = accountObj.Save(nil)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*LoginResponse]()
	}

	// create a session and set its value as the same as the token
	userSession := session.New(tools.ParseStringI(req.Context().Value("ip"))).WithUser(accountObj)
	userSession.Key = token
	err = userSession.Save()
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*LoginResponse]()
	}
	/*
		bgCtx := context.WithoutCancel(req.Context())
		go func(accountObj *account.Account) {
			fullJoined, err := account.GetJoined(bgCtx, accountObj.ID())
			if err != nil {
				log.ErrorContext(err, bgCtx)
				return
			}
			err = slacknotifications.Login(bgCtx, fullJoined, slacknotifications.LOGIN_TYPE_OAUTH)
			if err != nil {
				log.ErrorContext(err, bgCtx)
			}
		}(accountObj)
	*/

	SendSessionCookie(res, environment.GetConfig().Server.SessionKey, userSession.Key)
	SendOrgCookie(res, accountObj.OrganizationID.Get().String())
	successResponse := &LoginResponse{
		Token:    userSession.Key,
		IsSignup: false,
	}
	return response.Success(successResponse)
}

func logout(res http.ResponseWriter, req *http.Request) (bool, int, error) {
	data, err := request.GetJSONPostAs[*TokenInput](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[bool]()
	}

	if !tools.Empty(data.Token) {
		userSession := session.Load(data.Token)

		if !tools.Empty(userSession) {
			err := userSession.Invalidate()
			if err != nil {
				log.ErrorContext(err, req.Context())
			}
		}

		if helpers.CustomAdminKeyValid(tools.ParseStringI(data.Token)) {
			DeleteSessionCookie(res, environment.GetConfig().Server.AdminSessionKey)
		} else {
			DeleteSessionCookie(res, environment.GetConfig().Server.SessionKey)
		}

	}

	return response.Success(true)
}
