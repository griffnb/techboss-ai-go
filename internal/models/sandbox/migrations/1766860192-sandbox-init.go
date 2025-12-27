package migrations

import (
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

const TABLE = "sandboxes"

func init() {
	model.AddMigration(&model.Migration{
		ID:          1766860192,
		Table:       TABLE,
		TableStruct: &SandboxV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type SandboxV1 struct {
	base.Structure
	OrganizationID *fields.UUIDField             `column:"organization_id" type:"uuid"     default:"null" index:"true" null:"true" public:"view"`
	AccountID      *fields.UUIDField             `column:"account_id"      type:"uuid"     default:"null" index:"true" null:"true" public:"view"`
	MetaData       *fields.StructField[any]      `column:"meta_data"       type:"jsonb"    default:"{}"`
	Provider       *fields.IntConstantField[int] `column:"type"            type:"smallint" default:"0"`
	AgentID        *fields.UUIDField             `column:"agent_id"        type:"uuid"     default:"null" index:"true" null:"true" public:"view"`
}
