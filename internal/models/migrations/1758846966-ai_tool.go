package migrations

import (
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/models/ai_tool"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

func init() {
	model.AddMigration(&model.Migration{
		ID:          1758846966,
		Table:       ai_tool.TABLE,
		TableStruct: &AiToolV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type AiToolV1 struct {
	base.Structure
	ToolName                *fields.StringField                                   `column:"tool_name"                 type:"text"     default:""`
	WebsiteURL              *fields.StringField                                   `column:"website_url"               type:"text"     default:""`
	AffiliateURL            *fields.StringField                                   `column:"affiliate_url"             type:"text"     default:""`
	HeroSection             *fields.StructField[*ai_tool.HeroSection]             `column:"hero_section"              type:"jsonb"    default:"{}"`
	Overview                *fields.StructField[*ai_tool.Overview]                `column:"overview"                  type:"jsonb"    default:"{}"`
	FeaturesAndCapabilities *fields.StructField[*ai_tool.FeaturesAndCapabilities] `column:"features_and_capabilities" type:"jsonb"    default:"{}"`
	UseCases                *fields.StructField[*ai_tool.UseCases]                `column:"use_cases"                 type:"jsonb"    default:"{}"`
	PricingAndPlans         *fields.StructField[*ai_tool.PricingAndPlans]         `column:"pricing_and_plans"         type:"jsonb"    default:"{}"`
	TargetAudience          *fields.StringField                                   `column:"target_audience"           type:"text"     default:""`
	Tags                    *fields.StructField[[]string]                         `column:"tags"                      type:"jsonb"    default:"[]"`
	Categorization          *fields.StringField                                   `column:"categorization"            type:"text"     default:""`
	BusinessFunction        *fields.StringField                                   `column:"business_function"         type:"text"     default:""`
	Affiliate               *fields.IntField                                      `column:"affiliate"                 type:"smallint" default:"0"`
	IsFeatured              *fields.IntField                                      `column:"is_featured"               type:"smallint" default:"0"`
	SearchBlobTSV           *fields.StringField                                   `column:"search_blob_tsv"           type:"tsvector" default:"null" null:"true" index:"true"`
}
