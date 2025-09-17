package admins

import (
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
)

func addSearch(parameters *model.Options, query string) {
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
