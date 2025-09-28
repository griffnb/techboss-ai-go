//go:generate core_generate model AgentAttribute

package agent_attribute

import (
	"context"

	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/common"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

const (
	TABLE        string = "agent_attributes"
	CHANGE_LOGS         = true
	IS_VERSIONED        = false
	CLIENT              = environment.CLIENT_DEFAULT
)

type Structure struct {
	DBColumns
	JoinData
}

type DBColumns struct {
	base.Structure
	AccountID *fields.IntField `column:"account_id" type:"integer" default:"0" index:"true"`
	AgentID   *fields.IntField `column:"agent_id"   type:"integer" default:"0" index:"true"`
}

type JoinData struct{}

type AgentAttribute struct {
	model.BaseModel
	DBColumns
}

type AgentAttributeJoined struct {
	AgentAttribute
	JoinData
}

func (this *AgentAttribute) beforeSave(ctx context.Context) error {
	this.BaseBeforeSave(ctx)
	common.GenerateURN(this)
	common.SetDisabledDeleted(this)
	return this.ValidateSubStructs()
}

func (this *AgentAttribute) afterSave(ctx context.Context) {
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
