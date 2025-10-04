package ai_tool

type MetaData struct {
	Logo             string         `json:"logo"`
	Tagline          string         `json:"tagline"`
	KeyBenefits      []string       `json:"keyBenefits"`
	Introduction     string         `json:"introduction"`
	HowItWorks       string         `json:"howItWorks"`
	CoreFeatures     []*CoreFeature `json:"coreFeatures"`
	Applications     []string       `json:"applications"`
	FreeTier         bool           `json:"freeTier"`
	PricingRange     string         `json:"pricingRange"`
	PricingOptions   string         `json:"pricingOptions"`
	TargetAudience   string         `json:"targetAudience"`
	Categorization   string         `json:"categorization"`
	BusinessFunction string         `json:"businessFunction"`
}

type CoreFeature struct {
	Feature     string `json:"feature"`
	Description string `json:"description"`
}
