package billing_plan_price

// This file contains additional helper functions for the BillingPlanPrice model
func (this *BillingPlanPrice) HasStripeChanges() bool {
	return this.Price.HasChanged()
}
