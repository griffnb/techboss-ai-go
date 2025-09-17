package worker_jobs

import (
	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/queue"
	"github.com/griffnb/techboss-ai-go/internal/environment"
)

const DYNAMO_THROTTLE_RETRY = "dynamo_throttle_retry"

type DynamoThrottleRetryJob struct {
	ObjectTable string `json:"object_table"`
	ObjData     any    `json:"obj_data"`
}

func QueueDynamoThrottleRetryJob(objTable string, data any) error {
	jobData := &DynamoThrottleRetryJob{
		ObjectTable: objTable,
		ObjData:     data,
	}
	job := &queue.Job{
		Type: DYNAMO_THROTTLE_RETRY,
		Data: jobData,
	}
	err := environment.GetQueue().Push(environment.QUEUE_THROTTLES, job)
	if err != nil {
		log.Error(err)
	}
	return nil
}
