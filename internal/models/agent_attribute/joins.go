package agent_attribute

import (
	"github.com/CrowdShield/go-core/lib/model"
)

// AddJoinData adds in the join data
func AddJoinData(options *model.Options) {
	options.WithPrependJoins([]string{}...)
	options.WithIncludeFields([]string{}...)
}
