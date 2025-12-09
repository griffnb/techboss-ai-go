package admin_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/testtools"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	testmodel "github.com/griffnb/techboss-ai-go/internal/models/admin"
)

func init() {
	system_testing.BuildSystem()
}

const (
	UNIT_TEST_FIELD         = "first_name"
	UNIT_TEST_VALUE         = "UNIT_TEST_VALUE"
	UNIT_TEST_CHANGED_VALUE = "UNIT_TEST_CHANGED_VALUE"
)

func TestNew(_ *testing.T) {
	obj := testmodel.New()
	obj.Set(UNIT_TEST_FIELD, UNIT_TEST_VALUE)
}

func TestSave(t *testing.T) {
	obj := testmodel.New()
	obj.Email.Set(tools.RandString(10))
	obj.Set(UNIT_TEST_FIELD, UNIT_TEST_VALUE)

	err := obj.Save(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer testtools.CleanupModel(obj)

	objFromDb, err := testmodel.Get(context.Background(), obj.ID())
	if err != nil {
		t.Fatal(err)
	}

	if objFromDb.GetString(UNIT_TEST_FIELD) != UNIT_TEST_VALUE {
		t.Fatalf(`Didnt Save`)
	}

	obj.Set(UNIT_TEST_FIELD, UNIT_TEST_CHANGED_VALUE)
	err = obj.Save(nil)
	if err != nil {
		t.Fatal(err)
	}

	updatedObjFromDb, err := testmodel.Get(context.Background(), obj.ID())
	if err != nil {
		t.Fatal(err)
	}

	if updatedObjFromDb.GetString(UNIT_TEST_FIELD) != UNIT_TEST_CHANGED_VALUE {
		t.Fatalf(`UNIT_TEST_FIELD Didnt Update`)
	}

	bookMarks, _ := updatedObjFromDb.Bookmarks.Get()
	bookMarks.Pages = []*testmodel.Bookmark{
		{
			Name: "some test",
		},
	}
	updatedObjFromDb.Bookmarks.Set(bookMarks)
	err = updatedObjFromDb.Save(nil)
	if err != nil {
		t.Fatal(err)
	}

	fromDB2, err := testmodel.Get(context.Background(), obj.ID())
	if err != nil {
		t.Fatal(err)
	}

	log.PrintEntity(fromDB2)
}

func TestFindAll(t *testing.T) {
	obj := testmodel.New()
	obj.Email.Set(tools.RandString(10))
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

	log.PrintEntity(objs)
}

func TestFindFirst(t *testing.T) {
	obj := testmodel.New()
	obj.Email.Set(tools.RandString(10))
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

func TestFindFirstJoined(t *testing.T) {
	obj := testmodel.New()
	obj.Email.Set(tools.RandString(10))
	err := obj.Save(nil)
	if err != nil {
		t.Fatalf(`Save Err %v`, err)
	}

	defer testtools.CleanupModel(obj)

	options := &model.Options{
		Conditions: fmt.Sprintf("%s.id = :id:", testmodel.TABLE),
		Params: map[string]interface{}{
			":id:": obj.ID(),
		},
	}
	obj2, err := testmodel.FindFirstJoined(context.Background(), options)
	if err != nil {
		t.Fatalf(`Get Err %v`, err)
	}

	if tools.Empty(obj2) {
		t.Fatalf(`Get Err  couldnt find`)
	}
}

func TestDupe(t *testing.T) {
	/*
	   email
	*/

	random := tools.RandString(10)

	obj := testmodel.New()
	obj.Email.Set(random)

	err := obj.Save(nil)
	if err != nil {
		t.Fatalf(`Save Err %v`, err)
	}

	defer testtools.CleanupModel(obj)

	obj2 := testmodel.New()
	obj2.Email.Set(random)

	err = obj2.Save(nil)
	if err == nil {
		t.Fatalf(`Should have failed for dupe`)
	}

	defer testtools.CleanupModel(obj2)

	obj.Status.Set(constants.STATUS_DELETED)
	err = obj.Save(nil)
	if err != nil {
		t.Fatalf(`Save Err %v`, err)
	}

	err = obj2.Save(nil)
	if err != nil {
		t.Fatalf(`Save Err %v`, err)
	}
}
