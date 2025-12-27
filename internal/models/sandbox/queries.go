package sandbox

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/types"
)

type Mocker struct {
	// Standard Functions
	Get              func(ctx context.Context, id types.UUID) (*Sandbox, error)
	GetJoined        func(ctx context.Context, id types.UUID) (*SandboxJoined, error)
	FindAll          func(ctx context.Context, options *model.Options) ([]*Sandbox, error)
	FindAllJoined    func(ctx context.Context, options *model.Options) ([]*SandboxJoined, error)
	FindFirst        func(ctx context.Context, options *model.Options) (*Sandbox, error)
	FindFirstJoined  func(ctx context.Context, options *model.Options) (*SandboxJoined, error)
	FindResultsCount func(ctx context.Context, options *model.Options) (int64, error)
	// Custom Functions
}
