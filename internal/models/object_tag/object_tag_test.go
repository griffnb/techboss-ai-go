package object_tag_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/testtools"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	testmodel "github.com/griffnb/techboss-ai-go/internal/models/object_tag"
	"github.com/griffnb/techboss-ai-go/internal/models/tag"
)

func init() {
	system_testing.BuildSystem()
}

const (
	UNIT_TEST_FIELD         = "status"
	UNIT_TEST_VALUE         = 1
	UNIT_TEST_CHANGED_VALUE = 2
)

func TestNew(_ *testing.T) {
	obj := testmodel.New()
	obj.Set(UNIT_TEST_FIELD, UNIT_TEST_VALUE)
}

func TestSave(t *testing.T) {
	tagObj := tag.New()
	tagObj.Name.Set(tools.RandString(10))
	err := tagObj.Save(nil)
	if err != nil {
		t.Fatal(err)
	}

	obj := testmodel.New()
	obj.Set(UNIT_TEST_FIELD, UNIT_TEST_VALUE)
	obj.ObjectURN.Set("urn:object:tag:" + string(tools.GUID()))
	obj.TagID.Set(tagObj.ID())
	err = obj.Save(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer testtools.CleanupModel(obj)

	objFromDb, err := testmodel.Get(context.Background(), obj.ID())
	if err != nil {
		t.Fatal(err)
	}

	if objFromDb.GetInt(UNIT_TEST_FIELD) != UNIT_TEST_VALUE {
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

	if updatedObjFromDb.GetInt(UNIT_TEST_FIELD) != UNIT_TEST_CHANGED_VALUE {
		t.Fatalf(`UNIT_TEST_FIELD Didnt Update`)
	}
}

func TestFindAll(t *testing.T) {
	tagObj := tag.New()
	tagObj.Name.Set(tools.RandString(10))
	err := tagObj.Save(nil)
	if err != nil {
		t.Fatal(err)
	}

	obj := testmodel.New()
	obj.ObjectURN.Set("urn:object:tag:" + string(tools.GUID()))
	obj.TagID.Set(tagObj.ID())
	err = obj.Save(nil)
	if err != nil {
		t.Fatalf(`Save Err %v`, err)
	}

	defer testtools.CleanupModel(obj)

	options := &model.Options{}
	objs, err := testmodel.FindAll(context.Background(), options)
	if err != nil {
		t.Errorf(`FindAll Err %v`, err)
	}

	if len(objs) <= 0 {
		t.Errorf(`FindAll Err nothing found`)
	}
}

func TestFindFirst(t *testing.T) {
	tagObj := tag.New()
	tagObj.Name.Set(tools.RandString(10))
	err := tagObj.Save(nil)
	if err != nil {
		t.Fatal(err)
	}

	obj := testmodel.New()
	obj.ObjectURN.Set("urn:object:tag:" + string(tools.GUID()))
	obj.TagID.Set(tagObj.ID())
	err = obj.Save(nil)
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
	tagObj := tag.New()
	tagObj.Name.Set(tools.RandString(10))
	err := tagObj.Save(nil)
	if err != nil {
		t.Fatal(err)
	}

	obj := testmodel.New()
	obj.ObjectURN.Set("urn:object:tag:" + string(tools.GUID()))
	obj.TagID.Set(tagObj.ID())
	err = obj.Save(nil)
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
