package billing_plan_price

type BillingCycle int

const (
	BILLING_CYCLE_MONTHLY BillingCycle = iota + 1
	BILLING_CYCLE_QUARTERLY
	BILLING_CYCLE_ANNUALLY
)

func (bc BillingCycle) ToStripe() string {
	switch bc {
	case BILLING_CYCLE_MONTHLY:
		return "month"
	case BILLING_CYCLE_QUARTERLY:
		return "quarter"
	case BILLING_CYCLE_ANNUALLY:
		return "year"
	default:
		return "month"
	}
}
