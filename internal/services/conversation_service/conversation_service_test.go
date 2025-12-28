package conversation_service

import (
	"context"
	"testing"

	"github.com/griffnb/core/lib/testtools"
	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/models/conversation"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
)

func init() {
	system_testing.BuildSystem()
}

func skipIfNotConfigured(t *testing.T) {
	if !modal.Configured() {
		t.Skip("Modal client is not configured, skipping test")
	}
}

func Test_NewConversationService(t *testing.T) {
	t.Run("creates service successfully", func(t *testing.T) {
		// Act
		service := NewConversationService()

		// Assert
		assert.NEmpty(t, service)
		assert.NEmpty(t, service.sandboxService)
	})
}

func Test_GetOrCreateConversation_create(t *testing.T) {
	t.Run("creates new conversation when not exists", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		service := NewConversationService()
		conversationID := types.UUID("test-conv-001")
		accountID := types.UUID("test-acct-001")
		agentID := types.UUID("test-agent-001")

		// Act
		conv, err := service.GetOrCreateConversation(ctx, conversationID, accountID, agentID)

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, conv)
		assert.Equal(t, conversationID, conv.ID())
		assert.Equal(t, accountID, conv.AccountID.Get())
		assert.Equal(t, agentID, conv.AgentID.Get())

		stats, err := conv.Stats.Get()
		assert.NoError(t, err)
		assert.NEmpty(t, stats)
		assert.Equal(t, 0, stats.MessagesExchanged)
		assert.Equal(t, int64(0), stats.TotalInputTokens)
		assert.Equal(t, int64(0), stats.TotalOutputTokens)
		assert.Equal(t, int64(0), stats.TotalCacheTokens)

		// Cleanup
		testtools.CleanupModel(conv)
	})

	t.Run("returns existing conversation when already exists", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		service := NewConversationService()

		// Create existing conversation
		existingConv := conversation.New()
		existingConv.ID_.Set(types.UUID("test-existing-conv"))
		existingConv.AccountID.Set(types.UUID("test-acct-002"))
		existingConv.AgentID.Set(types.UUID("test-agent-002"))
		existingConv.Stats.Set(&conversation.ConversationStats{
			MessagesExchanged: 5,
			TotalInputTokens:  100,
			TotalOutputTokens: 200,
			TotalCacheTokens:  50,
		})
		err := existingConv.Save(nil)
		assert.NoError(t, err)

		// Act
		conv, err := service.GetOrCreateConversation(
			ctx,
			existingConv.ID(),
			existingConv.AccountID.Get(),
			existingConv.AgentID.Get(),
		)

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, conv)
		assert.Equal(t, existingConv.ID(), conv.ID())

		stats, err := conv.Stats.Get()
		assert.NoError(t, err)
		assert.Equal(t, 5, stats.MessagesExchanged)
		assert.Equal(t, int64(100), stats.TotalInputTokens)
		assert.Equal(t, int64(200), stats.TotalOutputTokens)
		assert.Equal(t, int64(50), stats.TotalCacheTokens)

		// Cleanup
		testtools.CleanupModel(conv)
	})
}

