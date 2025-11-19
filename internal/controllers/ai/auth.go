package ai

import (
	"net/http"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/techboss-ai-go/internal/services/ai_proxies/openai"
)

// SetupAdmin sets up admin routes

func authRun(w http.ResponseWriter, req *http.Request) {
	service, err := openai.NewServiceFromEnv()
	if err != nil {
		log.ErrorContext(err, req.Context())

		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	err = service.ProxyNonStreaming(req.Context(), req, w)
	if err != nil {
		log.ErrorContext(err, req.Context())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func authStream(w http.ResponseWriter, req *http.Request) {
	service, err := openai.NewServiceFromEnv()
	if err != nil {
		log.ErrorContext(err, req.Context())

		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	err = service.ProxyStreaming(req.Context(), req, w)
	if err != nil {
		log.ErrorContext(err, req.Context())

		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
