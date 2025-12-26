package object_tag

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/types"
)

type Mocker struct {
	// Standard Functions
	Get              func(ctx context.Context, id types.UUID) (*ObjectTag, error)
	GetJoined        func(ctx context.Context, id types.UUID) (*ObjectTagJoined, error)
	FindAll          func(ctx context.Context, options *model.Options) ([]*ObjectTag, error)
	FindAllJoined    func(ctx context.Context, options *model.Options) ([]*ObjectTagJoined, error)
	FindFirst        func(ctx context.Context, options *model.Options) (*ObjectTag, error)
	FindFirstJoined  func(ctx context.Context, options *model.Options) (*ObjectTagJoined, error)
	FindResultsCount func(ctx context.Context, options *model.Options) (int64, error)
}
