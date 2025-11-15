package migrations

import (
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
	"github.com/griffnb/techboss-ai-go/internal/models/subscription"
)

func init() {
	model.AddMigration(&model.Migration{
		ID:          1763231248,
		Table:       subscription.TABLE,
		TableStruct: &SubscriptionV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type SubscriptionV1 struct {
	base.Structure
	OrganizationID  *fields.UUIDField             `column:"organization_id" type:"uuid" default:"null" index:"true" null:"true" public:"view"`
	BillingPlanID   *fields.UUIDField             `column:"billing_plan_id" type:"uuid" default:"null" index:"true" null:"true" public:"view"`
	Level           *fields.IntField              `column:"level"            type:"int"      default:"0"`
	BillingProvider *fields.IntConstantField[int] `column:"billing_provider" type:"smallint" default:"0" index:"true"`
	SubscriptionID  *fields.StringField           `column:"subscription_id"  type:"text"     default:""`
	PriceOrPlanID   *fields.StringField           `column:"price_or_plan_id" type:"text"     default:""`
	StartTS         *fields.IntField              `column:"start_ts"         type:"bigint"   default:"0"  public:"view"`
	EndTS           *fields.IntField              `column:"end_ts"           type:"bigint"   default:"0"  public:"view"`
	NextBillingTS   *fields.IntField              `column:"next_billing_ts"  type:"bigint"   default:"0"  public:"view"`
	BillingCycle    *fields.IntConstantField[int] `column:"billing_cycle"    type:"smallint" default:"0"`
	Amount          *fields.DecimalField          `column:"amount"           type:"numeric"  default:"0"  public:"view" scale:"4" precision:"10"`
	CouponCode      *fields.StringField           `column:"coupon_code"      type:"text"     default:""   public:"view"`
	BillingInfo     *fields.StructField[any]      `column:"billing_info"     type:"jsonb"    default:"{}" public:"view"`
	MetaData        *fields.StructField[any]      `column:"meta_data"        type:"jsonb"    default:"{}" public:"view"`
}
