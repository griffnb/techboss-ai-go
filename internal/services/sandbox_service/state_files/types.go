package state_files

import "time"

const (
	// StateFileName is the name of the state file stored in volumes and S3
	StateFileName = ".sandbox-state"
	// StateFileVersion is the current state file schema version
	StateFileVersion = "1.0"
)

// StateFile represents the .sandbox-state file format used to track file synchronization
// state between local volumes and S3. This file acts like .git for maintaining perfect
// sync state, enabling efficient incremental synchronization and detecting deleted files.
type StateFile struct {
	Version      string      `json:"version"`        // Schema version (e.g., "1.0")
	LastSyncedAt int64       `json:"last_synced_at"` // Unix timestamp of last sync
	Files        []FileEntry `json:"files"`          // Array of tracked files
}

// FileEntry represents a single file tracked in the state file.
// Each entry contains the information needed to determine if the file has changed
// by comparing checksums, sizes, and modification times between local and S3 versions.
type FileEntry struct {
	Path       string `json:"path"`        // Relative path from volume root
	Checksum   string `json:"checksum"`    // MD5 hash (matches S3 ETag format)
	Size       int64  `json:"size"`        // File size in bytes
	ModifiedAt int64  `json:"modified_at"` // Unix timestamp of last modification
}

// StateDiff represents the differences between local and S3 state files.
// It identifies which files need to be downloaded from S3, which local files
// should be deleted to maintain sync, and which files are unchanged and can be skipped.
type StateDiff struct {
	FilesToDownload []FileEntry // Files to download from S3 (new or updated)
	FilesToDelete   []string    // Local file paths to delete (removed from S3)
	FilesToSkip     []FileEntry // Files that match (no action needed)
}

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
