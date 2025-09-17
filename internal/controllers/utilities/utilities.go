package utilities

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/router"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/pkg/errors"
)

func uploadURL(_ http.ResponseWriter, req *http.Request) (any, int, error) {
	fileName := req.URL.Query().Get("name")
	fileType := req.URL.Query().Get("type")

	uploadName := fmt.Sprintf("assets/%s/%s", tools.SessionKey(), fileName)

	url, err := environment.GetS3().GetPreSignedPutURL(environment.GetConfig().S3Config.Buckets["assets"], uploadName, fileType)
	if err != nil {
		log.ErrorContext(err, req.Context())
		return helpers.AdminBadRequestError[any](err)
	}
	return helpers.Success(map[string]string{"url": url})
}

func testError(_ http.ResponseWriter, req *http.Request) (any, int, error) {
	log.ErrorContext(errors.Errorf("This is a test error"), req.Context())
	fmt.Println("This is a test fatal error")
	return nil, http.StatusInternalServerError, errors.Errorf("This is a test error")
}

func hookLog(res http.ResponseWriter, req *http.Request) {
	rawdata := router.GetJSONPostData(req)
	params := router.FixParams(req.URL.Query())

	logData := map[string]interface{}{
		"post_data": rawdata,
		"params":    params,
		"headers":   req.Header,
	}

	log.PrintEntity(logData)

	log.Slack("dev-errors", logData)
	log.Info("Hook Log", logData)

	responseData := map[string]interface{}{
		"response": "im an AI response i think from local",
	}

	bytes, err := json.Marshal(responseData)
	if err != nil {
		log.ErrorContext(err, req.Context())
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = res.Write(bytes)
}
