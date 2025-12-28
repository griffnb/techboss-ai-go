package sandbox_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/griffnb/core/lib/model"
	"github.com/griffnb/core/lib/testtools"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
)

func init() {
	system_testing.BuildSystem()
}

const (
	UNIT_TEST_FIELD         = "external_id"
	UNIT_TEST_VALUE         = "sb-test-123"
	UNIT_TEST_CHANGED_VALUE = "sb-test-456"
)

func TestNew(_ *testing.T) {
	obj := sandbox.New()
	obj.Set(UNIT_TEST_FIELD, UNIT_TEST_VALUE)
}

func TestSave(t *testing.T) {
	obj := sandbox.New()
	obj.Set(UNIT_TEST_FIELD, UNIT_TEST_VALUE)

	err := obj.Save(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer testtools.CleanupModel(obj)

	objFromDb, err := sandbox.Get(context.Background(), obj.ID())
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

	updatedObjFromDb, err := sandbox.Get(context.Background(), obj.ID())
	if err != nil {
		t.Fatal(err)
	}

	if updatedObjFromDb.GetString(UNIT_TEST_FIELD) != UNIT_TEST_CHANGED_VALUE {
		t.Fatalf("UNIT_TEST_FIELD Didnt Update")
	}
}

func TestFindAll(t *testing.T) {
	obj := sandbox.New()
	err := obj.Save(nil)
	if err != nil {
		t.Fatalf("Save Err %v", err)
	}

	defer testtools.CleanupModel(obj)

	options := &model.Options{
		Conditions: "disabled = 0 AND deleted = 0",
	}
	objs, err := sandbox.FindAll(context.Background(), options)
	if err != nil {
		t.Errorf("FindAll Err %v", err)
	}

	if len(objs) <= 0 {
		t.Errorf("FindAll Err nothing found")
	}
}

func TestFindFirst(t *testing.T) {
	obj := sandbox.New()
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
	obj2, err := sandbox.FindFirst(context.Background(), options)
	if err != nil {
		t.Fatalf("Get Err %v", err)
	}

	if tools.Empty(obj2) {
		t.Fatalf("Get Err couldnt find")
	}
}

func TestFindFirstJoined(t *testing.T) {
	obj := sandbox.New()
	err := obj.Save(nil)
	if err != nil {
		t.Fatalf("Save Err %v", err)
	}

	defer testtools.CleanupModel(obj)

	options := &model.Options{
		Conditions: fmt.Sprintf("%s.id = :id:", sandbox.TABLE),
		Params: map[string]any{
			":id:": obj.ID(),
		},
	}
	obj2, err := sandbox.FindFirstJoined(context.Background(), options)
	if err != nil {
		t.Fatalf("Get Err %v", err)
	}

	if tools.Empty(obj2) {
		t.Fatalf("Get Err  couldnt find")
	}
}

func Test_Sandbox_SaveWithMetaData(t *testing.T) {
	t.Run("saves sandbox with minimal metadata", func(t *testing.T) {
		// Arrange - Create sandbox with minimal metadata
		obj := sandbox.New()
		obj.ExternalID.Set("sb-test-minimal")
		obj.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		obj.MetaData.Set(&sandbox.MetaData{}) // Empty metadata

		// Act
		err := obj.Save(nil)
		// Assert
		if err != nil {
			t.Fatalf("Failed to save sandbox: %v", err)
		}
		defer testtools.CleanupModel(obj)

		// Verify sandbox was saved and can be retrieved
		retrieved, err := sandbox.Get(context.Background(), obj.ID())
		if err != nil {
			t.Fatalf("Failed to retrieve sandbox: %v", err)
		}

		if tools.Empty(retrieved) {
			t.Fatal("Retrieved sandbox is empty")
		}

		// Verify fields
		if retrieved.ExternalID.Get() != "sb-test-minimal" {
			t.Fatalf("Expected external_id 'sb-test-minimal', got '%s'", retrieved.ExternalID.Get())
		}

		if retrieved.Provider.Get() != sandbox.PROVIDER_CLAUDE_CODE {
			t.Fatalf("Expected provider %d, got %d", sandbox.PROVIDER_CLAUDE_CODE, retrieved.Provider.Get())
		}

		// Verify metadata exists and is empty
		metadata, err := retrieved.MetaData.Get()
		if err != nil {
			t.Fatalf("Failed to get metadata: %v", err)
		}
		if metadata == nil {
			t.Fatal("Metadata should not be nil")
		}

		if metadata.LastS3Sync != nil {
			t.Fatal("LastS3Sync should be nil for minimal metadata")
		}

		if metadata.SyncStats != nil {
			t.Fatal("SyncStats should be nil for minimal metadata")
		}
	})

	t.Run("saves sandbox with populated metadata", func(t *testing.T) {
		// Arrange - Create sandbox with populated metadata
		obj := sandbox.New()
		obj.ExternalID.Set("sb-test-populated")
		obj.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)

		metadata := &sandbox.MetaData{}
		metadata.UpdateLastSync(10, 2, 5, 1024, 500)
		obj.MetaData.Set(metadata)

		// Act
		err := obj.Save(nil)
		// Assert
		if err != nil {
			t.Fatalf("Failed to save sandbox: %v", err)
		}
		defer testtools.CleanupModel(obj)

		// Verify sandbox was saved and can be retrieved
		retrieved, err := sandbox.Get(context.Background(), obj.ID())
		if err != nil {
			t.Fatalf("Failed to retrieve sandbox: %v", err)
		}

		if tools.Empty(retrieved) {
			t.Fatal("Retrieved sandbox is empty")
		}

		// Verify metadata was persisted
		retrievedMetadata, err := retrieved.MetaData.Get()
		if err != nil {
			t.Fatalf("Failed to get metadata: %v", err)
		}
		if retrievedMetadata == nil {
			t.Fatal("Metadata should not be nil")
		}

		if retrievedMetadata.LastS3Sync == nil {
			t.Fatal("LastS3Sync should not be nil")
		}

		if retrievedMetadata.SyncStats == nil {
			t.Fatal("SyncStats should not be nil")
		}

		if retrievedMetadata.SyncStats.FilesDownloaded != 10 {
			t.Fatalf("Expected FilesDownloaded 10, got %d", retrievedMetadata.SyncStats.FilesDownloaded)
		}

		if retrievedMetadata.SyncStats.FilesDeleted != 2 {
			t.Fatalf("Expected FilesDeleted 2, got %d", retrievedMetadata.SyncStats.FilesDeleted)
		}

		if retrievedMetadata.SyncStats.FilesSkipped != 5 {
			t.Fatalf("Expected FilesSkipped 5, got %d", retrievedMetadata.SyncStats.FilesSkipped)
		}

		if retrievedMetadata.SyncStats.BytesTransferred != 1024 {
			t.Fatalf("Expected BytesTransferred 1024, got %d", retrievedMetadata.SyncStats.BytesTransferred)
		}

		if retrievedMetadata.SyncStats.DurationMs != 500 {
			t.Fatalf("Expected DurationMs 500, got %d", retrievedMetadata.SyncStats.DurationMs)
		}
	})
}
