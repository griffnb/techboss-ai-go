package billing_plan

type Properties struct {
	PricingText         string `public:"view" json:"pricing_text"`
	StripePriceID       string `public:"view" json:"stripe_price_id"`       // The plan that will be billed
	DefaultDiscountCode string `              json:"default_discount_code"` // Used if we want to use a standard plan but say its a discount
}
