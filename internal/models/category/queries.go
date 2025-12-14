package category

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/types"
)

type Mocker struct {
	// Standard Functions
	Get              func(ctx context.Context, id types.UUID) (*Category, error)
	GetJoined        func(ctx context.Context, id types.UUID) (*CategoryJoined, error)
	FindAll          func(ctx context.Context, options *model.Options) ([]*Category, error)
	FindAllJoined    func(ctx context.Context, options *model.Options) ([]*CategoryJoined, error)
	FindFirst        func(ctx context.Context, options *model.Options) (*Category, error)
	FindFirstJoined  func(ctx context.Context, options *model.Options) (*CategoryJoined, error)
	FindResultsCount func(ctx context.Context, options *model.Options) (int64, error)
}
