package migrations

import (
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
	"github.com/griffnb/techboss-ai-go/internal/models/billing_plan"
)

func init() {
	model.AddMigration(&model.Migration{
		ID:          1763233248,
		Table:       billing_plan.TABLE,
		TableStruct: &BillingPlanV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type BillingPlanV1 struct {
	base.Structure
}
