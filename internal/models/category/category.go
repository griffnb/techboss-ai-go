//go:generate core_generate model Category
package category

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
	TABLE        = "categories"
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
	Name             *fields.StringField `column:"name" type:"text" default:"" public:"view|edit"`
	Slug             *fields.StringField `column:"slug" type:"text" default:"" unique:"true" public:"view|edit"`
	Description      *fields.StringField `column:"description" type:"text" default:"" public:"view|edit"`
	ParentCategoryID *fields.UUIDField   `column:"parent_category_id" type:"uuid" default:"null" null:"true" public:"view|edit"`
}

type JoinData struct {
	CreatedByName *fields.StringField `json:"created_by_name" type:"text"`
	UpdatedByName *fields.StringField `json:"updated_by_name" type:"text"`
}

// Category - Database model
type Category struct {
	model.BaseModel
	DBColumns
}

type CategoryJoined struct {
	Category
	JoinData
}

func (this *Category) beforeSave(ctx context.Context) error {
	this.BaseBeforeSave(ctx)
	common.GenerateURN(this)
	common.SetDisabledDeleted(this)
	return this.ValidateSubStructs()

}

func (this *Category) afterSave(ctx context.Context) {
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
