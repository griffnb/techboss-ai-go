package helpers

import (
	"fmt"
	"strings"

	"github.com/CrowdShield/go-core/lib/model"
)

type SearchConfig struct {
	TableName       string
	DocumentColumns []string            // All the columns to text search across using TSVector for matches
	RankColumns     map[string][]string // Combined columns to rank search results by
	RankOrder       []string            // Order to rank columns or combined columns by
}

func AddGenericSearch(parameters *model.Options, query string, config *SearchConfig) {
	parameters.WithParam(":q:", strings.ToLower(query))
	parameters.WithJoins(`CROSS JOIN websearch_to_tsquery(:q:) _query`)
	documentColumns := formatColumns(config.DocumentColumns, config.TableName)
	parameters.WithJoins(`CROSS JOIN to_tsvector(` + strings.Join(documentColumns, " || ' ' || ") + `) _document`)
	parameters.WithJoins(`CROSS JOIN word_similarity(:q:, ` + strings.Join(documentColumns, " || ' ' || ") + `) _similarity`)

	for rankColumn, rawColumns := range config.RankColumns {
		columns := formatColumns(rawColumns, config.TableName)
		parameters.WithJoins(
			fmt.Sprintf(`CROSS JOIN  NULLIF(ts_rank_cd(to_tsvector(%s), _query), 0) _rank_%s`, strings.Join(columns, " || ' ' || "), rankColumn),
		)
	}

	parameters.WithCondition(`( _query @@ _document OR _similarity > 0 )`)

	orders := []string{}
	for _, rankColumn := range config.RankOrder {
		orders = append(orders, `_rank_`+rankColumn)
	}
	orders = append(orders, "_similarity")
	parameters.Order = strings.Join(orders, ",") + " DESC NULLS LAST"
}

func AddV1Search(parameters *model.Options, query string, config *SearchConfig) {
	parameters.WithParam(":q:", strings.ToLower(query))
	parameters.WithJoins(`CROSS JOIN websearch_to_tsquery(:q:) _query`)
	documentColumns := formatColumns(config.DocumentColumns, config.TableName)
	parameters.WithJoins(`CROSS JOIN to_tsvector(` + strings.Join(documentColumns, " || ' ' || ") + `) _document`)

	for rankColumn, rawColumns := range config.RankColumns {
		columns := formatColumns(rawColumns, config.TableName)
		parameters.WithJoins(
			fmt.Sprintf(`CROSS JOIN  NULLIF(ts_rank_cd(to_tsvector(%s), _query), 0) _rank_%s`, strings.Join(columns, " || ' ' || "), rankColumn),
		)
	}

	parameters.WithCondition(`( _query @@ _document )`)

	orders := []string{}
	for _, rankColumn := range config.RankOrder {
		orders = append(orders, `_rank_`+rankColumn)
	}

	parameters.Order = strings.Join(orders, ",") + " DESC NULLS LAST"
}

func formatColumns(columns []string, tableName string) []string {
	formattedColumns := []string{}
	for _, column := range columns {
		if strings.Contains(column, ".") {
			formattedColumns = append(formattedColumns, column)
		} else {
			formattedColumns = append(formattedColumns, fmt.Sprintf("%s.%s", tableName, column))
		}
	}
	return formattedColumns
}
