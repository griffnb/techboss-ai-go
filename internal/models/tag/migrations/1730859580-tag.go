package migrations

import (
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

const TABLE string = "tags"

func init() {
	model.AddMigration(&model.Migration{
		ID:          1730859580,
		Table:       TABLE,
		TableStruct: &TagV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})

	model.AddMigration(&model.Migration{
		ID:    1730859581,
		Table: TABLE,
		DataTransform: func() error {
			return environment.DB().DB.Insert(`
			ALTER TABLE tags ADD CONSTRAINT tag_name_type_unique UNIQUE (name, type)
			`, map[string]interface{}{})
		},
	})
}

type TagV1 struct {
	base.Structure
	Name     *fields.StringField                         `column:"name"     type:"text"     default:""`
	Key      *fields.StringConstantField[string]         `column:"key"      type:"text"     default:"null" unique:"true" null:"true"`
	Internal *fields.IntField                            `column:"internal" type:"smallint" default:"0"                              index:"true"`
	Type     *fields.IntConstantField[constants.TagType] `column:"type"     type:"smallint" default:"0"                              index:"true"`
}
