//go:generate core_gen model GithubInstallation
package github_installation

import (
	"context"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/common"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
	_ "github.com/griffnb/techboss-ai-go/internal/models/github_installation/migrations"
)

// Constants for the model
const (
	TABLE        = "github_installations"
	CHANGE_LOGS  = true
	CLIENT       = environment.CLIENT_DEFAULT
	IS_VERSIONED = false
)

// GithubAccountType represents the type of GitHub account
type GithubAccountType int

const (
	ACCOUNT_TYPE_USER GithubAccountType = iota
	ACCOUNT_TYPE_ORGANIZATION
)

type Structure struct {
	DBColumns
	JoinData
}

type DBColumns struct {
	base.Structure
	AccountID         *fields.UUIDField                           `column:"account_id"          type:"uuid"     default:"null" index:"true" null:"true"`
	InstallationID    *fields.StringField                         `column:"installation_id"     type:"text"                    index:"true"             unique:"true"`
	GithubAccountID   *fields.StringField                         `column:"github_account_id"   type:"text"                    index:"true"`
	GithubAccountType *fields.IntConstantField[GithubAccountType] `column:"github_account_type" type:"smallint" default:"0"`
	GithubAccountName *fields.StringField                         `column:"github_account_name" type:"text"`
	RepositoryAccess  *fields.StringField                         `column:"repository_access"   type:"text"     default:"all"`
	Permissions       *fields.StructField[map[string]any]         `column:"permissions"         type:"jsonb"    default:"{}"`
	Suspended         *fields.IntField                            `column:"suspended"           type:"smallint" default:"0"    index:"true"`
	AppSlug           *fields.StringField                         `column:"app_slug"            type:"text"`
}

// GithubInstallation - Database model
type GithubInstallation struct {
	model.BaseModel
	DBColumns
}

type GithubInstallationJoined struct {
	GithubInstallation
	JoinData
}

func (this *GithubInstallation) beforeSave(ctx context.Context) error {
	this.BaseBeforeSave(ctx)
	common.GenerateURN(this)
	common.SetDisabledDeleted(this)
	return this.ValidateSubStructs()
}

func (this *GithubInstallation) afterSave(ctx context.Context) {
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
