package migrations

import (
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
	"github.com/griffnb/techboss-ai-go/internal/models/billing_plan"
	"github.com/griffnb/techboss-ai-go/internal/models/organization"
)

func init() {
	model.AddMigration(&model.Migration{
		ID:          1759012428,
		Table:       organization.TABLE,
		TableStruct: &OrganizationV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type OrganizationV1 struct {
	base.Structure
	Name                *fields.StringField                           `column:"name"                  type:"text"  default:""`
	ExternalID          *fields.StringField                           `column:"external_id"           type:"text"  default:"null" null:"true" index:"true" unique:"true"`
	BillingPlanID       *fields.UUIDField                             `column:"billing_plan_id"       type:"uuid"  default:"null" null:"true" index:"true"               public:"view"`
	StripeID            *fields.StringField                           `column:"stripe_id"             type:"text"  default:"null" null:"true" index:"true"`
	Properties          *fields.StructField[map[string]any]           `column:"properties"            type:"jsonb" default:"{}"                                          public:"view"`
	MetaData            *fields.StructField[map[string]any]           `column:"meta_data"             type:"jsonb" default:"{}"                                          public:"view"`
	EmailDomains        *fields.StructField[[]string]                 `column:"email_domains"         type:"jsonb" default:"[]"               index:"true"`
	Subdomain           *fields.StringField                           `column:"subdomain"             type:"text"  default:""`
	FeatureSetOverrides *fields.StructField[*billing_plan.FeatureSet] `column:"feature_set_overrides" type:"jsonb" default:"{}"`
}
