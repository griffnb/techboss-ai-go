package account

import (
	"context"

	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/model/coremodel"
	"github.com/CrowdShield/go-core/lib/sanitize"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/CrowdShield/go-core/lib/types"
)

// FindAllRestrictedJoined returns all joined records with restrictions for the session account
// TODO: Implement specific access restrictions for this model
func FindAllRestrictedJoined(ctx context.Context, options *model.Options, _ coremodel.Model) ([]*AccountJoined, error) {
	// Uncomment and adjust the following lines to implement proper restrictions
	// options.WithCondition("%s = :account_id:", Columns.AccountID.Column())
	// options.WithParam(":account_id:", sessionAccount.ID())
	return FindAllJoined(ctx, options)
}

func FindAllRestricted(ctx context.Context, options *model.Options, _ coremodel.Model) ([]*Account, error) {
	// Uncomment and adjust the following lines to implement proper restrictions
	// options.WithCondition("%s = :account_id:", Columns.AccountID.Column())
	// options.WithParam(":account_id:", sessionAccount.ID())
	return FindAll(ctx, options)
}

// GetRestrictedJoined gets a specific record with joined data and restrictions for the session account
// TODO: Adjust the restrictions to match your access control requirements
func GetRestrictedJoined(ctx context.Context, id types.UUID, sessionAccount coremodel.Model) (*AccountJoined, error) {
	options := model.NewOptions().
		WithCondition("%s.id = :id:", TABLE).
		WithParam(":id:", id).
		WithCondition("%s = :account_id:", Columns.ID_.Column()).
		WithParam(":account_id:", sessionAccount.ID())

	return FindFirstJoined(ctx, options)
}

func GetRestricted(ctx context.Context, id types.UUID, sessionAccount coremodel.Model) (*Account, error) {
	options := model.NewOptions().
		WithCondition("%s.id = :id:", TABLE).
		WithParam(":id:", id).
		WithCondition("%s = :account_id:", Columns.ID_.Column()).
		WithParam(":account_id:", sessionAccount.ID())

	return FindFirst(ctx, options)
}

// NewPublic creates a new model instance with sanitized input and session account context
// TODO: Add any session-specific initialization
func NewPublic(data map[string]any, _ coremodel.Model) *Account {
	obj := New()
	data = sanitize.SanitizeModelInput(data, obj, &Structure{})
	obj.MergeData(data)
	// obj.AccountID.Set(sessionAccount.ID())
	return obj
}

func UpdatePublic(obj *Account, data map[string]any, _ coremodel.Model) {
	data = sanitize.SanitizeModelInput(data, obj, &Structure{})
	obj.MergeData(data)
}

// MergeSignup Bypasses the filter for signup_properties, we only want to do this on the signup page
func MergeSignup(obj *Account, data map[string]interface{}, sessionAccount *Account) {
	signupProperties := data["signup_properties"]
	password := data["password"]
	cleanData := sanitize.SanitizeModelInput(data, obj, &Structure{})
	cleanData["signup_properties"] = signupProperties
	cleanData["password"] = password
	obj.MergeData(cleanData)

	if !tools.Empty(sessionAccount) {
		obj.OrganizationID.Set(sessionAccount.OrganizationID.Get())
	}
}
