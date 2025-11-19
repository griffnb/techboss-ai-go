package agent_attribute

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/types"
)

func GetByAgentID(ctx context.Context, agentID types.UUID) ([]*AgentAttribute, error) {
	options := model.NewOptions().
		WithCondition("%s = :agent_id: AND %s = 0", Columns.AgentID.Column(), Columns.Disabled.Column()).
		WithParam(":agent_id:", agentID)

	return FindAll(ctx, options)
}
