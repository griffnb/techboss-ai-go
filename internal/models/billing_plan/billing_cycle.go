package billing_plan

type BillingCycle int

const (
	BILLING_CYCLE_MONTHLY BillingCycle = iota + 1
	BILLING_CYCLE_QUARTERLY
	BILLING_CYCLE_ANNUALLY
)
