//go:generate core_gen model Lead
package lead

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/common"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

// Constants for the model
const (
	TABLE        = "leads"
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
	Name       *fields.StringField        `column:"name"        type:"text"  default:""     public:"edit"`
	Email      *fields.StringField        `column:"email"       type:"text"  default:"null" public:"edit" null:"true" unique:"true" index:"true"`
	Phone      *fields.StringField        `column:"phone"       type:"text"  default:""     public:"edit"`
	ExternalID *fields.StringField        `column:"external_id" type:"text"  default:""                                             index:"true"`
	Utms       *fields.StructField[*Utms] `column:"utms"        type:"jsonb" default:"{}"`
}

type JoinData struct {
	CreatedByName *fields.StringField `json:"created_by_name" type:"text"`
	UpdatedByName *fields.StringField `json:"updated_by_name" type:"text"`
}

// Lead - Database model
type Lead struct {
	model.BaseModel
	DBColumns
}

type LeadJoined struct {
	Lead
	JoinData
}

func (this *Lead) beforeSave(ctx context.Context) error {
	this.BaseBeforeSave(ctx)
	common.GenerateURN(this)
	common.SetDisabledDeleted(this)
	return this.ValidateSubStructs()
}

func (this *Lead) afterSave(ctx context.Context) {
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
