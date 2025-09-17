package common

import (
	"regexp"

	"github.com/CrowdShield/go-core/lib/model/coremodel"
	"github.com/griffnb/techboss-ai-go/internal/constants"
)

// SetDisabledDeleted sets disabled/deleted flags
func SetDisabledDeleted(obj coremodel.Model) {
	if !obj.IsEmpty("status") {
		status := obj.GetInt("status")
		if status >= 100 && status < 200 {
			if obj.GetInt("disabled") == 1 {
				obj.Set("disabled", 0)
			}
			if obj.GetInt("deleted") == 1 {
				obj.Set("deleted", 0)
			}
		} else if status >= 200 && status < 300 {
			if obj.GetInt("disabled") == 0 {
				obj.Set("disabled", 1)
			}
			if obj.GetInt("deleted") == 1 {
				obj.Set("deleted", 0)
			}
		} else if status >= 300 && status < 400 {
			if obj.GetInt("disabled") == 0 {
				obj.Set("disabled", 1)
			}
			if obj.GetInt("deleted") == 0 {
				obj.Set("deleted", 1)
			}
		}
	} else {
		obj.Set("status", constants.STATUS_ACTIVE)
	}
}

func SanitizeString(inputString string) string {
	// Compile the regex pattern that matches anything not a letter, number, or underscore
	regex := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	// Replace occurrences of the pattern with an empty string
	return regex.ReplaceAllString(inputString, "")
}
