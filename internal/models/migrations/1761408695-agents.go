package migrations

import (
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/models/agent"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

func init() {
	model.AddMigration(&model.Migration{
		ID:          1761408695,
		Table:       agent.TABLE,
		TableStruct: &AgentV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type AgentV1 struct {
	base.Structure
	Name     *fields.StringField           `column:"name"     type:"text"     default:""`
	Type     *fields.IntConstantField[int] `column:"type"     type:"smallint" default:"0"  index:"true"`
	Settings *fields.StructField[any]      `column:"settings" type:"jsonb"    default:"{}"`
}
