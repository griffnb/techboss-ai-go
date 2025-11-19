package billing_plan

import "github.com/griffnb/core/lib/sanitize"

// ToPublicJSON converts the model to a sanitized JSON representation for public consumption
func (this *BillingPlan) ToPublicJSON() any {
	return sanitize.SanitizeModel(this, &Structure{})
}
