package ai_tools

import (
	"net/http"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/router/request"
	"github.com/CrowdShield/go-core/lib/router/response"
	"github.com/griffnb/techboss-ai-go/internal/models/ai_tool"
)

func authCount(_ http.ResponseWriter, req *http.Request) (int64, int, error) {
	userSession := request.GetReqSession(req)
	user := userSession.User
	parameters := request.BuildIndexParams(req.Context(), req.URL.Query(), TABLE_NAME)
	ai_tool.AddJoinData(parameters)
	count, err := ai_tool.CountRestricted(req.Context(), parameters, user)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[int64](err)
	}

	return response.Success(count)
}
