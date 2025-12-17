package migrations

import (
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
	"github.com/griffnb/techboss-ai-go/internal/models/billing_plan_price"
)

const TABLE string = "subscriptions"

func init() {
	model.AddMigration(&model.Migration{
		ID:          1763231248,
		Table:       TABLE,
		TableStruct: &SubscriptionV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type SubscriptionV1 struct {
	base.Structure
	OrganizationID       *fields.UUIDField                                         `column:"organization_id"        type:"uuid"     default:"null" index:"true" null:"true" public:"view"`
	BillingPlanPriceID   *fields.UUIDField                                         `column:"billing_plan_price_id"  type:"uuid"     default:"null" index:"true" null:"true" public:"view"`
	Level                *fields.IntField                                          `column:"level"                  type:"smallint" default:"0"`
	BillingProvider      *fields.IntConstantField[int]                             `column:"billing_provider"       type:"smallint" default:"0"    index:"true"`
	StripeCustomerID     *fields.StringField                                       `column:"stripe_customer_id"     type:"text"     default:""     index:"true"`
	StripeSubscriptionID *fields.StringField                                       `column:"stripe_subscription_id" type:"text"     default:""     index:"true"`
	StripePriceID        *fields.StringField                                       `column:"stripe_price_id"        type:"text"     default:""     index:"true"`
	StartTS              *fields.IntField                                          `column:"start_ts"               type:"bigint"   default:"0"                             public:"view"`
	EndTS                *fields.IntField                                          `column:"end_ts"                 type:"bigint"   default:"0"                             public:"view"`
	TrialEndTS           *fields.IntField                                          `column:"trial_end_ts"           type:"bigint"   default:"0"                             public:"view"`
	InTrial              *fields.IntField                                          `column:"in_trial"               type:"smallint" default:"0"`
	NextBillingTS        *fields.IntField                                          `column:"next_billing_ts"        type:"bigint"   default:"0"                             public:"view"`
	BillingCycle         *fields.IntConstantField[billing_plan_price.BillingCycle] `column:"billing_cycle"          type:"smallint" default:"0"`
	Amount               *fields.DecimalField                                      `column:"amount"                 type:"numeric"  default:"0"                             public:"view" scale:"4" precision:"10"`
	CouponCode           *fields.StringField                                       `column:"coupon_code"            type:"text"     default:""                              public:"view"`
	BillingInfo          *fields.StructField[any]                                  `column:"billing_info"           type:"jsonb"    default:"{}"                            public:"view"`
	MetaData             *fields.StructField[any]                                  `column:"meta_data"              type:"jsonb"    default:"{}"                            public:"view"`
}
