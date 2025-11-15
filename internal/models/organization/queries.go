package organization

import (
	"context"

	"github.com/CrowdShield/go-core/lib/model"
)

type mocker interface {
	GetByExternalID(ctx context.Context, externalID string) (*Organization, error)
}

// GetByExternalID finds an organization by its external ID (e.g., Stripe organization ID).
// Returns the first non-disabled organization matching the external ID.
func GetByExternalID(ctx context.Context, externalID string) (*Organization, error) {
	mocker, ok := model.GetMocker[mocker](ctx)
	if ok {
		return mocker.GetByExternalID(ctx, externalID)
	}

	return FindFirst(ctx, model.NewOptions().
		WithCondition("%s = :external_id:", Columns.ExternalID.Column()).
		WithCondition("%s = 0", Columns.Disabled.Column()).
		WithParam(":external_id:", externalID))
}

func GetWhitelistDomainOrgs(ctx context.Context) ([]*Organization, error) {
	return FindAll(ctx, model.NewOptions().WithCondition("%s != :email_domains:", Columns.EmailDomains.Column()).
		WithParam(":email_domains:", "[]"))
}
