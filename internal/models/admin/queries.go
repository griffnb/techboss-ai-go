package admin

import (
	"context"

	"github.com/griffnb/core/lib/model"
)

func GetByExternalID(ctx context.Context, externalID string) (*Admin, error) {
	return FindFirst(ctx, model.NewOptions().
		WithCondition("%s = :external_id:", Columns.ExternalID.Column()).
		WithCondition("%s = 0", Columns.Disabled.Column()).
		WithParam(":external_id:", externalID))
}
