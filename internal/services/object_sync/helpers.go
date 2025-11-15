package object_sync

import (
	"github.com/griffnb/assettradingdesk-go/lib/model/coremodel"
	"github.com/griffnb/assettradingdesk-go/lib/tools"
	"github.com/griffnb/assettradingdesk-go/lib/types"
)

// getIDsFromMap returns the IDs from a map of models
func getIDsFromMap(m map[types.UUID]coremodel.Model) []types.UUID {
	ids := make([]types.UUID, 0)
	for id := range m {
		ids = append(ids, id)
	}
	return ids
}

func Lookup(objMap map[string]map[types.UUID]coremodel.Model, otherObj coremodel.Model) coremodel.Model {
	packageName := otherObj.GetPackage()
	id := otherObj.ID()
	if objMap == nil {
		return nil
	}
	packageMap, ok := objMap[packageName]
	if !ok {
		return nil
	}
	obj, ok := packageMap[id]
	if !ok {
		return nil
	}

	return obj
}

func RemoveNilFromRecords(records []coremodel.Model) []any {
	cleaned := make([]any, 0)
	for _, record := range records {
		obj, err := tools.StructToMapJSON(record)
		if err != nil {
			continue
		}
		cleaned = append(cleaned, removeNils(obj))
	}
	return cleaned
}

func removeNils(v any) any {
	switch val := v.(type) {
	case map[string]any:
		cleanedMap := make(map[string]any)
		for k, v2 := range val {
			cleaned := removeNils(v2)
			if cleaned != nil {
				cleanedMap[k] = cleaned
			}
		}
		if len(cleanedMap) == 0 {
			return nil
		}
		return cleanedMap
	default:
		if v == nil {
			return nil
		}
		return v
	}
}
