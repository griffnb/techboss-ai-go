package message_test

import (
	"context"
	"testing"
	"time"

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
