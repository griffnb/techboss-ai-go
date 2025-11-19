package helpers

import (
	"net/http"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/session"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/admin"

	"github.com/pkg/errors"
)

const (
	API_KEY_HEADER   = "x-api-key"
	API_EMAIL_HEADER = "x-admin-email"
)

// Uses OAUTH to login users across the system
func apiLoginAdmin(req *http.Request) (*session.Session, error) {
	if tools.Empty(req.Header.Get(API_KEY_HEADER)) {
		return nil, nil
	}

	if req.Header.Get(API_KEY_HEADER) != environment.GetConfig().InternalAPIKey {
		return nil, errors.Errorf("invalid api key")
	}

	if tools.Empty(req.Header.Get(API_EMAIL_HEADER)) {
		return nil, errors.Errorf("missing admin email")
	}

	adminEmail := req.Header.Get(API_EMAIL_HEADER)

	adminObj, err := admin.GetByEmail(req.Context(), adminEmail)
	// Query error occured
	if err != nil {
		log.ErrorContext(err, req.Context())
		return nil, errors.Wrap(err, "Failed to get admin")
	}

	if tools.Empty(adminObj) {
		return nil, errors.Errorf("No admin for email %s", adminEmail)
	}

	// create a session and set its value as the same as the token
	userSession := session.New(tools.ParseStringI(req.Context().Value("ip"))).WithUser(adminObj)
	userSession.Key = tools.SessionKey()
	err = userSession.Save()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to save session")
	}

	return userSession, nil
}
