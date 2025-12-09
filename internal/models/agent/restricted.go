package agent

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/coremodel"
	"github.com/griffnb/core/lib/types"
)

// FindAllRestrictedJoined returns all joined records with restrictions for the session account
// TODO: Implement specific access restrictions for this model
func FindAllRestrictedJoined(ctx context.Context, options *model.Options, _ coremodel.Model) ([]*AgentJoined, error) {
	// Uncomment and adjust the following lines to implement proper restrictions
	// options.WithCondition("%s = :account_id:", Columns.AccountID.Column())
	// options.WithParam(":account_id:", sessionAccount.ID())
	return FindAllJoined(ctx, options)
}

func FindAllRestricted(ctx context.Context, options *model.Options, _ coremodel.Model) ([]*Agent, error) {
	// Uncomment and adjust the following lines to implement proper restrictions
	// options.WithCondition("%s = :account_id:", Columns.AccountID.Column())
	// options.WithParam(":account_id:", sessionAccount.ID())
	return FindAll(ctx, options)
}

// GetRestrictedJoined gets a specific record with joined data and restrictions for the session account
// TODO: Adjust the restrictions to match your access control requirements
func GetRestrictedJoined(ctx context.Context, id types.UUID, _ coremodel.Model) (*AgentJoined, error) {
	options := model.NewOptions().
		WithCondition("%s.id = :id:", TABLE).
		WithParam(":id:", id)
		// WithCondition("%s = :account_id:", Columns.ID_.Column()).
		// WithParam(":account_id:", sessionAccount.ID())

	return FindFirstJoined(ctx, options)
}

func GetRestricted(ctx context.Context, id types.UUID, _ coremodel.Model) (*Agent, error) {
	options := model.NewOptions().
		WithCondition("%s.id = :id:", TABLE).
		WithParam(":id:", id)
		// WithCondition("%s = :account_id:", Columns.ID_.Column()).
		// WithParam(":account_id:", sessionAccount.ID())

	return FindFirst(ctx, options)
}
