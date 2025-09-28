//go:generate core_generate model Account
package account

import (
	"context"

	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/common"
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
	Name           *fields.StringField `column:"name"            type:"text"     default:""`
	Email          *fields.StringField `column:"email"           type:"text"     default:""     unique:"true"`
	Phone          *fields.StringField `column:"phone"           type:"text"     default:""`
	ExternalID     *fields.StringField `column:"external_id"     type:"text"     default:""                   index:"true"`
	PlanID         *fields.StringField `column:"plan_id"         type:"text"     default:""                   index:"true"`
	TestUserType   *fields.IntField    `column:"test_user_type"  type:"smallint" default:"0"`
	OrganizationID *fields.UUIDField   `column:"organization_id" type:"uuid"     default:"null"               index:"true" null:"true"`
}

type JoinData struct {
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
