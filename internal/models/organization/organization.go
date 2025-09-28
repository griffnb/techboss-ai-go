//go:generate core_generate model Organization
package organization

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
	TABLE        = "organizations"
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
	Name       *fields.StringField                 `column:"name"       type:"text"  default:""   public:"view"`
	Properties *fields.StructField[map[string]any] `column:"properties" type:"jsonb" default:"{}" public:"view"`
	MetaData   *fields.StructField[map[string]any] `column:"meta_data"  type:"jsonb" default:"{}" public:"view"`
	PlanID     *fields.StringField                 `column:"plan_id"    type:"text"  default:""   public:"view" index:"true"`
}

type JoinData struct {
	CreatedByName *fields.StringField `json:"created_by_name" type:"text"`
	UpdatedByName *fields.StringField `json:"updated_by_name" type:"text"`
}

// Organization - Database model
type Organization struct {
	model.BaseModel
	DBColumns
}

type OrganizationJoined struct {
	Organization
	JoinData
}

func (this *Organization) beforeSave(ctx context.Context) error {
	this.BaseBeforeSave(ctx)
	common.GenerateURN(this)
	common.SetDisabledDeleted(this)
	return this.ValidateSubStructs()
}

func (this *Organization) afterSave(ctx context.Context) {
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
