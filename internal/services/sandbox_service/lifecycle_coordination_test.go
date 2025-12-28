package sandbox_service

import (
	"context"
	"errors"
	"testing"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/models/message"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service/lifecycle"
)

// Test_ExecuteColdStartHook tests the critical cold start hook execution
func Test_ExecuteColdStartHook(t *testing.T) {
	t.Run("executes hook and returns error on failure", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()
		conversationID := types.UUID("test-conversation-id")
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox-id",
			Config: &modal.SandboxConfig{
				S3Config: &modal.S3MountConfig{
					BucketName: "test-bucket",
				},
			},
		}

		hookCalled := false
		expectedErr := errors.New("cold start failed")
		template := &SandboxTemplate{
			Hooks: &lifecycle.LifecycleHooks{
				OnColdStart: func(_ context.Context, hookData *lifecycle.HookData) error {
					hookCalled = true
					assert.Equal(t, conversationID, hookData.ConversationID)
					assert.Equal(t, sandboxInfo, hookData.SandboxInfo)
					return expectedErr
				},
			},
		}

		// Act
		err := service.ExecuteColdStartHook(ctx, conversationID, sandboxInfo, template)

		// Assert
		assert.True(t, hookCalled)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("returns nil when hook succeeds", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()
		conversationID := types.UUID("test-conversation-id")
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox-id",
		}

		hookCalled := false
		template := &SandboxTemplate{
			Hooks: &lifecycle.LifecycleHooks{
				OnColdStart: func(_ context.Context, _ *lifecycle.HookData) error {
					hookCalled = true
					return nil
				},
			},
		}

		// Act
		err := service.ExecuteColdStartHook(ctx, conversationID, sandboxInfo, template)

		// Assert
		assert.True(t, hookCalled)
		assert.NoError(t, err)
	})

	t.Run("returns nil when no hook is registered", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()
		conversationID := types.UUID("test-conversation-id")
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox-id",
		}
		template := &SandboxTemplate{
			Hooks: nil,
		}

		// Act
		err := service.ExecuteColdStartHook(ctx, conversationID, sandboxInfo, template)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("returns nil when hooks struct exists but OnColdStart is nil", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()
		conversationID := types.UUID("test-conversation-id")
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox-id",
		}
		template := &SandboxTemplate{
			Hooks: &lifecycle.LifecycleHooks{
				OnColdStart: nil,
			},
		}

		// Act
		err := service.ExecuteColdStartHook(ctx, conversationID, sandboxInfo, template)

		// Assert
		assert.NoError(t, err)
	})
}

// Test_ExecuteMessageHook tests the non-critical message hook execution
func Test_ExecuteMessageHook(t *testing.T) {
	t.Run("executes hook and ignores error", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()
		conversationID := types.UUID("test-conversation-id")
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox-id",
		}
		msg := &message.Message{
			ConversationID: conversationID,
			Body:           "test message",
		}

		hookCalled := false
		expectedErr := errors.New("message save failed")
		template := &SandboxTemplate{
			Hooks: &lifecycle.LifecycleHooks{
				OnMessage: func(_ context.Context, hookData *lifecycle.HookData) error {
					hookCalled = true
					assert.Equal(t, conversationID, hookData.ConversationID)
					assert.Equal(t, sandboxInfo, hookData.SandboxInfo)
					assert.Equal(t, msg, hookData.Message)
					return expectedErr // Error should be ignored
				},
			},
		}

		// Act
		service.ExecuteMessageHook(ctx, conversationID, sandboxInfo, template, msg)

		// Assert
		assert.True(t, hookCalled)
		// No error returned - hook swallows errors
	})

	t.Run("executes hook when it succeeds", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()
		conversationID := types.UUID("test-conversation-id")
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox-id",
		}
		msg := &message.Message{
			ConversationID: conversationID,
			Body:           "test message",
		}

		hookCalled := false
		template := &SandboxTemplate{
			Hooks: &lifecycle.LifecycleHooks{
				OnMessage: func(_ context.Context, _ *lifecycle.HookData) error {
					hookCalled = true
					return nil
				},
			},
		}

		// Act
		service.ExecuteMessageHook(ctx, conversationID, sandboxInfo, template, msg)

		// Assert
		assert.True(t, hookCalled)
	})

	t.Run("does nothing when no hook is registered", func(_ *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()
		conversationID := types.UUID("test-conversation-id")
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox-id",
		}
		msg := &message.Message{
			ConversationID: conversationID,
			Body:           "test message",
		}
		template := &SandboxTemplate{
			Hooks: nil,
		}

		// Act - should not panic
		service.ExecuteMessageHook(ctx, conversationID, sandboxInfo, template, msg)

		// Assert - no crash
	})

	t.Run("does nothing when hooks struct exists but OnMessage is nil", func(_ *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()
		conversationID := types.UUID("test-conversation-id")
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox-id",
		}
		msg := &message.Message{
			ConversationID: conversationID,
			Body:           "test message",
		}
		template := &SandboxTemplate{
			Hooks: &lifecycle.LifecycleHooks{
				OnMessage: nil,
			},
		}

		// Act - should not panic
		service.ExecuteMessageHook(ctx, conversationID, sandboxInfo, template, msg)

		// Assert - no crash
	})
}

