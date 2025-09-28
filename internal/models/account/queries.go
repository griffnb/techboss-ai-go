package account

import (
	"context"

	"github.com/CrowdShield/go-core/lib/model"
)

func GetByExternalID(ctx context.Context, externalID string) (*Account, error) {
	return FindFirst(ctx, model.NewOptions().
		WithCondition("%s = :external_id:", Columns.ExternalID.Column()).
		WithCondition("%s = 0", Columns.Disabled.Column()).
		WithParam(":external_id:", externalID))
}
