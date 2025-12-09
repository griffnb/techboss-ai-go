package account

import (
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/models/billing_plan"
)

// AddJoinData adds in the join data
func AddJoinData(options *model.Options) {
	options.WithPrependJoins([]string{
		"LEFT JOIN organizations ON organizations.id = accounts.organization_id",
	}...)
	options.WithIncludeFields([]string{
		"concat(accounts.first_name, ' ', accounts.last_name) as name",
		"organizations.name as organization_name",
	}...)
}

type PlanJoins struct {
	// Only on with AddPlans to join
	BillingPlanLevel      *fields.IntField                                       `public:"view" json:"billing_plan_level"       type:"smallint"`
	BillingPlanPrice      *fields.DecimalField                                   `public:"view" json:"billing_plan_price"       type:"numeric"`
	BillingPlanFeatureSet *fields.StructField[*billing_plan.FeatureSet]          `              json:"billing_plan_feature_set" type:"jsonb"`
	FeatureSetOverrides   *fields.StructField[*billing_plan.MergeableFeatureSet] `              json:"feature_set_overrides"    type:"jsonb"    default:"{}"`
	FeatureSet            *fields.StructField[*billing_plan.FeatureSet]          `public:"view" json:"feature_set"              type:"jsonb"`
}

func AddPlans(options *model.Options) *model.Options {
	options.WithPrependJoins([]string{
		"LEFT JOIN organization_subscription_plans ON organization_subscription_plans.id = families.organization_subscription_plan_id",
	}...)
	options.WithIncludeFields([]string{
		"billing_plans.name AS billing_plan_name",
		"billing_plans.id AS billing_plan_id",
		"billing_plans.internal_id AS billing_plan_internal_id",
		"billing_plans.price AS billing_plan_price",
		"billing_plans.feature_set AS billing_plan_feature_set",
		"billing_plans.properties AS billing_plan_properties",
		"billing_plans.level AS billing_plan_level",
		"organizations.feature_set_overrides as feature_set_overrides",
	}...)
	return options
}
