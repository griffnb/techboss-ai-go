package helpers

import (
	"net/http"
	"strings"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/session"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/models/admin"
	clerk_service "github.com/griffnb/techboss-ai-go/internal/services/clerk"
	"github.com/pkg/errors"
)

// GetAdminSession Gets the admin session from the request checking header and cookie, priority is cookie
func HasClerkAdminSession(req *http.Request) bool {
	claims, ok := clerk.SessionClaimsFromContext(req.Context())
	if !ok {
		return false
	}

	return !tools.Empty(claims.Actor)
}

// GetClerkAdminSession Gets the admin session from clerk claims in the request
func getClerkAdminSession(req *http.Request) *session.Session {
	claims, ok := clerk.SessionClaimsFromContext(req.Context())
	if !ok {
		log.Debugf("No session claims found %s", req.URL.Path)
		return nil
	}

	customClaims, err := clerk_service.CustomClaims(claims)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return nil
	}

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
		log.Debugf("Admin role is too low: %d", customClaims.AdminRole)
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
	log.Debugf("Session loaded")
	return clerkSession
}

// GetClerkSession Gets the session from the request checking header and cookie, priority is cookie
func getClerkSession(req *http.Request) (*session.Session, error) {
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

func loadClerkAdmin(req *http.Request, adminSession *session.Session) (*admin.Admin, error) {
	admn, err := admin.Get(req.Context(), adminSession.User.ID())
	if err != nil {
		return nil, err
	}

	if tools.Empty(admn) {
		emailAdmin, err := admin.GetByEmail(req.Context(), adminSession.User.GetString("email"))
		if err != nil {
			return nil, err
		}

		// Fixes the case where we've wiped the admins, and clerk sees something diff for ID
		if !tools.Empty(emailAdmin) {
			admn = emailAdmin
			err := admin.RepairID(req.Context(), emailAdmin.ID(), adminSession.User.ID())
			if err != nil {
				return nil, err
			}
		} else {
			log.ErrorContext(errors.Errorf("admin not found %s", adminSession.User.ID()), req.Context())
			return nil, errors.New("Unauthorized")
		}
	}

	return admn, nil
}
