package documents

import (
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/router/route_helpers"
	"github.com/CrowdShield/go-core/lib/tools"
)

func addSearch(parameters *model.Options, query string) {
	if tools.IsAnyValidUUID(query) {
		parameters.WithCondition("%s.id = :id:", TABLE_NAME)
		parameters.WithParam(":id:", query)
		return
	}

	config := &route_helpers.SearchConfig{
		TableName: TABLE_NAME,
		DocumentColumns: []string{
			"name",
		},
		RankColumns: map[string][]string{
			"name": {"name"},
		},
		RankOrder: []string{"name"},
	}

	route_helpers.AddGenericSearch(parameters, query, config)
}
