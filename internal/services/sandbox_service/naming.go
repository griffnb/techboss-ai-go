package sandbox_service

import (
	"fmt"

	"github.com/griffnb/core/lib/types"
)

// GenerateAppName generates a Modal app name scoped to an account.
// Format: "app-{accountID}"
func GenerateAppName(accountID types.UUID) string {
	return fmt.Sprintf("app-%s", accountID.String())
}

// GenerateVolumeName generates a Modal volume name scoped to an account.
// Format: "volume-{accountID}"
func GenerateVolumeName(accountID types.UUID) string {
	return fmt.Sprintf("volume-%s", accountID.String())
}
