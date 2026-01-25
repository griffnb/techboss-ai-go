package modal

import "time"

// CompareStateFiles compares local and S3 state files to determine which sync actions are needed.
// It identifies files to download (new or updated), files to skip (unchanged), and files to delete
// (removed from S3). This enables efficient incremental synchronization with minimal data transfer.
//
// Requirements satisfied:
// - 3.4: Files in S3 but not local -> download
// - 3.5: Files with matching checksums -> skip
// - 3.6: Files with different checksums -> download
// - 3.7: Files in local but not S3 -> delete
//
// Nil state files are treated as empty states, allowing comparison with missing state files.
func CompareStateFiles(localState *StateFile, s3State *StateFile) *StateDiff {
	// Initialize result
	diff := &StateDiff{
		FilesToDownload: []FileEntry{},
		FilesToDelete:   []string{},
		FilesToSkip:     []FileEntry{},
	}

	// Create maps for O(1) lookups
	localFiles := make(map[string]FileEntry)
	s3Files := make(map[string]FileEntry)

	// Build local file map
	if localState != nil {
		for _, f := range localState.Files {
			localFiles[f.Path] = f
		}
	}

	// Build S3 file map
	if s3State != nil {
		for _, f := range s3State.Files {
			s3Files[f.Path] = f
		}
	}

	// Process S3 files: determine if download or skip
	for path, s3File := range s3Files {
		localFile, existsLocal := localFiles[path]

		if existsLocal {
			// File exists in both states - compare checksums
			if localFile.Checksum == s3File.Checksum {
				// Matching checksum -> skip (Requirement 3.5)
				diff.FilesToSkip = append(diff.FilesToSkip, s3File)
			} else {
				// Different checksum -> download (Requirement 3.6)
				diff.FilesToDownload = append(diff.FilesToDownload, s3File)
			}
		} else {
			// File only in S3 -> download (Requirement 3.4)
			diff.FilesToDownload = append(diff.FilesToDownload, s3File)
		}
	}

	// Process local files: identify files to delete
	for path := range localFiles {
		if _, existsS3 := s3Files[path]; !existsS3 {
			// File only in local -> delete (Requirement 3.7)
			diff.FilesToDelete = append(diff.FilesToDelete, path)
		}
	}

	return diff
}

// CheckIfStale determines if a state file is stale based on the given threshold in seconds.
// A state file is considered stale if:
// - The state file is nil (never synced)
// - LastSyncedAt is 0 (never synced)
// - The age (current time - LastSyncedAt) exceeds the threshold
//
// This is used to determine if a cold start sync is needed based on the configured
// SyncStaleThreshold (default: 1 hour / 3600 seconds).
func CheckIfStale(stateFile *StateFile, thresholdSeconds int64) bool {
	// Nil state or zero LastSyncedAt means never synced -> stale
	if stateFile == nil || stateFile.LastSyncedAt == 0 {
		return true
	}

	// Calculate age in seconds
	age := time.Now().Unix() - stateFile.LastSyncedAt

	// Stale if age exceeds threshold
	return age > thresholdSeconds
}
