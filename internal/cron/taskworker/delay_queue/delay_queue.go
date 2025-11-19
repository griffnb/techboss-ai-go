package delay_queue

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/griffnb/core/lib/queue"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/pkg/errors"
)

const MAX_RETRIES = 5

type DelayQueueItem struct {
	ID        string `column:"id"        json:"id"`
	Queued    int64  `column:"queued"    json:"queued"`
	Type      string `column:"type"      json:"type"`
	Timestamp int64  `column:"timestamp" json:"timestamp"`
	Data      any    `column:"data"      json:"data"`

	// Static dont set this, used as a hash key because dynamo doesnt allow empty keys
	Static int `column:"_static" json:"_static"`
}

func NewItem(itemType string, delayUntilTS int64, data any) *DelayQueueItem {
	return &DelayQueueItem{
		ID:        tools.SessionKey(),
		Type:      itemType,
		Timestamp: delayUntilTS,
		Data:      data,
	}
}

func (this *DelayQueueItem) Save(ctx context.Context) (err error) {
	if tools.Empty(this.ID) {
		this.ID = tools.SessionKey()
	}
	if tools.Empty(this.Timestamp) {
		this.Timestamp = time.Now().Unix()
	}

	return this.saveWithRetries(ctx, MAX_RETRIES)
}

func (this *DelayQueueItem) SaveWithRetries(ctx context.Context) error {
	return this.saveWithRetries(ctx, MAX_RETRIES)
}

func (this *DelayQueueItem) Delete(ctx context.Context) (bool, error) {
	return DeleteDynamoRecord(ctx, this.ID)
}

func (this *DelayQueueItem) saveWithRetries(ctx context.Context, retries int) error {
	if retries <= 0 {
		bytes, _ := json.Marshal(this)
		return errors.Errorf("max retries exceeded for event %s", string(bytes))
	}

	throttled, err := environment.GetDynamo().PutWithContext(ctx, TABLE_NAME, this)
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

func (this *DelayQueueItem) PushToSQS(_ context.Context) error {
	job := &queue.Job{
		Type:           this.Type,
		Data:           this.Data,
		DelayQueueID:   this.ID,
		DelayedUntilTS: this.Timestamp,
	}
	err := environment.GetQueue().Push(environment.QUEUE_THROTTLES, job)
	if err != nil {
		return err
	}
	return nil
}

func (this *DelayQueueItem) CheckLock(ctx context.Context) (bool, error) {
	success, throttled, err := environment.GetDynamo().IncrementLockFieldWithContext(ctx, TABLE_NAME, "id", this.ID, "queued", 1)
	if err != nil {
		return false, err
	}
	if throttled {
		return false, errors.Errorf("CheckLock throttled")
	}
	return success, err
}
