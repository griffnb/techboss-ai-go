package subscription

type BillingProvider int

const (
	BILLING_PROVIDER_STRIPE BillingProvider = iota + 1
)

func (bp BillingProvider) String() string {
	switch bp {

	case BILLING_PROVIDER_STRIPE:
		return "Stripe"
	default:
		return "Unknown"
	}
}
