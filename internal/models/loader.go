package models

import (
	"sync"

	"github.com/CrowdShield/go-core/lib/model"
	"github.com/griffnb/techboss-ai-go/internal/cron/taskworker/delay_queue"
	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/admin"
	"github.com/griffnb/techboss-ai-go/internal/models/ai_tool"
	"github.com/griffnb/techboss-ai-go/internal/models/change_log"
	"github.com/griffnb/techboss-ai-go/internal/models/global_config"
	"github.com/griffnb/techboss-ai-go/internal/models/migrations"

	"github.com/pkg/errors"
)

var (
	modelsLoaded = false
	mx           sync.RWMutex
)

func setModelsLoaded() {
	mx.Lock()
	defer mx.Unlock()
	modelsLoaded = true
}

func getModelsLoaded() bool {
	mx.RLock()
	defer mx.RUnlock()
	return modelsLoaded
}

// LoadModels loads the models into the table properties
func LoadModels() (err error) {
	defaultClient := environment.GetDBClient(environment.CLIENT_DEFAULT)

	err = defaultClient.AddTableToProperties(global_config.TABLE, &global_config.Structure{})
	if err != nil {
		return err
	}

	err = defaultClient.AddTableToProperties(admin.TABLE, &admin.Structure{})
	if err != nil {
		return err
	}

	err = defaultClient.AddTableToProperties(ai_tool.TABLE, &ai_tool.Structure{})
	if err != nil {
		return err
	}

	return nil
}

func LoadModelsAndValidate() error {
	err := LoadModels()
	if err != nil {
		return err
	}

	err = environment.GetDBClient(environment.CLIENT_DEFAULT).ValidateMigrationsAgainstModels()
	if err != nil {
		return err
	}

	setModelsLoaded()

	setupChangeLogs()

	return nil
}

func LoadModelsOnly() error {
	err := LoadModels()
	if err != nil {
		return err
	}
	setModelsLoaded()
	setupChangeLogs()

	return nil
}

// RunMigration runs the migrations
func RunMigration() error {
	migrations.BuildDynamo()
	delay_queue.AddDelayQueueTable()

	return environment.GetDBClient(environment.CLIENT_DEFAULT).MigrateUp()
}

func RunPostMigration() error {
	if !getModelsLoaded() {
		return errors.Errorf("Models must be loaded before running post migration")
	}

	err := environment.GetDBClient(environment.CLIENT_DEFAULT).PostMigrateUp()
	if err != nil {
		return err
	}
	return nil
}

func setupChangeLogs() {
	//Database Change Log
	/*
		model.RegisterChangeLogFunction(change_log.DBChangeLog)
		err = defaultClient.AddTableToProperties(change_log.TABLE, &change_log.ChangeLog{})
		if err != nil {
			return err
		}
	*/

	// Dynamo Change Logs
	change_log.AddChangeLogTable()
	model.RegisterChangeLogFunction(change_log.DynamoChangeLog)
}
