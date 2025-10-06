package ai_tool

type MetaData struct {
	Logo           string         `json:"logo"`
	Tagline        string         `json:"tagline"`
	KeyBenefits    []string       `json:"benefits"`
	Introduction   string         `json:"introduction"`
	HowItWorks     string         `json:"how_it_works"`
	Features       []*CoreFeature `json:"features"`
	Applications   []string       `json:"applications"`
	FreeTier       bool           `json:"free_tier"`
	PricingRange   string         `json:"price_range"`
	PricingOptions string         `json:"price_options"`
	TargetAudience string         `json:"target_audience"`
}

type CoreFeature struct {
	Feature     string `json:"feature"`
	Description string `json:"description"`
}
