package ai_tools

import (
	"net/http"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/router"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/models/ai_tool"
)

func authCount(_ http.ResponseWriter, req *http.Request) (int64, int, error) {
	userSession := helpers.GetReqSession(req)
	user := userSession.User
	parameters := router.BuildIndexParams(req.Context(), req.URL.Query(), TABLE_NAME)
	ai_tool.AddJoinData(parameters)
	count, err := ai_tool.CountRestricted(req.Context(), parameters, user)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return helpers.AdminBadRequestError[int64](err)
	}

	return helpers.Success(count)
}
