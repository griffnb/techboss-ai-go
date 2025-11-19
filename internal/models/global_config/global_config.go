//go:generate core_generate model GlobalConfig -options=base,queries,marshaler
package global_config

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/common"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

const TABLE string = "global_configs"

const (
	IS_VERSIONED = false
	CHANGE_LOGS  = true
	CLIENT       = environment.CLIENT_DEFAULT
)

type Structure struct {
	DBColumns
}
type DBColumns struct {
	base.Structure
	Key *fields.StringConstantField[constants.GlobalConfigKey] `column:"key"   type:"text" default:"" index:"true"`
	Val *fields.StringField                                    `column:"value" type:"text" default:""`
}

type GlobalConfig struct {
	model.BaseModel
	DBColumns
}

type GlobalConfigJoined struct {
	GlobalConfig
}

func (this *GlobalConfig) beforeSave(ctx context.Context) error {
	this.BaseBeforeSave(ctx)
	common.GenerateURN(this)
	common.SetDisabledDeleted(this)
	return nil
}

func (this *GlobalConfig) afterSave(ctx context.Context) {
	this.BaseAfterSave(ctx)
}
