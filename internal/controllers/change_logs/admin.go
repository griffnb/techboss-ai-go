package change_logs

import (
	"net/http"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/router"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/models/admin"
	"github.com/griffnb/techboss-ai-go/internal/models/change_log"
)

const DYNAMO_CHANGE_LOGS = true

func adminIndex(_ http.ResponseWriter, req *http.Request) (interface{}, int, error) {
	if !DYNAMO_CHANGE_LOGS {
		parameters := router.BuildIndexParams(req.Context(), req.URL.Query(), TABLE_NAME)

		if tools.Empty(parameters.Limit) {
			parameters.Limit = constants.SYSTEM_LIMIT
		}

		coreModels, err := change_log.FindAllJoinedWithContext(req.Context(), parameters)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return helpers.AdminBadRequestError[any](err)

		}

		return helpers.Success(coreModels)
	}

	limit := constants.SYSTEM_LIMIT

	logs, err := change_log.GetChangeLogsByObjectURN(req.Context(), req.URL.Query().Get("object_urn"), int32(limit))
	if err != nil {
		log.ErrorContext(err, req.Context())
		return helpers.AdminBadRequestError[any](err)
	}

	logRecords, err := change_log.BuildChangeLogs(req.Context(), logs)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return helpers.AdminBadRequestError[any](err)
	}

	return helpers.Success(logRecords)
}

func adminCount(_ http.ResponseWriter, req *http.Request) (interface{}, int, error) {
	if !DYNAMO_CHANGE_LOGS {
		parameters := router.BuildIndexParams(req.Context(), req.URL.Query(), TABLE_NAME)
		admin.AddJoinData(parameters)
		count, err := change_log.FindResultsCount(req.Context(), parameters)
		if err != nil {
			log.ErrorContext(err, req.Context())
			return helpers.AdminBadRequestError[any](err)
		}

		return helpers.Success(count)
	}

	return helpers.Success(constants.SYSTEM_LIMIT)
}
