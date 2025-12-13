package login

import (
	"net/http"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/session"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/integrations/cloudflare"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	emailsender "github.com/griffnb/techboss-ai-go/internal/services/email_sender"
	"github.com/griffnb/techboss-ai-go/internal/services/magic_link"
)

// Logs in with a magic link session
// @link {models}/src/models/account/services/_link.ts:loginLink
func loginMagicLink(res http.ResponseWriter, req *http.Request) (*TokenResponse, int, error) {
	tokenInput, err := request.GetJSONPostAs[*TokenInput](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*TokenResponse]()
	}

	magicLink, err := magic_link.GetSession(req.Context(), tokenInput.Token)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*TokenResponse]()
	}

	if tools.Empty(magicLink) || tools.Empty(magicLink.AccountID) {
		return response.PublicNotFoundError[*TokenResponse]()
	}

	accountObj, err := account.Get(req.Context(), magicLink.AccountID)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*TokenResponse]()
	}
	if tools.Empty(accountObj) {
		return response.PublicNotFoundError[*TokenResponse]()
	}

	if accountObj.Disabled.Bool() {
		return response.PublicCustomError[*TokenResponse]("Account is disabled", http.StatusForbidden)
	}

	userSession, ok := request.CheckReqSession(req)
	if !ok {
		userSession = session.New(tools.ParseStringI(req.Context().Value("ip"))).WithUser(accountObj)
		err = userSession.Save()
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[*TokenResponse]()
		}

	}

	/*
		bgCtx := context.WithoutCancel(req.Context())
		go func(accountObj *account.Account) {
			fullJoined, err := account.GetJoined(bgCtx, accountObj.ID())
			if err != nil {
				log.ErrorContext(err, bgCtx)
				return
			}
			err = slacknotifications.Login(bgCtx, fullJoined, slacknotifications.LOGIN_TYPE_MAGIC_LINK)
			if err != nil {
				log.ErrorContext(err, bgCtx)
			}
			accountObj.LastLoginTS.Set(time.Now().Unix())
			err = accountObj.Save(nil)
			if err != nil {
				log.ErrorContext(err, bgCtx)
			}

			emailObj, err := email_address.GetPrimaryByAccountID(bgCtx, accountObj.ID())
			if err != nil {
				log.ErrorContext(err, bgCtx)
				return
			}
			// Auto verify email if not already verified since it was sent by a magic link to their email
			if !tools.Empty(emailObj) && !emailObj.IsVerified.Bool() {
				emailObj.IsVerified.Set(1)
				err = emailObj.SaveWithLockBypass(bgCtx, accountObj)
				if err != nil {
					log.ErrorContext(err, bgCtx)
					return
				}
			}
		}(accountObj)
	*/

	SendSessionCookie(res, environment.GetConfig().Server.SessionKey, userSession.Key)
	SendOrgCookie(res, accountObj.OrganizationID.Get().String())

	return response.Success(&TokenResponse{
		Token: userSession.Key,
	})
}

type SendLinkInput struct {
	Email         string `json:"email"`
	CfToken       string `json:"cf_token"`
	AutoToken     string `json:"auto_token"`
	PreviousToken string `json:"previous_token"`
	RedirectURL   string `json:"redirect_url"`
}

// @link {models}/src/models/account/services/_link.ts:sendLink

type SendMagicLinkResponse struct {
	Email     string `json:"email"                public:"view"`
	OrgDomain string `json:"org_domain,omitempty" public:"view"`
}

func sendMagicLink(_ http.ResponseWriter, req *http.Request) (*SendMagicLinkResponse, int, error) {
	input, err := request.GetJSONPostAs[*SendLinkInput](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*SendMagicLinkResponse]()
	}

	if input.Email == "" && input.AutoToken == "" {
		return response.PublicCustomError[*SendMagicLinkResponse]("Email is required", http.StatusBadRequest)
	}

	var accountObj *account.Account

	if !tools.Empty(input.AutoToken) {
		magicLink, err := magic_link.GetSession(req.Context(), input.AutoToken)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[*SendMagicLinkResponse]()
		}

		if tools.Empty(magicLink) || tools.Empty(magicLink.AccountID) {
			return response.PublicNotFoundError[*SendMagicLinkResponse]()
		}

		accountObj, err = account.Get(req.Context(), magicLink.AccountID)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[*SendMagicLinkResponse]()
		}
		if tools.Empty(accountObj) {
			return response.PublicNotFoundError[*SendMagicLinkResponse]()
		}

		if accountObj.Disabled.Bool() {
			return response.PublicCustomError[*SendMagicLinkResponse]("Account is disabled", http.StatusForbidden)
		}

	} else {

		turnstileToken := input.CfToken

		if cloudflare.Configured() {
			if tools.Empty(turnstileToken) && !tools.Empty(cloudflare.Client().TurnstileKey) {
				return response.PublicBadRequestError[*SendMagicLinkResponse]()
			}

			if !tools.Empty(cloudflare.Client().TurnstileKey) {
				resp, err := cloudflare.Client().ValidateTurnstileResponse(tools.ParseStringI(turnstileToken), req.RemoteAddr)
				if err != nil {
					log.ErrorContext(err, req.Context())
					return response.PublicBadRequestError[*SendMagicLinkResponse]()
				}

				if !resp {
					return response.PublicBadRequestError[*SendMagicLinkResponse]()
				}
			}
		}

		accountObj, err = account.GetByEmail(req.Context(), input.Email)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.PublicBadRequestError[*SendMagicLinkResponse]()
		}

		if tools.Empty(accountObj) {
			return response.PublicNotFoundError[*SendMagicLinkResponse]()
		}

		if accountObj.Disabled.Bool() {
			return response.PublicCustomError[*SendMagicLinkResponse]("Account is disabled", http.StatusForbidden)
		}

	}

	key, err := magic_link.CreateSession(req.Context(), accountObj, &magic_link.LinkOptions{
		PreviousToken: input.PreviousToken,
	})
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*SendMagicLinkResponse]()
	}

	err = emailsender.SendLoginLinkEmail(accountObj, key, input.RedirectURL)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.PublicBadRequestError[*SendMagicLinkResponse]()
	}

	return response.Success(&SendMagicLinkResponse{
		Email: accountObj.Email.Get(),
	})
}
