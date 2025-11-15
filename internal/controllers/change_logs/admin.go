package change_logs

import (
	"net/http"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/router/request"
	"github.com/CrowdShield/go-core/lib/router/response"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/models/admin"
	"github.com/griffnb/techboss-ai-go/internal/models/change_log"
)

const DYNAMO_CHANGE_LOGS = true

func adminIndex(_ http.ResponseWriter, req *http.Request) (interface{}, int, error) {
	if !DYNAMO_CHANGE_LOGS {
		parameters := request.BuildIndexParams(req.Context(), req.URL.Query(), TABLE_NAME)

		if tools.Empty(parameters.Limit) {
			parameters.Limit = constants.SYSTEM_LIMIT
		}

		coreModels, err := change_log.FindAllJoined(req.Context(), parameters)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.AdminBadRequestError[any](err)

		}

		return response.Success(coreModels)
	}

	limit := constants.SYSTEM_LIMIT

	logs, err := change_log.GetChangeLogsByObjectURN(req.Context(), req.URL.Query().Get("object_urn"), int32(limit))
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[any](err)
	}

	logRecords, err := change_log.BuildChangeLogs(req.Context(), logs)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[any](err)
	}

	return response.Success(logRecords)
}

func adminCount(_ http.ResponseWriter, req *http.Request) (interface{}, int, error) {
	if !DYNAMO_CHANGE_LOGS {
		parameters := request.BuildIndexParams(req.Context(), req.URL.Query(), TABLE_NAME)
		admin.AddJoinData(parameters)
		count, err := change_log.FindResultsCount(req.Context(), parameters)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return response.AdminBadRequestError[any](err)
		}

		return response.Success(count)
	}

	return response.Success(constants.SYSTEM_LIMIT)
}
