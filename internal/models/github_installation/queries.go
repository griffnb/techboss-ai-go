package github_installation

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/types"
)

type Mocker struct {
	// Standard Functions
	Get              func(ctx context.Context, id types.UUID) (*GithubInstallation, error)
	GetJoined        func(ctx context.Context, id types.UUID) (*GithubInstallationJoined, error)
	FindAll          func(ctx context.Context, options *model.Options) ([]*GithubInstallation, error)
	FindAllJoined    func(ctx context.Context, options *model.Options) ([]*GithubInstallationJoined, error)
	FindFirst        func(ctx context.Context, options *model.Options) (*GithubInstallation, error)
	FindFirstJoined  func(ctx context.Context, options *model.Options) (*GithubInstallationJoined, error)
	FindResultsCount func(ctx context.Context, options *model.Options) (int64, error)
	// Custom Functions
	GetByInstallationID  func(ctx context.Context, installationID string) (*GithubInstallation, error)
	GetByAccountID       func(ctx context.Context, accountID types.UUID) ([]*GithubInstallation, error)
	GetActiveByAccountID func(ctx context.Context, accountID types.UUID) ([]*GithubInstallation, error)
}

// GetByInstallationID finds a GitHub installation by its installation ID.
// Returns the first non-disabled installation matching the installation ID.
func GetByInstallationID(ctx context.Context, installationID string) (*GithubInstallation, error) {
	mocker, ok := model.GetMocker[*Mocker](ctx, PACKAGE)
	if ok {
		return mocker.GetByInstallationID(ctx, installationID)
	}

	return FindFirst(ctx, model.NewOptions().
		WithCondition("%s = :installation_id:", Columns.InstallationID.Column()).
		WithCondition("%s = 0", Columns.Disabled.Column()).
		WithParam(":installation_id:", installationID))
}

// GetByAccountID finds all GitHub installations for a specific account.
// Returns all non-disabled installations for the account.
func GetByAccountID(ctx context.Context, accountID types.UUID) ([]*GithubInstallation, error) {
	mocker, ok := model.GetMocker[*Mocker](ctx, PACKAGE)
	if ok {
		return mocker.GetByAccountID(ctx, accountID)
	}

	return FindAll(ctx, model.NewOptions().
		WithCondition("%s = :account_id:", Columns.AccountID.Column()).
		WithCondition("%s = 0", Columns.Disabled.Column()).
		WithParam(":account_id:", accountID))
}

// GetActiveByAccountID finds all active (non-suspended) GitHub installations for a specific account.
// Returns all non-disabled, non-suspended installations for the account.
func GetActiveByAccountID(ctx context.Context, accountID types.UUID) ([]*GithubInstallation, error) {
	mocker, ok := model.GetMocker[*Mocker](ctx, PACKAGE)
	if ok {
		return mocker.GetActiveByAccountID(ctx, accountID)
	}

	return FindAll(ctx, model.NewOptions().
		WithCondition("%s = :account_id:", Columns.AccountID.Column()).
		WithCondition("%s = 0", Columns.Disabled.Column()).
		WithCondition("%s = 0", Columns.Suspended.Column()).
		WithParam(":account_id:", accountID))
}
