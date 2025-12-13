package billing_plan

type Prices struct {
	Monthly  *Price `public:"view" json:"monthly"`
	Annually *Price `public:"view" json:"annually"`
}

type Price struct {
	StripePriceID string `public:"view" json:"stripe_price_id"` // The plan that will be billed
	Amount        int64  `public:"view" json:"amount"`          // in cents
	Currency      string `public:"view" json:"currency"`        // USD, etc.
}
