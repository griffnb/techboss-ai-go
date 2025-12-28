package lifecycle

import (
	"context"

	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/models/message"
)

// HookFunc defines the signature for lifecycle hooks.
// Hooks are executed at specific points in the sandbox lifecycle.
// A hook determines its own criticality by returning an error (critical) or nil (success/non-critical).
// Context can be used for cancellation and timeout control.
type HookFunc func(ctx context.Context, hookData *HookData) error

// HookData contains context data passed to lifecycle hooks.
// It provides access to conversation, sandbox, message, and token information
// depending on which lifecycle phase is executing.
type HookData struct {
	// ConversationID links the hook execution to a specific conversation
	ConversationID types.UUID

	// SandboxInfo provides access to the sandbox environment and configuration
	SandboxInfo *modal.SandboxInfo

	// Message is populated for OnMessage hook, contains the message being processed
	Message *message.Message

	// TokenUsage is populated for OnStreamFinish hook, contains token consumption data
	TokenUsage *TokenUsage
}

// TokenUsage tracks token consumption from Claude streaming responses.
// These values are parsed from Claude's final summary event at the end of streaming.
type TokenUsage struct {
	// InputTokens represents tokens consumed for input (prompt and context)
	InputTokens int64

	// OutputTokens represents tokens generated in the response
	OutputTokens int64

	// CacheTokens represents tokens served from cache (reduces cost)
	CacheTokens int64
}

// LifecycleHooks defines the complete set of hooks for sandbox lifecycle management.
// Each hook is optional and can be nil if not needed.
// Templates register hooks to customize behavior at each lifecycle phase.
type LifecycleHooks struct {
	// OnColdStart is executed when a new sandbox is created before first use.
	// This is typically used for initialization like syncing from S3.
	// Errors from this hook are considered critical and will fail sandbox creation.
	OnColdStart HookFunc

	// OnMessage is executed when a message is saved to the conversation.
	// This is typically used for message persistence and statistics tracking.
	// Errors from this hook should be swallowed by the hook implementation if non-critical.
	OnMessage HookFunc

	// OnStreamFinish is executed after Claude streaming completes successfully.
	// This is typically used for S3 sync and token usage tracking.
	// Errors from this hook should be swallowed by the hook implementation if non-critical.
	OnStreamFinish HookFunc

	// OnTerminate is executed when a sandbox is being shut down.
	// This is typically used for cleanup operations.
	// Errors from this hook should be swallowed by the hook implementation if non-critical.
	OnTerminate HookFunc
}
