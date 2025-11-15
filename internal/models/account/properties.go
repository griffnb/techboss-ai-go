package account

import (
	"golang.org/x/oauth2"
)

type Properties struct {
	InviteKey        string        `json:"invite_key,omitempty"`
	InviteTS         int64         `json:"invite_ts,omitempty"`
	LastSeen         int64         `json:"last_seen,omitempty"`
	OauthToken       *oauth2.Token `json:"oauth_token,omitempty"`
	ExternalUserInfo any           `json:"external_user_info,omitempty"`
	VerifyEmailKey   string        `json:"verify_email_key,omitempty"`
}
