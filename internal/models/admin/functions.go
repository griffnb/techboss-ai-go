package admin

import (
	"context"
	"fmt"

	"github.com/CrowdShield/go-core/lib/model"
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
