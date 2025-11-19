package helpers

import (
	"net/http"

	"github.com/griffnb/core/lib/session"
	"github.com/griffnb/core/lib/tools"
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

	if tools.Empty(admn) {
		return nil, errors.Errorf("Admin not found: %s", adminSession.User.ID())
	}

	return admn, nil
}
