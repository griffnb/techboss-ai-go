package main

import (
	"os"
	"time"

	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models"

	"github.com/griffnb/core/lib/log"
)

func main() {
	environment.CreateEnvironment()

	err := models.RunMigration()
	if err != nil {
		log.Error(err)
		time.Sleep(2 * time.Second)
		os.Exit(1)
	}
	log.Debug("Finished Migrations")
	err = models.LoadModelsAndValidate()
	if err != nil {
		log.Error(err)
		time.Sleep(2 * time.Second)
		os.Exit(1)
	}

	err = models.RunPostMigration()
	if err != nil {
		log.Error(err)
		time.Sleep(2 * time.Second)
		os.Exit(1)
	}

	log.Debug("Finished")
}
