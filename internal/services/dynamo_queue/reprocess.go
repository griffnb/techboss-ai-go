package dynamoqueue

import (
	"context"
	"encoding/json"

	"github.com/griffnb/techboss-ai-go/internal/cron/taskworker/worker_jobs"
	"github.com/griffnb/techboss-ai-go/internal/models/change_log"
	"github.com/pkg/errors"
)

func ReprocessDynamoThrottle(ctx context.Context, jobData *worker_jobs.DynamoThrottleRetryJob) error {
	rawJobData, ok := jobData.ObjData.(*json.RawMessage)
	if !ok {
		return errors.Errorf("failed to cast ObjData to json.RawMessage is %T", jobData.ObjData)
	}

	switch jobData.ObjectTable {
	case change_log.TABLE:
		changeLog := &change_log.ChangeLogDynamo{}
		err := json.Unmarshal(*rawJobData, changeLog)
		if err != nil {
			return errors.WithStack(err)
		}

		err = changeLog.Save(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
