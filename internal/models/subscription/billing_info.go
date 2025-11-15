package subscription

type BillingInfo struct {
	CardType     string `public:"view" json:"card_type"`
	CardLast4    string `public:"view" json:"card_last4"`
	CardExpMonth int    `public:"view" json:"card_exp_month"`
	CardExpYear  int    `public:"view" json:"card_exp_year"`
	CardAddress1 string `public:"view" json:"card_address1"`
	CardAddress2 string `public:"view" json:"card_address2"`
	CardCity     string `public:"view" json:"card_city"`
	CardState    string `public:"view" json:"card_state"`
	CardZip      string `public:"view" json:"card_zip"`
	CardCountry  string `public:"view" json:"card_country"`
}
