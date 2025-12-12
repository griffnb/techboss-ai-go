package subscription

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/types"
)

type mocker interface {
	GetBySubscriptionID(ctx context.Context, subscriptionID string) (*Subscription, error)
	GetActiveByOrganizationID(ctx context.Context, organizationID types.UUID) (*Subscription, error)
	GetByOrganizationAndPlanID(ctx context.Context, organizationID types.UUID, planID types.UUID) (*Subscription, error)
}

// GetBySubscriptionID finds a subscription by its Stripe subscription ID.
// Returns the first non-disabled subscription matching the subscription ID.
func GetBySubscriptionID(ctx context.Context, subscriptionID string) (*Subscription, error) {
	mocker, ok := model.GetMocker[mocker](ctx, PACKAGE)
	if ok {
		return mocker.GetBySubscriptionID(ctx, subscriptionID)
	}

	return FindFirst(ctx, model.NewOptions().
		WithCondition("%s = :subscription_id:", Columns.SubscriptionID.Column()).
		WithCondition("%s = 0", Columns.Disabled.Column()).
		WithParam(":subscription_id:", subscriptionID))
}

// GetActiveByOrganizationID finds the active subscription for an organization.
// Returns the first non-disabled subscription with no end_ts (active) for the organization.
func GetActiveByOrganizationID(ctx context.Context, organizationID types.UUID) (*Subscription, error) {
	mocker, ok := model.GetMocker[mocker](ctx, PACKAGE)
	if ok {
		return mocker.GetActiveByOrganizationID(ctx, organizationID)
	}

	return FindFirst(ctx, model.NewOptions().
		WithCondition("%s = :organization_id:", Columns.OrganizationID.Column()).
		WithCondition("%s = 0", Columns.EndTS.Column()).
		WithCondition("%s = 0", Columns.Disabled.Column()).
		WithParam(":organization_id:", organizationID))
}

func GetByOrganizationAndPlanID(ctx context.Context, organizationID types.UUID, planID types.UUID) (*Subscription, error) {
	mocker, ok := model.GetMocker[mocker](ctx, PACKAGE)
	if ok {
		return mocker.GetByOrganizationAndPlanID(ctx, organizationID, planID)
	}

	return FindFirst(ctx, model.NewOptions().
		WithCondition("%s = :organization_id:", Columns.OrganizationID.Column()).
		WithCondition("%s = :plan_id:", Columns.BillingPlanID.Column()).
		WithCondition("%s = 0", Columns.Disabled.Column()).
		WithParam(":organization_id:", organizationID).
		WithParam(":plan_id:", planID))
}
