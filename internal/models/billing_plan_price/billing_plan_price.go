//go:generate core_gen model BillingPlanPrice
package billing_plan_price

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/common"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
	_ "github.com/griffnb/techboss-ai-go/internal/models/billing_plan_price/migrations"
)

// Constants for the model
const (
	TABLE        = "billing_plan_prices"
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
	BillingPlanID *fields.UUIDField                      `public:"view" column:"billing_plan_id" type:"uuid"     default:"null" index:"true" null:"true"`
	Name          *fields.StringField                    `public:"view" column:"name"            type:"text"     default:""`
	InternalName  *fields.StringField                    `              column:"internal_name"   type:"text"     default:""`
	StripePriceID *fields.StringField                    `public:"view" column:"stripe_price_id" type:"text"     default:""     index:"true"`
	Price         *fields.DecimalField                   `public:"view" column:"price"           type:"numeric"  default:"0"                             scale:"4" precision:"10"`
	TrialDays     *fields.IntField                       `public:"view" column:"trial_days"      type:"smallint" default:"0"`
	Currency      *fields.StringField                    `public:"view" column:"currency"        type:"text"     default:"USD"`
	BillingCycle  *fields.IntConstantField[BillingCycle] `public:"view" column:"billing_cycle"   type:"smallint" default:"0"`
}

type JoinData struct {
	BillingPlanStripeProductID *fields.StringField `json:"billing_plan_stripe_product_id" type:"text"`
}

// BillingPlanPrice - Database model
type BillingPlanPrice struct {
	model.BaseModel
	DBColumns
}

type BillingPlanPriceJoined struct {
	BillingPlanPrice
	JoinData
}

func (this *BillingPlanPrice) beforeSave(ctx context.Context) error {
	this.BaseBeforeSave(ctx)
	common.GenerateURN(this)
	common.SetDisabledDeleted(this)
	return this.ValidateSubStructs()
}

func (this *BillingPlanPrice) afterSave(ctx context.Context) {
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
