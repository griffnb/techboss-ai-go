//go:generate core_generate model Admin
package admin

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/common"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

const TABLE string = "admins"

const (
	IS_VERSIONED = false
	CHANGE_LOGS  = true
	CLIENT       = environment.CLIENT_DEFAULT
)

type Structure struct {
	DBColumns
	JoinData
}
type DBColumns struct {
	base.Structure
	FirstName  *fields.StringField                      `column:"first_name"  type:"text"     default:""`
	LastName   *fields.StringField                      `column:"last_name"   type:"text"     default:""`
	Email      *fields.StringField                      `column:"email"       type:"text"     default:""   unique:"true"`
	ExternalID *fields.StringField                      `column:"external_id" type:"text"     default:""                 index:"true"`
	Role       *fields.IntConstantField[constants.Role] `column:"role"        type:"smallint" default:"0"`
	SlackID    *fields.StringField                      `column:"slack_id"    type:"text"     default:""`
	Bookmarks  *fields.StructField[*Bookmarks]          `column:"bookmarks"   type:"jsonb"    default:"{}"`
}

type JoinData struct{}

type Admin struct {
	model.BaseModel
	DBColumns
}

type AdminJoined struct {
	Admin
	JoinData
}

func (this *Admin) beforeSave(ctx context.Context) error {
	this.BaseBeforeSave(ctx)
	common.GenerateURN(this)

	common.SetDisabledDeleted(this)
	return this.ValidateSubStructs()
}

func (this *Admin) afterSave(ctx context.Context) {
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
