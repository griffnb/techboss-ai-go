//go:generate core_generate model AiTool
package ai_tool

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
	TABLE        = "ai_tools"
	CHANGE_LOGS  = true
	CLIENT       = environment.CLIENT_DEFAULT
	IS_VERSIONED = false
)

// Struct types for JSONB fields
type HeroSection struct {
	Logo        string `json:"logo"`
	Tagline     string `json:"tagline"`
	ToolName    string `json:"toolName"`
	Description string `json:"description"`
}

type Overview struct {
	KeyBenefits  []string `json:"keyBenefits"`
	Introduction string   `json:"introduction"`
}

type CoreFeature struct {
	Feature     string `json:"feature"`
	Description string `json:"description"`
}

type FeaturesAndCapabilities struct {
	HowItWorks   string        `json:"howItWorks"`
	CoreFeatures []CoreFeature `json:"coreFeatures"`
}

type UseCases struct {
	Applications []string `json:"applications"`
}

type PricingAndPlans struct {
	FreeTier       bool   `json:"freeTier"`
	PricingRange   string `json:"pricingRange"`
	PricingOptions string `json:"pricingOptions"`
}

type Structure struct {
	DBColumns
	JoinData
}

type DBColumns struct {
	base.Structure
	ToolName                *fields.StringField                           `column:"tool_name"                 type:"text"     default:""`
	WebsiteURL              *fields.StringField                           `column:"website_url"               type:"text"     default:""`
	HeroSection             *fields.StructField[*HeroSection]             `column:"hero_section"              type:"jsonb"    default:"{}"`
	Overview                *fields.StructField[*Overview]                `column:"overview"                  type:"jsonb"    default:"{}"`
	FeaturesAndCapabilities *fields.StructField[*FeaturesAndCapabilities] `column:"features_and_capabilities" type:"jsonb"    default:"{}"`
	UseCases                *fields.StructField[*UseCases]                `column:"use_cases"                 type:"jsonb"    default:"{}"`
	PricingAndPlans         *fields.StructField[*PricingAndPlans]         `column:"pricing_and_plans"         type:"jsonb"    default:"{}"`
	TargetAudience          *fields.StringField                           `column:"target_audience"           type:"text"     default:""`
	Tags                    *fields.StructField[[]string]                 `column:"tags"                      type:"jsonb"    default:"[]"`
	Categorization          *fields.StringField                           `column:"categorization"            type:"text"     default:""`
	BusinessFunction        *fields.StringField                           `column:"business_function"         type:"text"     default:""`
	Affiliate               *fields.IntField                              `column:"affiliate"                 type:"smallint" default:"0"`
	IsFeatured              *fields.IntField                              `column:"is_featured"               type:"smallint" default:"0"`
}

type JoinData struct {
	CreatedByName *fields.StringField `json:"created_by_name" type:"text"`
	UpdatedByName *fields.StringField `json:"updated_by_name" type:"text"`
}

// AiTool - Database model
type AiTool struct {
	model.BaseModel
	DBColumns
}

type AiToolJoined struct {
	AiTool
	JoinData
}

func (this *AiTool) beforeSave(ctx context.Context) error {
	this.BaseBeforeSave(ctx)
	common.GenerateURN(this)
	common.SetDisabledDeleted(this)
	return this.ValidateSubStructs()
}

func (this *AiTool) afterSave(ctx context.Context) {
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
