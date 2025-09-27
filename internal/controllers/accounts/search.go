package accounts

import (
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
)

func addSearch(parameters *model.Options, query string) {
	if tools.IsAnyValidUUID(query) {
		parameters.WithCondition("%s.id = :id:", TABLE_NAME)
		parameters.WithParam(":id:", query)
		return
	}

	config := &helpers.SearchConfig{
		TableName: TABLE_NAME,
		DocumentColumns: []string{
			"name",
		},
		RankColumns: map[string][]string{
			"name": {"name"},
		},
		RankOrder: []string{"name"},
	}

	helpers.AddGenericSearch(parameters, query, config)
}
