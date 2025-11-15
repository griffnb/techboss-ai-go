package object_sync

import (
	"context"
	"fmt"

	"github.com/CrowdShield/go-core/lib/log"
	"github.com/CrowdShield/go-core/lib/model/coremodel"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/CrowdShield/go-core/lib/tools/ptr"
	"github.com/CrowdShield/go-core/lib/tools/set"

	"github.com/griffnb/techboss-ai-go/internal/environment"
	"github.com/griffnb/techboss-ai-go/internal/models/base/caller"
	"github.com/griffnb/techboss-ai-go/internal/task_update"

	"github.com/pkg/errors"
)

var skipTables = []string{
	"admins",
	"change_logs",
}

type TaskUpdate struct {
	Model              string  `json:"model,omitempty"`
	Total              *int    `json:"total,omitempty"`
	Current            *int    `json:"current,omitempty"`
	CurrentModelNumber *int    `json:"current_model_number,omitempty"`
	TotalModels        *int    `json:"total_models,omitempty"`
	Error              *string `json:"error,omitempty"`
}

func FullSync(ctx context.Context, sessionKeyOrEmail string, onlyPackages []string) error {
	if environment.IsProduction() {
		return errors.Errorf("cant full sync in production")
	}

	modelPackageCallers := caller.Registry().GetAll()

	taskUpdater, updatesActive := task_update.GetTaskUpdater[*TaskUpdate](ctx)

	modelList := []string{}

	// These have foreign key constraints so need to go last
	lastObjects := set.New("")
	skipObjects := set.New(skipTables...)

	for modelPackageName := range modelPackageCallers {
		if !lastObjects.Contains(modelPackageName) && !skipObjects.Contains(modelPackageName) {
			modelList = append(modelList, modelPackageName)
		}
	}

	modelList = append(modelList, lastObjects.Values...)

	if !tools.Empty(onlyPackages) {
		modelList = onlyPackages
	}

	if updatesActive {
		taskUpdater.Send(&TaskUpdate{
			TotalModels: ptr.To(len(modelList)),
		})
	}

	for i, modelPackageName := range modelList {
		if updatesActive {
			taskUpdater.Send(&TaskUpdate{
				Model:              modelPackageName,
				CurrentModelNumber: ptr.To(i + 1),
			})
		}

		err := truncateTable(ctx, modelPackageName)
		if err != nil {
			return err
		}

		if err := importModel(ctx, sessionKeyOrEmail, modelPackageName); err != nil {
			return err
		}

	}

	return nil
}

func truncateTable(ctx context.Context, modelPackageName string) error {
	if environment.IsProduction() {
		return errors.Errorf("cant truncate in production")
	}
	registry := caller.Registry().Get(modelPackageName)
	fmt.Printf("Truncating table for model package %s\n", modelPackageName)
	tableName := registry.New().(coremodel.Model).GetTable()

	err := environment.DB().DB.InsertWithContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", tableName), nil)
	if err != nil {
		return errors.Wrapf(err, "failed to truncate table %s", tableName)
	}

	return nil
}

func importModel(ctx context.Context, sessionKeyOrEmail, modelPackageName string) error {
	if environment.IsProduction() {
		return errors.Errorf("cant import in production")
	}

	taskUpdater, updatesActive := task_update.GetTaskUpdater[*TaskUpdate](ctx)
	registry := caller.Registry().Get(modelPackageName)

	var limit int64 = 10000
	var offset int64
	var total int
	for {

		remoteRecords, err := GetAllRemoteRecords(sessionKeyOrEmail, modelPackageName, registry, limit, offset)
		if err != nil {
			return err
		}

		if updatesActive {
			taskUpdater.Send(&TaskUpdate{
				Model:   modelPackageName,
				Total:   ptr.To(total + len(remoteRecords)),
				Current: ptr.To(total),
			})
		}

		if tools.Empty(remoteRecords) {
			log.Debugf("No remote records found for model %s", modelPackageName)
			break
		}

		for i, remoteRecord := range remoteRecords {
			remoteData := remoteRecord.GetDataCopy()
			localObj, err := registry.Get(context.Background(), remoteRecord.ID())
			if err != nil {
				return err
			}
			if tools.Empty(localObj) {
				localObj = registry.New().(coremodel.Model)
			}
			localObj.MergeRawData(remoteData)
			err = localObj.Save(nil)
			if err != nil {
				return errors.Wrapf(err, "failed to save remote record %s", remoteRecord.ID())
			}

			if updatesActive && i%10 == 0 {
				taskUpdater.Send(&TaskUpdate{
					Model:   modelPackageName,
					Current: ptr.To(total + i + 1),
				})
			}
		}

		total += len(remoteRecords)

		if len(remoteRecords) < int(limit) {
			break
		}
		offset += limit

	}

	if updatesActive {
		log.Debugf("Sending Final task update for model %s Total %d", modelPackageName, total)
		taskUpdater.Send(&TaskUpdate{
			Model:   modelPackageName,
			Current: ptr.To(total),
		})
	}

	return nil
}
