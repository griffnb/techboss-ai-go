package migrations

import (
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
)

func init() {
	model.AddMigration(&model.Migration{
		ID:    1759071465,
		Table: account.TABLE,
		ColumnMigrations: []*model.ColumnMigration{
			{
				Type:       model.ADD_COLUMN,
				ColumnName: "organization_id",
				Properties: &model.Property{
					Type:     model.TYPE_UUID,
					Default:  "null",
					Nullable: true,
					Indexed:  true,
				},
			},
		},
	})
}
