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
	Model    string `json:"model"` // Model to use (optional, default handled by service)
	Messages []struct {
		Role    string `json:"role"` // Role of the message (e.g., "user", "assistant")
		Content []struct {
			Type string `json:"type"` // Type of content (e.g., "text")
			Text string `json:"text"`
		} `json:"content"` // Content of the message
	} `json:"messages"` // Conversation messages
	// Prompt string `json:"prompt"` // Prompt to send to Claude Code CLI
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
func adminStreamClaude(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	// Parse request body for prompt
	data, err := request.GetJSONPostAs[*ClaudeRequest](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		http.Error(w, "failed to parse request body", http.StatusBadRequest)
		return
	}

	var prompt string
	if len(data.Messages) > 0 && len(data.Messages[0].Content) > 0 {
		prompt = data.Messages[0].Content[0].Text
	}

	// Validate prompt not empty
	if tools.Empty(prompt) {
		http.Error(w, "prompt is required", http.StatusBadRequest)
		return
	}

	log.Infof("streamClaude called for sandbox ID: %s, prompt: %s", id, prompt)

	// Get sandbox from database and verify ownership
	sandboxModel, err := sandbox.Get(req.Context(), types.UUID(id))
	if err != nil || tools.Empty(sandboxModel) {
		log.ErrorContext(err, req.Context())
		http.Error(w, "sandbox not found", http.StatusNotFound)
		return
	}

	// Reconstruct SandboxInfo from model
	sandboxInfo, err := sandbox_service.ReconstructSandboxInfo(req.Context(), sandboxModel, sandboxModel.AccountID.Get())
	if err != nil {
		log.ErrorContext(err, req.Context())
		http.Error(w, "failed to reconstruct sandbox info", http.StatusInternalServerError)
		return
	}

	log.PrintEntity(sandboxInfo, "Reconstructed SandboxInfo")

	// Build ClaudeExecConfig with prompt
	// stream-json: structured JSON output format
	// SkipPermissions: prevents interactive prompts in sandbox (safe with claudeuser setup)
	// Verbose: detailed logging for debugging
	claudeConfig := &modal.ClaudeExecConfig{
		Prompt:          prompt,
		OutputFormat:    "stream-json",
		SkipPermissions: true,
		Verbose:         true,
	}

	// Execute Claude and stream via service layer
	service := sandbox_service.NewSandboxService()
	claudeProcess, err := service.ExecuteClaudeStream(req.Context(), sandboxInfo, claudeConfig, w)
	if err != nil {
		log.ErrorContext(err, req.Context())
		http.Error(w, "failed to execute Claude", http.StatusInternalServerError)
		return
	}

	// ClaudeProcess contains token usage information (InputTokens, OutputTokens, CacheTokens)
	// populated during streaming by parsing the final summary event.
	// For now, we just log it. Future enhancements could save this to database.
	log.Infof("Claude execution completed - Input: %d, Output: %d, Cache: %d tokens",
		claudeProcess.InputTokens, claudeProcess.OutputTokens, claudeProcess.CacheTokens)
}
