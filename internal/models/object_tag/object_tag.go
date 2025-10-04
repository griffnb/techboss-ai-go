//go:generate core_generate model ObjectTag
package object_tag

import (
	"context"

	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/model/coremodel"
	"github.com/CrowdShield/go-core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/common"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/environment"
)

const TABLE string = "object_tags"

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
	// conflict with underlying ID() method
	ID_          *fields.UUIDField                          `public:"view" column:"id"             type:"uuid"     default:"gen_random_uuid()" index:"true"`
	TagID        *fields.UUIDField                          `              column:"tag_id"         type:"uuid"                                              pk:"true"`
	ObjectURN    *fields.StringField                        `public:"view" column:"object_urn"     type:"text"     default:""                               pk:"true"`
	CreatedByURN *fields.StringField                        `public:"view" column:"created_by_urn" type:"text"     default:"null"              index:"true"           null:"true"`
	UpdatedByURN *fields.StringField                        `public:"view" column:"updated_by_urn" type:"text"     default:"null"              index:"true"           null:"true"`
	Status       *fields.IntConstantField[constants.Status] `public:"view" column:"status"         type:"smallint" default:"0"                 index:"true"`
	UpdatedAt    *fields.TimeField                          `public:"view" column:"updated_at"     type:"tswtz"    default:"CURRENT_TIMESTAMP" index:"true"`
	CreatedAt    *fields.TimeField                          `public:"view" column:"created_at"     type:"tswtz"    default:"CURRENT_TIMESTAMP" index:"true"`
}

type JoinData struct{}

type ObjectTag struct {
	model.BaseModel
	DBColumns
}

type ObjectTagJoined struct {
	ObjectTag
	JoinData
}

func (this *ObjectTag) beforeSave(ctx context.Context) error {
	this.BaseBeforeSave(ctx)
	common.GenerateURN(this)

	common.SetDisabledDeleted(this)
	return this.ValidateSubStructs()
}

func (this *ObjectTag) afterSave(ctx context.Context) {
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

func (this *ObjectTag) SaveIfNotExists(savingUser coremodel.Model) error {
	this.SetAnyConflict()
	return this.SaveWithContext(context.Background(), savingUser)
}
