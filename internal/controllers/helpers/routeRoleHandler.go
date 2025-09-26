package helpers

import (
	"context"
	"net/http"
	"strings"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/router"
	"github.com/CrowdShield/go-core/lib/session"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/models/admin"
	"github.com/pkg/errors"
)

// RoleHandlerMap defines a mapping from roles to http.HandlerFunc
type RoleHandlerMap map[constants.Role]http.HandlerFunc

// RoleHandler takes a map of roles to handler functions and returns a http.HandlerFunc
// TODO this is weird with the admin stuff, extract it out to be smarter
func RoleHandler(roleHandlers RoleHandlerMap) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		handleAdminRoute(res, req, roleHandlers)
	}
}

// GetAdminSession Gets the admin session from the request checking header and cookie, priority is cookie
func GetAdminSession(req *http.Request) *session.Session {
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

	if !AdminKeyValid(sessionKey) {
		return nil
	}

	return session.Load(sessionKey)
}

func AdminKeyValid(sessionKey string) bool {
	return strings.HasPrefix(sessionKey, "admn:")
}

func CreateAdminKey(sessionKey string) string {
	return "admn:" + sessionKey
}

func handleAdminRoute(res http.ResponseWriter, req *http.Request, roleHandlers RoleHandlerMap) {
	adminSession := GetAdminSession(req)

	if tools.Empty(adminSession) {
		if handler, ok := roleHandlers[constants.ROLE_UNAUTHORIZED]; ok {
			handler(res, req)
			return
		}
		ErrorWrapper(res, req, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// sets to the session context the admin session
	ctx := context.WithValue(req.Context(), router.SessionContextKey("session"), adminSession)

	admn, err := admin.Get(req.Context(), adminSession.User.ID())
	if err != nil {
		log.ErrorContext(err, req.Context())
		ErrorWrapper(res, req, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if tools.Empty(admn) {
		log.ErrorContext(errors.Errorf("admin not found %s", adminSession.User.ID()), req.Context())
		ErrorWrapper(res, req, "Unauthorized", http.StatusUnauthorized)
		return
	}

	role := admn.Role.Get()
	adminSession.LoadedUser = admn

	if recorder, ok := res.(*router.ResponseRecorder); ok {
		recorder.Trace.Admin = admn.Email.Get()
		recorder.Trace.User = admn
		recorder.Trace.SessionID = adminSession.Key
	}

	req = req.WithContext(ctx)

	// Is there a specific endpoint for my role
	if handler, ok := roleHandlers[role]; ok {
		handler(res, req)
		return
	}

	// If i dont have a specific one, iterate through the roles to find one im allowed to do
	for _, possibleRole := range constants.DescOrderedAdminRoles {
		if possibleRole <= role {
			if handler, ok := roleHandlers[possibleRole]; ok {
				handler(res, req)
				return
			}
		}
	}
}

func GetReqSession(req *http.Request) *session.Session {
	return req.Context().Value(router.SessionContextKey("session")).(*session.Session)
}

func IsSuperUpdate(req *http.Request) bool {
	if !environment.IsProduction() {
		return false
	}

	if !HasAdminSession(req) {
		return false
	}

	userSession := req.Context().Value(router.SessionContextKey("session")).(*session.Session)
	accountObj := userSession.LoadedUser.(*account.Account)

	return !accountObj.IsInternal()
}

// GetAdminSession Gets the admin session from the request checking header and cookie, priority is cookie
func HasAdminSession(req *http.Request) bool {
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

	if AdminKeyValid(sessionKey) {
		return true
	}

	return false
}
