//go:generate core_generate model Subscription
package subscription

import (
	"context"

	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/common"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

// Constants for the model
const (
	TABLE        = "subscriptions"
	CHANGE_LOGS  = true
	CLIENT       = environment.CLIENT_DEFAULT
	IS_VERSIONED = false
)

type Structure struct {
	DBColumns
	JoinData
}

type DBColumns struct {
	base.Structure
	OrganizationID *fields.UUIDField `column:"organization_id" type:"uuid" default:"null" index:"true" null:"true" public:"view"`

	Level           *fields.IntField                          `column:"level"            type:"int"      default:"0"`
	BillingProvider *fields.IntConstantField[BillingProvider] `column:"billing_provider" type:"smallint" default:"0" index:"true"`
	SubscriptionID  *fields.StringField                       `column:"subscription_id"  type:"text"     default:""`

	PriceOrPlanID *fields.StringField                    `column:"price_or_plan_id" type:"text"     default:""`
	StartTS       *fields.IntField                       `column:"start_ts"         type:"bigint"   default:"0"  public:"view"`
	EndTS         *fields.IntField                       `column:"end_ts"           type:"bigint"   default:"0"  public:"view"`
	NextBillingTS *fields.IntField                       `column:"next_billing_ts"  type:"bigint"   default:"0"  public:"view"`
	BillingCycle  *fields.IntConstantField[BillingCycle] `column:"billing_cycle"    type:"smallint" default:"0"`
	Amount        *fields.DecimalField                   `column:"amount"           type:"numeric"  default:"0"  public:"view" scale:"4" precision:"10"`
	CouponCode    *fields.StringField                    `column:"coupon_code"      type:"text"     default:""   public:"view"`
	BillingInfo   *fields.StructField[*BillingInfo]      `column:"billing_info"     type:"jsonb"    default:"{}" public:"view"`
	MetaData      *fields.StructField[*MetaData]         `column:"meta_data"        type:"jsonb"    default:"{}" public:"view"`
}

type JoinData struct {
	CreatedByName *fields.StringField `json:"created_by_name" type:"text"`
	UpdatedByName *fields.StringField `json:"updated_by_name" type:"text"`
}

// Subscription - Database model
type Subscription struct {
	model.BaseModel
	DBColumns
}

type SubscriptionJoined struct {
	Subscription
	JoinData
}

func (this *Subscription) beforeSave(ctx context.Context) error {
	this.BaseBeforeSave(ctx)
	common.GenerateURN(this)
	common.SetDisabledDeleted(this)
	return this.ValidateSubStructs()
}

func (this *Subscription) afterSave(ctx context.Context) {
	this.BaseAfterSave(ctx)
	/*
		go func() {
			err := this.UpdateCache()
			if err != nil {
				log.Error(err)
			}
		}()
	*/
}
