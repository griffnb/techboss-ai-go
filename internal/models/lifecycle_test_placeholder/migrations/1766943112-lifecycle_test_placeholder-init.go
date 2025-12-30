package migrations

import (
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

const TABLE = "lifecycle_test_placeholders"

func init() {
	model.AddMigration(&model.Migration{
		ID:          1766943112,
		Table:       TABLE,
		TableStruct: &LifecycleTestPlaceholderV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type LifecycleTestPlaceholderV1 struct {
	base.Structure
}
