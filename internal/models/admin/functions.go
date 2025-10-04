package admin

import (
	"context"
	"fmt"

	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/common"
	"github.com/griffnb/techboss-ai-go/internal/environment"
)

func GetByEmail(ctx context.Context, email string) (*Admin, error) {
	return FindFirst(ctx, &model.Options{
		Conditions: fmt.Sprintf("LOWER(%s) = :email: AND %s = 0",
			Columns.Email.Column(),
			Columns.Deleted.Column()),
		Params: map[string]interface{}{
			":email:": email,
		},
	})
}

func RepairID(ctx context.Context, oldID, newID types.UUID) error {
	// TODO: We actually will call this in production
	// if env.Env().IsProduction() {
	// 	return errors.Errorf("DEVChangeID called in production")
	// }

	urn := common.IDToURN(TABLE, newID)
	return environment.DB().GetDB().InsertWithContext(ctx, fmt.Sprintf(`
			UPDATE %s SET id = :id:, urn = :urn: WHERE id = :old_id:
			`, TABLE), map[string]any{
		":id:":     newID,
		":old_id:": oldID,
		":urn:":    urn,
	})
}
