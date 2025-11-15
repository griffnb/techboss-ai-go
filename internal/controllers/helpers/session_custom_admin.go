package helpers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/oauth"
	"github.com/CrowdShield/go-core/lib/session"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/admin"

	"github.com/pkg/errors"
)

func GetAdminSession(req *http.Request) *session.Session {
	return getCustomAdminSession(req)
}

// GetCustomAdminSession Gets the admin session from the request checking header and cookie, priority is cookie
func getCustomAdminSession(req *http.Request) *session.Session {
	key := environment.GetConfig().Server.AdminSessionKey

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

	if !CustomAdminKeyValid(sessionKey) {
		return nil
	}

	userSession := session.Load(sessionKey)

	if tools.Empty(userSession) || tools.Empty(userSession.User) {
		userSession, err := autoLoginOauthAdmin(req, sessionKey)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return nil
		}
		return userSession
	}
	return userSession
}

func CustomAdminKeyValid(sessionKey string) bool {
	return strings.HasPrefix(sessionKey, "admn:")
}

func CreateCustomAdminKey(sessionKey string) string {
	return "admn:" + sessionKey
}

// Uses OAUTH to login users across the system
func autoLoginOauthAdmin(req *http.Request, sessionKey string) (*session.Session, error) {
	token := strings.Split(sessionKey, ":")[1]

	domain := req.Header.Get("x-domain")
	domainKey := req.Header.Get("x-domain-key")

	if tools.Empty(domain) || tools.Empty(domainKey) {
		return nil, nil
	}

	if !strings.HasSuffix(domain, "auth0.com") {
		return nil, errors.Errorf("domain %s is not auth0", domain)
	}

	if tools.Empty(domainKey) || domainKey != tools.Sha256(fmt.Sprintf("%s:%s", domain, "atlas")) {
		return nil, errors.Errorf("domain key %s is not valid", domainKey)
	}

	profile, err := oauth.VerifyIDTokenString(req.Context(), domain, token)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to verify id token")
	}

	if tools.Empty(profile) {
		return nil, errors.Errorf("No profile")
	}

	adminObj, err := admin.GetByEmail(req.Context(), profile.Email)
	// Query error occured
	if err != nil {
		log.ErrorContext(err, req.Context())
		return nil, errors.Wrap(err, "Failed to get admin")
	}

	if tools.Empty(adminObj) {
		return nil, errors.Errorf("No admin for email %s", profile.Email)
	}

	// create a session and set its value as the same as the token
	userSession := session.New(tools.ParseStringI(req.Context().Value("ip"))).WithUser(adminObj)
	userSession.Key = sessionKey
	err = userSession.Save()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to save session")
	}

	return userSession, nil
}

// GetAdminSession Gets the admin session from the request checking header and cookie, priority is cookie
func HasCustomAdminSession(req *http.Request) bool {
	key := environment.GetConfig().Server.AdminSessionKey

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
		return false
	}

	if CustomAdminKeyValid(sessionKey) {
		return true
	}

	return false
}
