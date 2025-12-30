package lifecycle_test_placeholder_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/testtools"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/models/lifecycle_test_placeholder"
)

func init() {
	system_testing.BuildSystem()
}

const (
	UNIT_TEST_FIELD         = "name"
	UNIT_TEST_VALUE         = "UNIT_TEST_VALUE"
	UNIT_TEST_CHANGED_VALUE = "UNIT_TEST_CHANGED_VALUE"
)

func TestNew(_ *testing.T) {
	obj := lifecycle_test_placeholder.New()
	obj.Set(UNIT_TEST_FIELD, UNIT_TEST_VALUE)
}

func TestSave(t *testing.T) {
	obj := lifecycle_test_placeholder.New()
	obj.Set(UNIT_TEST_FIELD, UNIT_TEST_VALUE)

	err := obj.Save(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer testtools.CleanupModel(obj)

	objFromDb, err := lifecycle_test_placeholder.Get(context.Background(), obj.ID())
	if err != nil {
		t.Fatal(err)
	}

	if objFromDb.GetString(UNIT_TEST_FIELD) != UNIT_TEST_VALUE {
		t.Fatalf("Didnt Save")
	}

	obj.Set(UNIT_TEST_FIELD, UNIT_TEST_CHANGED_VALUE)
	err = obj.Save(nil)
	if err != nil {
		t.Fatal(err)
	}

	updatedObjFromDb, err := lifecycle_test_placeholder.Get(context.Background(), obj.ID())
	if err != nil {
		t.Fatal(err)
	}

	if updatedObjFromDb.GetString(UNIT_TEST_FIELD) != UNIT_TEST_CHANGED_VALUE {
		t.Fatalf("UNIT_TEST_FIELD Didnt Update")
	}
}

func TestFindAll(t *testing.T) {
	obj := lifecycle_test_placeholder.New()
	err := obj.Save(nil)
	if err != nil {
		t.Fatalf("Save Err %v", err)
	}

	defer testtools.CleanupModel(obj)

	options := &model.Options{
		Conditions: "disabled = 0 AND deleted = 0",
	}
	objs, err := lifecycle_test_placeholder.FindAll(context.Background(), options)
	if err != nil {
		t.Errorf("FindAll Err %v", err)
	}

	if len(objs) <= 0 {
		t.Errorf("FindAll Err nothing found")
	}
}

func TestFindFirst(t *testing.T) {
	obj := lifecycle_test_placeholder.New()
	err := obj.Save(nil)
	if err != nil {
		t.Fatalf("Save Err %v", err)
	}

	defer testtools.CleanupModel(obj)

	options := &model.Options{
		Conditions: "id = :id:",
		Params: map[string]any{
			":id:": obj.ID(),
		},
	}
	obj2, err := lifecycle_test_placeholder.FindFirst(context.Background(), options)
	if err != nil {
		t.Fatalf("Get Err %v", err)
	}

	if tools.Empty(obj2) {
		t.Fatalf("Get Err couldnt find")
	}
}

func TestFindFirstJoined(t *testing.T) {
	obj := lifecycle_test_placeholder.New()
	err := obj.Save(nil)
	if err != nil {
		t.Fatalf("Save Err %v", err)
	}

	defer testtools.CleanupModel(obj)

	options := &model.Options{
		Conditions: fmt.Sprintf("%s.id = :id:", lifecycle_test_placeholder.TABLE),
		Params: map[string]any{
			":id:": obj.ID(),
		},
	}
	obj2, err := lifecycle_test_placeholder.FindFirstJoined(context.Background(), options)
	if err != nil {
		t.Fatalf("Get Err %v", err)
	}

	if tools.Empty(obj2) {
		t.Fatalf("Get Err  couldnt find")
	}
}
