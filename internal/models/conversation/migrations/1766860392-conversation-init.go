package migrations

import (
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

const TABLE = "conversations"

func init() {
	model.AddMigration(&model.Migration{
		ID:          1766860392,
		Table:       TABLE,
		TableStruct: &ConversationV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type ConversationV1 struct {
	base.Structure
	AccountID      *fields.UUIDField        `column:"account_id"      type:"uuid"  default:"null" null:"true" index:"true"`
	OrganizationID *fields.UUIDField        `column:"organization_id" type:"uuid"  default:"null" null:"true" index:"true"`
	AgentID        *fields.UUIDField        `column:"agent_id"        type:"uuid"  default:"null" null:"true" index:"true"`
	SandboxID      *fields.UUIDField        `column:"sandbox_id"      type:"uuid"  default:"null" null:"true" index:"true"`
	Stats          *fields.StructField[any] `column:"stats"           type:"jsonb" default:"{}"`
}
