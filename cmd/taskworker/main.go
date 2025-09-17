package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/griffnb/techboss-ai-go/internal/cron/taskworker"
	"github.com/griffnb/techboss-ai-go/internal/environment"

	"github.com/griffnb/techboss-ai-go/internal/models"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/CrowdShield/go-core/lib/workers/puller"
	"github.com/CrowdShield/go-core/lib/workers/worker"
	"github.com/coreos/go-systemd/daemon"
)

func main() {
	env := environment.CreateEnvironment()
	sysConfigObj := environment.GetConfig()

	// Amount of processors

	taskPullers := 1
	taskWorkers := 5

	dispatchers := make([]*worker.QueueDispatcher, 0)

	go func() {
		// wait for everythign to boot
		time.Sleep(60 * time.Second)

		if sysConfigObj.SQS.Hold {
			log.Debug("SQS Hold is set, not starting task workers")
			log.Slack("dev-test", "SQS Hold is set, not starting task workers")
			return
		}

		for key := range sysConfigObj.SQS.Queues {
			taskWorker := taskworker.New(env.Queue, key)
			taskDispatcher := worker.NewQueueDispatcher(
				env.Queue,
				taskWorkers,
				taskPullers,
				key,
				taskWorker,
				puller.SQSTASK,
			)
			taskDispatcher.Run()
			dispatchers = append(dispatchers, taskDispatcher)
		}
	}()

	// Simple watchdog notifier
	go func() {
		interval, err := daemon.SdWatchdogEnabled(false)
		if err != nil || interval == 0 {
			return
		}
		for {
			_, _ = daemon.SdNotify(false, "WATCHDOG=1")
			time.Sleep(interval / 3)
		}
	}()
	_, _ = daemon.SdNotify(false, "READY=1")

	err := models.LoadModelsAndValidate()
	if err != nil {
		log.Error(err)
		// time.Sleep(5 * time.Second)
		// os.Exit(1)
	}
	//-----------------

	shutdown := make(chan struct{})

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		// log.Debug("Recieved Kill, waiting for shutdown")
		// log.Slack("dev-test", "TaskWorker shutting down")

		// Stop the dispatchers so they workers can drain
		if !tools.Empty(dispatchers) {
			for _, dispatcher := range dispatchers {
				dispatcher.Stop()
			}
		}

		go func() {
			time.Sleep(25 * time.Second)
			// log.Slack("dev-test", "TaskWorker delayed shutdown for 25 sec")
			// time.Sleep(2 * time.Second)
			close(shutdown)
			os.Exit(0)
		}()
	}()

	// Wait until shutdown signal completes
	<-shutdown
	log.Debug("TaskWorker exited cleanly")
}
