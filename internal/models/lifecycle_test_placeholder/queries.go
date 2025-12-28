package lifecycle_test_placeholder

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/types"
)

type Mocker struct {
	// Standard Functions
	Get              func(ctx context.Context, id types.UUID) (*LifecycleTestPlaceholder, error)
	GetJoined        func(ctx context.Context, id types.UUID) (*LifecycleTestPlaceholderJoined, error)
	FindAll          func(ctx context.Context, options *model.Options) ([]*LifecycleTestPlaceholder, error)
	FindAllJoined    func(ctx context.Context, options *model.Options) ([]*LifecycleTestPlaceholderJoined, error)
	FindFirst        func(ctx context.Context, options *model.Options) (*LifecycleTestPlaceholder, error)
	FindFirstJoined  func(ctx context.Context, options *model.Options) (*LifecycleTestPlaceholderJoined, error)
	FindResultsCount func(ctx context.Context, options *model.Options) (int64, error)
	// Custom Functions
}
