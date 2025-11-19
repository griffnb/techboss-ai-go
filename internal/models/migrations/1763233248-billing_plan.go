package migrations

import (
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
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
	Name         *fields.StringField           `column:"name"          type:"text"     default:""`
	InternalName *fields.StringField           `column:"internal_name" type:"text"     default:""`
	BillingCycle *fields.IntConstantField[int] `column:"billing_cycle" type:"smallint" default:"0"`
	Price        *fields.DecimalField          `column:"price"         type:"numeric"  default:"0"  public:"view" scale:"4" precision:"10"`
	FeatureSet   *fields.StructField[any]      `column:"feature_set"   type:"jsonb"    default:"{}" public:"view"`
	Properties   *fields.StructField[any]      `column:"properties"    type:"jsonb"    default:"{}" public:"view"`
	Level        *fields.IntField              `column:"level"         type:"smallint" default:"0"  public:"view"`
	IsDefault    *fields.IntField              `column:"is_default"    type:"smallint" default:"0"                                         index:"true"`
}
