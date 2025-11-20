//go:generate core_gen model Tag
package tag

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/common"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

const TABLE string = "tags"

const (
	CHANGE_LOGS       = true
	CLIENT            = environment.CLIENT_DEFAULT
	IS_VERSIONED bool = false
)

type Structure struct {
	DBColumns
	JoinData
}

type DBColumns struct {
	base.Structure
	Name     *fields.StringField                         `column:"name"     type:"text"     default:""     unique:"true"`
	Key      *fields.StringConstantField[string]         `column:"key"      type:"text"     default:"null" unique:"true" null:"true"`
	Internal *fields.IntField                            `column:"internal" type:"smallint" default:"0"                              index:"true"`
	Type     *fields.IntConstantField[constants.TagType] `column:"type"     type:"smallint" default:"0"    unique:"true"             index:"true"`
}

type JoinData struct{}

type Tag struct {
	model.BaseModel
	DBColumns
}

type TagJoined struct {
	Tag
	JoinData
}

func (this *Tag) beforeSave(ctx context.Context) error {
	this.BaseBeforeSave(ctx)
	common.GenerateURN(this)

	common.SetDisabledDeleted(this)
	return this.ValidateSubStructs()
}

func (this *Tag) afterSave(ctx context.Context) {
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
