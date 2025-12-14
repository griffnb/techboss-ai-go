package subscription

import (
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
)

type JoinData struct {
	BillingPlanPricePrice    *fields.DecimalField `json:"billing_plan_price_price"    type:"numeric"`
	BillingPlanPriceCurrency *fields.StringField  `json:"billing_plan_price_currency" type:"text"`
	BillingPlanName          *fields.StringField  `json:"billing_plan_name"           type:"text"`
	BillingPlanLevel         *fields.IntField     `json:"billing_plan_level"          type:"smallint"`
}

// AddJoinData adds in the join data
func AddJoinData(options *model.Options) {
	options.WithPrependJoins([]string{
		"LEFT JOIN billing_plan_prices ON subscription.billing_plan_price_id = billing_plan_prices.id",
		"LEFT JOIN billing_plans ON billing_plan_prices.billing_plan_id = billing_plans.id",
	}...)
	options.WithIncludeFields([]string{
		"billing_plan_prices.price AS billing_plan_price_price",
		"billing_plan_prices.currency AS billing_plan_price_currency",
		"billing_plans.name AS billing_plan_name",
		"billing_plans.level AS billing_plan_level",
	}...)
}
