package migrations

import (
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

func init() {
	model.AddMigration(&model.Migration{
		ID:          1758894768,
		Table:       account.TABLE,
		TableStruct: &AccountV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type AccountV1 struct {
	base.Structure
	FirstName      *fields.StringField                      `column:"first_name"      type:"text"     default:""`
	LastName       *fields.StringField                      `column:"last_name"       type:"text"     default:""`
	Email          *fields.StringField                      `column:"email"           type:"text"     default:""     unique:"true"`
	Phone          *fields.StringField                      `column:"phone"           type:"text"     default:""`
	ExternalID     *fields.StringField                      `column:"external_id"     type:"text"     default:""                   index:"true"`
	PlanID         *fields.StringField                      `column:"plan_id"         type:"text"     default:""                   index:"true"`
	TestUserType   *fields.IntField                         `column:"test_user_type"  type:"smallint" default:"0"`
	OrganizationID *fields.UUIDField                        `column:"organization_id" type:"uuid"     default:"null"               index:"true" null:"true"`
	Role           *fields.IntConstantField[constants.Role] `column:"role"            type:"smallint" default:"1"                  index:"true"             public:"view"`
}
