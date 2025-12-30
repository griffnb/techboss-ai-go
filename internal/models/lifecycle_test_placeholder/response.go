package lifecycle_test_placeholder

import "github.com/griffnb/core/lib/sanitize"

// ToPublicJSON converts the model to a sanitized JSON representation for public consumption
func (this *LifecycleTestPlaceholder) ToPublicJSON() any {
	return sanitize.SanitizeModel(this, &Structure{})
}
