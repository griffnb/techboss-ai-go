//go:generate core_gen model Sandbox
package sandbox

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/common"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
	_ "github.com/griffnb/techboss-ai-go/internal/models/sandbox/migrations"
)

// Constants for the model
const (
	TABLE        = "sandboxes"
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
	OrganizationID *fields.UUIDField                  `column:"organization_id" type:"uuid"     default:"null" index:"true" null:"true" public:"view"`
	AccountID      *fields.UUIDField                  `column:"account_id"      type:"uuid"     default:"null" index:"true" null:"true" public:"view"`
	MetaData       *fields.StructField[*MetaData]     `column:"meta_data"       type:"jsonb"    default:"{}"`
	Provider       *fields.IntConstantField[Provider] `column:"type"            type:"smallint" default:"0"`
	AgentID        *fields.UUIDField                  `column:"agent_id"        type:"uuid"     default:"null" index:"true" null:"true" public:"view"`
	ExternalID     *fields.StringField                `column:"external_id"     type:"text"     default:"null" index:"true" null:"true" public:"view"`
}

// Sandbox - Database model
type Sandbox struct {
	model.BaseModel
	DBColumns
}

type SandboxJoined struct {
	Sandbox
	JoinData
}

func (this *Sandbox) beforeSave(ctx context.Context) error {
	this.BaseBeforeSave(ctx)
	common.GenerateURN(this)
	common.SetDisabledDeleted(this)
	return this.ValidateSubStructs()
}

func (this *Sandbox) afterSave(ctx context.Context) {
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
