package conversation_service

import (
	"context"
	"time"

	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/models/message"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service"
)

// Message role constants define the types of conversation participants.
// These constants are used in the Message.Role field to identify who sent the message.
//
// Satisfies:
// - Requirement 2.1: Role identification for conversation messages
// - Design Phase 7.2: Role constant definitions
const (
	ROLE_USER      = 1 // User-generated messages (prompts from the user)
	ROLE_ASSISTANT = 2 // Assistant-generated messages (responses from Claude)
	ROLE_TOOL      = 3 // Tool execution results and tool calls
)

// SaveUserMessage creates a user message and executes the OnMessage lifecycle hook.
// The OnMessage hook is NON-CRITICAL - failures are logged but don't prevent message creation.
// The message is not persisted to DynamoDB by this function; persistence happens in the hook.
//
// Parameters:
// - ctx: Context for cancellation and timeout control
// - conversationID: UUID of the conversation this message belongs to
// - sandboxInfo: Sandbox information for hook execution context
// - template: Sandbox template containing lifecycle hooks
// - prompt: The user's message content
//
// Returns:
// - *message.Message: The created message object
// - error: Always returns nil (hook errors are swallowed)
//
// Satisfies:
// - Requirement 2.1: User message creation with role identification
// - Requirement 2.2: Message timestamp generation
// - Requirement 2.3: Message conversation linkage
// - Requirement 2.4: Zero token count for user messages
// - Requirement 2.5: OnMessage hook execution for persistence
// - Requirement 2.6: Non-critical hook failure handling
// - Design Phase 7.2: SaveUserMessage implementation
func (s *ConversationService) SaveUserMessage(
	ctx context.Context,
	conversationID types.UUID,
	sandboxInfo *modal.SandboxInfo,
	template *sandbox_service.SandboxTemplate,
	prompt string,
) (*message.Message, error) {
	msg := &message.Message{
		Key:            tools.SessionKey(),
		ConversationID: conversationID,
		Body:           prompt,
		Role:           ROLE_USER,
		Timestamp:      time.Now().Unix(),
		Tokens:         0, // User messages don't consume tokens
		ToolCalls:      nil,
	}

	// Execute OnMessage hook (non-critical)
	// The hook handles message persistence to DynamoDB
	// Failures are logged by the hook but don't prevent message creation
	s.sandboxService.ExecuteMessageHook(ctx, conversationID, sandboxInfo, template, msg)

	return msg, nil
}

// SaveAssistantMessage creates an assistant message with token usage and executes the OnMessage hook.
// The OnMessage hook is NON-CRITICAL - failures are logged but don't prevent message creation.
// The message is not persisted to DynamoDB by this function; persistence happens in the hook.
//
// Parameters:
// - ctx: Context for cancellation and timeout control
// - conversationID: UUID of the conversation this message belongs to
// - sandboxInfo: Sandbox information for hook execution context
// - template: Sandbox template containing lifecycle hooks
// - response: The assistant's message content
// - tokens: Total tokens consumed for this response (input + output + cache)
//
// Returns:
// - *message.Message: The created message object
// - error: Always returns nil (hook errors are swallowed)
//
// Satisfies:
// - Requirement 2.1: Assistant message creation with role identification
// - Requirement 2.2: Message timestamp generation
// - Requirement 2.3: Message conversation linkage
// - Requirement 2.4: Token tracking per message
// - Requirement 2.5: OnMessage hook execution for persistence
// - Requirement 2.6: Non-critical hook failure handling
// - Design Phase 7.2: SaveAssistantMessage implementation
func (s *ConversationService) SaveAssistantMessage(
	ctx context.Context,
	conversationID types.UUID,
	sandboxInfo *modal.SandboxInfo,
	template *sandbox_service.SandboxTemplate,
	response string,
	tokens int64,
) (*message.Message, error) {
	msg := &message.Message{
		Key:            tools.SessionKey(),
		ConversationID: conversationID,
		Body:           response,
		Role:           ROLE_ASSISTANT,
		Timestamp:      time.Now().Unix(),
		Tokens:         tokens,
		ToolCalls:      nil,
	}

	// Execute OnMessage hook (non-critical)
	// The hook handles message persistence to DynamoDB and stats updates
	// Failures are logged by the hook but don't prevent message creation
	s.sandboxService.ExecuteMessageHook(ctx, conversationID, sandboxInfo, template, msg)

	return msg, nil
}
