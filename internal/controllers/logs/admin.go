package logs

import (
	"net/http"
	"time"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/router"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/go-chi/chi/v5"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/environment"
)

// { $.meta.objs[*].url = %a=VNgENE% } // find ad id
// { $.meta.objs[*].url = %pbk=[\-a-zA-Z_0-9]*-97063a6f970c% } // find postback key
func search(_ http.ResponseWriter, req *http.Request) (any, int, error) {
	rawdata := router.GetJSONPostData(req)
	data := helpers.ConvertPost(rawdata)

	logGroup := chi.URLParam(req, "logGroup")
	startDate := tools.ParseStringI(data["start_date"])
	endDate := tools.ParseStringI(data["end_date"])
	query := tools.ParseStringI(data["q"])

	startTime, err := time.Parse(tools.TIME_YYYY_MM_DD, startDate)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return helpers.AdminBadRequestError[any](err)
	}
	endTime, err := time.Parse(tools.TIME_YYYY_MM_DD, endDate)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return helpers.AdminBadRequestError[any](err)
	}

	if startTime.Equal(endTime) {
		endTime = endTime.AddDate(0, 0, 1)
	}

	results, err := environment.GetLogReader().QueryAndWait(req.Context(), logGroup, query, startTime.Unix(), endTime.Unix())
	if err != nil {
		log.ErrorContext(err, req.Context())
		return helpers.AdminBadRequestError[any](err)
	}

	return helpers.Success(results)
}

func searchRecursive(_ http.ResponseWriter, req *http.Request) (any, int, error) {
	rawdata := router.GetJSONPostData(req)
	data := helpers.ConvertPost(rawdata)

	logGroup := chi.URLParam(req, "logGroup")
	startDate := tools.ParseStringI(data["start_date"])
	endDate := tools.ParseStringI(data["end_date"])
	query := tools.ParseStringI(data["q"])

	startTime, err := time.Parse(tools.TIME_YYYY_MM_DD, startDate)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return helpers.AdminBadRequestError[any](err)
	}
	endTime, err := time.Parse(tools.TIME_YYYY_MM_DD, endDate)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return helpers.AdminBadRequestError[any](err)
	}

	if startTime.Equal(endTime) {
		endTime = endTime.AddDate(0, 0, 1)
	}

	results, err := environment.GetLogReader().SearchRecursive(req.Context(), logGroup, query, startTime.Unix(), endTime.Unix(), 0)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return helpers.AdminBadRequestError[any](err)
	}

	response := map[string]interface{}{
		"results": results,
	}
	return helpers.Success(response)
}

/*
func searchStream(res http.ResponseWriter, req *http.Request) {

	rawdata := router.GetJSONPostData(req)
	data := helpers.ConvertPost(rawdata)

	logGroup := chi.URLParam(req, "logGroup")
	startDate := tools.ParseStringI(data["start_date"])
	endDate := tools.ParseStringI(data["end_date"])
	query := tools.ParseStringI(data["q"])

	limit := tools.ParseIntI(data["limit"])
	if limit == 0 {
		limit = 1000
	}

	startTime, err := time.Parse(tools.TIME_YYYY_MM_DD, startDate)
	if err != nil {
		log.ErrorContext(err, req.Context())
		helpers.ErrorWrapper(res, req, err.Error(), http.StatusInternalServerError)
		return
	}
	endTime, err := time.Parse(tools.TIME_YYYY_MM_DD, endDate)
	if err != nil {
		log.ErrorContext(err, req.Context())
		helpers.ErrorWrapper(res, req, err.Error(), http.StatusInternalServerError)
		return
	}

	if startTime.Equal(endTime) {
		endTime = endTime.AddDate(0, 0, 1)
	}

	results, err := environment.GetLogReader().QueryAndWait(req.Context(), logGroup, query, startTime.Unix(), endTime.Unix(), 0)
	if err != nil {
		log.ErrorContext(err, req.Context())
		helpers.ErrorWrapper(res, req, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ensure the response supports streaming
	res.Header().Set("Content-Type", "text/plain")
	res.Header().Set("Transfer-Encoding", "chunked")

	flusher, ok := res.(http.Flusher)
	if !ok {
		http.Error(res, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Simulate streaming data
	for i := 1; i <= 5; i++ {
		fmt.Fprintf(res, "Chunk %d: This is part of the streamed response\n", i)
		flusher.Flush() // Flush the buffer to send data to the client immediately
		time.Sleep(1 * time.Second)
	}

	fmt.Fprintln(res, "Streaming complete!")
	return helpers.Success(response)
}
*/
