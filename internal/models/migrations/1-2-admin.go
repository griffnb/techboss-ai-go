package migrations

import (
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/model/fields"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/admin"
	"github.com/griffnb/techboss-ai-go/internal/models/base"
)

func init() {
	model.AddMigration(&model.Migration{
		ID:          1,
		Table:       admin.TABLE,
		TableStruct: &AdminV1{},
		TableMigration: &model.TableMigration{
			Type: model.CREATE_TABLE,
		},
	})

	model.AddMigration(&model.Migration{
		ID:    2,
		Table: admin.TABLE,
		DataTransform: func() error {
			return environment.DB().DB.Insert(`
			CREATE UNIQUE INDEX admins_email_unique_idx
			ON admins (email)
			WHERE deleted = 0;
			`, map[string]interface{}{})
		},
	})

	model.AddMigration(&model.Migration{
		ID:    10,
		Table: admin.TABLE,
		PostMigrationTransform: func() error {
			emails := []string{"griffnb@gmail.com", "pearson@techboss.ai"}

			for _, email := range emails {

				adminObj := admin.New()
				adminObj.Email.Set(email)
				adminObj.Role.Set(constants.ROLE_ADMIN)
				adminObj.Status.Set(constants.STATUS_ACTIVE)
				err := adminObj.Save(nil)
				if err != nil {
					return err
				}
			}

			return nil
		},
	})
}

type AdminV1 struct {
	base.Structure
	FirstName  *fields.StringField                      `column:"first_name"  type:"text"     default:""`
	LastName   *fields.StringField                      `column:"last_name"   type:"text"     default:""`
	Email      *fields.StringField                      `column:"email"       type:"text"     default:""   unique:"true"`
	ExternalID *fields.StringField                      `column:"external_id" type:"text"     default:""                 index:"true"`
	Role       *fields.IntConstantField[constants.Role] `column:"role"        type:"smallint" default:"0"`
	SlackID    *fields.StringField                      `column:"slack_id"    type:"text"     default:""`
	Bookmarks  *fields.StructField[map[string]any]      `column:"bookmarks"   type:"jsonb"    default:"{}"`
}
