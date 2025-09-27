package lead

// Utms represents the UTM parameters stored in JSONB
type Utms struct {
	UtmCampaign string `json:"utm_campaign,omitempty"`
	UtmSource   string `json:"utm_source,omitempty"`
	UtmKeyword  string `json:"utm_keyword,omitempty"`
	UtmContent  string `json:"utm_content,omitempty"`
}
