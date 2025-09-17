package taskmaster

import (
	"context"
	"time"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/cron/taskworker/delay_queue"
	"github.com/robfig/cron/v3"
)

var QUEUES = []string{"priority1", "priority2", "priority3"}

func Run() *cron.Cron {
	loc, err := time.LoadLocation(constants.DEFAULT_LOCATION)
	if err != nil {
		log.Error(err)
		return nil
	}

	c := cron.New(cron.WithLocation(loc))

	// hourly
	_, _ = c.AddFunc("0 * * * *", func() {
		go func() {
			// Run delay queue hourly for now
			err := delay_queue.RunDelayQueue(context.Background())
			if err != nil {
				log.Error(err)
			}
		}()
	})
	// daily
	_, _ = c.AddFunc("0 1 * * *", func() {
	})

	// weekly
	_, _ = c.AddFunc("0 1 * * 1", func() { // monday hour 1
	})

	// Monthly
	_, _ = c.AddFunc("0 0 1 * *", func() {
	})

	c.Start()

	return c
}
