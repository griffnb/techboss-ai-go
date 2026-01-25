package conversation_service

import (
	"context"
	"net/http"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service/lifecycle"
	"github.com/pkg/errors"
)

// StreamClaudeWithHooks coordinates Claude streaming with full lifecycle hook support.
// It handles the complete flow: save user message, execute streaming, track tokens, call hooks.
//
// This function implements the conversation-centric streaming approach where:
// 1. User messages are saved (non-critical - logged if fails)
// 2. Claude streaming executes (critical - returns error if fails)
// 3. Token usage is extracted from ClaudeProcess (Phase 9.1 - COMPLETE)
// 4. Response body is captured during streaming (Phase 9.2 - COMPLETE)
// 5. Assistant message is saved with response and tokens
// 6. OnStreamFinish hook is called (non-critical - logged if fails)
//
// Satisfies:
// - Requirement 5.1-5.10: Stream coordination with hooks
// - Requirement 2.1-2.6: Message management during streaming
// - Design Phase 7.3: Streaming coordination
// - Design Phase 9.1: Token extraction via ClaudeProcess
// - Design Phase 9.2: Response capture during streaming
func (s *ConversationService) StreamClaudeWithHooks(
	ctx context.Context,
	conversationID types.UUID,
	sandboxInfo *modal.SandboxInfo,
	template *sandbox_service.SandboxTemplate,
	prompt string,
	responseWriter http.ResponseWriter,
) error {
	// Validate inputs
	if sandboxInfo == nil {
		return errors.New("sandboxInfo cannot be nil")
	}
	if template == nil {
		return errors.New("template cannot be nil")
	}
	if prompt == "" {
		return errors.New("prompt cannot be empty")
	}
	if responseWriter == nil {
		return errors.New("responseWriter cannot be nil")
	}

	// 1. Save user message (non-critical - log but continue if fails)
	_, err := s.SaveUserMessage(ctx, conversationID, sandboxInfo, template, prompt)
	if err != nil {
		log.Errorf(err, "Failed to save user message for conversation %s", conversationID)
		// Continue - message save is non-critical for streaming
	}

	// 2. Build Claude config for streaming
	claudeConfig := &modal.ClaudeExecConfig{
		Prompt:          prompt,
		OutputFormat:    "stream-json",
		SkipPermissions: true,
		Verbose:         true,
	}

	// 3. Execute and stream Claude (critical - return error if fails)
	// This returns ClaudeProcess with captured response and token usage
	claudeProcess, err := s.sandboxService.ExecuteClaudeStream(ctx, sandboxInfo, claudeConfig, responseWriter)
	if err != nil {
		return errors.Wrapf(err, "failed to stream Claude for conversation %s", conversationID)
	}

	// 4. Extract token usage from Claude process
	// ExecuteClaudeStream populates the token fields when it parses the final summary event
	tokenUsage := &lifecycle.TokenUsage{
		InputTokens:  claudeProcess.InputTokens,
		OutputTokens: claudeProcess.OutputTokens,
		CacheTokens:  claudeProcess.CacheTokens,
	}

	// 5. Calculate total tokens for assistant message
	// Total tokens is sum of output tokens (what Claude generated)
	totalTokens := claudeProcess.OutputTokens

	// 6. Save assistant message with captured response body and token count
	// ResponseBody is captured during streaming by ExecuteClaudeStream
	_, err = s.SaveAssistantMessage(ctx, conversationID, sandboxInfo, template, claudeProcess.ResponseBody, totalTokens)
	if err != nil {
		log.Errorf(err, "Failed to save assistant message for conversation %s", conversationID)
		// Continue - message save is non-critical for streaming success
	}

	// 7. Execute OnStreamFinish hook (non-critical - logged but doesn't fail request)
	s.sandboxService.ExecuteStreamFinishHook(ctx, conversationID, sandboxInfo, template, tokenUsage)

	return nil
}
