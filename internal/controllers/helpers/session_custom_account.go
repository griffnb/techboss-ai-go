package helpers

import (
	"net/http"

	"github.com/griffnb/core/lib/session"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/environment"
)

// getCustomAdminSession Gets the account session from the request checking header and cookie, priority is cookie
func getCustomAccountSession(req *http.Request) *session.Session {
	key := environment.GetConfig().Server.SessionKey

	cookieSessionKey := ""
	cookie, _ := req.Cookie(key)
	if !tools.Empty(cookie) {
		cookieSessionKey = cookie.Value
	}

	headerSessionKey := req.Header.Get(key)

	var sessionKey string
	if !tools.Empty(cookieSessionKey) {
		sessionKey = cookieSessionKey
	} else if !tools.Empty(headerSessionKey) {
		sessionKey = headerSessionKey
	} else {
		return nil
	}

	userSession := session.Load(sessionKey)

	if tools.Empty(userSession) || tools.Empty(userSession.User) {
		return nil
	}

	return userSession
}
