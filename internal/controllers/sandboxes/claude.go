package sandboxes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service"
	"github.com/pkg/errors"
)

// ContentItem represents a single content block with type and text.
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// MessageContent handles both string and array content formats.
type MessageContent []ContentItem

// UnmarshalJSON implements custom unmarshaling to handle content as either string or array.
func (mc *MessageContent) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*mc = MessageContent{{Type: "text", Text: str}}
		return nil
	}

	// If not a string, try as array of ContentItem
	var items []ContentItem
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}
	*mc = items
	return nil
}

// ClaudeRequest holds request data for Claude execution.
// It contains the prompt that will be sent to Claude Code CLI.
type ClaudeRequest struct {
	Model    string `json:"model"` // Model to use (optional, default handled by service)
	Messages []struct {
		Role    string         `json:"role"`    // Role of the message (e.g., "user", "assistant")
		Content MessageContent `json:"content"` // Content of the message (string or array)
	} `json:"messages"` // Conversation messages
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
// 4. Call service.ExecuteClaudeStream which sets SSE headers and streams formatted output
// 5. Service layer calls claude.ProcessStream to emit typed SSE events per Vercel AI SDK spec
func adminStreamClaude(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	// Parse request body for prompt
	data, err := request.GetJSONPostAs[*ClaudeRequest](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		http.Error(w, "failed to parse request body", http.StatusBadRequest)
		return
	}

	// Find the last message with role "user"
	var prompt string
	for i := len(data.Messages) - 1; i >= 0; i-- {
		if data.Messages[i].Role == "user" && len(data.Messages[i].Content) > 0 {
			prompt = data.Messages[i].Content[0].Text
			break
		}
	}

	// Validate prompt not empty
	if tools.Empty(prompt) {
		log.ErrorContext(errors.Errorf("prompt is required"), req.Context())
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

// streamClaude executes Claude Code CLI in a sandbox and streams output using SSE.
// It parses the request body for the prompt, validates inputs, retrieves sandbox info
// from the database with ownership verification, and streams the Claude output to the
// client in real-time.
//
// The streaming uses Server-Sent Events (SSE) with the following flow:
// 1. Parse request and validate prompt
// 2. Retrieve sandboxInfo from database with ownership verification
// 3. Build ClaudeExecConfig
// 4. Call service.ExecuteClaudeStream which sets SSE headers and streams formatted output
// 5. Service layer calls claude.ProcessStream to emit typed SSE events per Vercel AI SDK spec
func authStreamClaude(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")

	usr := helpers.GetLoadedUser(req)

	// Parse request body for prompt
	data, err := request.GetJSONPostAs[*ClaudeRequest](req)
	if err != nil {
		log.ErrorContext(err, req.Context())
		http.Error(w, "failed to parse request body", http.StatusBadRequest)
		return
	}

	// Find the last message with role "user"
	var prompt string
	for i := len(data.Messages) - 1; i >= 0; i-- {
		if data.Messages[i].Role == "user" && len(data.Messages[i].Content) > 0 {
			prompt = data.Messages[i].Content[0].Text
			break
		}
	}

	// Validate prompt not empty
	if tools.Empty(prompt) {
		log.ErrorContext(errors.Errorf("prompt is required"), req.Context())
		http.Error(w, "prompt is required", http.StatusBadRequest)
		return
	}

	log.Infof("streamClaude called for sandbox ID: %s, prompt: %s", id, prompt)

	// Get sandbox from database and verify ownership
	sandboxModel, err := sandbox.GetRestricted(req.Context(), types.UUID(id), usr)
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

	// log.PrintEntity(sandboxInfo, "Reconstructed SandboxInfo")

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
