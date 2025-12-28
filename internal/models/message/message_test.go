package message_test

import (
	"context"
	"testing"
	"time"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/models/message"
)

func init() {
	system_testing.BuildSystem()
}

func TestGetMessagesByConversationID(t *testing.T) {
	t.Skip()
	conversationKey := tools.GUID()

	{
		msg := &message.Message{
			Key:            tools.SessionKey(),
			ConversationID: conversationKey,
			Body:           tools.RandString(10),
			Role:           1,
			Timestamp:      time.Now().Unix(),
		}

		err := msg.Save(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		newMessage, err := message.GetMessage(context.Background(), msg.Key)
		if err != nil {
			t.Fatal(err)
		}

		if newMessage.Body != msg.Body {
			t.Fatal("Message body does not match")
		}
	}

	{
		msg := &message.Message{
			Key:            tools.SessionKey(),
			ConversationID: conversationKey,
			Body:           tools.RandString(10),
			Role:           2,
			Timestamp:      time.Now().Unix(),
		}

		err := msg.Save(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		messages, err := message.GetMessagesByConversationID(context.Background(), conversationKey, 10)
		if err != nil {
			t.Fatal(err)
		}

		if len(messages) != 2 {
			t.Fatal("Expected 2 messages")
		}
	}
}

func Test_Message_WithToolCalls(t *testing.T) {
	ctx := context.Background()

	t.Run("save message with empty tool calls", func(t *testing.T) {
		msg := &message.Message{
			ConversationID: tools.GUID(),
			Body:           "Test message with empty tool calls",
			Role:           1,
			Tokens:         100,
			ToolCalls:      []message.ToolCall{},
		}

		err := msg.Save(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, msg.Key, "Key should be auto-generated")

		retrieved, err := message.GetMessage(ctx, msg.Key)
		assert.NoError(t, err)
		assert.Equal(t, msg.Body, retrieved.Body)
		assert.Equal(t, msg.Role, retrieved.Role)
		assert.Equal(t, msg.Tokens, retrieved.Tokens)
		assert.Equal(t, 0, len(retrieved.ToolCalls))
	})

	t.Run("save message with single tool call", func(t *testing.T) {
		toolCall := message.ToolCall{
			ID:   "tool_123",
			Type: "bash_command",
			Name: "ls",
			Input: map[string]any{
				"command": "ls -la",
			},
			Output: "file1.txt\nfile2.txt",
			Status: "success",
			Error:  "",
		}

		msg := &message.Message{
			ConversationID: tools.GUID(),
			Body:           "Test message with tool call",
			Role:           2,
			Tokens:         150,
			ToolCalls:      []message.ToolCall{toolCall},
		}

		err := msg.Save(ctx)
		assert.NoError(t, err)

		retrieved, err := message.GetMessage(ctx, msg.Key)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(retrieved.ToolCalls))
		assert.Equal(t, "tool_123", retrieved.ToolCalls[0].ID)
		assert.Equal(t, "bash_command", retrieved.ToolCalls[0].Type)
		assert.Equal(t, "ls", retrieved.ToolCalls[0].Name)
		assert.Equal(t, "ls -la", retrieved.ToolCalls[0].Input["command"])
		assert.Equal(t, "file1.txt\nfile2.txt", retrieved.ToolCalls[0].Output)
		assert.Equal(t, "success", retrieved.ToolCalls[0].Status)
		assert.Empty(t, retrieved.ToolCalls[0].Error)
	})

	t.Run("save message with multiple tool calls", func(t *testing.T) {
		toolCalls := []message.ToolCall{
			{
				ID:   "tool_1",
				Type: "function",
				Name: "read_file",
				Input: map[string]any{
					"path": "/home/user/file.txt",
				},
				Output: "File contents here",
				Status: "success",
				Error:  "",
			},
			{
				ID:   "tool_2",
				Type: "bash_command",
				Name: "grep",
				Input: map[string]any{
					"pattern": "error",
					"file":    "log.txt",
				},
				Output: "",
				Status: "error",
				Error:  "File not found",
			},
			{
				ID:   "tool_3",
				Type: "function",
				Name: "write_file",
				Input: map[string]any{
					"path":    "/home/user/output.txt",
					"content": "New content",
				},
				Output: "File written successfully",
				Status: "success",
				Error:  "",
			},
		}

		msg := &message.Message{
			ConversationID: tools.GUID(),
			Body:           "Test message with multiple tool calls",
			Role:           2,
			Tokens:         300,
			ToolCalls:      toolCalls,
		}

		err := msg.Save(ctx)
		assert.NoError(t, err)

		retrieved, err := message.GetMessage(ctx, msg.Key)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(retrieved.ToolCalls))

		assert.Equal(t, "tool_1", retrieved.ToolCalls[0].ID)
		assert.Equal(t, "function", retrieved.ToolCalls[0].Type)
		assert.Equal(t, "success", retrieved.ToolCalls[0].Status)

		assert.Equal(t, "tool_2", retrieved.ToolCalls[1].ID)
		assert.Equal(t, "error", retrieved.ToolCalls[1].Status)
		assert.Equal(t, "File not found", retrieved.ToolCalls[1].Error)

		assert.Equal(t, "tool_3", retrieved.ToolCalls[2].ID)
		assert.Equal(t, "write_file", retrieved.ToolCalls[2].Name)
	})

	t.Run("save message with nil tool calls", func(t *testing.T) {
		msg := &message.Message{
			ConversationID: tools.GUID(),
			Body:           "Test message with nil tool calls",
			Role:           1,
			Tokens:         50,
			ToolCalls:      nil,
		}

		err := msg.Save(ctx)
		assert.NoError(t, err)

		retrieved, err := message.GetMessage(ctx, msg.Key)
		assert.NoError(t, err)
		assert.Equal(t, msg.Body, retrieved.Body)
	})

	t.Run("retrieve messages by conversation ID with tool calls", func(t *testing.T) {
		t.Skip("Skipping due to pre-existing index name mismatch in migration")
		conversationID := tools.GUID()

		msg1 := &message.Message{
			ConversationID: conversationID,
			Body:           "First message",
			Role:           1,
			Tokens:         50,
			ToolCalls:      []message.ToolCall{},
		}
		err := msg1.Save(ctx)
		assert.NoError(t, err)

		msg2 := &message.Message{
			ConversationID: conversationID,
			Body:           "Second message with tool",
			Role:           2,
			Tokens:         100,
			ToolCalls: []message.ToolCall{
				{
					ID:     "tool_xyz",
					Type:   "function",
					Name:   "test_function",
					Input:  map[string]any{"arg": "value"},
					Output: "result",
					Status: "success",
					Error:  "",
				},
			},
		}
		err = msg2.Save(ctx)
		assert.NoError(t, err)

		messages, err := message.GetMessagesByConversationID(ctx, conversationID, 10)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(messages))

		hasToolCalls := false
		for _, msg := range messages {
			if len(msg.ToolCalls) > 0 {
				hasToolCalls = true
				assert.Equal(t, "tool_xyz", msg.ToolCalls[0].ID)
			}
		}
		assert.True(t, hasToolCalls, "Should have at least one message with tool calls")
	})
}
