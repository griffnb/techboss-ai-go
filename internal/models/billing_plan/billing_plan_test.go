package billing_plan_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/testtools"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/models/billing_plan"
)

func init() {
	system_testing.BuildSystem()
}

const UNIT_TEST_FIELD = "name"
const UNIT_TEST_VALUE = "UNIT_TEST_VALUE"
const UNIT_TEST_CHANGED_VALUE = "UNIT_TEST_CHANGED_VALUE"

func TestNew(_ *testing.T) {
	obj := billing_plan.New()
	obj.Set(UNIT_TEST_FIELD, UNIT_TEST_VALUE)
}

func TestSave(t *testing.T) {

	obj := billing_plan.New()
	obj.Set(UNIT_TEST_FIELD, UNIT_TEST_VALUE)

	err := obj.Save(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer testtools.CleanupModel(obj)

	objFromDb, err := billing_plan.Get(context.Background(), obj.ID())
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

	updatedObjFromDb, err := billing_plan.Get(context.Background(), obj.ID())
	if err != nil {
		t.Fatal(err)
	}

	if updatedObjFromDb.GetString(UNIT_TEST_FIELD) != UNIT_TEST_CHANGED_VALUE {
		t.Fatalf("UNIT_TEST_FIELD Didnt Update")
	}

}

func TestFindAll(t *testing.T) {

	obj := billing_plan.New()
	err := obj.Save(nil)
	if err != nil {
		t.Fatalf("Save Err %v", err)
	}

	defer testtools.CleanupModel(obj)

	options := &model.Options{
		Conditions: "disabled = 0 AND deleted = 0",
	}
	objs, err := billing_plan.FindAll(context.Background(), options)
	if err != nil {
		t.Errorf("FindAll Err %v", err)
	}

	if len(objs) <= 0 {
		t.Errorf("FindAll Err nothing found")
	}

}

func TestFindFirst(t *testing.T) {

	obj := billing_plan.New()
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
	obj2, err := billing_plan.FindFirst(context.Background(), options)
	if err != nil {
		t.Fatalf("Get Err %v", err)
	}

	if tools.Empty(obj2) {
		t.Fatalf("Get Err couldnt find")
	}
}

func TestFindFirstJoined(t *testing.T) {

	obj := billing_plan.New()
	err := obj.Save(nil)
	if err != nil {
		t.Fatalf("Save Err %v", err)
	}

	defer testtools.CleanupModel(obj)

	options := &model.Options{
		Conditions: fmt.Sprintf("%s.id = :id:", billing_plan.TABLE),
		Params: map[string]any{
			":id:": obj.ID(),
		},
	}
	obj2, err := billing_plan.FindFirstJoined(context.Background(), options)
	if err != nil {
		t.Fatalf("Get Err %v", err)
	}

	if tools.Empty(obj2) {
		t.Fatalf("Get Err  couldnt find")
	}
}
