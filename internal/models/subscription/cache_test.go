package subscription_test

import (
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
)

func init() {
	system_testing.BuildSystem()
}

/*
func TestGetWithCache(t *testing.T) {

	ctx := context.Background()
	obj, err := subscription.New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	obj.UNIT_TEST_FIELD.Set("UNIT_TEST_VALUE")
	err = obj.Save(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer testtools.CleanupModel(obj)

	t.Run("Get object from database and cache it", func(t *testing.T) {
		objCache, err := subscription.GetWithCache(ctx, obj.ID())
		if err != nil {
			t.Fatal(err)
		}

		assert.Eq(t, objCache.UNIT_TEST_FIELD.Get(), "UNIT_TEST_VALUE", "Should match the original value")
	})

	t.Run("Get object from cache", func(t *testing.T) {
		objCache, err := subscription.GetWithCache(ctx, obj.ID())
		if err != nil {
			t.Fatal(err)
		}
		assert.Eq(t, objCache.UNIT_TEST_FIELD.Get(), "UNIT_TEST_VALUE", "Should match the original value")

		obj.UNIT_TEST_FIELD.Set("UNIT_TEST_CHANGED_VALUE")
		err = obj.Save(nil)
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(1 * time.Second)
	})

	t.Run("Get updated object after cache expiration", func(t *testing.T) {
		objCache, err := subscription.GetWithCache(ctx, obj.ID())
		if err != nil {
			t.Fatal(err)
		}

		assert.Eq(t, objCache.UNIT_TEST_FIELD.Get(), "UNIT_TEST_CHANGED_VALUE", "Should match the updated value")
	})
}
*/
