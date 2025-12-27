package sandbox

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/types"
)

type Mocker struct {
	// Standard Functions
	Get              func(ctx context.Context, id types.UUID) (*Sandbox, error)
	GetJoined        func(ctx context.Context, id types.UUID) (*SandboxJoined, error)
	FindAll          func(ctx context.Context, options *model.Options) ([]*Sandbox, error)
	FindAllJoined    func(ctx context.Context, options *model.Options) ([]*SandboxJoined, error)
	FindFirst        func(ctx context.Context, options *model.Options) (*Sandbox, error)
	FindFirstJoined  func(ctx context.Context, options *model.Options) (*SandboxJoined, error)
	FindResultsCount func(ctx context.Context, options *model.Options) (int64, error)
	// Custom Functions
	FindByExternalID func(ctx context.Context, externalID string, accountID types.UUID) (*Sandbox, error)
	FindAllByAccount func(ctx context.Context, accountID types.UUID) ([]*Sandbox, error)
	CountByAccount   func(ctx context.Context, accountID types.UUID) (int64, error)
}

// FindByExternalID finds a sandbox by its Modal external ID and AccountID.
// This ensures users can only access their own sandboxes.
// Returns nil if no sandbox is found (not an error).
func FindByExternalID(ctx context.Context, externalID string, accountID types.UUID) (*Sandbox, error) {
	mocker, ok := model.GetMocker[*Mocker](ctx, PACKAGE)
	if ok {
		return mocker.FindByExternalID(ctx, externalID, accountID)
	}

	options := model.NewOptions().
		WithCondition(
			"%s = :external_id: AND %s = :account_id: AND %s = 0 AND %s = 0",
			Columns.ExternalID.Column(),
			Columns.AccountID.Column(),
			Columns.Deleted.Column(),
			Columns.Disabled.Column(),
		).
		WithParam(":external_id:", externalID).
		WithParam(":account_id:", accountID)

	return FindFirst(ctx, options)
}

// FindAllByAccount returns all active sandboxes for a specific account.
// Excludes deleted and disabled sandboxes.
func FindAllByAccount(ctx context.Context, accountID types.UUID) ([]*Sandbox, error) {
	mocker, ok := model.GetMocker[*Mocker](ctx, PACKAGE)
	if ok {
		return mocker.FindAllByAccount(ctx, accountID)
	}

	options := model.NewOptions().
		WithCondition("%s = :account_id: AND %s = 0 AND %s = 0",
			Columns.AccountID.Column(),
			Columns.Deleted.Column(),
			Columns.Disabled.Column()).
		WithParam(":account_id:", accountID)

	return FindAll(ctx, options)
}

// CountByAccount returns the count of active sandboxes for an account.
func CountByAccount(ctx context.Context, accountID types.UUID) (int64, error) {
	mocker, ok := model.GetMocker[*Mocker](ctx, PACKAGE)
	if ok {
		return mocker.CountByAccount(ctx, accountID)
	}

	options := model.NewOptions().
		WithCondition("%s = :account_id: AND %s = 0 AND %s = 0",
			Columns.AccountID.Column(),
			Columns.Deleted.Column(),
			Columns.Disabled.Column()).
		WithParam(":account_id:", accountID)

	return FindResultsCount(ctx, options)
}
