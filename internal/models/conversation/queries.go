package conversation

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/types"
)

type Mocker struct {
	// Standard Functions
	Get              func(ctx context.Context, id types.UUID) (*Conversation, error)
	GetJoined        func(ctx context.Context, id types.UUID) (*ConversationJoined, error)
	FindAll          func(ctx context.Context, options *model.Options) ([]*Conversation, error)
	FindAllJoined    func(ctx context.Context, options *model.Options) ([]*ConversationJoined, error)
	FindFirst        func(ctx context.Context, options *model.Options) (*Conversation, error)
	FindFirstJoined  func(ctx context.Context, options *model.Options) (*ConversationJoined, error)
	FindResultsCount func(ctx context.Context, options *model.Options) (int64, error)
}
