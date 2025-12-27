package sandbox

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	modalService "github.com/griffnb/techboss-ai-go/internal/services/modal"
	"github.com/pkg/errors"
)

// ClaudeRequest holds request data for Claude execution.
// It contains the prompt that will be sent to Claude Code CLI.
type ClaudeRequest struct {
	Prompt string `json:"prompt"`
}

// streamClaude executes Claude Code CLI in a sandbox and streams output using SSE.
// It parses the request body for the prompt, validates inputs, retrieves sandbox info,
// and streams the Claude output to the client in real-time.
//
// Currently uses in-memory cache for Phase 1 testing. Phase 2 will retrieve
// sandboxInfo from the database using the sandboxID parameter.
//
// The streaming uses Server-Sent Events (SSE) with the following flow:
// 1. Parse request and validate prompt
// 2. Retrieve sandboxInfo from cache (temporary)
// 3. Build ClaudeExecConfig
// 4. Call service.ExecuteClaudeStream which handles SSE headers and streaming
// 5. Service layer streams output line by line with [DONE] event at completion
func streamClaude(w http.ResponseWriter, req *http.Request) {
	// Get sandboxID from URL params
	sandboxID := chi.URLParam(req, "sandboxID")

	// Parse request body for prompt
	data, err := request.GetJSONPostAs[*ClaudeRequest](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		http.Error(w, "failed to parse request body", http.StatusBadRequest)
		return
	}

	// Validate prompt not empty
	if tools.Empty(data.Prompt) {
		http.Error(w, "prompt is required", http.StatusBadRequest)
		return
	}

	log.Infof("streamClaude called with sandboxID: %s, prompt: %s", sandboxID, data.Prompt)

	// Retrieve from in-memory cache (temporary solution)
	// TODO (Phase 2): Retrieve sandboxInfo from database instead of memory cache
	// This will enable sandbox retrieval across server restarts and multiple instances
	value, ok := sandboxCache.Load(sandboxID)
	if !ok {
		err := errors.Errorf("sandbox not found: %s", sandboxID)
		log.ErrorContext(err, req.Context())
		http.Error(w, "sandbox not found", http.StatusNotFound)
		return
	}

	sandboxInfo := value.(*modal.SandboxInfo)

	// Build ClaudeExecConfig with prompt
	// stream-json: structured JSON output format
	// SkipPermissions: bypasses permission prompts (safe in sandboxed environment)
	// Verbose: detailed logging for debugging
	claudeConfig := &modal.ClaudeExecConfig{
		Prompt:          data.Prompt,
		OutputFormat:    "stream-json",
		SkipPermissions: true,
		Verbose:         true,
	}

	// Execute Claude and stream via service layer
	service := modalService.NewSandboxService()
	err = service.ExecuteClaudeStream(req.Context(), sandboxInfo, claudeConfig, w)
	if err != nil {
		log.ErrorContext(err, req.Context())
		http.Error(w, "failed to execute Claude", http.StatusInternalServerError)
		return
	}
}
