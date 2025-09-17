package common

import (
	"fmt"
	"strings"

	"github.com/CrowdShield/go-core/lib/model/coremodel"
	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/CrowdShield/go-core/lib/types"
)

func GenerateURN(obj coremodel.Model) {
	key := obj.ID()
	if tools.Empty(key) {
		key = tools.GUID()
	}

	urn := IDToURN(obj.GetTable(), key)

	if obj.IsEmpty("urn") {
		obj.Set("urn", urn)
	}
	if obj.ID() == "" {
		obj.Set("id", key)
	}
}

func IDToURN(table string, id types.UUID) string {
	return strings.ToLower(fmt.Sprintf("boss:%s:%s", table, id))
}

func CheckURNType(urn, tableName string) bool {
	return strings.HasPrefix(urn, fmt.Sprintf("boss:%s:", tableName))
}
