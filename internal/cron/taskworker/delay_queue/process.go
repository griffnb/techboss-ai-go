package delay_queue

import (
	"context"
	"time"

	"github.com/CrowdShield/go-core/lib/dynamo"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/pkg/errors"
)

const LIMIT = 10

func RunDelayQueue(ctx context.Context) error {
	items, err := PopFromDynamo(ctx, LIMIT)
	if err != nil {
		return err
	}

	for _, item := range items {
		success, err := item.CheckLock(ctx)
		if err != nil {
			return err
		}
		if success {
			err = item.PushToSQS(ctx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func PopFromDynamo(ctx context.Context, limit int64) ([]*DelayQueueItem, error) {
	out := []*DelayQueueItem{}
	query := &dynamo.DynamoQuery{
		IndexName:        "timestamp-index",
		Limit:            limit,
		KeyCondition:     "#_static = :zero AND #timestamp < :timestamp",
		FilterExpression: "#queued = :zero",
		AttributeMap: map[string]string{
			"#timestamp": "timestamp",
			"#queued":    "queued",
			"#_static":   "_static",
		},
		Params: map[string]interface{}{
			":timestamp": time.Now().Unix(),
			":zero":      0,
		},
	}
	throttled, err := environment.GetDynamo().Query(ctx, TABLE_NAME, query, &out)
	if err != nil {
		return nil, err
	}

	if throttled {
		return nil, errors.Errorf("PopFromQueue throttled")
	}

	return out, nil
}

func DeleteDynamoRecord(_ context.Context, id string) (bool, error) {
	return environment.GetDynamo().Delete(TABLE_NAME, "id", id)
}

func GetItem(ctx context.Context, id string) (*DelayQueueItem, error) {
	item := &DelayQueueItem{}
	throttled, err := environment.GetDynamo().GetWithContext(ctx, TABLE_NAME, "id", id, item, true)
	if err != nil {
		return nil, err
	}
	if throttled {
		return nil, errors.Errorf("event throttled by key")
	}
	return item, nil
}
