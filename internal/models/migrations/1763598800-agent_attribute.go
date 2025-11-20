package migrations

import (
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/models/agent_attribute"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

func init() {
	model.AddMigration(&model.Migration{
		ID:          1763598800,
		Table:       agent_attribute.TABLE,
		TableStruct: &AgentAttributeV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type AgentAttributeV1 struct {
	base.Structure
	AccountID *fields.IntField `column:"account_id" type:"integer" default:"0" index:"true"`
	AgentID   *fields.IntField `column:"agent_id"   type:"integer" default:"0" index:"true"`
}
