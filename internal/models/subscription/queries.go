package subscription

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/types"
)

type mocker interface {
	GetByStripeSubscriptionID(ctx context.Context, subscriptionID string) (*Subscription, error)
	GetByStripeCustomerID(ctx context.Context, customerID string) (*Subscription, error)
	GetActiveByOrganizationID(ctx context.Context, organizationID types.UUID) (*Subscription, error)
	GetByOrganizationAndPlanPriceID(ctx context.Context, organizationID types.UUID, planPriceID types.UUID) (*Subscription, error)
}

// GetByStripeSubscriptionID finds a subscription by its Stripe subscription ID.
// Returns the first non-disabled subscription matching the subscription ID.
func GetByStripeSubscriptionID(ctx context.Context, subscriptionID string) (*Subscription, error) {
	mocker, ok := model.GetMocker[mocker](ctx, PACKAGE)
	if ok {
		return mocker.GetByStripeSubscriptionID(ctx, subscriptionID)
	}

	return FindFirst(ctx, model.NewOptions().
		WithCondition("%s = :subscription_id:", Columns.StripeSubscriptionID.Column()).
		WithCondition("%s = 0", Columns.Disabled.Column()).
		WithParam(":subscription_id:", subscriptionID))
}

func GetByStripeCustomerID(ctx context.Context, customerID string) (*Subscription, error) {
	mocker, ok := model.GetMocker[mocker](ctx, PACKAGE)
	if ok {
		return mocker.GetByStripeCustomerID(ctx, customerID)
	}

	return FindFirst(ctx, model.NewOptions().
		WithCondition("%s = :customer_id:", Columns.StripeCustomerID.Column()).
		WithCondition("%s = 0", Columns.Disabled.Column()).
		WithParam(":customer_id:", customerID))
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

func GetByOrganizationAndPlanPriceID(ctx context.Context, organizationID types.UUID, planPriceID types.UUID) (*Subscription, error) {
	mocker, ok := model.GetMocker[mocker](ctx, PACKAGE)
	if ok {
		return mocker.GetByOrganizationAndPlanPriceID(ctx, organizationID, planPriceID)
	}

	return FindFirst(ctx, model.NewOptions().
		WithCondition("%s = :organization_id:", Columns.OrganizationID.Column()).
		WithCondition("%s = :plan_price_id:", Columns.BillingPlanPriceID.Column()).
		WithCondition("%s = 0", Columns.Disabled.Column()).
		WithParam(":organization_id:", organizationID).
		WithParam(":plan_price_id:", planPriceID))
}
