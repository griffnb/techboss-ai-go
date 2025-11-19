package lead

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/coremodel"
	"github.com/griffnb/core/lib/sanitize"
	"github.com/griffnb/core/lib/types"
)

// FindAllRestrictedJoined returns all joined records with restrictions for the session account
// TODO: Implement specific access restrictions for this model
func FindAllRestrictedJoined(ctx context.Context, options *model.Options, _ coremodel.Model) ([]*LeadJoined, error) {
	// Uncomment and adjust the following lines to implement proper restrictions
	// options.WithCondition("%s = :account_id:", Columns.AccountID.Column())
	// options.WithParam(":account_id:", sessionAccount.ID())
	return FindAllJoined(ctx, options)
}

func FindAllRestricted(ctx context.Context, options *model.Options, _ coremodel.Model) ([]*Lead, error) {
	// Uncomment and adjust the following lines to implement proper restrictions
	// options.WithCondition("%s = :account_id:", Columns.AccountID.Column())
	// options.WithParam(":account_id:", sessionAccount.ID())
	return FindAll(ctx, options)
}

// GetRestrictedJoined gets a specific record with joined data and restrictions for the session account
// TODO: Adjust the restrictions to match your access control requirements
func GetRestrictedJoined(ctx context.Context, id types.UUID, _ coremodel.Model) (*LeadJoined, error) {
	options := model.NewOptions().
		WithCondition("%s.id = :id:", TABLE).
		WithParam(":id:", id)
	// TODO: Add appropriate restrictions for lead access
	// For now, allowing access to any lead - adjust based on business requirements

	return FindFirstJoined(ctx, options)
}

func GetRestricted(ctx context.Context, id types.UUID, _ coremodel.Model) (*Lead, error) {
	options := model.NewOptions().
		WithCondition("%s.id = :id:", TABLE).
		WithParam(":id:", id)
	// TODO: Add appropriate restrictions for lead access
	// For now, allowing access to any lead - adjust based on business requirements

	return FindFirst(ctx, options)
}

// NewPublic creates a new model instance with sanitized input and session account context
// TODO: Add any session-specific initialization
func NewPublic(data map[string]any, _ coremodel.Model) *Lead {
	obj := New()
	data = sanitize.SanitizeModelInput(data, obj, &Structure{})
	obj.MergeData(data)
	// obj.AccountID.Set(sessionAccount.ID())
	return obj
}

func UpdatePublic(obj *Lead, data map[string]any, _ coremodel.Model) {
	data = sanitize.SanitizeModelInput(data, obj, &Structure{})
	obj.MergeData(data)
}
