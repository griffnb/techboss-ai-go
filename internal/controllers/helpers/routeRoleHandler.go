package helpers

import (
	"context"
	"net/http"
	"strings"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/router"
	"github.com/CrowdShield/go-core/lib/router/response"
	"github.com/CrowdShield/go-core/lib/session"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/account"

	"github.com/pkg/errors"
)

// RoleHandlerMap defines a mapping from roles to http.HandlerFunc
type RoleHandlerMap map[constants.Role]http.HandlerFunc

// RoleHandler takes a map of roles to handler functions and returns a http.HandlerFunc
func RoleHandler(roleHandlers RoleHandlerMap) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.URL.Path, "/admin/") || strings.HasPrefix(req.URL.Path, "admin/") {
			handleAdminRoute(res, req, roleHandlers)
			return
		}
		handlePublicRoute(res, req, roleHandlers)
	}
}

func handleAdminRoute(res http.ResponseWriter, req *http.Request, roleHandlers RoleHandlerMap) {
	adminSession, err := getAdminSession(req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		response.ErrorWrapper(res, req, err.Error(), http.StatusUnauthorized)
		return
	}

	if tools.Empty(adminSession) {
		response.ErrorWrapper(res, req, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// sets to the session context the admin session
	ctx := context.WithValue(req.Context(), router.SessionContextKey("session"), adminSession)

	admn, err := loadAdmin(req, adminSession)
	if err != nil {
		log.ErrorContext(err, req.Context())
		response.ErrorWrapper(res, req, "Unauthorized", http.StatusUnauthorized)
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

func handlePublicRoute(res http.ResponseWriter, req *http.Request, roleHandlers RoleHandlerMap) {
	userSession, err := getAccountSession(req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		response.ErrorWrapper(res, req, "internal error", http.StatusBadRequest)
		return
	}

	if tools.Empty(userSession) {
		if handler, ok := roleHandlers[constants.ROLE_UNAUTHORIZED]; ok {
			handler(res, req)
			return
		}
		response.ErrorWrapper(res, req, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// sets to the session context the user session
	ctx := context.WithValue(req.Context(), router.SessionContextKey("session"), userSession)
	req = req.WithContext(ctx)

	// TODO might need to cache this
	accnt, err := loadAccount(req, userSession)
	if err != nil {
		log.ErrorContext(err, req.Context())
		response.ErrorWrapper(res, req, "Unauthorized", http.StatusUnauthorized)
		return
	}

	role := accnt.Role.Get()
	userSession.LoadedUser = accnt
	if recorder, ok := res.(*router.ResponseRecorder); ok {

		if HasAdminSession(req) {
			adminSession := GetAdminSession(req)
			if !tools.Empty(adminSession) {
				recorder.Trace.Admin = adminSession.User.GetString("email")
			}
		}

		recorder.Trace.AccountID = accnt.ID().String()
		recorder.Trace.User = accnt
		recorder.Trace.SessionID = userSession.Key

	}

	// Is there a specific endpoint for my role
	if handler, ok := roleHandlers[role]; ok {
		handler(res, req)
		return
	}

	// If i dont have a specific one, iterate through the roles to find one im allowed to do
	for _, possibleRole := range constants.DescOrderedAccountRoles {
		if possibleRole <= role {
			if handler, ok := roleHandlers[possibleRole]; ok {
				handler(res, req)
				return
			}
		}
	}
	log.ErrorContext(errors.Errorf("user does not have permission"), req.Context())
	response.ErrorWrapper(res, req, "Unauthorized", http.StatusUnauthorized)
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

func HasAdminSession(req *http.Request) bool {
	return HasCustomAdminSession(req)
}
