package billing_plan_price

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/coremodel"
	"github.com/griffnb/core/lib/types"
)

// FindAllRestrictedJoined returns all joined records with restrictions for the session account
// TODO: Implement specific access restrictions for this model
func FindAllRestrictedJoined(ctx context.Context, options *model.Options, _ coremodel.Model) ([]*BillingPlanPriceJoined, error) {
	return FindAllJoined(ctx, options)
}

func FindAllRestricted(ctx context.Context, options *model.Options, _ coremodel.Model) ([]*BillingPlanPrice, error) {
	return FindAll(ctx, options)
}

// CountRestricted returns the count of records with restrictions for the session account
// TODO: Implement specific access restrictions for this model
func CountRestricted(ctx context.Context, options *model.Options, _ coremodel.Model) (int64, error) {
	return FindResultsCount(ctx, options)
}

// GetRestrictedJoined gets a specific record with joined data and restrictions for the session account
// TODO: Adjust the restrictions to match your access control requirements
func GetRestrictedJoined(ctx context.Context, id types.UUID, _ coremodel.Model) (*BillingPlanPriceJoined, error) {
	return GetJoined(ctx, id)
}

func GetRestricted(ctx context.Context, id types.UUID, _ coremodel.Model) (*BillingPlanPrice, error) {
	return Get(ctx, id)
}
