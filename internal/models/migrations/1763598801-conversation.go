package migrations

import (
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
	"github.com/griffnb/techboss-ai-go/internal/models/conversation"
)

func init() {
	model.AddMigration(&model.Migration{
		ID:          1763598801,
		Table:       conversation.TABLE,
		TableStruct: &ConversationV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type ConversationV1 struct {
	base.Structure
	AccountID *fields.UUIDField `column:"account_id" type:"uuid" default:"null" null:"true" index:"true"`
	AgentID   *fields.UUIDField `column:"agent_id"   type:"uuid" default:"null" null:"true" index:"true"`
}
