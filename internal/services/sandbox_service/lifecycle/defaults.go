package lifecycle

import (
	"context"
	"fmt"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/models/conversation"
	"github.com/pkg/errors"
)

// DefaultOnColdStart performs S3 sync if configured.
// This hook is CRITICAL - it returns errors to the caller, which will fail sandbox creation.
// If S3 sync fails during cold start, the sandbox should not be created.
//
// Satisfies:
// - Requirement 2.1-2.5: Cold start synchronization with state tracking
// - Requirement 6.7: Default hook implementations
// - Design Phase 5.3: DefaultOnColdStart returns errors (critical)
func DefaultOnColdStart(ctx context.Context, hookData *HookData) error {
	if hookData.SandboxInfo.Config.S3Config == nil {
		// No S3 configured, skip sync
		return nil
	}

	// Perform state-based sync from S3
	// This is critical - if sync fails, sandbox should not be created
	client := modal.Client()
	_, err := client.InitVolumeFromS3WithState(ctx, hookData.SandboxInfo)
	if err != nil {
		// Return error to caller (critical failure)
		log.Error(errors.Wrapf(err, "Critical: Cold start S3 sync failed for conversation %s", string(hookData.ConversationID)))
		return err
	}

	log.Info(fmt.Sprintf("Cold start S3 sync completed successfully for conversation %s", string(hookData.ConversationID)))
	return nil
}

// DefaultOnMessage saves message to DynamoDB and updates conversation stats.
// This hook is NON-CRITICAL - it swallows errors and returns nil.
// If message save fails, we log the error but allow the operation to continue.
//
// Satisfies:
// - Requirement 1.3-1.5: Message storage and tracking
// - Requirement 6.7: Default hook implementations
// - Requirement 9.2: Non-critical hooks swallow errors
// - Design Phase 5.3: DefaultOnMessage swallows errors (non-critical)
func DefaultOnMessage(ctx context.Context, hookData *HookData) error {
	if hookData.Message == nil {
		// No message to save
		return nil
	}

	// Save message to DynamoDB
	err := hookData.Message.Save(ctx)
	if err != nil {
		log.Error(errors.Wrap(err, "Non-critical: Failed to save message but continuing"))
		return nil // Swallow error - message save is non-critical
	}

	// Increment message counter in conversation stats
	conv, err := conversation.Get(ctx, hookData.ConversationID)
	if err != nil {
		log.Error(errors.Wrap(err, "Non-critical: Failed to get conversation for stats update but continuing"))
		return nil // Swallow error
	}

	if conv == nil {
		log.Error(errors.New("Non-critical: Conversation not found for stats update but continuing"))
		return nil // Swallow error
	}

	// Update stats
	stats, _ := conv.Stats.Get()
	if stats == nil {
		stats = &conversation.ConversationStats{
			MessagesExchanged: 0,
			TotalInputTokens:  0,
			TotalOutputTokens: 0,
			TotalCacheTokens:  0,
		}
	}
	stats.IncrementMessages()
	conv.Stats.Set(stats)

	err = conv.Save(nil)
	if err != nil {
		log.Error(errors.Wrap(err, "Non-critical: Failed to save conversation stats but continuing"))
		return nil // Swallow error
	}

	log.Info(fmt.Sprintf("Message saved and stats updated for conversation %s", string(hookData.ConversationID)))
	return nil
}

// DefaultOnStreamFinish syncs to S3 and updates conversation stats with token usage.
// This hook is NON-CRITICAL - it swallows errors and returns nil.
// If S3 sync or stats update fails, we log the error but return success.
//
// Satisfies:
// - Requirement 5.1-5.3: Upload with state file tracking
// - Requirement 1.3: Token usage tracking
// - Requirement 6.7: Default hook implementations
// - Requirement 9.3: Non-critical hooks swallow errors
// - Design Phase 5.3: DefaultOnStreamFinish swallows errors (non-critical)
func DefaultOnStreamFinish(ctx context.Context, hookData *HookData) error {
	// Sync to S3 if configured
	if hookData.SandboxInfo.Config.S3Config != nil {
		client := modal.Client()
		_, err := client.SyncVolumeToS3WithState(ctx, hookData.SandboxInfo)
		if err != nil {
			log.Error(errors.Wrap(err, "Non-critical: Failed to sync to S3 but continuing"))
			// Continue to update stats even if sync fails
		} else {
			log.Info(fmt.Sprintf("S3 sync completed successfully for conversation %s", string(hookData.ConversationID)))
		}
	}

	// Update conversation stats with token usage
	if hookData.TokenUsage != nil {
		conv, err := conversation.Get(ctx, hookData.ConversationID)
		if err != nil {
			log.Error(errors.Wrap(err, "Non-critical: Failed to get conversation for token stats update but continuing"))
			return nil // Swallow error
		}

		if conv == nil {
			log.Error(errors.New("Non-critical: Conversation not found for token stats update but continuing"))
			return nil // Swallow error
		}

		// Update token stats
		stats, _ := conv.Stats.Get()
		if stats == nil {
			stats = &conversation.ConversationStats{
				MessagesExchanged: 0,
				TotalInputTokens:  0,
				TotalOutputTokens: 0,
				TotalCacheTokens:  0,
			}
		}
		stats.AddTokenUsage(
			hookData.TokenUsage.InputTokens,
			hookData.TokenUsage.OutputTokens,
			hookData.TokenUsage.CacheTokens,
		)
		conv.Stats.Set(stats)

		err = conv.Save(nil)
		if err != nil {
			log.Error(errors.Wrap(err, "Non-critical: Failed to save conversation token stats but continuing"))
			return nil // Swallow error
		}

		// Log both message and accumulated conversation token stats (Requirement 10.7)
		log.Info(fmt.Sprintf("[Token Stats] Message: input=%d output=%d cache=%d | Conversation total: input=%d output=%d cache=%d | conversation=%s",
			hookData.TokenUsage.InputTokens,
			hookData.TokenUsage.OutputTokens,
			hookData.TokenUsage.CacheTokens,
			stats.TotalInputTokens,
			stats.TotalOutputTokens,
			stats.TotalCacheTokens,
			string(hookData.ConversationID),
		))
	}

	return nil
}

// DefaultOnTerminate performs cleanup operations when a sandbox is terminated.
// This hook is NON-CRITICAL - it swallows errors and returns nil.
// The default implementation performs no cleanup, but if cleanup is added later,
// errors should be logged but not propagated.
//
// Satisfies:
// - Requirement 6.7: Default hook implementations
// - Requirement 9.3: Non-critical hooks swallow errors
// - Design Phase 5.3: DefaultOnTerminate returns nil (non-critical)
func DefaultOnTerminate(_ context.Context, hookData *HookData) error {
	// Default: no special cleanup needed
	// If cleanup is added in the future, swallow errors and log them
	log.Info(fmt.Sprintf("Sandbox terminating for conversation %s", string(hookData.ConversationID)))
	return nil
}
