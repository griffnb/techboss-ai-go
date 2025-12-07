package helpers

import (
	"net/http"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/pkg/errors"
)

func ApiAuthRequestWrapper(fn http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		receivedSignature := req.Header.Get("x-techboss-api")
		if receivedSignature == "" {
			log.ErrorContext(errors.Errorf("Received signature is empty %s:%s", "x-techboss-api", receivedSignature), req.Context())
			response.ErrorWrapper(res, req, "Internal Error", http.StatusInternalServerError)
			return
		}

		if receivedSignature != environment.GetConfig().InternalAPIKey {
			log.ErrorContext(errors.Errorf("Received signature does not match secret key"), req.Context())
			response.ErrorWrapper(res, req, "Internal Error", http.StatusInternalServerError)
			return
		}

		fn(res, req)
	})
}
