package helpers

import (
	"context"
	"fmt"

	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/router"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/CrowdShield/go-core/lib/tools/slice"
)

func BuildBulkParams(ctx context.Context, rawPostBody map[string]interface{}, tableName string) (*model.Options, error) {
	updatedIDs := make([]string, 0)
	idsInterface, idsExist := rawPostBody["ids"]
	if idsExist {
		updatedIDs = slice.Convert[string](idsInterface)
		if tools.Empty(updatedIDs) {
			return nil, fmt.Errorf("ids must be passed")
		}
	}

	if !tools.Empty(updatedIDs) {
		return &model.Options{
			Conditions: fmt.Sprintf("%s.id IN(:ids:)", tableName),
			Params: map[string]interface{}{
				":ids:": updatedIDs,
			},
		}, nil
	}

	if !tools.Empty(rawPostBody["query"]) {
		postQuery, _ := rawPostBody["query"].(map[string]interface{})
		return router.BuildParams(ctx, postQuery, tableName), nil
	}

	return nil, nil
}
