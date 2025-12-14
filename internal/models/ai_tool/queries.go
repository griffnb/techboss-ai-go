package ai_tool

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/types"
)

type Mocker struct {
	// Standard Functions
	Get              func(ctx context.Context, id types.UUID) (*AiTool, error)
	GetJoined        func(ctx context.Context, id types.UUID) (*AiToolJoined, error)
	FindAll          func(ctx context.Context, options *model.Options) ([]*AiTool, error)
	FindAllJoined    func(ctx context.Context, options *model.Options) ([]*AiToolJoined, error)
	FindFirst        func(ctx context.Context, options *model.Options) (*AiTool, error)
	FindFirstJoined  func(ctx context.Context, options *model.Options) (*AiToolJoined, error)
	FindResultsCount func(ctx context.Context, options *model.Options) (int64, error)
}
