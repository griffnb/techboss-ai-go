package category

import "github.com/CrowdShield/go-core/lib/sanitize"

// ToPublicJSON converts the model to a sanitized JSON representation for public consumption
func (this *Category) ToPublicJSON() any {
	return sanitize.SanitizeModel(this, &Structure{})
}
