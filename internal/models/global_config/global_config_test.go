package global_config_test

import (
	"context"
	"testing"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/testtools"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/global_config"
)

func TestNew(_ *testing.T) {
	obj := global_config.New()
	obj.Set("value", "54871")
}

func init() {
	system_testing.BuildSystem()
}

func TestFindAll(t *testing.T) {
	ctx := context.Background()

	obj := global_config.New()
	obj.Set("value", "54871")
	err := obj.Save(nil)
	if err != nil {
		t.Fatalf(`Save Err %v`, err)
	}
	defer testtools.CleanupModelWithClient(environment.GetDBClient(global_config.CLIENT), obj)

	options := &model.Options{
		Conditions: "disabled = 0",
	}
	objs, err := global_config.FindAll(ctx, options)
	if err != nil {
		t.Errorf(`FindAll Err %v`, err)
	}

	if len(objs) == 0 {
		t.Errorf(`FindAll Err nothing found`)
	}
}

func TestFindFirst(t *testing.T) {
	ctx := context.Background()

	obj := global_config.New()
	obj.Set("value", "54871")
	err := obj.Save(nil)
	if err != nil {
		t.Fatalf(`Save Err %v`, err)
	}
	defer testtools.CleanupModelWithClient(environment.GetDBClient(global_config.CLIENT), obj)

	options := &model.Options{
		Conditions: "id = :id:",
		Params: map[string]interface{}{
			":id:": obj.ID(),
		},
	}
	obj2, err := global_config.FindFirst(ctx, options)
	if err != nil {
		t.Errorf(`Get Err %v`, err)
	}

	if obj2 == nil {
		t.Errorf(`Get Err  couldnt find`)
	}
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	obj := global_config.New()
	obj.Set("value", "54871")
	err := obj.Save(nil)
	if err != nil {
		t.Fatalf(`Save Err %v`, err)
	}
	defer testtools.CleanupModelWithClient(environment.GetDBClient(global_config.CLIENT), obj)
	obj2, err := global_config.Get(ctx, obj.ID())
	if err != nil {
		t.Errorf(`Get Err %v`, err)
	}

	if obj2 == nil {
		t.Errorf(`Get Err  couldnt find`)
	}
}
