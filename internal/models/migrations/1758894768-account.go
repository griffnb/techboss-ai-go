package migrations

import (
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/model/fields"
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
	Name         *fields.StringField `column:"name"           type:"text"     default:""`
	Email        *fields.StringField `column:"email"          type:"text"     default:""  unique:"true"`
	Phone        *fields.StringField `column:"phone"          type:"text"     default:""`
	ExternalID   *fields.StringField `column:"external_id"    type:"text"     default:""                index:"true"`
	PlanID       *fields.StringField `column:"plan_id"        type:"text"     default:""                index:"true"`
	TestUserType *fields.IntField    `column:"test_user_type" type:"smallint" default:"0"                            nullable:"false"`
}
