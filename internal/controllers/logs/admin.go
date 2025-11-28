package logs

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/environment"
)

func searchRecursive(_ http.ResponseWriter, req *http.Request) (any, int, error) {
	data := request.GetModelPostData(req)

	logGroup := chi.URLParam(req, "logGroup")
	startDate := tools.ParseStringI(data["start_date"])
	endDate := tools.ParseStringI(data["end_date"])
	query := tools.ParseStringI(data["q"])

	startTime, err := time.Parse(tools.TIME_YYYY_MM_DD, startDate)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[any](err)
	}
	endTime, err := time.Parse(tools.TIME_YYYY_MM_DD, endDate)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[any](err)
	}

	if startTime.Equal(endTime) {
		endTime = endTime.AddDate(0, 0, 1)
	}

	results, err := environment.GetCloudwatch().SearchRecursive(req.Context(), logGroup, query, startTime.Unix(), endTime.Unix(), 0)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[any](err)
	}

	resp := map[string]interface{}{
		"results": results,
	}
	return response.Success(resp)
}
