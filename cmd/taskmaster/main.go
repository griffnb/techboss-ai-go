package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/griffnb/techboss-ai-go/internal/cron/taskmaster"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models"
	"github.com/robfig/cron/v3"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/coreos/go-systemd/daemon"
)

func main() {
	environment.CreateEnvironment()

	var crons *cron.Cron

	go func() {
		// wait for everythign to boot
		time.Sleep(60 * time.Second)
		crons = taskmaster.Run()
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

	go func() {
		err := models.RunMigration()
		if err != nil {
			log.Error(err)
			// time.Sleep(5 * time.Second)
			// os.Exit(1)
		}
		err = models.LoadModelsAndValidate()
		if err != nil {
			log.Error(err)
			// time.Sleep(5 * time.Second)
			// os.Exit(1)
		}

		err = models.RunPostMigration()
		if err != nil {
			log.Error(err)
			// time.Sleep(5 * time.Second)
			// os.Exit(1)
		}
	}()
	shutdown := make(chan struct{})
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		log.Debug("Recieved Kill, waiting for shutdown")
		if crons != nil {
			// Stop crons so they can drain
			crons.Stop()
		}
		go func() {
			time.Sleep(25 * time.Second)
			// log.Slack("dev-test", "TaskWorker delayed shutdown for 25 sec")
			// time.Sleep(2 * time.Second)
			close(shutdown)
			os.Exit(0)
		}()

		//_ = httpRouter.Server.Shutdown(context.Background())
		//os.Exit(0)
	}()
	log.Debug("TaskMaster Booted")

	// Wait until shutdown signal completes
	<-shutdown
}
