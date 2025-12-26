package billing_plan_price

import (
	"github.com/griffnb/core/lib/model"
)

// AddJoinData adds in the join data
func AddJoinData(options *model.Options) {
	options.WithPrependJoins([]string{
		"LEFT JOIN billing_plans ON billing_plans.id = billing_plan_prices.billing_plan_id",
	}...)
	options.WithIncludeFields([]string{
		"billing_plans.stripe_product_id AS billing_plan_stripe_product_id",
	}...)
}
