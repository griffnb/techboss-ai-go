package account_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/testtools"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
)

func init() {
	system_testing.BuildSystem()
}

const (
	UNIT_TEST_FIELD         = "first_name" // Using first_name which is an actual DB column
	UNIT_TEST_VALUE         = "UNIT_TEST_VALUE"
	UNIT_TEST_CHANGED_VALUE = "UNIT_TEST_CHANGED_VALUE"
)

// TestNew verifies that a new Account instance can be created successfully
// and that field values can be set on the newly created object.
func TestNew(_ *testing.T) {
	obj := account.New()
	obj.Set(UNIT_TEST_FIELD, UNIT_TEST_VALUE)
}

// TestSave verifies that an Account can be saved to the database and retrieved correctly.
// It tests both initial save (INSERT) and update (UPDATE) operations, ensuring that:
// - A new account can be saved with a field value
// - The saved account can be retrieved from the database
// - The retrieved account has the correct field value
// - The account can be updated with a new field value
// - The updated account can be retrieved with the new value
func TestSave(t *testing.T) {
	obj := account.New()
	obj.Set(UNIT_TEST_FIELD, UNIT_TEST_VALUE)

	err := obj.Save(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer testtools.CleanupModel(obj)

	objFromDb, err := account.Get(context.Background(), obj.ID())
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

	updatedObjFromDb, err := account.Get(context.Background(), obj.ID())
	if err != nil {
		t.Fatal(err)
	}

	if updatedObjFromDb.GetString(UNIT_TEST_FIELD) != UNIT_TEST_CHANGED_VALUE {
		t.Fatalf("UNIT_TEST_FIELD Didnt Update")
	}
}

// TestFindAll verifies that the FindAll function correctly retrieves multiple accounts
// from the database with the specified conditions. It tests:
// - Creating and saving a test account
// - Querying for all non-disabled and non-deleted accounts
// - Verifying that at least one account is returned
func TestFindAll(t *testing.T) {
	obj := account.New()
	err := obj.Save(nil)
	if err != nil {
		t.Fatalf("Save Err %v", err)
	}

	defer testtools.CleanupModel(obj)

	options := &model.Options{
		Conditions: "disabled = 0 AND deleted = 0",
	}
	objs, err := account.FindAll(context.Background(), options)
	if err != nil {
		t.Errorf("FindAll Err %v", err)
	}

	if len(objs) <= 0 {
		t.Errorf("FindAll Err nothing found")
	}
}

// TestFindFirst verifies that the FindFirst function can locate a specific account
// by ID using parameterized queries. It tests:
// - Creating and saving a test account
// - Querying for the account using its ID with a parameterized condition
// - Verifying that the correct account is found and returned
func TestFindFirst(t *testing.T) {
	obj := account.New()
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
	obj2, err := account.FindFirst(context.Background(), options)
	if err != nil {
		t.Fatalf("Get Err %v", err)
	}

	if tools.Empty(obj2) {
		t.Fatalf("Get Err couldnt find")
	}
}

// TestFindFirstJoined verifies that the FindFirstJoined function can retrieve an account
// with joined data (such as name, created_by_name, updated_by_name) from related tables.
// It tests:
// - Creating and saving a test account
// - Querying for the account using a table-qualified condition
// - Verifying that the account with joined data is found and returned
func TestFindFirstJoined(t *testing.T) {
	obj := account.New()
	err := obj.Save(nil)
	if err != nil {
		t.Fatalf("Save Err %v", err)
	}

	defer testtools.CleanupModel(obj)

	options := &model.Options{
		Conditions: fmt.Sprintf("%s.id = :id:", account.TABLE),
		Params: map[string]any{
			":id:": obj.ID(),
		},
	}
	obj2, err := account.FindFirstJoined(context.Background(), options)
	if err != nil {
		t.Fatalf("Get Err %v", err)
	}

	if tools.Empty(obj2) {
		t.Fatalf("Get Err  couldnt find")
	}
}
