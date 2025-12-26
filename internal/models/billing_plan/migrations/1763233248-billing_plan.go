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
	Name            *fields.StringField      `public:"view" column:"name"              type:"text"     default:""`
	Description     *fields.StringField      `public:"view" column:"description"       type:"text"     default:""`
	InternalName    *fields.StringField      `              column:"internal_name"     type:"text"     default:""`
	FeatureSet      *fields.StructField[any] `public:"view" column:"feature_set"       type:"jsonb"    default:"{}"`
	Properties      *fields.StructField[any] `public:"view" column:"properties"        type:"jsonb"    default:"{}"`
	StripeProductID *fields.StringField      `public:"view" column:"stripe_product_id" type:"text"     default:""   index:"true"`
	Level           *fields.IntField         `public:"view" column:"level"             type:"smallint" default:"0"`
	IsDefault       *fields.IntField         `public:"view" column:"is_default"        type:"smallint" default:"0"  index:"true"`
}
