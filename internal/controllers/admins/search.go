package admins

import (
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/router/route_helpers"
)

func addSearch(parameters *model.Options, query string) {
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
