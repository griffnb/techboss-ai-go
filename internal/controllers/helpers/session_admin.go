package helpers

import (
	"net/http"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/session"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/models/admin"
	"github.com/pkg/errors"
)

func getAdminSession(req *http.Request) (*session.Session, error) {
	apiSession, err := apiLoginAdmin(req)
	if err != nil {
		return nil, err
	}
	if !tools.Empty(apiSession) {
		return apiSession, nil
	}

	adminSession := getCustomAdminSession(req)
	if !tools.Empty(adminSession) {
		return adminSession, nil
	}

	return nil, nil
}

func loadAdmin(req *http.Request, adminSession *session.Session) (*admin.Admin, error) {
	admn, err := admin.Get(req.Context(), adminSession.User.ID())
	if err != nil {
		return nil, err
	}

	if tools.Empty(admn) && adminSession.IsClerk() {
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
