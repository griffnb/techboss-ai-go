package conversation

import (
	"github.com/griffnb/core/lib/model"
)

// AddJoinData adds in the join data
func AddJoinData(options *model.Options) {
	options.WithPrependJoins([]string{
		"LEFT JOIN agents ON agents.id = conversations.agent_id",
	}...)
	options.WithIncludeFields([]string{
		"agents.name AS agent_name",
	}...)
}
