package migrations

import (
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

const TABLE = "billing_plan_prices"

func init() {
	model.AddMigration(&model.Migration{
		ID:          1765642387,
		Table:       TABLE,
		TableStruct: &BillingPlanPriceV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type BillingPlanPriceV1 struct {
	base.Structure
	BillingPlanID *fields.UUIDField             `public:"view" column:"billing_plan_id" type:"uuid"     default:"null" index:"true" null:"true"`
	Name          *fields.StringField           `public:"view" column:"name"            type:"text"     default:""`
	InternalName  *fields.StringField           `              column:"internal_name"   type:"text"     default:""`
	StripePriceID *fields.StringField           `public:"view" column:"stripe_price_id" type:"text"     default:""     index:"true"`
	Price         *fields.DecimalField          `public:"view" column:"price"           type:"numeric"  default:"0"                             scale:"4" precision:"10"`
	Currency      *fields.StringField           `public:"view" column:"currency"        type:"text"     default:"USD"`
	TrialDays     *fields.IntField              `public:"view" column:"trial_days"      type:"smallint" default:"0"`
	BillingCycle  *fields.IntConstantField[int] `public:"view" column:"billing_cycle"   type:"smallint" default:"0"`
}
