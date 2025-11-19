package magic_link

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/griffnb/core/lib/session"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
)

type LinkOptions struct {
	ExpirationHours int    `json:"expiration_hours"`
	PreviousToken   string `json:"previous_token"`
	RedirectURL     string `json:"redirect_url"`
	AutoLogin       bool   `json:"auto_login"`
}

type ShortLinkOptions[T any] struct {
	LinkOptions
	Data T `json:"data"`
}

func CreateSession(_ context.Context, accountObj *account.Account, options *LinkOptions) (string, error) {
	magicSession := session.New("").WithUser(accountObj)

	// Refreshes a previous session if it exists
	if !tools.Empty(options.PreviousToken) {
		previousSession := session.LoadExpired(options.PreviousToken)
		if !tools.Empty(previousSession) && !tools.Empty(previousSession.Model) {
			magicSession.WithModel(previousSession.Model)
		}
	}

	var expirationTime int64
	if options.ExpirationHours > 0 {
		expirationTime = time.Now().Add(time.Hour * time.Duration(options.ExpirationHours)).Unix()
	} else {
		expirationTime = time.Now().Add(24 * time.Hour).Unix()
	}

	err := magicSession.SaveWithExpiration(expirationTime)
	if err != nil {
		return "", err
	}

	return magicSession.Key, nil
}

type MagicLink struct {
	AccountID types.UUID `json:"account_id"`
}

func GetSession(_ context.Context, token string) (*MagicLink, error) {
	magicSession := session.Load(token)

	if tools.Empty(magicSession) {
		return nil, nil
	}

	err := magicSession.Invalidate()
	if err != nil {
		return nil, err
	}

	magicLink := &MagicLink{
		AccountID: magicSession.User.ID(),
	}

	return magicLink, nil
}

func GetLoginLink(token, redirectURL string) string {
	baseURL := environment.GetConfig().Server.AppURL

	params := url.Values{}
	params.Add("token", token)
	if !tools.Empty(redirectURL) {
		params.Add("redirect_url", redirectURL)
	}

	return fmt.Sprintf("%s/login/link?%s", baseURL, params.Encode())
}

func GetAutoSendLink(token, redirectURL string) string {
	baseURL := environment.GetConfig().Server.AppURL

	params := url.Values{}
	params.Add("token", token)
	if !tools.Empty(redirectURL) {
		params.Add("redirect_url", redirectURL)
	}

	return fmt.Sprintf("%s/login/link/send?%s", baseURL, params.Encode())
}

func GetResendLink(email, token, redirectURL string) string {
	baseURL := environment.GetConfig().Server.AppURL

	params := url.Values{}
	params.Add("email", email)
	if !tools.Empty(token) {
		params.Add("token", token)
	}
	if !tools.Empty(redirectURL) {
		params.Add("redirect_url", redirectURL)
	}

	return fmt.Sprintf("%s/login/resend?%s", baseURL, params.Encode())
}
