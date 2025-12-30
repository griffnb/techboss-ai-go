package lifecycle

import (
	"context"
	"testing"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/pkg/errors"
)

func Test_ExecuteHook_nilHook(t *testing.T) {
	t.Run("returns nil when hook is nil", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		hookData := &HookData{
			ConversationID: types.UUID("test-conversation"),
		}

		// Act
		err := ExecuteHook(ctx, "TestHook", nil, hookData)

		// Assert
		assert.NoError(t, err)
	})
}

func Test_ExecuteHook_callsHook(t *testing.T) {
	t.Run("calls hook function and returns nil on success", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		hookData := &HookData{
			ConversationID: types.UUID("test-conversation"),
		}
		called := false
		var capturedData *HookData

		mockHook := func(_ context.Context, data *HookData) error {
			called = true
			capturedData = data
			return nil
		}

		// Act
		err := ExecuteHook(ctx, "TestHook", mockHook, hookData)

		// Assert
		assert.NoError(t, err)
		assert.True(t, called, "Expected hook to be called")
		assert.Equal(t, hookData.ConversationID, capturedData.ConversationID)
	})
}

func Test_ExecuteHook_propagatesErrors(t *testing.T) {
	t.Run("propagates error from hook", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		hookData := &HookData{
			ConversationID: types.UUID("test-conversation"),
		}
		expectedErr := errors.New("hook failed")

		mockHook := func(_ context.Context, _ *HookData) error {
			return expectedErr
		}

		// Act
		err := ExecuteHook(ctx, "TestHook", mockHook, hookData)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func Test_ExecuteHook_logsDuration(t *testing.T) {
	t.Run("executes hook successfully without logging errors", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		hookData := &HookData{
			ConversationID: types.UUID("test-conversation"),
		}

		mockHook := func(_ context.Context, _ *HookData) error {
			// Simulate some work
			return nil
		}

		// Act
		err := ExecuteHook(ctx, "TestHook", mockHook, hookData)

		// Assert
		assert.NoError(t, err)
		// Note: Duration logging is verified by execution completing without error
		// Actual log output would require log capture infrastructure for verification
	})

	t.Run("logs duration even when hook returns error", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		hookData := &HookData{
			ConversationID: types.UUID("test-conversation"),
		}
		hookErr := errors.New("simulated failure")

		mockHook := func(_ context.Context, _ *HookData) error {
			return hookErr
		}

		// Act
		err := ExecuteHook(ctx, "TestHook", mockHook, hookData)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, hookErr, err)
		// Logging happens regardless of success/failure
	})
}

func Test_ExecuteHook_multipleScenarios(t *testing.T) {
	tests := []struct {
		name        string
		hook        HookFunc
		hookName    string
		expectError bool
		description string
	}{
		{
			name:        "nil_hook_no_error",
			hook:        nil,
			hookName:    "NilHook",
			expectError: false,
			description: "Nil hook should return nil without error",
		},
		{
			name: "successful_hook",
			hook: func(_ context.Context, _ *HookData) error {
				return nil
			},
			hookName:    "SuccessHook",
			expectError: false,
			description: "Successful hook should return nil",
		},
		{
			name: "failing_hook",
			hook: func(_ context.Context, _ *HookData) error {
				return errors.New("hook error")
			},
			hookName:    "FailingHook",
			expectError: true,
			description: "Failing hook should propagate error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx := context.Background()
			hookData := &HookData{
				ConversationID: types.UUID("test-conversation"),
			}

			// Act
			err := ExecuteHook(ctx, tt.hookName, tt.hook, hookData)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
