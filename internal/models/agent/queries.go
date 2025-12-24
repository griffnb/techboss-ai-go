package agent

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/types"
)

type Mocker struct {
	// Standard Functions
	Get              func(ctx context.Context, id types.UUID) (*Agent, error)
	GetJoined        func(ctx context.Context, id types.UUID) (*AgentJoined, error)
	FindAll          func(ctx context.Context, options *model.Options) ([]*Agent, error)
	FindAllJoined    func(ctx context.Context, options *model.Options) ([]*AgentJoined, error)
	FindFirst        func(ctx context.Context, options *model.Options) (*Agent, error)
	FindFirstJoined  func(ctx context.Context, options *model.Options) (*AgentJoined, error)
	FindResultsCount func(ctx context.Context, options *model.Options) (int64, error)
}
