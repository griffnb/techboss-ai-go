package account

import (
	"context"
	"strings"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/core/lib/types"
)

type Mocker struct {
	// Standard Functions
	Get              func(ctx context.Context, id types.UUID) (*Account, error)
	GetJoined        func(ctx context.Context, id types.UUID) (*AccountJoined, error)
	FindAll          func(ctx context.Context, options *model.Options) ([]*Account, error)
	FindAllJoined    func(ctx context.Context, options *model.Options) ([]*AccountJoined, error)
	FindFirst        func(ctx context.Context, options *model.Options) (*Account, error)
	FindFirstJoined  func(ctx context.Context, options *model.Options) (*AccountJoined, error)
	FindResultsCount func(ctx context.Context, options *model.Options) (int64, error)
	// Custom Functions
	GetByExternalId    func(ctx context.Context, externalID string) (*Account, error)
	GetByEmail         func(ctx context.Context, email string) (*Account, error)
	GetExistingByEmail func(ctx context.Context, email string) (*Account, error)
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
	mocker, ok := model.GetMocker[*Mocker](ctx, PACKAGE)
	if ok {
		return mocker.GetByExternalId(ctx, externalID)
	}

	return FindFirst(ctx, model.NewOptions().
		WithCondition("%s = :external_id:", Columns.ExternalID.Column()).
		WithCondition("%s = 0", Columns.Disabled.Column()).
		WithParam(":external_id:", externalID))
}

// GetExistingByEmail finds an account by its email, including disabled accounts.
func GetExistingByEmail(ctx context.Context, email string) (*Account, error) {
	mocker, ok := model.GetMocker[*Mocker](ctx, PACKAGE)
	if ok {
		return mocker.GetExistingByEmail(ctx, email)
	}
	return FindFirst(ctx, model.NewOptions().WithCondition("lower(%s) = :email: AND %s = 0 ",
		Columns.Email.Column(),
		Columns.Deleted.Column()).
		WithParam(":email:", strings.ToLower(email)))
}

// GetByEmail finds an active (non-disabled) account by its email.
func GetByEmail(ctx context.Context, email string) (*Account, error) {
	mocker, ok := model.GetMocker[*Mocker](ctx, PACKAGE)
	if ok {
		return mocker.GetByEmail(ctx, email)
	}
	return FindFirst(ctx, model.NewOptions().WithCondition("lower(%s) = :email: AND %s = 0 ",
		Columns.Email.Column(),
		Columns.Disabled.Column()).
		WithParam(":email:", strings.ToLower(email)))
}
