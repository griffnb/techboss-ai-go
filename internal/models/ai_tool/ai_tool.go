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

type Structure struct {
	DBColumns
	JoinData
}

type DBColumns struct {
	base.Structure
	Name                       *fields.StringField            `column:"name"                          type:"text"     default:""`
	Description                *fields.StringField            `column:"description"                   type:"text"     default:""`
	WebsiteURL                 *fields.StringField            `column:"website_url"                   type:"text"     default:""`
	AffiliateURL               *fields.StringField            `column:"affiliate_url"                 type:"text"     default:""`
	MetaData                   *fields.StructField[*MetaData] `column:"meta_data"                     type:"jsonb"    default:"{}"`
	IsFeatured                 *fields.IntField               `column:"is_featured"                   type:"smallint" default:"0"`
	SearchBlobTSV              *fields.StringField            `column:"search_blob_tsv"               type:"tsvector" default:"null" null:"true" index:"true"`
	CategoryID                 *fields.UUIDField              `column:"category_id"                   type:"uuid"     default:"null" null:"true" index:"true"`
	BusinessFunctionCategoryID *fields.UUIDField              `column:"business_function_category_id" type:"uuid"     default:"null" null:"true" index:"true"`
}

type JoinData struct {
	CategoryName                 *fields.StringField `public:"view" json:"category_name"                   type:"text"`
	BusinessFunctionCategoryName *fields.StringField `public:"view" json:"business_function_category_name" type:"text"`
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
	this.UpdateSearchVector()
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
