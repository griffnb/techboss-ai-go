package conversation_test

import (
	"testing"

	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/models/conversation"
)

func init() {
	system_testing.BuildSystem()
}

func Test_ConversationStats_AddTokenUsage(t *testing.T) {
	t.Run("should add token usage correctly", func(t *testing.T) {
		// Arrange
		stats := &conversation.ConversationStats{
			MessagesExchanged: 0,
			TotalInputTokens:  0,
			TotalOutputTokens: 0,
			TotalCacheTokens:  0,
		}

		// Act
		stats.AddTokenUsage(100, 200, 50)

		// Assert
		if stats.TotalInputTokens != 100 {
			t.Fatalf("expected TotalInputTokens to be 100, got %d", stats.TotalInputTokens)
		}
		if stats.TotalOutputTokens != 200 {
			t.Fatalf("expected TotalOutputTokens to be 200, got %d", stats.TotalOutputTokens)
		}
		if stats.TotalCacheTokens != 50 {
			t.Fatalf("expected TotalCacheTokens to be 50, got %d", stats.TotalCacheTokens)
		}
	})

	t.Run("should accumulate token usage across multiple calls", func(t *testing.T) {
		// Arrange
		stats := &conversation.ConversationStats{
			MessagesExchanged: 0,
			TotalInputTokens:  100,
			TotalOutputTokens: 200,
			TotalCacheTokens:  50,
		}

		// Act
		stats.AddTokenUsage(50, 100, 25)

		// Assert
		if stats.TotalInputTokens != 150 {
			t.Fatalf("expected TotalInputTokens to be 150, got %d", stats.TotalInputTokens)
		}
		if stats.TotalOutputTokens != 300 {
			t.Fatalf("expected TotalOutputTokens to be 300, got %d", stats.TotalOutputTokens)
		}
		if stats.TotalCacheTokens != 75 {
			t.Fatalf("expected TotalCacheTokens to be 75, got %d", stats.TotalCacheTokens)
		}
	})

	t.Run("should handle zero values", func(t *testing.T) {
		// Arrange
		stats := &conversation.ConversationStats{
			MessagesExchanged: 0,
			TotalInputTokens:  100,
			TotalOutputTokens: 200,
			TotalCacheTokens:  50,
		}

		// Act
		stats.AddTokenUsage(0, 0, 0)

		// Assert
		if stats.TotalInputTokens != 100 {
			t.Fatalf("expected TotalInputTokens to remain 100, got %d", stats.TotalInputTokens)
		}
		if stats.TotalOutputTokens != 200 {
			t.Fatalf("expected TotalOutputTokens to remain 200, got %d", stats.TotalOutputTokens)
		}
		if stats.TotalCacheTokens != 50 {
			t.Fatalf("expected TotalCacheTokens to remain 50, got %d", stats.TotalCacheTokens)
		}
	})

	t.Run("should handle negative values as additions", func(t *testing.T) {
		// Arrange
		stats := &conversation.ConversationStats{
			MessagesExchanged: 0,
			TotalInputTokens:  100,
			TotalOutputTokens: 200,
			TotalCacheTokens:  50,
		}

		// Act - negative values should still be added (no validation, just arithmetic)
		stats.AddTokenUsage(-10, -20, -5)

		// Assert
		if stats.TotalInputTokens != 90 {
			t.Fatalf("expected TotalInputTokens to be 90, got %d", stats.TotalInputTokens)
		}
		if stats.TotalOutputTokens != 180 {
			t.Fatalf("expected TotalOutputTokens to be 180, got %d", stats.TotalOutputTokens)
		}
		if stats.TotalCacheTokens != 45 {
			t.Fatalf("expected TotalCacheTokens to be 45, got %d", stats.TotalCacheTokens)
		}
	})
}

func Test_ConversationStats_IncrementMessages(t *testing.T) {
	t.Run("should increment messages from zero", func(t *testing.T) {
		// Arrange
		stats := &conversation.ConversationStats{
			MessagesExchanged: 0,
		}

		// Act
		stats.IncrementMessages()

		// Assert
		if stats.MessagesExchanged != 1 {
			t.Fatalf("expected MessagesExchanged to be 1, got %d", stats.MessagesExchanged)
		}
	})

	t.Run("should increment messages multiple times", func(t *testing.T) {
		// Arrange
		stats := &conversation.ConversationStats{
			MessagesExchanged: 5,
		}

		// Act
		stats.IncrementMessages()
		stats.IncrementMessages()
		stats.IncrementMessages()

		// Assert
		if stats.MessagesExchanged != 8 {
			t.Fatalf("expected MessagesExchanged to be 8, got %d", stats.MessagesExchanged)
		}
	})
}

func Test_ConversationStats_BackwardCompatibility(t *testing.T) {
	t.Run("should maintain TotalTokensUsed field for backward compatibility", func(t *testing.T) {
		// Arrange
		stats := &conversation.ConversationStats{
			MessagesExchanged: 1,
			TotalTokensUsed:   500, // Legacy field
			TotalInputTokens:  100,
			TotalOutputTokens: 200,
			TotalCacheTokens:  50,
		}

		// Assert - just verify the field exists and can be read
		if stats.TotalTokensUsed != 500 {
			t.Fatalf("expected TotalTokensUsed to be 500, got %d", stats.TotalTokensUsed)
		}
	})
}
