package account

import (
	"context"
	"strings"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/tools"
)

type mocker interface {
	GetByExternalID(ctx context.Context, externalID string) (*Account, error)
	GetByEmail(ctx context.Context, email string) (*Account, error)
	GetExistingByEmail(ctx context.Context, email string) (*Account, error)
}

func Exists(ctx context.Context, email string) (bool, error) {
	accountObj, err := GetExistingByEmail(ctx, email)
	if err != nil {
		return false, err
	}
	return !tools.Empty(accountObj), nil
}

// GetByExternalID finds an account by its external ID (e.g., Stripe customer ID).
// Returns the first non-disabled account matching the external ID.
func GetByExternalID(ctx context.Context, externalID string) (*Account, error) {
	mocker, ok := model.GetMocker[mocker](ctx)
	if ok {
		return mocker.GetByExternalID(ctx, externalID)
	}

	return FindFirst(ctx, model.NewOptions().
		WithCondition("%s = :external_id:", Columns.ExternalID.Column()).
		WithCondition("%s = 0", Columns.Disabled.Column()).
		WithParam(":external_id:", externalID))
}

// GetExistingByEmail finds an account by its email, including disabled accounts.
func GetExistingByEmail(ctx context.Context, email string) (*Account, error) {
	return FindFirst(ctx, model.NewOptions().WithCondition("lower(%s) = :email: AND %s = 0 ",
		Columns.Email.Column(),
		Columns.Deleted.Column()).
		WithParam(":email:", strings.ToLower(email)))
}

// GetByEmail finds an active (non-disabled) account by its email.
func GetByEmail(ctx context.Context, email string) (*Account, error) {
	return FindFirst(ctx, model.NewOptions().WithCondition("lower(%s) = :email: AND %s = 0 ",
		Columns.Email.Column(),
		Columns.Disabled.Column()).
		WithParam(":email:", strings.ToLower(email)))
}
