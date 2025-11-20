package dynamoqueue_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/cron/taskworker"
	"github.com/griffnb/techboss-ai-go/internal/cron/taskworker/worker_jobs"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/change_log"
)

func init() {
	system_testing.BuildSystem()
}

func TestReprocessDynamoThrottle(t *testing.T) {
	t.Run("Change Log", func(t *testing.T) {
		changeLog := change_log.NewDynamo()
		changeLog.BeforeValues = map[string]any{
			"test": "test",
			"nested": map[string]any{
				"test": "test",
			},
		}
		changeLog.UserURN = string(tools.GUID())
		changeLog.ObjectURN = string(tools.GUID())

		changeLog.AfterValues = map[string]any{
			"test": "test2",
			"nested2": map[string]any{
				"test": "test2",
			},
		}

		err := worker_jobs.QueueDynamoThrottleRetryJob(change_log.TABLE, changeLog)
		if err != nil {
			// Skip if DynamoDB table doesn't exist (local DynamoDB not running)
			if strings.Contains(err.Error(), "Cannot do operations on a non-existent table") {
				t.Skip("Skipping test: DynamoDB table does not exist (local DynamoDB may not be running)")
			}
			t.Fatal(err)
		}

		time.Sleep(2 * time.Second)
		msgs, err := environment.GetQueue().Pull(environment.QUEUE_THROTTLES, 1)
		if err != nil {
			t.Fatal(err)
		}

		if len(msgs) != 1 {
			t.Fatal("Expected 1 job, got ", len(msgs))
		}

		jobs, err := msgs[0].GetRawJobs()
		if err != nil {
			t.Fatal(err)
		}

		err = taskworker.New(environment.GetQueue(), "priority1").
			ProcessJob(context.Background(), jobs[0])
		if err != nil {
			t.Fatal(err)
		}

		logObj, err := change_log.GetDynamoChangeLogByID(context.Background(), changeLog.ID)
		if err != nil {
			t.Fatal(err)
		}

		if logObj.BeforeValues["test"] != "test" {
			t.Fatal("Expected BeforeValues to be test, got ", logObj.BeforeValues["test"])
		}
	})
}
