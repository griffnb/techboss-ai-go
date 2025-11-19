package object_tag_test

import "github.com/griffnb/techboss-ai-go/internal/common/system_testing"

func init() {
	system_testing.BuildSystem()
}

/*
func TestGetWithCache(t *testing.T) {

	obj := object_tag.New()
	obj.Set(UNIT_TEST_FIELD, UNIT_TEST_VALUE)
	err := obj.Save(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer testtools.CleanupModel(obj)

	{
		objCache, err := object_tag.GetWithCache(obj.ID())
		if err != nil {
			t.Fatal(err)
		}

		if objCache.GetString(UNIT_TEST_FIELD) != UNIT_TEST_VALUE {
			t.Fatalf("Expect %v got %v", UNIT_TEST_VALUE, objCache.Get(UNIT_TEST_FIELD))
		}
	}

	{
		objCache, err := object_tag.GetWithCache(obj.ID())
		if err != nil {
			t.Fatal(err)
		}
		if objCache.GetString(UNIT_TEST_FIELD) != UNIT_TEST_VALUE {
			t.Fatalf("Expect %v got %v", UNIT_TEST_VALUE, objCache.Get(UNIT_TEST_FIELD))
		}

		obj.Set(UNIT_TEST_FIELD, UNIT_TEST_CHANGED_VALUE)
		err = obj.Save(nil)
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(1 * time.Second)
	}
	{
		objCache, err := object_tag.GetWithCache(obj.ID())
		if err != nil {
			t.Fatal(err)
		}

		if objCache.GetString(UNIT_TEST_FIELD) != UNIT_TEST_CHANGED_VALUE {
			t.Fatalf("Expect %v got %v", UNIT_TEST_VALUE, objCache.Get(UNIT_TEST_FIELD))
		}
	}

}
*/
