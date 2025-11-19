package system_testing

import (
	senv "github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models"
	"github.com/griffnb/techboss-ai-go/internal/models/migrations"

	"github.com/griffnb/core/lib/log"
)

const CHROME_USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

func BuildSystem() {
	senv.CreateEnvironment()
	err := models.LoadModels()
	if err != nil {
		log.Error(err)
	}

	// using migrations instead of building straight table
	_ = models.RunMigration()
	// senv.GetDBClient(senv.CLIENT_DEFAULT).BuildTablesIgnoreErrors()
	migrations.BuildDynamo()
	err = senv.GetDBClient(senv.CLIENT_DEFAULT).ValidateMigrationsAgainstModels()
	if err != nil {
		log.Error(err)
	}
}
