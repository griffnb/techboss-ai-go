package helpers

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/pkg/errors"
)

// TODO activate when ready
func CheckPayloadHash(payload map[string]interface{}, hash string) (bool, error) {
	sortedJson, err := sortedJsonString(payload)
	if err != nil {
		return false, err
	}

	currentTime := getCurrentTimeRounded()
	if hash == tools.Sha256(fmt.Sprintf("%s%d", sortedJson, currentTime)) {
		return true, nil
	}

	timeWindowMinutes := 5

	// Iterate over the time window, checking for +X and -X minutes
	for i := 1; i <= timeWindowMinutes; i++ {
		// +i minutes
		if hash == tools.Sha256(fmt.Sprintf("%s%d", sortedJson, currentTime+int64(i*60))) {
			return true, nil
		}
		// -i minutes
		if hash == tools.Sha256(fmt.Sprintf("%s%d", sortedJson, currentTime-int64(i*60))) {
			return true, nil
		}
	}

	return false, nil
}

func sortedJsonString(data map[string]interface{}) (string, error) {
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sortedMap := make(map[string]interface{})
	for _, k := range keys {
		sortedMap[k] = data[k]
	}

	jsonBytes, err := json.Marshal(sortedMap)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return string(jsonBytes), nil
}

func getCurrentTimeRounded() int64 {
	now := time.Now()
	rounded := now.Truncate(time.Minute).Unix()
	return rounded
}

/*
func AuthenticateAPIKey(req *http.Request) (*account.Account, error) {
	apiKey := req.Header.Get("Bearer")
	if apiKey == "" {
		return nil, fmt.Errorf("API Key not found")
	}

	acc, err := account.GetByAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	return acc, nil

}
*/
