package delay_queue_test

import (
	"context"
	"testing"
	"time"

	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/cron/taskworker/delay_queue"
)

func init() {
	system_testing.BuildSystem()
}

func TestRunDelayQueue(t *testing.T) {
	{
		item := delay_queue.NewItem("test", time.Now().Unix(), map[string]any{
			"test": "test",
		})

		err := item.Save(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		defer func() {
			_, _ = item.Delete(context.Background())
		}()
	}

	{
		item := delay_queue.NewItem("test", time.Now().Unix()-5, map[string]any{
			"test2": "test2",
		})

		err := item.Save(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			_, _ = item.Delete(context.Background())
		}()
	}

	{
		item := delay_queue.NewItem("test", time.Now().Unix()+25, map[string]any{
			"test3": "test3",
		})

		err := item.Save(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			_, _ = item.Delete(context.Background())
		}()
	}

	time.Sleep(1 * time.Second)

	items, err := delay_queue.PopFromDynamo(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	lockSuccesses := 0
	for _, item := range items {
		success, err := item.CheckLock(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		if success {
			lockSuccesses++
		}
	}

	if lockSuccesses != 2 {
		t.Fatalf("expected 2 lock successes, got %d", lockSuccesses)
	}

	lockSuccesses = 0
	for _, item := range items {
		success, err := item.CheckLock(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		if success {
			lockSuccesses++
		}
	}

	if lockSuccesses != 0 {
		t.Fatalf("expected 0 lock successes, got %d", lockSuccesses)
	}
}
