package taskworker

import (
	"context"
	"encoding/json"

	"github.com/griffnb/core/lib/queue"
	"github.com/griffnb/techboss-ai-go/internal/cron/taskworker/worker_jobs"
	dynamoqueue "github.com/griffnb/techboss-ai-go/internal/services/dynamo_queue"
)

func (this *TaskWorker) ProcessJob(ctx context.Context, job *queue.RawJob) error {
	switch job.Type {
	case worker_jobs.DYNAMO_THROTTLE_RETRY:
		jobData := &worker_jobs.DynamoThrottleRetryJob{
			ObjData: &json.RawMessage{},
		}
		err := job.GetData(jobData)
		if err != nil {
			return err
		}

		err = dynamoqueue.ReprocessDynamoThrottle(ctx, jobData)
		if err != nil {
			return err
		}
	}

	return nil
}
