package tag

import (
	"context"
	"sync"

	"github.com/CrowdShield/go-core/lib/cache"
	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/model"
	"github.com/CrowdShield/go-core/lib/types"
)

var (
	cacheInstance *tagCache
	once          sync.Once
)

func Cache() *tagCache {
	once.Do(func() {
		cacheInstance = &tagCache{
			ExpiringCache: cache.NewExpiringCache().
				WithNeverExpire().
				WithLoadFunction("all", func(ctx context.Context, dataCache *cache.ExpiringCache) {
					tags, err := FindAll(ctx, &model.Options{Conditions: "disabled = 0"})
					if err != nil {
						log.Error(err)
						return
					}

					for _, tag := range tags {
						dataCache.Set(string(tag.ID()), tag)
						dataCache.Set(string(tag.Key.Get()), tag)
					}
				}),
		}
		// Load Now
		cacheInstance.RefreshGlobalData()
	})

	return cacheInstance
}

type tagCache struct {
	*cache.ExpiringCache
}

// Lookup returns an organization by v1 internal ID, v2 ID, or v2 URN
func (this *tagCache) GetByID(id types.UUID) (*Tag, error) {
	tag, err := this.Get(string(id))
	if err != nil {
		return nil, err
	}

	if tag == nil {
		return nil, nil
	}

	tagObj, ok := tag.(*Tag)
	if !ok {
		return nil, nil
	}

	return tagObj, nil
}

// Lookup returns an organization by v1 internal ID, v2 ID, or v2 URN
func (this *tagCache) GetByKey(key string) (*Tag, error) {
	tag, err := this.Get(string(key))
	if err != nil {
		return nil, err
	}

	if tag == nil {
		return nil, nil
	}

	tagObj, ok := tag.(*Tag)
	if !ok {
		return nil, nil
	}

	return tagObj, nil
}
