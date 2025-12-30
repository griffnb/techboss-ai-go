package conversation_service

import (
	"context"
	"testing"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/models/message"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service/lifecycle"
)

// Test_SaveUserMessage_createsUserMessage verifies user message creation
func Test_SaveUserMessage_createsUserMessage(t *testing.T) {
	ctx := context.Background()
	service := NewConversationService()
	conversationID := types.UUID("test-conv-001")
	prompt := "Test user prompt"

	// Mock sandbox info and template
	sandboxInfo := &modal.SandboxInfo{
		SandboxID: "test-sandbox",
		Config: &modal.SandboxConfig{
			AccountID: types.UUID("test-account-001"),
		},
	}
	template := &sandbox_service.SandboxTemplate{
		Hooks: &lifecycle.LifecycleHooks{
			OnMessage: nil, // No hook for this test
		},
	}

	// Act
	msg, err := service.SaveUserMessage(ctx, conversationID, sandboxInfo, template, prompt)

	// Assert
	assert.NoError(t, err)
	assert.NEmpty(t, msg)
	assert.Equal(t, conversationID, msg.ConversationID)
	assert.Equal(t, prompt, msg.Body)
	assert.Equal(t, int64(ROLE_USER), msg.Role)
	assert.Equal(t, int64(0), msg.Tokens)
}

// Test_SaveUserMessage_withHook verifies OnMessage hook is executed
func Test_SaveUserMessage_withHook(t *testing.T) {
	ctx := context.Background()
	service := NewConversationService()
	conversationID := types.UUID("test-conv-002")
	prompt := "Test user prompt"

	hookCalled := false
	var hookMessage *message.Message

	// Mock sandbox info and template with hook
	sandboxInfo := &modal.SandboxInfo{
		SandboxID: "test-sandbox",
		Config: &modal.SandboxConfig{
			AccountID: types.UUID("test-account-002"),
		},
	}
	template := &sandbox_service.SandboxTemplate{
		Hooks: &lifecycle.LifecycleHooks{
			OnMessage: func(_ context.Context, hookData *lifecycle.HookData) error {
				hookCalled = true
				hookMessage = hookData.Message
				return nil
			},
		},
	}

	// Act
	msg, err := service.SaveUserMessage(ctx, conversationID, sandboxInfo, template, prompt)

	// Assert
	assert.NoError(t, err)
	assert.NEmpty(t, msg)
	assert.True(t, hookCalled)
	assert.NEmpty(t, hookMessage)
	assert.Equal(t, conversationID, hookMessage.ConversationID)
	assert.Equal(t, prompt, hookMessage.Body)
	assert.Equal(t, int64(ROLE_USER), hookMessage.Role)
}

// Test_SaveUserMessage_emptyPrompt verifies empty prompts are allowed
func Test_SaveUserMessage_emptyPrompt(t *testing.T) {
	ctx := context.Background()
	service := NewConversationService()
	conversationID := types.UUID("test-conv-003")
	prompt := ""

	sandboxInfo := &modal.SandboxInfo{
		SandboxID: "test-sandbox",
		Config: &modal.SandboxConfig{
			AccountID: types.UUID("test-account-003"),
		},
	}
	template := &sandbox_service.SandboxTemplate{
		Hooks: &lifecycle.LifecycleHooks{},
	}

	// Act
	msg, err := service.SaveUserMessage(ctx, conversationID, sandboxInfo, template, prompt)

	// Assert
	assert.NoError(t, err)
	assert.NEmpty(t, msg)
	assert.Equal(t, "", msg.Body)
}

// Test_SaveAssistantMessage_createsAssistantMessage verifies assistant message creation
func Test_SaveAssistantMessage_createsAssistantMessage(t *testing.T) {
	ctx := context.Background()
	service := NewConversationService()
	conversationID := types.UUID("test-conv-004")
	response := "Test assistant response"
	tokens := int64(150)

	// Mock sandbox info and template
	sandboxInfo := &modal.SandboxInfo{
		SandboxID: "test-sandbox",
		Config: &modal.SandboxConfig{
			AccountID: types.UUID("test-account-004"),
		},
	}
	template := &sandbox_service.SandboxTemplate{
		Hooks: &lifecycle.LifecycleHooks{
			OnMessage: nil, // No hook for this test
		},
	}

	// Act
	msg, err := service.SaveAssistantMessage(ctx, conversationID, sandboxInfo, template, response, tokens)

	// Assert
	assert.NoError(t, err)
	assert.NEmpty(t, msg)
	assert.Equal(t, conversationID, msg.ConversationID)
	assert.Equal(t, response, msg.Body)
	assert.Equal(t, int64(ROLE_ASSISTANT), msg.Role)
	assert.Equal(t, tokens, msg.Tokens)
}

// Test_SaveAssistantMessage_withHook verifies OnMessage hook is executed
func Test_SaveAssistantMessage_withHook(t *testing.T) {
	ctx := context.Background()
	service := NewConversationService()
	conversationID := types.UUID("test-conv-005")
	response := "Test assistant response"
	tokens := int64(200)

	hookCalled := false
	var hookMessage *message.Message

	// Mock sandbox info and template with hook
	sandboxInfo := &modal.SandboxInfo{
		SandboxID: "test-sandbox",
		Config: &modal.SandboxConfig{
			AccountID: types.UUID("test-account-005"),
		},
	}
	template := &sandbox_service.SandboxTemplate{
		Hooks: &lifecycle.LifecycleHooks{
			OnMessage: func(_ context.Context, hookData *lifecycle.HookData) error {
				hookCalled = true
				hookMessage = hookData.Message
				return nil
			},
		},
	}

	// Act
	msg, err := service.SaveAssistantMessage(ctx, conversationID, sandboxInfo, template, response, tokens)

	// Assert
	assert.NoError(t, err)
	assert.NEmpty(t, msg)
	assert.True(t, hookCalled)
	assert.NEmpty(t, hookMessage)
	assert.Equal(t, conversationID, hookMessage.ConversationID)
	assert.Equal(t, response, hookMessage.Body)
	assert.Equal(t, int64(ROLE_ASSISTANT), hookMessage.Role)
	assert.Equal(t, tokens, hookMessage.Tokens)
}

// Test_SaveAssistantMessage_zeroTokens verifies zero tokens are allowed
func Test_SaveAssistantMessage_zeroTokens(t *testing.T) {
	ctx := context.Background()
	service := NewConversationService()
	conversationID := types.UUID("test-conv-006")
	response := "Test response"
	tokens := int64(0)

	sandboxInfo := &modal.SandboxInfo{
		SandboxID: "test-sandbox",
		Config: &modal.SandboxConfig{
			AccountID: types.UUID("test-account-006"),
		},
	}
	template := &sandbox_service.SandboxTemplate{
		Hooks: &lifecycle.LifecycleHooks{},
	}

	// Act
	msg, err := service.SaveAssistantMessage(ctx, conversationID, sandboxInfo, template, response, tokens)

	// Assert
	assert.NoError(t, err)
	assert.NEmpty(t, msg)
	assert.Equal(t, int64(0), msg.Tokens)
}

// Test_RoleConstants_values verifies role constant values match design spec
func Test_RoleConstants_values(t *testing.T) {
	// Verify role constants match the design specification
	assert.Equal(t, int64(1), int64(ROLE_USER))
	assert.Equal(t, int64(2), int64(ROLE_ASSISTANT))
	assert.Equal(t, int64(3), int64(ROLE_TOOL))
}
