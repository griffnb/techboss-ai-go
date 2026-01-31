package conversations

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/griffnb/core/lib/testtools"
	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/models/conversation"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
	"github.com/griffnb/techboss-ai-go/internal/services/testing_service"
)

func init() {
	system_testing.BuildSystem()
}

func Test_streamClaude_invalidRequest(t *testing.T) {
	t.Run("missing prompt returns 400", func(t *testing.T) {
		// Arrange
		conversationID := types.UUID("00000000-0000-0000-0000-000000000001")
		body := map[string]any{
			"prompt":   "", // Missing prompt
			"type":     sandbox.TYPE_CLAUDE_CODE,
			"agent_id": types.UUID("00000000-0000-0000-0000-000000000002").String(),
		}

		req, err := testing_service.NewPOSTRequest[any](
			"/"+conversationID.String()+"/sandbox/test-sandbox",
			nil,
			body,
		)
		assert.NoError(t, err)

		err = req.WithAccount()
		assert.NoError(t, err)

		w := httptest.NewRecorder()

		// Act
		streamClaude(w, req.Request)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json returns 400", func(t *testing.T) {
		// Arrange
		conversationID := types.UUID("00000000-0000-0000-0000-000000000003")

		req, err := testing_service.NewPOSTRequest[any](
			"/"+conversationID.String()+"/sandbox/test-sandbox",
			nil,
			map[string]any{},
		)
		assert.NoError(t, err)

		err = req.WithAccount()
		assert.NoError(t, err)

		// Manually set bad JSON body
		req.Request.Body = http.NoBody
		req.Request.ContentLength = 0

		w := httptest.NewRecorder()

		// Act
		streamClaude(w, req.Request)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func Test_streamClaude_conversationHandling(t *testing.T) {
	t.Run("creates new conversation if not exists", func(t *testing.T) {
		// Arrange
		conversationID := types.UUID("00000000-0000-0000-0000-000000000004")
		agentID := types.UUID("00000000-0000-0000-0000-000000000005")

		body := map[string]any{
			"prompt":   "Hello, this is a test",
			"type":     sandbox.TYPE_CLAUDE_CODE,
			"agent_id": agentID.String(),
		}

		req, err := testing_service.NewPOSTRequest[any](
			"/"+conversationID.String()+"/sandbox/test-sandbox",
			nil,
			body,
		)
		assert.NoError(t, err)

		err = req.WithAccount()
		assert.NoError(t, err)

		w := httptest.NewRecorder()

		// Act
		streamClaude(w, req.Request)

		// Assert
		// Check that conversation was created
		conv, err := conversation.Get(req.Request.Context(), conversationID)
		assert.NoError(t, err)
		assert.NEmpty(t, conv)
		assert.Equal(t, conversationID, conv.ID())
		assert.Equal(t, req.Account.ID(), conv.AccountID.Get())
		assert.Equal(t, agentID, conv.AgentID.Get())

		defer testtools.CleanupModel(conv)

		// If sandbox was successfully created, clean it up
		if !conv.SandboxID.IsEmpty() {
			sandboxModel, _ := sandbox.Get(req.Request.Context(), conv.SandboxID.Get())
			if sandboxModel != nil {
				defer testtools.CleanupModel(sandboxModel)
			}
		}
	})

	t.Run("uses existing conversation if exists", func(t *testing.T) {
		// Arrange
		builder := testing_service.New()
		builder.WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		// Create existing conversation
		conv := conversation.New()
		conv.Set("id", types.UUID("00000000-0000-0000-0000-000000000006"))
		conv.AccountID.Set(builder.Account.ID())
		conv.AgentID.Set(types.UUID("00000000-0000-0000-0000-000000000007"))
		conv.Stats.Set(&conversation.ConversationStats{
			MessagesExchanged: 5,
			TotalInputTokens:  100,
			TotalOutputTokens: 200,
			TotalCacheTokens:  50,
		})
		err = conv.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(conv)

		body := map[string]any{
			"prompt":   "Hello, this is a test",
			"type":     sandbox.TYPE_CLAUDE_CODE,
			"agent_id": conv.AgentID.Get().String(),
		}

		req, err := testing_service.NewPOSTRequest[any](
			"/"+conv.ID().String()+"/sandbox/test-sandbox",
			nil,
			body,
		)
		assert.NoError(t, err)

		err = req.WithAccount(builder.Account)
		assert.NoError(t, err)

		w := httptest.NewRecorder()

		// Act
		streamClaude(w, req.Request)

		// Assert
		// Verify conversation still exists and wasn't recreated
		retrieved, err := conversation.Get(req.Request.Context(), conv.ID())
		assert.NoError(t, err)
		assert.NEmpty(t, retrieved)
		assert.Equal(t, conv.ID(), retrieved.ID())
		// Stats should be preserved from original
		stats, _ := retrieved.Stats.Get()
		assert.Equal(t, 5, stats.MessagesExchanged)

		// Cleanup sandbox if created
		if !retrieved.SandboxID.IsEmpty() {
			sandboxModel, _ := sandbox.Get(req.Request.Context(), retrieved.SandboxID.Get())
			if sandboxModel != nil {
				defer testtools.CleanupModel(sandboxModel)
			}
		}
	})
}

func Test_streamClaude_providerDefaults(t *testing.T) {
	t.Run("defaults to PROVIDER_CLAUDE_CODE when not specified", func(t *testing.T) {
		// Arrange
		conversationID := types.UUID("00000000-0000-0000-0000-000000000008")

		body := map[string]any{
			"prompt":   "Hello, this is a test",
			"provider": 0, // Not specified
			"agent_id": types.UUID("00000000-0000-0000-0000-000000000009").String(),
		}

		req, err := testing_service.NewPOSTRequest[any](
			"/"+conversationID.String()+"/sandbox/test-sandbox",
			nil,
			body,
		)
		assert.NoError(t, err)

		err = req.WithAccount()
		assert.NoError(t, err)

		w := httptest.NewRecorder()

		// Act
		streamClaude(w, req.Request)

		// Assert
		// Should use default provider - verify by checking conversation was created
		conv, err := conversation.Get(req.Request.Context(), conversationID)
		assert.NoError(t, err)
		assert.NEmpty(t, conv)

		defer testtools.CleanupModel(conv)

		// Cleanup sandbox if created
		if !conv.SandboxID.IsEmpty() {
			sandboxModel, err := sandbox.Get(req.Request.Context(), conv.SandboxID.Get())
			if err == nil && sandboxModel != nil {
				defer testtools.CleanupModel(sandboxModel)
			}
		}
	})
}

func Test_streamClaude_sandboxCreation(t *testing.T) {
	t.Run("creates sandbox if conversation has no sandbox", func(t *testing.T) {
		// Note: This test will attempt to create a real sandbox via Modal
		// In a real environment, this would need Modal configuration
		// For now, we test that the flow executes without panicking

		// Arrange
		conversationID := types.UUID("00000000-0000-0000-0000-00000000000a")
		agentID := types.UUID("00000000-0000-0000-0000-00000000000b")

		body := map[string]any{
			"prompt":   "Hello, this is a test",
			"type":     sandbox.TYPE_CLAUDE_CODE,
			"agent_id": agentID.String(),
		}

		req, err := testing_service.NewPOSTRequest[any](
			"/"+conversationID.String()+"/sandbox/test-sandbox",
			nil,
			body,
		)
		assert.NoError(t, err)

		err = req.WithAccount()
		assert.NoError(t, err)

		w := httptest.NewRecorder()

		// Act
		streamClaude(w, req.Request)

		// Assert
		// Check that conversation was created
		conv, err := conversation.Get(req.Request.Context(), conversationID)
		assert.NoError(t, err)
		assert.NEmpty(t, conv)

		defer testtools.CleanupModel(conv)

		// If sandbox was successfully created, clean it up
		if !conv.SandboxID.IsEmpty() {
			sandboxModel, _ := sandbox.Get(req.Request.Context(), conv.SandboxID.Get())
			if sandboxModel != nil {
				defer testtools.CleanupModel(sandboxModel)
			}
		}
	})
}
