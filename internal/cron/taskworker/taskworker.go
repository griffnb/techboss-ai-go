package taskworker

import (
	"context"
	"fmt"
	"time"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/queue"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/cron/taskworker/delay_queue"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/pkg/errors"
)

type TaskWorker struct {
	Queue         queue.Queue
	QueueLocation string
	// processing    bool
}

func New(queueObj queue.Queue, queueLocation string) *TaskWorker {
	return &TaskWorker{
		Queue:         queueObj,
		QueueLocation: queueLocation,
	}
}

func (this *TaskWorker) Finish() {
	// nothing to do here
	// Block until processing is complete
	//for this.processing {
	//	time.Sleep(1 * time.Second)
	//}
}

// Process Return true for task workers to let the system know it needs to delete the job
// Task workers only do 1 job at a time
func (this *TaskWorker) Process(sqsMessage *queue.SQSMessage) bool {
	// this.processing = true
	// Build a way to time these out right
	ctx := context.Background()

	jobs, err := sqsMessage.GetRawJobs()
	if err != nil {
		log.Error(err)
		return true
	}

	finished, _ := tools.FinishInTime(ctx, 40*time.Minute, func(ctx context.Context) any {
		for i, job := range jobs {

			if i%10 == 0 {
				log.Debug(fmt.Sprintf("Processed %v", i))
			}

			err := this.ProcessJob(ctx, job)
			if err != nil {
				log.Error(err)

				// dont erase the job in production, try again
				if environment.IsProduction() {
					continue
				}
			}

			// If the job was successful, remove it from the queue, otherwise it'll stay in the queue and attempt again after timeout
			err = finishJob(ctx, job)
			if err != nil {
				log.Error(err)
			}

		}

		return nil
	})

	if !finished {
		log.Error(
			errors.Errorf("Task Was canceled after 40 minutes, %+v", jobs[0].Type),
			jobs[0].Data,
		)
	}

	// this.processing = false
	log.Debug(fmt.Sprintf("Finished Jobs %v", this.QueueLocation))
	return true
}

// Removes it from the delay queue if it was successful

func finishJob(ctx context.Context, job *queue.RawJob) error {
	if tools.Empty(job.DelayQueueID) {
		return nil
	}

	_, err := delay_queue.DeleteDynamoRecord(ctx, job.DelayQueueID)
	if err != nil {
		return err
	}

	return nil
}
