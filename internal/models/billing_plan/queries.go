package billing_plan

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/types"
)

type Mocker struct {
	// Standard Functions
	Get              func(ctx context.Context, id types.UUID) (*BillingPlan, error)
	GetJoined        func(ctx context.Context, id types.UUID) (*BillingPlanJoined, error)
	FindAll          func(ctx context.Context, options *model.Options) ([]*BillingPlan, error)
	FindAllJoined    func(ctx context.Context, options *model.Options) ([]*BillingPlanJoined, error)
	FindFirst        func(ctx context.Context, options *model.Options) (*BillingPlan, error)
	FindFirstJoined  func(ctx context.Context, options *model.Options) (*BillingPlanJoined, error)
	FindResultsCount func(ctx context.Context, options *model.Options) (int64, error)
	// Custom Functions
	GetAllActivePlans func(ctx context.Context) ([]*BillingPlan, error)
}

func GetAllActivePlans(ctx context.Context) ([]*BillingPlan, error) {
	mocker, ok := model.GetMocker[*Mocker](ctx, PACKAGE)
	if ok {
		return mocker.GetAllActivePlans(ctx)
	}
	options := model.NewOptions().
		WithCondition("%s = 0", Columns.Disabled.Column()).WithOrder("%s asc", Columns.Level.Column())
	return FindAll(ctx, options)
}
