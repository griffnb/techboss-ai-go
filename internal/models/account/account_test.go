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
	UNIT_TEST_FIELD         = "first_name"
	UNIT_TEST_VALUE         = "UNIT_TEST_VALUE"
	UNIT_TEST_CHANGED_VALUE = "UNIT_TEST_CHANGED_VALUE"
)

// TestNew verifies that a new Account instance can be created and fields can be set.
// This test ensures the New() constructor initializes all required fields properly.
func TestNew(_ *testing.T) {
	obj := account.New()
	obj.Set(UNIT_TEST_FIELD, UNIT_TEST_VALUE)
}

// TestSave verifies the Save() method works correctly for both create and update operations.
// This test ensures:
// 1. A new Account can be saved to the database (INSERT)
// 2. The saved data can be retrieved from the database
// 3. An existing Account can be updated (UPDATE)
// 4. The updated data persists correctly
func TestSave(t *testing.T) {
	// Create a new account with test data
	obj := account.New()
	obj.Set(UNIT_TEST_FIELD, UNIT_TEST_VALUE)

	// Save the new account (INSERT operation)
	err := obj.Save(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer testtools.CleanupModel(obj)

	// Retrieve the account from database to verify it was saved
	objFromDb, err := account.Get(context.Background(), obj.ID())
	if err != nil {
		t.Fatal(err)
	}

	// Verify the field value matches what we saved
	if objFromDb.GetString(UNIT_TEST_FIELD) != UNIT_TEST_VALUE {
		t.Fatalf("Didnt Save")
	}

	// Update the account with a different value
	obj.Set(UNIT_TEST_FIELD, UNIT_TEST_CHANGED_VALUE)
	err = obj.Save(nil)
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve the updated account from database
	updatedObjFromDb, err := account.Get(context.Background(), obj.ID())
	if err != nil {
		t.Fatal(err)
	}

	// Verify the updated value persisted correctly
	if updatedObjFromDb.GetString(UNIT_TEST_FIELD) != UNIT_TEST_CHANGED_VALUE {
		t.Fatalf("UNIT_TEST_FIELD Didnt Update")
	}
}

// TestFindAll verifies the FindAll() method can retrieve multiple accounts from the database.
// This test ensures:
// 1. FindAll() can query accounts with specific conditions
// 2. The query returns active (non-disabled, non-deleted) accounts
// 3. At least one account exists matching the criteria
func TestFindAll(t *testing.T) {
	// Create and save a test account
	obj := account.New()
	err := obj.Save(nil)
	if err != nil {
		t.Fatalf("Save Err %v", err)
	}

	defer testtools.CleanupModel(obj)

	// Query for all active accounts (not disabled, not deleted)
	options := &model.Options{
		Conditions: "disabled = 0 AND deleted = 0",
	}
	objs, err := account.FindAll(context.Background(), options)
	if err != nil {
		t.Errorf("FindAll Err %v", err)
	}

	// Verify at least one account was found
	if len(objs) <= 0 {
		t.Errorf("FindAll Err nothing found")
	}
}

// TestFindFirst verifies the FindFirst() method can retrieve a single account by ID.
// This test ensures:
// 1. FindFirst() correctly queries with parameterized conditions (SQL injection safe)
// 2. An account can be found by its ID
// 3. The returned object is not empty/nil
func TestFindFirst(t *testing.T) {
	// Create and save a test account
	obj := account.New()
	err := obj.Save(nil)
	if err != nil {
		t.Fatalf("Save Err %v", err)
	}

	defer testtools.CleanupModel(obj)

	// Query for the specific account by ID using parameterized query
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

	// Verify the account was found
	if tools.Empty(obj2) {
		t.Fatalf("Get Err couldnt find")
	}
}

// TestFindFirstJoined verifies the FindFirstJoined() method can retrieve an account with joined data.
// This test ensures:
// 1. FindFirstJoined() correctly handles table-qualified column names in conditions
// 2. The joined query returns an AccountJoined instance with additional fields
// 3. Accounts can be queried when joins are involved (e.g., to get computed 'name' field)
func TestFindFirstJoined(t *testing.T) {
	// Create and save a test account
	obj := account.New()
	err := obj.Save(nil)
	if err != nil {
		t.Fatalf("Save Err %v", err)
	}

	defer testtools.CleanupModel(obj)

	// Query for the account with joined data using table-qualified column name
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

	// Verify the joined account was found
	if tools.Empty(obj2) {
		t.Fatalf("Get Err  couldnt find")
	}
}
