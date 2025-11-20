//go:generate core_generate model Document
package document

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/common"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

const (
	TABLE string = "documents"

	IS_VERSIONED = false
	CLIENT       = environment.CLIENT_DEFAULT
	CHANGE_LOGS  = true
)

type Structure struct {
	DBColumns
	JoinData
}

type DBColumns struct {
	base.Structure
	AccountID      *fields.IntField                       `column:"account_id"       type:"integer"  default:"0"  index:"true"`
	DocumentType   *fields.IntConstantField[DocumentType] `column:"document_type"    type:"smallint" default:"0"  index:"true"`
	Name           *fields.StringField                    `column:"name"             type:"text"     default:""`
	ProcessedS3URL *fields.StringField                    `column:"processed_s3_url" type:"text"     default:""`
	RawS3URL       *fields.StringField                    `column:"raw_s3_url"       type:"text"     default:""`
	MetaData       *fields.StructField[*MetaData]         `column:"meta_data"        type:"jsonb"    default:"{}"`
}

type JoinData struct {
	DocumentGroupName      *fields.StringField `public:"view" json:"document_group_name" type:"text"`
	DocumentGroupGroupType *fields.IntField    `public:"view" json:"document_group_group_type" type:"integer"`
}

type Document struct {
	model.BaseModel
	DBColumns
}

type DocumentJoined struct {
	Document
	JoinData
}

func (this *Document) beforeSave(ctx context.Context) error {
	this.BaseBeforeSave(ctx)
	common.GenerateURN(this)
	common.SetDisabledDeleted(this)
	return this.ValidateSubStructs()
}

func (this *Document) afterSave(ctx context.Context) {
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
