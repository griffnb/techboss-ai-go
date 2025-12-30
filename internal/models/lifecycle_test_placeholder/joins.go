package lifecycle_test_placeholder

import (
	"github.com/griffnb/core/lib/model"
)

type JoinData struct{}

// AddJoinData adds in the join data
func AddJoinData(options *model.Options) {
	options.WithPrependJoins([]string{}...)
	options.WithIncludeFields([]string{}...)
}
