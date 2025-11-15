package logs

import (
	"net/http"
	"time"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/router/request"
	"github.com/CrowdShield/go-core/lib/router/response"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/go-chi/chi/v5"
	"github.com/griffnb/techboss-ai-go/internal/environment"
)

// { $.meta.objs[*].url = %a=VNgENE% } // find ad id
// { $.meta.objs[*].url = %pbk=[\-a-zA-Z_0-9]*-97063a6f970c% } // find postback key
func search(_ http.ResponseWriter, req *http.Request) (any, int, error) {
	rawdata := request.GetJSONPostData(req)
	data := request.ConvertPost(rawdata)

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

	results, err := environment.GetLogReader().QueryAndWait(req.Context(), logGroup, query, startTime.Unix(), endTime.Unix())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[any](err)
	}

	return response.Success(results)
}

func searchRecursive(_ http.ResponseWriter, req *http.Request) (any, int, error) {
	rawdata := request.GetJSONPostData(req)
	data := request.ConvertPost(rawdata)

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

	results, err := environment.GetLogReader().SearchRecursive(req.Context(), logGroup, query, startTime.Unix(), endTime.Unix(), 0)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return response.AdminBadRequestError[any](err)
	}

	resp := map[string]interface{}{
		"results": results,
	}
	return response.Success(resp)
}
