package change_log

import (
	"context"
	"math/rand"
	"time"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/CrowdShield/go-core/lib/tools/maps"
	"github.com/CrowdShield/go-core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/cron/taskworker/worker_jobs"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/pkg/errors"
)

const MAX_RETRIES = 5

type ChangeLogDynamo struct {
	ID           string         `json:"id"            column:"id"            type:"text"`
	ObjectID     types.UUID     `json:"object_id"     column:"object_id"     type:"uuid"   index:"true" default:"gen_random_uuid()"`
	Type         string         `json:"type"          column:"type"          type:"text"   index:"true" default:""`
	UserURN      string         `json:"user_urn"      column:"user_urn"      type:"text"   index:"true" default:""`
	ObjectURN    string         `json:"object_urn"    column:"object_urn"    type:"text"   index:"true" default:""`
	BeforeValues map[string]any `json:"before_values" column:"before_values" type:"jsonb"               default:"{}"`
	AfterValues  map[string]any `json:"after_values"  column:"after_values"  type:"jsonb"               default:"{}"`
	Timestamp    int64          `json:"timestamp"     column:"timestamp"     type:"bigint"`
}

func NewDynamo() *ChangeLogDynamo {
	return &ChangeLogDynamo{
		Timestamp: time.Now().Unix(),
		ID:        tools.SessionKey(),
	}
}

func (this *ChangeLogDynamo) Save(ctx context.Context) (err error) {
	if tools.Empty(this.ID) {
		this.ID = tools.SessionKey()
	}
	if tools.Empty(this.Timestamp) {
		this.Timestamp = time.Now().Unix()
	}

	return this.saveWithRetries(ctx, MAX_RETRIES)
}

// TODO maybe pop to sqs for worst case
func (this *ChangeLogDynamo) saveWithRetries(ctx context.Context, retries int) error {
	if retries <= 0 {
		err := worker_jobs.QueueDynamoThrottleRetryJob(TABLE, this)
		if err != nil {
			return err
		}
		return nil
	}

	throttled, err := environment.GetDynamo().PutWithContext(ctx, TABLE, this)
	if err != nil && !throttled {
		return err
	}

	if !throttled {
		return nil
	}

	// Exponential backoff with jitter
	// #nosec
	backoff := time.Duration(1<<(5-retries))*time.Millisecond + time.Duration(rand.Intn(100))*time.Millisecond
	select {
	case <-time.After(backoff):
		return this.saveWithRetries(ctx, retries-1)
	case <-ctx.Done():
		return errors.WithStack(ctx.Err())
	}
}

func GetChangeLogsByObjectURN(ctx context.Context, objectURN string, limit int32) ([]*ChangeLogDynamo, error) {
	logs := []*ChangeLogDynamo{}
	throttled, err := environment.GetDynamo().GetByIndexWithContext(ctx, TABLE, "object_urn", objectURN, &logs, limit, false)
	if err != nil {
		return nil, err
	}
	if throttled {
		return nil, errors.Errorf("event throttled by key")
	}

	return logs, nil
}

func GetDynamoChangeLogByID(ctx context.Context, id string) (*ChangeLogDynamo, error) {
	log := &ChangeLogDynamo{}
	throttled, err := environment.GetDynamo().GetWithContext(ctx, TABLE, "id", id, log, true)
	if err != nil {
		return nil, err
	}
	if throttled {
		return nil, errors.Errorf("event throttled by key")
	}

	return log, nil
}

func DynamoChangeLog(ctx context.Context, change *model.Change) {
	changeLog := NewDynamo()
	changeLog.Timestamp = time.Now().Unix()
	// Unfolds the map
	preData, err := maps.RecursiveJSON(change.PreMap)
	if err != nil {
		log.Error(errors.WithMessage(err, "pre data error"))
		return
	}

	// Unfolds the map
	afterData, err := maps.RecursiveJSON(change.SaveMap)
	if err != nil {
		log.Error(errors.WithMessage(err, "after data error"))
		return
	}

	changeLog.BeforeValues = preData.(map[string]any)
	changeLog.AfterValues = afterData.(map[string]any)
	changeLog.Type = change.Table
	if !tools.Empty(change.SavingUser) && !tools.Empty(change.SavingUser.GetString("urn")) {
		changeLog.UserURN = change.SavingUser.GetString("urn")
	} else {
		changeLog.UserURN = "atl:system"
	}
	changeLog.ObjectID = change.ID
	changeLog.ObjectURN = change.URN
	err = changeLog.Save(ctx)
	if err != nil {
		log.Error(errors.WithMessagef(err, "Saving Change Log"))
		_ = worker_jobs.QueueDynamoThrottleRetryJob(TABLE, changeLog)
		return
	}
}
