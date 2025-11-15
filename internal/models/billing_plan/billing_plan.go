//go:generate core_generate model BillingPlan
package billing_plan

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
	TABLE        = "billing_plans"
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
	Name         *fields.StringField              `column:"name"          type:"text"     default:""`
	InternalName *fields.StringField              `column:"internal_name" type:"text"     default:""`
	Price        *fields.DecimalField             `column:"price"         type:"numeric"  default:"0"  public:"view" scale:"4" precision:"10"`
	FeatureSet   *fields.StructField[*FeatureSet] `column:"feature_set"   type:"jsonb"    default:"{}" public:"view"`
	Properties   *fields.StructField[*Properties] `column:"properties"    type:"jsonb"    default:"{}" public:"view"`
	Level        *fields.IntField                 `column:"level"         type:"smallint" default:"0"  public:"view"`
	IsDefault    *fields.IntField                 `column:"is_default"    type:"smallint" default:"0"                                         index:"true"`
}

type JoinData struct {
	CreatedByName *fields.StringField `json:"created_by_name" type:"text"`
	UpdatedByName *fields.StringField `json:"updated_by_name" type:"text"`
}

// BillingPlan - Database model
type BillingPlan struct {
	model.BaseModel
	DBColumns
}

type BillingPlanJoined struct {
	BillingPlan
	JoinData
}

func (this *BillingPlan) beforeSave(ctx context.Context) error {
	this.BaseBeforeSave(ctx)
	common.GenerateURN(this)
	common.SetDisabledDeleted(this)
	return this.ValidateSubStructs()
}

func (this *BillingPlan) afterSave(ctx context.Context) {
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