// Test_ExecuteStreamFinishHook tests the non-critical stream finish hook execution
func Test_ExecuteStreamFinishHook(t *testing.T) {
	t.Run("executes hook and ignores error", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()
		conversationID := types.UUID("test-conversation-id")
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox-id",
		}
		tokenUsage := &lifecycle.TokenUsage{
			InputTokens:  100,
			OutputTokens: 200,
			CacheTokens:  50,
		}

		hookCalled := false
		expectedErr := errors.New("stream finish failed")
		template := &SandboxTemplate{
			Hooks: &lifecycle.LifecycleHooks{
				OnStreamFinish: func(_ context.Context, hookData *lifecycle.HookData) error {
					hookCalled = true
					assert.Equal(t, conversationID, hookData.ConversationID)
					assert.Equal(t, sandboxInfo, hookData.SandboxInfo)
					assert.Equal(t, tokenUsage, hookData.TokenUsage)
					return expectedErr // Error should be ignored
				},
			},
		}

		// Act
		service.ExecuteStreamFinishHook(ctx, conversationID, sandboxInfo, template, tokenUsage)

		// Assert
		assert.True(t, hookCalled)
		// No error returned - hook swallows errors
	})

	t.Run("executes hook when it succeeds", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()
		conversationID := types.UUID("test-conversation-id")
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox-id",
		}
		tokenUsage := &lifecycle.TokenUsage{
			InputTokens:  100,
			OutputTokens: 200,
			CacheTokens:  50,
		}

		hookCalled := false
		template := &SandboxTemplate{
			Hooks: &lifecycle.LifecycleHooks{
				OnStreamFinish: func(_ context.Context, _ *lifecycle.HookData) error {
					hookCalled = true
					return nil
				},
			},
		}

		// Act
		service.ExecuteStreamFinishHook(ctx, conversationID, sandboxInfo, template, tokenUsage)

		// Assert
		assert.True(t, hookCalled)
	})

	t.Run("does nothing when no hook is registered", func(_ *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()
		conversationID := types.UUID("test-conversation-id")
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox-id",
		}
		tokenUsage := &lifecycle.TokenUsage{
			InputTokens:  100,
			OutputTokens: 200,
			CacheTokens:  50,
		}
		template := &SandboxTemplate{
			Hooks: nil,
		}

		// Act - should not panic
		service.ExecuteStreamFinishHook(ctx, conversationID, sandboxInfo, template, tokenUsage)

		// Assert - no crash
	})

	t.Run("does nothing when hooks struct exists but OnStreamFinish is nil", func(_ *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()
		conversationID := types.UUID("test-conversation-id")
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox-id",
		}
		tokenUsage := &lifecycle.TokenUsage{
			InputTokens:  100,
			OutputTokens: 200,
			CacheTokens:  50,
		}
		template := &SandboxTemplate{
			Hooks: &lifecycle.LifecycleHooks{
				OnStreamFinish: nil,
			},
		}

		// Act - should not panic
		service.ExecuteStreamFinishHook(ctx, conversationID, sandboxInfo, template, tokenUsage)

		// Assert - no crash
	})
}

// Test_ExecuteTerminateHook tests the terminate hook execution
func Test_ExecuteTerminateHook(t *testing.T) {
	t.Run("executes hook and returns error on failure", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()
		conversationID := types.UUID("test-conversation-id")
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox-id",
		}

		hookCalled := false
		expectedErr := errors.New("terminate failed")
		template := &SandboxTemplate{
			Hooks: &lifecycle.LifecycleHooks{
				OnTerminate: func(_ context.Context, hookData *lifecycle.HookData) error {
					hookCalled = true
					assert.Equal(t, conversationID, hookData.ConversationID)
					assert.Equal(t, sandboxInfo, hookData.SandboxInfo)
					return expectedErr
				},
			},
		}

		// Act
		err := service.ExecuteTerminateHook(ctx, conversationID, sandboxInfo, template)

		// Assert
		assert.True(t, hookCalled)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("returns nil when hook succeeds", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()
		conversationID := types.UUID("test-conversation-id")
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox-id",
		}

		hookCalled := false
		template := &SandboxTemplate{
			Hooks: &lifecycle.LifecycleHooks{
				OnTerminate: func(_ context.Context, _ *lifecycle.HookData) error {
					hookCalled = true
					return nil
				},
			},
		}

		// Act
		err := service.ExecuteTerminateHook(ctx, conversationID, sandboxInfo, template)

		// Assert
		assert.True(t, hookCalled)
		assert.NoError(t, err)
	})

	t.Run("returns nil when no hook is registered", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()
		conversationID := types.UUID("test-conversation-id")
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox-id",
		}
		template := &SandboxTemplate{
			Hooks: nil,
		}

		// Act
		err := service.ExecuteTerminateHook(ctx, conversationID, sandboxInfo, template)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("returns nil when hooks struct exists but OnTerminate is nil", func(t *testing.T) {
		// Arrange
		service := NewSandboxService()
		ctx := context.Background()
		conversationID := types.UUID("test-conversation-id")
		sandboxInfo := &modal.SandboxInfo{
			SandboxID: "test-sandbox-id",
		}
		template := &SandboxTemplate{
			Hooks: &lifecycle.LifecycleHooks{
				OnTerminate: nil,
			},
		}

		// Act
		err := service.ExecuteTerminateHook(ctx, conversationID, sandboxInfo, template)

		// Assert
		assert.NoError(t, err)
	})
}
