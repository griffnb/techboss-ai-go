package ai_tool

import (
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/griffnb/techboss-ai-go/internal/models/admin"
)

// AddJoinData adds in the join data
func AddJoinData(options *model.Options) {
	options.WithPrependJoins([]string{
		admin.JoinCreatedUpdatedQuery(TABLE),
	}...)
	options.WithIncludeFields(append([]string{}, admin.JoinCreatedUpdatedField()...)...)
}
