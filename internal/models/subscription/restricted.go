package subscription

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/coremodel"
	"github.com/griffnb/core/lib/types"
)

// FindAllRestrictedJoined returns all joined records with restrictions for the session account
// TODO: Implement specific access restrictions for this model
func FindAllRestrictedJoined(ctx context.Context, options *model.Options, sessionAccount coremodel.Model) ([]*SubscriptionJoined, error) {
	// Uncomment and adjust the following lines to implement proper restrictions
	options.WithCondition("%s = :organization_id:", Columns.OrganizationID.Column())
	options.WithParam(":organization_id:", sessionAccount.GetString("organization_id"))
	return FindAllJoined(ctx, options)
}

func FindAllRestricted(ctx context.Context, options *model.Options, sessionAccount coremodel.Model) ([]*Subscription, error) {
	// Uncomment and adjust the following lines to implement proper restrictions
	options.WithCondition("%s = :organization_id:", Columns.OrganizationID.Column())
	options.WithParam(":organization_id:", sessionAccount.GetString("organization_id"))
	return FindAll(ctx, options)
}

// CountRestricted returns the count of records with restrictions for the session account
// TODO: Implement specific access restrictions for this model
func CountRestricted(ctx context.Context, options *model.Options, sessionAccount coremodel.Model) (int64, error) {
	// Uncomment and adjust the following lines to implement proper restrictions
	options.WithCondition("%s = :organization_id:", Columns.OrganizationID.Column())
	options.WithParam(":organization_id:", sessionAccount.GetString("organization_id"))
	return FindResultsCount(ctx, options)
}

// GetRestrictedJoined gets a specific record with joined data and restrictions for the session account
// TODO: Adjust the restrictions to match your access control requirements
func GetRestrictedJoined(ctx context.Context, id types.UUID, sessionAccount coremodel.Model) (*SubscriptionJoined, error) {
	options := model.NewOptions().
		WithCondition("%s.id = :id:", TABLE).
		WithParam(":id:", id).
		WithCondition("%s = :organization_id:", Columns.OrganizationID.Column()).
		WithParam(":organization_id:", sessionAccount.GetString("organization_id"))

	return FindFirstJoined(ctx, options)
}

func GetRestricted(ctx context.Context, id types.UUID, sessionAccount coremodel.Model) (*Subscription, error) {
	options := model.NewOptions().
		WithCondition("%s.id = :id:", TABLE).
		WithParam(":id:", id).
		WithCondition("%s = :organization_id:", Columns.OrganizationID.Column()).
		WithParam(":organization_id:", sessionAccount.ID())

	return FindFirst(ctx, options)
}

// NewPublic creates a new model instance with sanitized input and session account context
