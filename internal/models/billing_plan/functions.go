package billing_plan

// This file contains additional helper functions for the BillingPlan model

func (this *BillingPlan) HasStripeChanges() bool {
	if this.Name.HasChanged() || this.Description.HasChanged() {
		return true
	}
	return false
}
