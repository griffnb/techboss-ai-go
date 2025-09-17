package admins

import (
	"net/http"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/router"
	"github.com/CrowdShield/go-core/lib/session"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/models/admin"
)

func adminMe(_ http.ResponseWriter, req *http.Request) (*admin.AdminJoined, int, error) {
	userSession := req.Context().Value(router.SessionContextKey("session")).(*session.Session)
	coreModel, err := admin.GetJoined(req.Context(), userSession.User.ID())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return helpers.AdminBadRequestError[*admin.AdminJoined](err)
	}

	return helpers.Success(coreModel)
}
