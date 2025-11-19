package object_tag

import (
	"github.com/griffnb/core/lib/sanitize"
)

func (this *ObjectTag) ToPublicJSON() any {
	return sanitize.SanitizeModel(this, &Structure{})
}
