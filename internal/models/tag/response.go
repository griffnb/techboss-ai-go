package tag

import (
	"github.com/griffnb/core/lib/sanitize"
)

func (this *Tag) ToPublicJSON() any {
	return sanitize.SanitizeModel(this, &Structure{})
}