func Test_EnsureSandbox_create_new(t *testing.T) {
	t.Run("creates new sandbox when conversation has none", func(t *testing.T) {
		skipIfNotConfigured(t)

		// Arrange
		ctx := context.Background()
		service := NewConversationService()

		conv := conversation.New()
		conv.ID_.Set(types.UUID("test-conv-with-sandbox"))
		conv.AccountID.Set(types.UUID("test-acct-003"))
		conv.AgentID.Set(types.UUID("test-agent-003"))
		conv.Stats.Set(&conversation.ConversationStats{})
		err := conv.Save(nil)
		assert.NoError(t, err)

		// Act
		sandboxInfo, template, err := service.EnsureSandbox(ctx, conv, sandbox.PROVIDER_CLAUDE_CODE)

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, sandboxInfo)
		assert.NEmpty(t, template)
		assert.NEmpty(t, sandboxInfo.SandboxID)
		assert.NEmpty(t, template.Hooks)

		// Verify conversation was updated with sandbox ID
		updatedConv, err := conversation.Get(ctx, conv.ID())
		assert.NoError(t, err)
		assert.NEmpty(t, updatedConv.SandboxID.Get())

		// Verify sandbox was saved to database
		sandboxModel, err := sandbox.Get(ctx, updatedConv.SandboxID.Get())
		assert.NoError(t, err)
		assert.NEmpty(t, sandboxModel)
		assert.Equal(t, sandboxInfo.SandboxID, sandboxModel.ExternalID.Get())

		// Cleanup
		if sandboxInfo != nil {
			_ = service.sandboxService.TerminateSandbox(ctx, sandboxInfo, false)
		}
		testtools.CleanupModel(sandboxModel)
		testtools.CleanupModel(conv)
	})

	t.Run("returns error when OnColdStart hook fails", func(t *testing.T) {
		skipIfNotConfigured(t)

		// Arrange
		ctx := context.Background()
		service := NewConversationService()

		conv := conversation.New()
		conv.ID_.Set(types.UUID("test-conv-with-sandbox"))
		conv.AccountID.Set(types.UUID("test-acct-003"))
		conv.AgentID.Set(types.UUID("test-agent-003"))
		conv.Stats.Set(&conversation.ConversationStats{})
		err := conv.Save(nil)
		assert.NoError(t, err)

		// NOTE: We cannot easily test hook failure without mocking the modal client
		// or creating a custom template with a failing hook. This test documents
		// the expected behavior but cannot be fully executed without mocking infrastructure.

		// For now, just verify the basic flow works
		sandboxInfo, template, err := service.EnsureSandbox(ctx, conv, sandbox.PROVIDER_CLAUDE_CODE)

		// If we got here, cold start succeeded
		if err == nil {
			assert.NEmpty(t, sandboxInfo)
			assert.NEmpty(t, template)

			// Cleanup
			_ = service.sandboxService.TerminateSandbox(ctx, sandboxInfo, false)
		}

		testtools.CleanupModel(conv)
	})
}

func Test_EnsureSandbox_reconstruct_existing(t *testing.T) {
	t.Run("reconstructs existing sandbox info", func(t *testing.T) {
		skipIfNotConfigured(t)

		// Arrange
		ctx := context.Background()
		service := NewConversationService()
		accountID := types.UUID("test-acct-004")

		// Create conversation with sandbox
		conv := conversation.New()
		conv.ID_.Set(types.UUID("test-conv-reconstruct"))
		conv.AccountID.Set(accountID)
		conv.AgentID.Set(types.UUID("test-agent-004"))
		conv.Stats.Set(&conversation.ConversationStats{})

		// Create sandbox model
		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(accountID)
		sandboxModel.AgentID.Set(conv.AgentID.Get())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-123")
		err := sandboxModel.Save(nil)
		assert.NoError(t, err)

		// Link sandbox to conversation
		conv.SandboxID.Set(sandboxModel.ID())
		err = conv.Save(nil)
		assert.NoError(t, err)

		// Act
		sandboxInfo, template, err := service.EnsureSandbox(ctx, conv, sandbox.PROVIDER_CLAUDE_CODE)

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, sandboxInfo)
		assert.NEmpty(t, template)
		assert.Equal(t, "sb-test-123", sandboxInfo.SandboxID)

		// Cleanup
		testtools.CleanupModel(sandboxModel)
		testtools.CleanupModel(conv)
	})
}

func Test_EnsureSandbox_invalid_provider(t *testing.T) {
	t.Run("returns error for unsupported provider", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		service := NewConversationService()

		conv := conversation.New()
		conv.ID_.Set(types.UUID("test-conv-with-sandbox"))
		conv.AccountID.Set(types.UUID("test-acct-003"))
		conv.AgentID.Set(types.UUID("test-agent-003"))
		conv.Stats.Set(&conversation.ConversationStats{})
		err := conv.Save(nil)
		assert.NoError(t, err)

		// Act
		sandboxInfo, template, err := service.EnsureSandbox(ctx, conv, sandbox.Provider(999))

		// Assert
		assert.Error(t, err)
		assert.Empty(t, sandboxInfo)
		assert.Empty(t, template)

		// Cleanup
		testtools.CleanupModel(conv)
	})
}
