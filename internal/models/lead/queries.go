package lead

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/types"
)

type Mocker struct {
	// Standard Functions
	Get              func(ctx context.Context, id types.UUID) (*Lead, error)
	GetJoined        func(ctx context.Context, id types.UUID) (*LeadJoined, error)
	FindAll          func(ctx context.Context, options *model.Options) ([]*Lead, error)
	FindAllJoined    func(ctx context.Context, options *model.Options) ([]*LeadJoined, error)
	FindFirst        func(ctx context.Context, options *model.Options) (*Lead, error)
	FindFirstJoined  func(ctx context.Context, options *model.Options) (*LeadJoined, error)
	FindResultsCount func(ctx context.Context, options *model.Options) (int64, error)
}
