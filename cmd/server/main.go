package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coreos/go-systemd/daemon"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/router"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/CrowdShield/go-core/lib/workers/puller"
	"github.com/CrowdShield/go-core/lib/workers/worker"
	"github.com/griffnb/techboss-ai-go/internal/controllers"
	"github.com/griffnb/techboss-ai-go/internal/cron/taskmaster"
	"github.com/griffnb/techboss-ai-go/internal/cron/taskworker"
	"github.com/griffnb/techboss-ai-go/internal/environment"

	"github.com/griffnb/techboss-ai-go/internal/models"
)

func main() {
	env := environment.CreateEnvironment()
	sysConfigObj := environment.GetConfig()

	dispatchers := make([]*worker.QueueDispatcher, 0)
	shutdown := make(chan struct{})

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

	port := sysConfigObj.Server.Port

	httpRouter := router.Setup(port, sysConfigObj.Server.SessionKey, sysConfigObj.Server.Cors)
	httpRouter.Router.Use(middleware.NoCache)

	httpRouter.SetupBasics()
	if !env.IsLocalDev() {
		httpRouter.LogType = router.LOG_TYPE_FULL
	} else {
		httpRouter.LogType = router.LOG_TYPE_PATH
	}
	httpRouter.SetLogger(env.RequestLog)

	httpRouter.SetupNoLogPaths([]string{
		"/p/",
	})

	// Controllers
	controllers.Setup(httpRouter)

	noMigrate := os.Getenv("NO_MIGRATE")
	// These get run on the taskmaster now, and on the server only in dev/stage
	if (!tools.Empty(sysConfigObj.Server.Migrate) || !env.IsProduction()) && tools.Empty(noMigrate) {
		log.Debug("Running Migrations")
		err := models.RunMigration()
		if err != nil {
			log.Error(err)
			// time.Sleep(5 * time.Second)
			// os.Exit(1)
		}
		log.Debug("Finished Migrations")
	}

	err := models.LoadModelsAndValidate()
	if err != nil {
		log.Error(err)
		// time.Sleep(5 * time.Second)
		// os.Exit(1)
	}
	// These get run on the taskmaster now, and on the server only in dev/stage
	if !tools.Empty(sysConfigObj.Server.Migrate) || !env.IsProduction() {
		err = models.RunPostMigration()
		if err != nil {
			log.Error(err)
			// time.Sleep(2 * time.Second)
			// os.Exit(1)
		}
	}

	//-----------------

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		log.Debug("Received Kill, waiting for shutdown")

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

		_ = httpRouter.Server.Shutdown(context.Background())
	}()
	/*
		if sEnv.IsProduction() {
			taskPullers := 1
			taskWorkers := 5

			go func() {
				// wait for everythign to boot
				time.Sleep(60 * time.Second)

				if senv.GetConfig().SQS.Hold {
					log.Debug("SQS Hold is set, not starting task workers")
					log.Slack("dev-test", "SQS Hold is set, not starting task workers")
					return
				}

				log.Slack("dev-test", "go-litigate taskworker booting up")
				for key := range sysConfigObj.SQS.Queues {
					taskWorker := taskworker.New(sEnv.Queue, key)
					taskDispatcher := worker.NewQueueDispatcher(
						sEnv.Queue,
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

			// if SQS is enabled and then local endpoint is set, then its local mode
		} else
	*/
	if !tools.Empty(environment.GetConfig().SQS.LocalEndpoint) {
		go func() {
			// wait for everythign to boot

			taskmaster.Run()

			for key := range sysConfigObj.SQS.Queues {
				taskWorker := taskworker.New(env.Queue, key)
				taskDispatcher := worker.NewQueueDispatcher(
					env.Queue,
					1,
					1,
					key,
					taskWorker,
					puller.SQSTASK,
				)
				taskDispatcher.SetLocalDev()
				taskDispatcher.Run()
			}

			log.Debug("Launched local task workers")
		}()
	}

	log.Debug(fmt.Sprintf("Server booted on %v", port))
	httpRouter.StartLong()
}
