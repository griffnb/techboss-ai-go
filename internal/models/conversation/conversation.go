//go:generate core_gen model Conversation

package conversation

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/common"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
	_ "github.com/griffnb/techboss-ai-go/internal/models/conversation/migrations"
)

const (
	TABLE         string = "conversations"
	EXTERNAL_TYPE string = "conversation"

	IS_VERSIONED = false
	CLIENT       = environment.CLIENT_DEFAULT
	CHANGE_LOGS  = false
)

type Structure struct {
	DBColumns
	JoinData
}

type DBColumns struct {
	base.Structure
	AccountID      *fields.UUIDField                       `column:"account_id"      type:"uuid"  default:"null" null:"true" index:"true"`
	OrganizationID *fields.UUIDField                       `column:"organization_id" type:"uuid"  default:"null" null:"true" index:"true"`
	AgentID        *fields.UUIDField                       `column:"agent_id"        type:"uuid"  default:"null" null:"true" index:"true"`
	SandboxID      *fields.UUIDField                       `column:"sandbox_id"      type:"uuid"  default:"null" null:"true" index:"true"`
	Stats          *fields.StructField[*ConversationStats] `column:"stats"           type:"jsonb" default:"{}"`
}

type JoinData struct{}

type Conversation struct {
	model.BaseModel
	DBColumns
}

type ConversationJoined struct {
	Conversation
	JoinData
}

func (this *Conversation) beforeSave(ctx context.Context) error {
	this.BaseBeforeSave(ctx)
	common.GenerateURN(this)
	common.SetDisabledDeleted(this)
	return this.ValidateSubStructs()
}

func (this *Conversation) afterSave(ctx context.Context) {
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
