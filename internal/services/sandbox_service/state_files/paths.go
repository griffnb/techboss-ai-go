package state_files

import (
	"fmt"
	"time"

	"github.com/griffnb/core/lib/types"
)

// GenerateS3Path generates an S3 path for storing account-scoped sandbox state.
// Format: s3://{bucket}/docs/{accountID}/{timestamp}/
func GenerateS3Path(bucketName string, accountID types.UUID, timestamp int64) string {
	return fmt.Sprintf("s3://%s/docs/%s/%d/", bucketName, accountID.String(), timestamp)
}

// GetCurrentTimestamp returns the current Unix timestamp (seconds since epoch).
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}
