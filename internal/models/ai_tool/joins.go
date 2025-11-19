package ai_tool

import (
	"github.com/griffnb/core/lib/model"
)

// AddJoinData adds in the join data
func AddJoinData(options *model.Options) {
	options.WithPrependJoins([]string{
		"LEFT JOIN categories ON ai_tools.category_id = categories.id",
		"LEFT JOIN categories business_function_categories ON ai_tools.business_function_category_id = business_function_categories.id",
	}...)
	options.WithIncludeFields([]string{
		"categories.name AS category_name",
		"business_function_categories.name AS business_function_category_name",
	}...)
}
