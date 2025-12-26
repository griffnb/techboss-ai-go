package subscriptions

import (
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/router/route_helpers"
	"github.com/griffnb/core/lib/tools"
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
