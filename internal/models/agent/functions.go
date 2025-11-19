package agent

import (
	"context"
	"fmt"

	"github.com/griffnb/core/lib/model"
)

func GetByKey(ctx context.Context, key string) (*Agent, error) {
	options := &model.Options{
		Conditions: fmt.Sprintf("%s.key = :key:", TABLE),
		Params: map[string]interface{}{
			":key:": key,
		},
	}
	return FindFirst(ctx, options)
}
