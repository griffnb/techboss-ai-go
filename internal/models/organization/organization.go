//go:generate core_gen model Organization
package organization

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/common"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
	"github.com/griffnb/techboss-ai-go/internal/models/billing_plan"
	_ "github.com/griffnb/techboss-ai-go/internal/models/organization/migrations"
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
	Name                *fields.StringField                           `public:"edit" column:"name"                  type:"text"  default:""`
	ExternalID          *fields.StringField                           `              column:"external_id"           type:"text"  default:"null" null:"true" index:"true" unique:"true"`
	StripeID            *fields.StringField                           `              column:"stripe_id"             type:"text"  default:"null" null:"true" index:"true"`
	BillingPlanID       *fields.UUIDField                             `public:"view" column:"billing_plan_id"       type:"uuid"  default:"null" null:"true" index:"true"`
	Properties          *fields.StructField[*Properties]              `public:"view" column:"properties"            type:"jsonb" default:"{}"`
	MetaData            *fields.StructField[*MetaData]                `public:"view" column:"meta_data"             type:"jsonb" default:"{}"`
	EmailDomains        *fields.StructField[[]string]                 `              column:"email_domains"         type:"jsonb" default:"[]"               index:"true"`
	Subdomain           *fields.StringField                           `              column:"subdomain"             type:"text"  default:""`
	FeatureSetOverrides *fields.StructField[*billing_plan.FeatureSet] `              column:"feature_set_overrides" type:"jsonb" default:"{}"`
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
