package migrations

import (
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
	"github.com/griffnb/techboss-ai-go/internal/models/document"
)

func init() {
	model.AddMigration(&model.Migration{
		ID:          1763598802,
		Table:       document.TABLE,
		TableStruct: &DocumentV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type DocumentV1 struct {
	base.Structure
	AccountID      *fields.IntField                       `column:"account_id"       type:"integer"  default:"0"  index:"true"`
	DocumentType   *fields.IntConstantField[document.DocumentType] `column:"document_type"    type:"smallint" default:"0"  index:"true"`
	Name           *fields.StringField                    `column:"name"             type:"text"     default:""`
	ProcessedS3URL *fields.StringField                    `column:"processed_s3_url" type:"text"     default:""`
	RawS3URL       *fields.StringField                    `column:"raw_s3_url"       type:"text"     default:""`
	MetaData       *fields.StructField[any]               `column:"meta_data"        type:"jsonb"    default:"{}"`
}
