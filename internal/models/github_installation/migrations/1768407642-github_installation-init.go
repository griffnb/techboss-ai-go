package migrations

import (
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

const TABLE = "github_installations"

type GithubAccountType int

const (
	ACCOUNT_TYPE_USER GithubAccountType = iota
	ACCOUNT_TYPE_ORGANIZATION
)

func init() {
	model.AddMigration(&model.Migration{
		ID:          1768407642,
		Table:       TABLE,
		TableStruct: &GithubInstallationV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type GithubInstallationV1 struct {
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
