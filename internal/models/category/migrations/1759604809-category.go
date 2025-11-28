package migrations

import (
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

const TABLE string = "categories"

func init() {
	model.AddMigration(&model.Migration{
		ID:          1759604809,
		Table:       TABLE,
		TableStruct: &CategoryV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type CategoryV1 struct {
	base.Structure
	Name             *fields.StringField `column:"name"               type:"text" default:""`
	Slug             *fields.StringField `column:"slug"               type:"text" default:""     unique:"true"`
	Description      *fields.StringField `column:"description"        type:"text" default:""`
	ParentCategoryID *fields.UUIDField   `column:"parent_category_id" type:"uuid" default:"null"               null:"true"`
}
