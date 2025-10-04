package tag

import (
	"github.com/CrowdShield/go-core/lib/sanitize"
)

func (this *Tag) ToPublicJSON() any {
	return sanitize.SanitizeModel(this, &Structure{})
}
