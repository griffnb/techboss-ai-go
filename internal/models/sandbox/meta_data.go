package sandbox

import (
	"time"
)

// MetaData stores minimal sandbox metadata in JSONB format
// Most data is now stored in dedicated columns (external_id, status, provider, agent_id)
// All fields use snake_case for JSON marshaling
type MetaData struct {
	LastS3Sync *int64     `json:"last_s3_sync"` // Last S3 sync unix timestamp (nullable)
	SyncStats  *SyncStats `json:"sync_stats"`   // Last sync statistics (nullable)
}

// SyncStats stores the results of the last S3 sync operation.
// Updated per design phase 3.1 to include detailed sync metrics:
// files downloaded, deleted, and skipped for state-file-based synchronization.
type SyncStats struct {
	FilesDownloaded  int   `json:"files_downloaded"`
	FilesDeleted     int   `json:"files_deleted"`
	FilesSkipped     int   `json:"files_skipped"`
	BytesTransferred int64 `json:"bytes_transferred"`
	DurationMs       int64 `json:"duration_ms"`
}

// UpdateLastSync updates the last sync timestamp and stats
func (m *MetaData) UpdateLastSync(filesDownloaded, filesDeleted, filesSkipped int, bytesTransferred int64, durationMs int64) {
	now := time.Now().Unix()
	m.LastS3Sync = &now
	m.SyncStats = &SyncStats{
		FilesDownloaded:  filesDownloaded,
		FilesDeleted:     filesDeleted,
		FilesSkipped:     filesSkipped,
		BytesTransferred: bytesTransferred,
		DurationMs:       durationMs,
	}
}
