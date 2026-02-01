package modal

import "time"

// SyncStats tracks the results of a sync operation between local volume and S3.
// It provides metrics for monitoring, billing, and debugging sync performance.
// Non-fatal errors are collected in the Errors slice to allow partial sync success.
type SyncStats struct {
	FilesDownloaded  int           // Number of files downloaded from S3
	FilesDeleted     int           // Number of local files deleted
	FilesSkipped     int           // Number of files unchanged (skipped)
	BytesTransferred int64         // Total bytes transferred during sync
	Duration         time.Duration // Total operation duration
	Errors           []error       // Non-fatal errors encountered during sync
}
