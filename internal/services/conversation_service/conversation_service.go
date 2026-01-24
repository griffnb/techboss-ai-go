// Package conversation_service provides business logic for conversation management.
// It coordinates conversation lifecycle, sandbox creation, and lifecycle hook execution.
// This service acts as the orchestrator between conversations, sandboxes, and messages.
package conversation_service

import (
	"context"

	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/models/conversation"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service"
	"github.com/pkg/errors"
)

// ConversationService handles conversation business logic and lifecycle management.
// It provides high-level operations for creating conversations, managing sandboxes,
// and coordinating hook execution throughout the conversation lifecycle.
type ConversationService struct {
	sandboxService *sandbox_service.SandboxService
}

// NewConversationService creates a new conversation service with initialized dependencies.
// The service uses the sandbox service for all sandbox-related operations.
func NewConversationService() *ConversationService {
	return &ConversationService{
		sandboxService: sandbox_service.NewSandboxService(),
	}
}

// GetOrCreateConversation retrieves an existing conversation or creates a new one.
// If the conversation exists, it returns the existing record with current stats.
// If it doesn't exist, a new conversation is created with zero stats.
//
// This method is idempotent - calling it multiple times with the same ID returns
// the same conversation without modification.
//
// Satisfies:
// - Requirement 1.1: Conversation creation with account and agent linkage
// - Requirement 1.3: Initialize stats structure with zero values
// - Design Phase 7.1: GetOrCreateConversation implementation
func (s *ConversationService) GetOrCreateConversation(
	ctx context.Context,
	conversationID types.UUID,
	accountID types.UUID,
	agentID types.UUID,
) (*conversation.Conversation, error) {
	// Try to get existing conversation
	conv, err := conversation.Get(ctx, conversationID)
	if err == nil && conv != nil {
		return conv, nil
	}

	// Create new conversation with zero stats
	conv = conversation.New()
	conv.ID_.Set(conversationID)
	conv.AccountID.Set(accountID)
	conv.AgentID.Set(agentID)
	conv.Stats.Set(&conversation.ConversationStats{
		MessagesExchanged: 0,
		TotalInputTokens:  0,
		TotalOutputTokens: 0,
		TotalCacheTokens:  0,
		TotalTokensUsed:   0, // Deprecated but set for backward compatibility
	})

	err = conv.Save(nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to save new conversation")
	}

	return conv, nil
}

// EnsureSandbox ensures the conversation has an active sandbox.
// If the conversation already has a sandbox, it reconstructs the sandbox info from the database.
// If not, it creates a new sandbox and executes the OnColdStart hook.
//
// CRITICAL: The OnColdStart hook is executed after sandbox creation. If the hook fails,
// the sandbox is terminated and cleaned up to prevent orphaned resources. This ensures
// sandboxes are only created if initialization (like S3 sync) succeeds.
//
// The returned SandboxInfo and SandboxTemplate can be used for subsequent operations
// like message handling and Claude execution.
//
// Satisfies:
// - Requirement 1.2: Link sandbox to conversation
// - Requirement 4.1-4.10: Execute OnColdStart hook for sandbox initialization
// - Requirement 9.1: Clean up sandbox on hook failure
// - Design Phase 7.1: EnsureSandbox implementation with OnColdStart
func (s *ConversationService) EnsureSandbox(
	ctx context.Context,
	conv *conversation.Conversation,
	provider sandbox.Provider,
) (*modal.SandboxInfo, *sandbox_service.SandboxTemplate, error) {
	// If conversation already has a sandbox, reconstruct it
	if !conv.SandboxID.IsEmpty() {
		sandboxModel, err := sandbox.Get(ctx, conv.SandboxID.Get())
		if err == nil && sandboxModel != nil {
			sandboxInfo, err := sandbox_service.ReconstructSandboxInfo(ctx, sandboxModel, conv.AccountID.Get())
			if err != nil {
				return nil, nil, errors.Wrap(err, "failed to reconstruct sandbox info")
			}
			template, err := sandbox_service.GetSandboxTemplate(provider, conv.AgentID.Get())
			if err != nil {
				return nil, nil, errors.Wrap(err, "failed to get sandbox template")
			}
			return sandboxInfo, template, nil
		}
		// If sandbox model not found, fall through to create new one
	}

	// Get template for provider
	template, err := sandbox_service.GetSandboxTemplate(provider, conv.AgentID.Get())
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get sandbox template")
	}

	// Build sandbox config from template
	config := template.BuildSandboxConfig(conv.AccountID.Get())

	// Create sandbox via sandbox service
	sandboxInfo, err := s.sandboxService.CreateSandbox(ctx, conv.AccountID.Get(), config)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create sandbox")
	}

	// Execute OnColdStart hook (CRITICAL - must succeed)
	err = s.sandboxService.ExecuteColdStartHook(ctx, conv.ID(), sandboxInfo, template)
	if err != nil {
		// Clean up sandbox on cold start failure
		cleanupErr := s.sandboxService.TerminateSandbox(ctx, sandboxInfo, false)
		if cleanupErr != nil {
			// Log cleanup error but return original error
			return nil, nil, errors.Wrapf(err, "OnColdStart failed and cleanup also failed: %v", cleanupErr)
		}
		return nil, nil, errors.Wrap(err, "OnColdStart hook failed")
	}

	// Save sandbox to database
	sandboxModel := sandbox.New()
	sandboxModel.AccountID.Set(conv.AccountID.Get())
	sandboxModel.AgentID.Set(conv.AgentID.Get())
	sandboxModel.ExternalID.Set(sandboxInfo.SandboxID)
	sandboxModel.Provider.Set(provider)
	err = sandboxModel.Save(nil)
	if err != nil {
		// Try to clean up sandbox on database save failure
		_ = s.sandboxService.TerminateSandbox(ctx, sandboxInfo, false)
		return nil, nil, errors.Wrap(err, "failed to save sandbox to database")
	}

	// Link sandbox to conversation
	conv.SandboxID.Set(sandboxModel.ID())
	err = conv.Save(nil)
	if err != nil {
		// Sandbox is created and saved, so we don't clean it up
		// Just return the error
		return nil, nil, errors.Wrap(err, "failed to link sandbox to conversation")
	}

	return sandboxInfo, template, nil
}
