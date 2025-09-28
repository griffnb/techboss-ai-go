package migrations

import (
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
	"github.com/griffnb/techboss-ai-go/internal/models/lead"
)

func init() {
	model.AddMigration(&model.Migration{
		ID:          1759010685,
		Table:       lead.TABLE,
		TableStruct: &LeadV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type LeadV1 struct {
	base.Structure
	Name       *fields.StringField                 `column:"name"        type:"text"  default:""     public:"edit"`
	Email      *fields.StringField                 `column:"email"       type:"text"  default:"null" public:"edit" null:"true" unique:"true" index:"true"`
	Phone      *fields.StringField                 `column:"phone"       type:"text"  default:""     public:"edit"`
	ExternalID *fields.StringField                 `column:"external_id" type:"text"  default:""                                             index:"true"`
	Utms       *fields.StructField[map[string]any] `column:"utms"        type:"jsonb" default:"{}"`
}
