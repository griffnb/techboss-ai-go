package organization

import "github.com/griffnb/techboss-ai-go/internal/common"

func TESTCreateOrganization() *Organization {
	obj := New()
	common.GenerateURN(obj)
	return obj
}
