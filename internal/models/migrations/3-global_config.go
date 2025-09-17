package migrations

import (
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

// -------------- Listings------------------
func init() {
	model.AddMigrationWithClientName(environment.CLIENT_DEFAULT, &model.Migration{
		ID:          3,
		Table:       "global_configs",
		TableStruct: &GlobalConfigV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})
}

type GlobalConfigV1 struct {
	base.Structure
	Key   string `column:"key"   type:"text" default:"" index:"true"`
	Value string `column:"value" type:"text" default:""`
}
