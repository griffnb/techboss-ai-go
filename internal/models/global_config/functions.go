package global_config

import (
	"context"

	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/pkg/errors"
)

func GetByKey(ctx context.Context, key constants.GlobalConfigKey) (*GlobalConfig, error) {
	return FindFirst(ctx, &model.Options{
		Conditions: "global_configs.key = :key:",
		Params: map[string]interface{}{
			":key:": key,
		},
	})
}

func GetValueByKey(ctx context.Context, key constants.GlobalConfigKey) (string, error) {
	configObj, err := GetByKey(ctx, key)
	if err != nil {
		return "", err
	}
	if tools.Empty(configObj) {
		return "", errors.Errorf("no config found for key %s", key)
	}
	return configObj.Val.Get(), nil
}
