package tag

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/types"
)

type Mocker struct {
	// Standard Functions
	Get              func(ctx context.Context, id types.UUID) (*Tag, error)
	GetJoined        func(ctx context.Context, id types.UUID) (*TagJoined, error)
	FindAll          func(ctx context.Context, options *model.Options) ([]*Tag, error)
	FindAllJoined    func(ctx context.Context, options *model.Options) ([]*TagJoined, error)
	FindFirst        func(ctx context.Context, options *model.Options) (*Tag, error)
	FindFirstJoined  func(ctx context.Context, options *model.Options) (*TagJoined, error)
	FindResultsCount func(ctx context.Context, options *model.Options) (int64, error)
}
