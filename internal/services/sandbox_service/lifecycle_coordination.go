package sandbox_service

import (
	"context"

	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/models/message"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service/lifecycle"
)

// ExecuteColdStartHook runs the OnColdStart hook for sandbox initialization.
// This hook is CRITICAL - it returns errors to the caller, which will fail sandbox creation.
// If the hook fails, the sandbox should not be created as files would be out of sync.
//
// Satisfies:
// - Requirement 4.1-4.10: Cold start hook execution with critical error handling
// - Requirement 6.2: Hook context passing for OnColdStart
// - Requirement 6.3: Hook execution when registered
// - Requirement 6.4: Error logging with full context
// - Requirement 9.1: Critical hook failures prevent sandbox creation
// - Design Phase 6.2: ExecuteColdStartHook implementation
func (s *SandboxService) ExecuteColdStartHook(
	ctx context.Context,
	conversationID types.UUID,
	sandboxInfo *modal.SandboxInfo,
	template *SandboxTemplate,
) error {
	if template.Hooks == nil || template.Hooks.OnColdStart == nil {
		return nil
	}

	hookData := &lifecycle.HookData{
		ConversationID: conversationID,
		SandboxInfo:    sandboxInfo,
	}

	// Hook determines criticality - if it returns error, we propagate it
	return lifecycle.ExecuteHook(ctx, "OnColdStart", template.Hooks.OnColdStart, hookData)
}

// ExecuteMessageHook runs the OnMessage hook for message persistence.
// This hook is NON-CRITICAL - it ignores return values.
// The hook implementation decides whether to return an error (critical) or swallow it (non-critical).
// For message saving, errors are typically swallowed since message content is in the stream.
//
// Satisfies:
// - Requirement 6.2: Hook context passing for OnMessage
// - Requirement 6.3: Hook execution when registered
// - Requirement 9.2: Non-critical hooks don't fail operations
// - Design Phase 6.2: ExecuteMessageHook ignores return
func (s *SandboxService) ExecuteMessageHook(
	ctx context.Context,
	conversationID types.UUID,
	sandboxInfo *modal.SandboxInfo,
	template *SandboxTemplate,
	msg *message.Message,
) {
	if template.Hooks == nil || template.Hooks.OnMessage == nil {
		return
	}

	hookData := &lifecycle.HookData{
		ConversationID: conversationID,
		SandboxInfo:    sandboxInfo,
		Message:        msg,
	}

	// Hook swallows errors if non-critical, returns nil
	// We ignore the return value - hook decides its own criticality
	_ = lifecycle.ExecuteHook(ctx, "OnMessage", template.Hooks.OnMessage, hookData)
}

// ExecuteStreamFinishHook runs the OnStreamFinish hook after streaming completes.
// This hook is NON-CRITICAL - it ignores return values.
// The hook implementation decides whether to return an error (critical) or swallow it (non-critical).
// For stream finish operations like S3 sync and stats updates, errors are typically swallowed
// since the user has already received their streaming response.
//
// Satisfies:
// - Requirement 5.1-5.10: Stream finish hook with S3 sync and token tracking
// - Requirement 6.2: Hook context passing for OnStreamFinish
// - Requirement 6.3: Hook execution when registered
// - Requirement 9.3: Non-critical hooks don't fail streaming
// - Design Phase 6.2: ExecuteStreamFinishHook ignores return
func (s *SandboxService) ExecuteStreamFinishHook(
	ctx context.Context,
	conversationID types.UUID,
	sandboxInfo *modal.SandboxInfo,
	template *SandboxTemplate,
	tokenUsage *lifecycle.TokenUsage,
) {
	if template.Hooks == nil || template.Hooks.OnStreamFinish == nil {
		return
	}

	hookData := &lifecycle.HookData{
		ConversationID: conversationID,
		SandboxInfo:    sandboxInfo,
		TokenUsage:     tokenUsage,
	}

	// Hook swallows errors if non-critical, returns nil
	// We ignore the return value - hook decides its own criticality
	_ = lifecycle.ExecuteHook(ctx, "OnStreamFinish", template.Hooks.OnStreamFinish, hookData)
}

// ExecuteTerminateHook runs the OnTerminate hook when a sandbox is terminated.
// This hook returns errors if the hook determines the failure is critical.
// The hook implementation decides criticality by its return value.
// Default implementation is non-critical, but custom hooks may need critical cleanup.
//
// Satisfies:
// - Requirement 6.2: Hook context passing for OnTerminate
// - Requirement 6.3: Hook execution when registered
// - Requirement 6.4: Error logging with full context
// - Design Phase 6.2: ExecuteTerminateHook returns errors if hook determines critical
func (s *SandboxService) ExecuteTerminateHook(
	ctx context.Context,
	conversationID types.UUID,
	sandboxInfo *modal.SandboxInfo,
	template *SandboxTemplate,
) error {
	if template.Hooks == nil || template.Hooks.OnTerminate == nil {
		return nil
	}

	hookData := &lifecycle.HookData{
		ConversationID: conversationID,
		SandboxInfo:    sandboxInfo,
	}

	// Hook determines criticality - if it returns error, we propagate it
	return lifecycle.ExecuteHook(ctx, "OnTerminate", template.Hooks.OnTerminate, hookData)
}
