package migrations

import (
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

const TABLE string = "accounts"

func init() {
	model.AddMigration(&model.Migration{
		ID:          1758894768,
		Table:       TABLE,
		TableStruct: &AccountV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type AccountV1 struct {
	base.Structure
	FirstName           *fields.StringField                      `public:"edit" column:"first_name"             type:"text"     default:""`
	LastName            *fields.StringField                      `public:"edit" column:"last_name"              type:"text"     default:""`
	Email               *fields.StringField                      `public:"edit" column:"email"                  type:"text"     default:""     unique:"true"`
	Phone               *fields.StringField                      `public:"edit" column:"phone"                  type:"text"     default:""`
	ExternalID          *fields.StringField                      `public:"view" column:"external_id"            type:"text"     default:""                   index:"true"`
	TestUserType        *fields.IntField                         `public:"view" column:"test_user_type"         type:"smallint" default:"0"`
	OrganizationID      *fields.UUIDField                        `public:"view" column:"organization_id"        type:"uuid"     default:"null"               index:"true" null:"true"`
	Role                *fields.IntConstantField[constants.Role] `public:"view" column:"role"                   type:"smallint" default:"1"                  index:"true"`
	Properties          *fields.StructField[any]                 `              column:"properties"             type:"jsonb"    default:"{}"`
	SignupProperties    *fields.StructField[any]                 `              column:"signup_properties"      type:"jsonb"    default:"{}"`
	HashedPassword      *fields.StringField                      `              column:"hashed_password"        type:"text"     default:""`
	PasswordUpdatedAtTS *fields.IntField                         `              column:"password_updated_at_ts" type:"bigint"   default:"0"`
	EmailVerifiedAtTS   *fields.IntField                         `              column:"email_verified_at_ts"   type:"bigint"   default:"0"`
	LastLoginTS         *fields.IntField                         `              column:"last_login_ts"          type:"bigint"   default:"0"                  index:"true"`
}
