package sandboxes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service"
)

// ClaudeRequest holds request data for Claude execution.
// It contains the prompt that will be sent to Claude Code CLI.
type ClaudeRequest struct {
	Prompt string `json:"prompt"`
}

// streamClaude executes Claude Code CLI in a sandbox and streams output using SSE.
// It parses the request body for the prompt, validates inputs, retrieves sandbox info
// from the database with ownership verification, and streams the Claude output to the
// client in real-time.
//
// The streaming uses Server-Sent Events (SSE) with the following flow:
// 1. Parse request and validate prompt
// 2. Retrieve sandboxInfo from database with ownership verification
// 3. Build ClaudeExecConfig
// 4. Call service.ExecuteClaudeStream which handles SSE headers and streaming
// 5. Service layer streams output line by line with [DONE] event at completion
func streamClaude(w http.ResponseWriter, req *http.Request) {
	userSession := request.GetReqSession(req)
	accountID := userSession.User.ID()
	id := chi.URLParam(req, "id")

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

	log.Infof("streamClaude called for sandbox ID: %s, prompt: %s", id, data.Prompt)

	// Get sandbox from database and verify ownership
	sandboxModel, err := sandbox.Get(req.Context(), types.UUID(id))
	if err != nil || sandboxModel == nil || sandboxModel.AccountID.Get() != accountID {
		log.ErrorContext(err, req.Context())
		http.Error(w, "sandbox not found", http.StatusNotFound)
		return
	}

	// Reconstruct SandboxInfo from model
	sandboxInfo := sandbox_service.ReconstructSandboxInfo(sandboxModel, accountID)

	// Build ClaudeExecConfig with prompt
	// stream-json: structured JSON output format
	// SkipPermissions: prevents interactive prompts in sandbox (safe with claudeuser setup)
	// Verbose: detailed logging for debugging
	claudeConfig := &modal.ClaudeExecConfig{
		Prompt:          data.Prompt,
		OutputFormat:    "stream-json",
		SkipPermissions: true,
		Verbose:         true,
	}

	// Execute Claude and stream via service layer
	service := sandbox_service.NewSandboxService()
	err = service.ExecuteClaudeStream(req.Context(), sandboxInfo, claudeConfig, w)
	if err != nil {
		log.ErrorContext(err, req.Context())
		http.Error(w, "failed to execute Claude", http.StatusInternalServerError)
		return
	}
}
