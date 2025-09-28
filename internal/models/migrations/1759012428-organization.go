package migrations

import (
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
	"github.com/griffnb/techboss-ai-go/internal/models/organization"
)

func init() {
	model.AddMigration(&model.Migration{
		ID:          1759012428,
		Table:       organization.TABLE,
		TableStruct: &OrganizationV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type OrganizationV1 struct {
	base.Structure
	Name       *fields.StringField                 `column:"name"       type:"text"  default:""`
	Properties *fields.StructField[map[string]any] `column:"properties" type:"jsonb" default:"{}"`
	MetaData   *fields.StructField[map[string]any] `column:"meta_data"  type:"jsonb" default:"{}"`
	PlanID     *fields.StringField                 `column:"plan_id"    type:"text"  default:""   index:"true"`
}
