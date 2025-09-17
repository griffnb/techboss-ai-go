package admin

import (
	"fmt"

	"github.com/CrowdShield/go-core/lib/model"
)

// AddJoinData adds in the join data
func AddJoinData(options *model.Options) {
	options.WithPrependJoins([]string{}...)
	options.WithIncludeFields([]string{}...)
}

func JoinCreatedUpdatedQuery(targetTable string) string {
	return fmt.Sprintf(`
		LEFT JOIN admins updated_admin on %s.updated_by_urn = updated_admin.urn
		LEFT JOIN admins created_admin on %s.created_by_urn = created_admin.urn
	
	`, targetTable, targetTable)
}

func JoinCreatedUpdatedField() []string {
	return []string{
		"updated_admin.name as updated_by_name",
		"created_admin.name as created_by_name",
	}
}
