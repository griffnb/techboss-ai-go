package main

import (
	"os"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models"
	"github.com/pkg/errors"
)

func main() {
	env := environment.CreateEnvironment()

	if env.Environment != "unit_test" {
		log.Error(errors.Errorf("Rebuild Unit Database can only be run in the 'unit_test' environment"))

		os.Exit(1)
	}

	err := environment.DB().DB.Insert(
		`DROP SCHEMA public CASCADE;
			CREATE SCHEMA public;`,

		nil)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	err = models.RunMigration()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	log.Debug("Finished Migrations")
	err = models.LoadModelsAndValidate()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	err = models.RunPostMigration()
	if err != nil {
		log.Error(err)

		os.Exit(1)
	}

	log.Debug("Finished")
}
