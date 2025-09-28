//go:generate core_generate model Agent

package agent

import (
	"context"

	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/common"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

const (
	TABLE        string = "agents"
	CLIENT              = environment.CLIENT_DEFAULT
	CHANGE_LOGS         = true
	IS_VERSIONED        = false
)

type Structure struct {
	DBColumns
	JoinData
}

type DBColumns struct {
	base.Structure
	Name     *fields.StringField            `column:"name"     type:"text"  default:""`
	Settings *fields.StructField[*Settings] `column:"settings" type:"jsonb" default:"{}"`
}

type JoinData struct{}

type Agent struct {
	model.BaseModel
	DBColumns
}

type AgentJoined struct {
	Agent
	JoinData
}

func (this *Agent) beforeSave(ctx context.Context) error {
	this.BaseBeforeSave(ctx)
	common.GenerateURN(this)
	common.SetDisabledDeleted(this)
	return this.ValidateSubStructs()
}

func (this *Agent) afterSave(ctx context.Context) {
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
