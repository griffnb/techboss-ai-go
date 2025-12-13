package migrations

import (
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

const TABLE string = "billing_plans"

func init() {
	model.AddMigration(&model.Migration{
		ID:          1763233248,
		Table:       TABLE,
		TableStruct: &BillingPlanV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type BillingPlanV1 struct {
	base.Structure
	Name         *fields.StringField      `column:"name"          type:"text"     default:""`
	InternalName *fields.StringField      `column:"internal_name" type:"text"     default:""`
	FeatureSet   *fields.StructField[any] `column:"feature_set"   type:"jsonb"    default:"{}" public:"view"`
	Properties   *fields.StructField[any] `column:"properties"    type:"jsonb"    default:"{}" public:"view"`
	Prices       *fields.StructField[any] `column:"prices"        type:"jsonb"    default:"{}" public:"view"`
	Level        *fields.IntField         `column:"level"         type:"smallint" default:"0"  public:"view"`
	IsDefault    *fields.IntField         `column:"is_default"    type:"smallint" default:"0"                index:"true"`
}
