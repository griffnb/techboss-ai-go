package billing_plan_price

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/types"
)

type Mocker struct {
	// Standard Functions
	Get              func(ctx context.Context, id types.UUID) (*BillingPlanPrice, error)
	GetJoined        func(ctx context.Context, id types.UUID) (*BillingPlanPriceJoined, error)
	FindAll          func(ctx context.Context, options *model.Options) ([]*BillingPlanPrice, error)
	FindAllJoined    func(ctx context.Context, options *model.Options) ([]*BillingPlanPriceJoined, error)
	FindFirst        func(ctx context.Context, options *model.Options) (*BillingPlanPrice, error)
	FindFirstJoined  func(ctx context.Context, options *model.Options) (*BillingPlanPriceJoined, error)
	FindResultsCount func(ctx context.Context, options *model.Options) (int64, error)
	// Custom Functions
	GetAllActivePrices func(ctx context.Context) ([]*BillingPlanPrice, error)
}

func GetAllActivePrices(ctx context.Context) ([]*BillingPlanPrice, error) {
	mocker, ok := model.GetMocker[*Mocker](ctx, PACKAGE)
	if ok {
		return mocker.GetAllActivePrices(ctx)
	}
	options := model.NewOptions().WithCondition("%s = 0", Columns.Disabled.Column())
	return FindAll(ctx, options)
}
