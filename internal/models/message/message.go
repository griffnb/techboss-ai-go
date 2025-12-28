package message

import (
	"context"
	"time"

	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/pkg/errors"
)

// Message represents a conversation message with optional tool calls
type Message struct {
	Key            string     `json:"key"`
	ConversationID types.UUID `json:"conversation_id"`
	Body           string     `json:"body"`
	Role           int64      `json:"role"`
	Timestamp      int64      `json:"timestamp"`
	Tokens         int64      `json:"tokens"`
	ToolCalls      []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall represents a tool invocation within a message
type ToolCall struct {
	ID     string         `json:"id"`
	Type   string         `json:"type"`
	Name   string         `json:"name"`
	Input  map[string]any `json:"input"`
	Output string         `json:"output"`
	Status string         `json:"status"`
	Error  string         `json:"error"`
}

func (this *Message) Save(ctx context.Context) error {
	if tools.Empty(this.Key) {
		this.Key = tools.SessionKey()
	}
	if tools.Empty(this.Timestamp) {
		this.Timestamp = time.Now().Unix()
	}

	throttled, err := environment.GetDynamo().PutWithContext(ctx, TABLE_NAME, this)
	if err != nil {
		return err
	}
	if throttled {
		return errors.Errorf("Message Throttled by key")
	}
	return nil
}

func GetMessage(ctx context.Context, key string) (*Message, error) {
	msg := &Message{}
	throttled, err := environment.GetDynamo().GetWithContext(ctx, TABLE_NAME, "key", key, msg, true)
	if err != nil {
		return nil, err
	}
	if throttled {
		return nil, errors.Errorf("Message Throttled by key")
	}
	return msg, nil
}

func GetMessagesByConversationID(ctx context.Context, conversationID types.UUID, limit int64) ([]*Message, error) {
	messages := []*Message{}
	throttled, err := environment.GetDynamo().
		// nolint: gosec
		GetByIndexWithContext(ctx, TABLE_NAME, "conversation_id", conversationID, &messages, int32(limit), true)
	if err != nil {
		return nil, err
	}
	if throttled {
		return nil, errors.Errorf("Message Throttled by conversationID")
	}

	reverseArray(messages)

	return messages, nil
}

func reverseArray(arr []*Message) {
	left := 0
	right := len(arr) - 1

	for left < right {
		// Swap the elements at left and right indices
		arr[left], arr[right] = arr[right], arr[left]
		left++
		right--
	}
}
