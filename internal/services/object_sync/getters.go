package object_sync

import (
	"net/url"

	"github.com/CrowdShield/go-core/lib/model/coremodel"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/CrowdShield/go-core/lib/tools/slice"
	"github.com/griffnb/techboss-ai-go/internal/models/base/caller"
	"github.com/griffnb/techboss-ai-go/internal/services/system_proxy"
)

func GetAllRemoteRecords(sessionKeyOrEmail, modelType string, factory caller.Caller, limit, offset int64) ([]coremodel.Model, error) {
	// Fetch from remote
	sliceTypePtr := factory.NewSlicePtr()
	_, err := system_proxy.RemoteGetType(sessionKeyOrEmail, modelType, url.Values{
		"limit":  []string{tools.ParseStringI(limit)},
		"offset": []string{tools.ParseStringI(offset)},
	}, sliceTypePtr)
	if err != nil {
		return nil, err
	}

	remoteRecords := make([]coremodel.Model, 0)
	_ = slice.IterateReflect(sliceTypePtr, func(_ int, record any) {
		coreRecord := record.(coremodel.Model)
		remoteRecords = append(remoteRecords, coreRecord)
	})

	return remoteRecords, nil
}
