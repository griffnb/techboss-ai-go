package helpers

import (
	"context"
	"net/http"
	"strings"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/router"
	"github.com/CrowdShield/go-core/lib/session"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/models/admin"
	clerk_service "github.com/griffnb/techboss-ai-go/internal/services/clerk"
	"github.com/pkg/errors"
)

// RoleHandlerMap defines a mapping from roles to http.HandlerFunc
type RoleHandlerMap map[constants.Role]http.HandlerFunc

// RoleHandler takes a map of roles to handler functions and returns a http.HandlerFunc
func RoleHandler(roleHandlers RoleHandlerMap) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		clerkhttp.WithHeaderAuthorization(clerk_service.WithCustomClaimsConstructor)(
			http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				if strings.HasPrefix(req.URL.Path, "/admin/") || strings.HasPrefix(req.URL.Path, "admin/") {
					handleAdminRoute(res, req, roleHandlers)
					return
				}
				handlePublicRoute(res, req, roleHandlers)
			}),
		).ServeHTTP(res, req)
	}
}

// GetAdminSession Gets the admin session from clerk claims in the request
func GetAdminSession(req *http.Request) *session.Session {
	claims, ok := clerk.SessionClaimsFromContext(req.Context())
	if !ok {
		return nil
	}

	customClaims, err := clerk_service.CustomClaims(claims)
	if err != nil {
		return nil
	}
	log.Debugf("custom claims: %+v", customClaims)

	if tools.Empty(customClaims.AdminRole) {
		adminObj, err := clerk_service.SyncAdmin(req.Context(), claims)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return nil
		}

		if tools.Empty(adminObj) {
			log.ErrorContext(errors.New("failed to sync admin, adminObj is nil"), req.Context())
			return nil
		}

		log.Debugf("synced admin: %+v", adminObj.GetData())
		customClaims.AdminRole = adminObj.Role.Get()
		customClaims.Email = adminObj.Email.Get()
		customClaims.AdminID = adminObj.ID()
	}

	if customClaims.AdminRole < constants.ROLE_READ_ADMIN {
		return nil
	}

	clerkSession := session.New("")

	token := strings.Split(req.Header.Get("Authorization"), "Bearer ")
	if len(token) != 2 {
		clerkSession.Key = token[1]
	} else {
		clerkSession.Key = tools.SessionKey()
	}

	coreModel := &model.BaseModel{}
	coreModel.Initialize(&model.InitializeOptions{
		Table: "session",
		Model: "Session",
	})
	coreModel.MergeData(map[string]any{
		"id":    customClaims.AdminID,
		"role":  customClaims.AdminRole,
		"email": customClaims.Email,
	})
	clerkSession.WithUser(coreModel)

	return clerkSession
}

func handleAdminRoute(res http.ResponseWriter, req *http.Request, roleHandlers RoleHandlerMap) {
	adminSession := GetAdminSession(req)

	if tools.Empty(adminSession) {
		log.Debugf("Empty admin session")
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
		emailAdmin, err := admin.GetByEmail(req.Context(), adminSession.User.GetString("email"))
		if err != nil {
			log.ErrorContext(err, req.Context())
			ErrorWrapper(res, req, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Fixes the case where we've wiped the admins, and clerk sees something diff for ID
		if !tools.Empty(emailAdmin) {
			admn = emailAdmin
			err := admin.RepairID(req.Context(), emailAdmin.ID(), adminSession.User.ID())
			if err != nil {
				log.ErrorContext(err, req.Context())
				ErrorWrapper(res, req, "Internal Error", http.StatusInternalServerError)
				return
			}
		} else {
			log.ErrorContext(errors.Errorf("admin not found %s", adminSession.User.ID()), req.Context())
			ErrorWrapper(res, req, "Unauthorized", http.StatusUnauthorized)
			return
		}
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
	claims, ok := clerk.SessionClaimsFromContext(req.Context())
	if !ok {
		return false
	}

	return !tools.Empty(claims.Actor)
}

func handlePublicRoute(res http.ResponseWriter, req *http.Request, roleHandlers RoleHandlerMap) {
	userSession, err := GetSession(req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		ErrorWrapper(res, req, "internal error", http.StatusBadRequest)
		return
	}

	if tools.Empty(userSession) {
		if handler, ok := roleHandlers[constants.ROLE_UNAUTHORIZED]; ok {
			handler(res, req)
			return
		}
		ErrorWrapper(res, req, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// sets to the session context the user session
	ctx := context.WithValue(req.Context(), router.SessionContextKey("session"), userSession)
	req = req.WithContext(ctx)

	// TODO might need to cache this
	accnt, err := account.Get(req.Context(), userSession.User.ID())
	if err != nil {
		log.ErrorContext(err, req.Context())
		ErrorWrapper(res, req, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if tools.Empty(accnt) {
		log.ErrorContext(errors.Errorf("userid not found %s", userSession.User.ID()), req.Context())
		ErrorWrapper(res, req, "Unauthorized", http.StatusUnauthorized)
		return
	}

	role := accnt.Role.Get()
	userSession.LoadedUser = accnt
	if recorder, ok := res.(*router.ResponseRecorder); ok {
		/*
			if HasAdminSession(req) {
				adminSession := GetAdminSession(req)
				if !tools.Empty(adminSession) {
					recorder.Trace.Admin = adminSession.User.GetString("email")
				}
			} else {
				// only do last seen if there isnt an admin session
				go checkLastSeen(ctx, accnt)
			}
		*/

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
	ErrorWrapper(res, req, "Unauthorized", http.StatusUnauthorized)
}

// GetSession Gets the session from the request checking header and cookie, priority is cookie
func GetSession(req *http.Request) (*session.Session, error) {
	claims, ok := clerk.SessionClaimsFromContext(req.Context())
	if !ok {
		log.Debugf("no claims found")
		return nil, nil
	}

	customClaims, err := clerk_service.CustomClaims(claims)
	if err != nil {
		return nil, err
	}

	log.Debugf("custom claims: %+v", customClaims)

	if tools.Empty(customClaims.AccountID) {
		accountObj, err := clerk_service.CreateAccount(req.Context(), claims)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return nil, nil
		}

		if tools.Empty(accountObj) {
			return nil, errors.Errorf("failed to create account from claims")
		}

		log.Debugf("synced account: %+v", accountObj.GetData())
		customClaims.Role = accountObj.Role.Get()
		customClaims.Email = accountObj.Email.Get()
		customClaims.AccountID = accountObj.ID()
	}

	clerkSession := session.New("")

	token := strings.Split(req.Header.Get("Authorization"), "Bearer ")
	if len(token) != 2 {
		clerkSession.Key = token[1]
	} else {
		clerkSession.Key = tools.SessionKey()
	}

	coreModel := &model.BaseModel{}
	coreModel.Initialize(&model.InitializeOptions{
		Table: "session",
		Model: "Session",
	})
	coreModel.MergeData(map[string]any{
		"id":    customClaims.AccountID,
		"role":  customClaims.Role,
		"email": customClaims.Email,
	})
	clerkSession.WithUser(coreModel)

	return clerkSession, nil
}
