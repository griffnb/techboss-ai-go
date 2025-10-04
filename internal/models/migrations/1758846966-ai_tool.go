package migrations

import (
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/models/ai_tool"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

func init() {
	model.AddMigration(&model.Migration{
		ID:          1758846966,
		Table:       ai_tool.TABLE,
		TableStruct: &AiToolV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type AiToolV1 struct {
	base.Structure
	Name          *fields.StringField      `column:"name"            type:"text"     default:""`
	WebsiteURL    *fields.StringField      `column:"website_url"     type:"text"     default:""`
	AffiliateURL  *fields.StringField      `column:"affiliate_url"   type:"text"     default:""`
	MetaData      *fields.StructField[any] `column:"meta_data"       type:"jsonb"    default:"{}"`
	IsFeatured    *fields.IntField         `column:"is_featured"     type:"smallint" default:"0"`
	SearchBlobTSV *fields.StringField      `column:"search_blob_tsv" type:"tsvector" default:"null" null:"true" index:"true"`
}
