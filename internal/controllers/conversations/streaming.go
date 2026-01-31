package conversations

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/router/request"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
	"github.com/griffnb/techboss-ai-go/internal/services/conversation_service"
)

// StreamRequest holds request data for Claude streaming.
// This struct defines the expected JSON body for streaming endpoints.
//
// Satisfies:
// - Design Phase 8.1: StreamRequest struct definition
// - Requirement 1.2: Support conversation-centric streaming
type StreamRequest struct {
	Prompt  string       `json:"prompt"`   // User prompt to send to Claude
	Type    sandbox.Type `json:"type"`     // Sandbox provider (defaults to PROVIDER_CLAUDE_CODE)
	AgentID types.UUID   `json:"agent_id"` // Agent ID for conversation
}

// streamClaude handles POST /conversation/{conversationId}/sandbox/{sandboxId}
//
// This endpoint implements the conversation-centric streaming architecture where:
// 1. Conversations are created or retrieved
// 2. Sandboxes are ensured to exist (with OnColdStart hook)
// 3. Streaming executes with full lifecycle hook coordination
// 4. OnColdStart failure returns error (CRITICAL)
//
// Satisfies:
// - Design Phase 8.1: Conversation streaming endpoint
// - Requirement 1.2-1.4: Conversation-centric architecture
// - Requirement 2.2: OnColdStart must succeed
//
// URL Parameters:
//
//	conversationId - Conversation UUID (created if doesn't exist)
//	sandboxId - Sandbox UUID (currently unused, reserved for future)
//
// Request Body (JSON):
//
//	prompt - User message to stream
//	provider - Sandbox provider type (optional, defaults to PROVIDER_CLAUDE_CODE)
//	agent_id - Agent ID for the conversation
//
// Response:
//
//	200 OK - Server-Sent Events stream
//	400 Bad Request - Invalid request (missing prompt, invalid JSON)
//	500 Internal Server Error - Conversation/sandbox creation failure
func streamClaude(w http.ResponseWriter, req *http.Request) {
	// Extract user session and URL parameters
	userSession := request.GetReqSession(req)
	accountID := userSession.User.ID()
	conversationID := types.UUID(chi.URLParam(req, "conversationId"))
	// sandboxID := types.UUID(chi.URLParam(req, "sandboxId")) // Reserved for future use

	// Parse request body
	data, err := request.GetJSONPostAs[*StreamRequest](req)
	if err != nil || data.Prompt == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Default provider to PROVIDER_CLAUDE_CODE if not specified
	if data.Type == 0 {
		data.Type = sandbox.TYPE_CLAUDE_CODE
	}

	// Initialize conversation service
	service := conversation_service.NewConversationService()

	// Get or create conversation
	conv, err := service.GetOrCreateConversation(req.Context(), conversationID, accountID, data.AgentID)
	if err != nil {
		log.ErrorContext(err, req.Context())
		http.Error(w, "failed to get conversation", http.StatusInternalServerError)
		return
	}

	// Ensure sandbox exists and is initialized
	// CRITICAL: OnColdStart failure will cause this to return error
	sandboxInfo, template, err := service.EnsureSandbox(req.Context(), conv, data.Type)
	if err != nil {
		log.ErrorContext(err, req.Context())
		http.Error(w, "failed to initialize sandbox", http.StatusInternalServerError)
		return
	}

	// Stream with full hook coordination
	// This handles: user message save, streaming, token tracking, OnStreamFinish hook
	err = service.StreamClaudeWithHooks(req.Context(), conversationID, sandboxInfo, template, data.Prompt, w)
	if err != nil {
		log.ErrorContext(err, req.Context())
		// Don't return error here - streaming may have already started
		// The error has been logged and the client will see connection close
	}
}
