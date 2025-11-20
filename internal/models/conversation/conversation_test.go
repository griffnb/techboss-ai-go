package conversation_test

import (
	"context"
	"testing"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/testtools"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	testmodel "github.com/griffnb/techboss-ai-go/internal/models/conversation"
)

func init() {
	system_testing.BuildSystem()
}

func TestNew(_ *testing.T) {
	obj := testmodel.New()
	obj.AccountID.Set(tools.GUID())
}

func TestSave(t *testing.T) {
	testAccountID := tools.GUID()
	changedAccountID := tools.GUID()
	
	obj := testmodel.New()
	obj.AccountID.Set(testAccountID)

	err := obj.Save(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer testtools.CleanupModel(obj)

	objFromDb, err := testmodel.Get(context.Background(), obj.ID())
	if err != nil {
		t.Fatal(err)
	}

	if objFromDb.AccountID.Get() != testAccountID {
		t.Fatalf("Didnt Save")
	}

	obj.AccountID.Set(changedAccountID)
	err = obj.Save(nil)
	if err != nil {
		t.Fatal(err)
	}

	updatedObjFromDb, err := testmodel.Get(context.Background(), obj.ID())
	if err != nil {
		t.Fatal(err)
	}

	if updatedObjFromDb.AccountID.Get() != changedAccountID {
		t.Fatalf("UNIT_TEST_FIELD Didnt Update")
	}
}

func TestFindAll(t *testing.T) {
	obj := testmodel.New()
	err := obj.Save(nil)
	if err != nil {
		t.Fatalf(`Save Err %v`, err)
	}

	defer testtools.CleanupModel(obj)

	options := &model.Options{
		Conditions: "disabled =0 AND deleted = 0",
	}
	objs, err := testmodel.FindAll(context.Background(), options)
	if err != nil {
		t.Errorf(`FindAll Err %v`, err)
	}

	if len(objs) <= 0 {
		t.Errorf(`FindAll Err nothing found`)
	}
}

func TestFindFirst(t *testing.T) {
	obj := testmodel.New()
	err := obj.Save(nil)
	if err != nil {
		t.Fatalf(`Save Err %v`, err)
	}

	defer testtools.CleanupModel(obj)

	options := &model.Options{
		Conditions: "id = :id:",
		Params: map[string]interface{}{
			":id:": obj.ID(),
		},
	}
	obj2, err := testmodel.FindFirst(context.Background(), options)
	if err != nil {
		t.Fatalf(`Get Err %v`, err)
	}

	if tools.Empty(obj2) {
		t.Fatalf(`Get Err  couldnt find`)
	}
}
