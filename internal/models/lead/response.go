package lead

import "github.com/CrowdShield/go-core/lib/sanitize"

// ToPublicJSON converts the model to a sanitized JSON representation for public consumption
func (this *Lead) ToPublicJSON() any {
	return sanitize.SanitizeModel(this, &Structure{})
}
