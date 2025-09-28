package conversation

import (
	"context"

	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/model/coremodel"
	"github.com/CrowdShield/go-core/lib/sanitize"
)

func FindAllRestrictedJoined(ctx context.Context, options *model.Options, sessionUser coremodel.Model) ([]*ConversationJoined, error) {
	options.WithCondition("%s = :account_id:", Columns.AccountID.Column()).WithParam(":account_id:", sessionUser.ID())
	return FindAllJoined(ctx, options)
}

func GetRestrictedJoined(ctx context.Context, id int64, sessionUser coremodel.Model) (*ConversationJoined, error) {
	options := model.NewOptions().
		WithCondition("%s = :id: AND %s = :account_id:", Columns.ID_.Column(), Columns.AccountID.Column()).
		WithParam(":id:", id).
		WithParam(":account_id:", sessionUser.ID())

	return FindFirstJoined(ctx, options)
}

// NewPublic creates a new model instance with sanitized input and session account context
// TODO: Add any session-specific initialization
func NewPublic(data map[string]any, _ coremodel.Model) *Conversation {
	obj := New()
	data = sanitize.SanitizeModelInput(data, obj, &Structure{})
	obj.MergeData(data)
	// obj.AccountID.Set(sessionAccount.ID())
	return obj
}

func UpdatePublic(obj *Conversation, data map[string]any, _ coremodel.Model) {
	data = sanitize.SanitizeModelInput(data, obj, &Structure{})
	obj.MergeData(data)
}
