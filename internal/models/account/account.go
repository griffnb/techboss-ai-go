//go:generate core_generate model Account
package account

import (
	"context"

	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/common"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

// Constants for the model
const (
	TABLE        = "accounts"
	CHANGE_LOGS  = true
	CLIENT       = environment.CLIENT_DEFAULT
	IS_VERSIONED = false
)

type Structure struct {
	DBColumns
	JoinData
}

type DBColumns struct {
	base.Structure
	FirstName           *fields.StringField                      `public:"edit" column:"first_name"             type:"text"     default:""`
	LastName            *fields.StringField                      `public:"edit" column:"last_name"              type:"text"     default:""`
	Email               *fields.StringField                      `public:"edit" column:"email"                  type:"text"     default:""     unique:"true"`
	Phone               *fields.StringField                      `public:"edit" column:"phone"                  type:"text"     default:""`
	ExternalID          *fields.StringField                      `public:"view" column:"external_id"            type:"text"     default:""                   index:"true"`
	TestUserType        *fields.IntField                         `public:"view" column:"test_user_type"         type:"smallint" default:"0"`
	OrganizationID      *fields.UUIDField                        `public:"view" column:"organization_id"        type:"uuid"     default:"null"               index:"true" null:"true"`
	Role                *fields.IntConstantField[constants.Role] `public:"view" column:"role"                   type:"smallint" default:"1"                  index:"true"`
	Properties          *fields.StructField[*Properties]         `column:"properties"             type:"jsonb"    default:"{}"`
	SignupProperties    *fields.StructField[*SignupProperties]   `column:"signup_properties"      type:"jsonb"    default:"{}"`
	HashedPassword      *fields.StringField                      `              column:"hashed_password"        type:"text"     default:""`
	PasswordUpdatedAtTS *fields.IntField                         `              column:"password_updated_at_ts" type:"bigint"   default:"0"`
	EmailVerifiedAtTS   *fields.IntField                         `              column:"email_verified_at_ts"   type:"bigint"   default:"0"`
	LastLoginTS         *fields.IntField                         `              column:"last_login_ts"          type:"bigint"   default:"0"                  index:"true"`
}

type JoinData struct {
	Name          *fields.StringField `json:"name"            type:"text"`
	CreatedByName *fields.StringField `json:"created_by_name" type:"text"`
	UpdatedByName *fields.StringField `json:"updated_by_name" type:"text"`
}

// Account - Database model
type Account struct {
	model.BaseModel
	DBColumns
}

type AccountJoined struct {
	Account
	JoinData
}

func (this *Account) beforeSave(ctx context.Context) error {
	this.BaseBeforeSave(ctx)
	common.GenerateURN(this)
	common.SetDisabledDeleted(this)
	return this.ValidateSubStructs()
}

func (this *Account) afterSave(ctx context.Context) {
	this.BaseAfterSave(ctx)
	/*
		go func() {
			err := this.UpdateCache()
			if err != nil {
				log.Error(err)
			}
		}()
	*/
}
