package document

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/types"
)

type Mocker struct {
	// Standard Functions
	Get              func(ctx context.Context, id types.UUID) (*Document, error)
	GetJoined        func(ctx context.Context, id types.UUID) (*DocumentJoined, error)
	FindAll          func(ctx context.Context, options *model.Options) ([]*Document, error)
	FindAllJoined    func(ctx context.Context, options *model.Options) ([]*DocumentJoined, error)
	FindFirst        func(ctx context.Context, options *model.Options) (*Document, error)
	FindFirstJoined  func(ctx context.Context, options *model.Options) (*DocumentJoined, error)
	FindResultsCount func(ctx context.Context, options *model.Options) (int64, error)
}
