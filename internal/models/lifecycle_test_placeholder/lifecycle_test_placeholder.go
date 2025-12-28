//go:generate core_gen model LifecycleTestPlaceholder
package lifecycle_test_placeholder

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/common"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
	_ "github.com/griffnb/techboss-ai-go/internal/models/lifecycle_test_placeholder/migrations"
)

// Constants for the model
const (
	TABLE        = "lifecycle_test_placeholders"
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
	Name *fields.StringField `column:"name" type:"text" default:""`
}

// LifecycleTestPlaceholder - Database model
type LifecycleTestPlaceholder struct {
	model.BaseModel
	DBColumns
}

type LifecycleTestPlaceholderJoined struct {
	LifecycleTestPlaceholder
	JoinData
}

func (this *LifecycleTestPlaceholder) beforeSave(ctx context.Context) error {
	this.BaseBeforeSave(ctx)
	common.GenerateURN(this)
	common.SetDisabledDeleted(this)
	return this.ValidateSubStructs()
}

func (this *LifecycleTestPlaceholder) afterSave(ctx context.Context) {
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
