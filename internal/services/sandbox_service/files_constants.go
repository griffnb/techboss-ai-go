package sandbox_service

import "time"

// File operation constants define limits and timeouts for sandbox file operations.
// These constants ensure consistent behavior and prevent resource exhaustion.
const (
	// MaxFilesPerRequest is the maximum number of files that can be returned in a single paginated request.
	// This prevents memory exhaustion when dealing with large file listings.
	MaxFilesPerRequest = 1000

	// MaxFileContentSize is the maximum file size that can be read in bytes (100MB).
	// Files larger than this limit will have their content truncated to this size.
	MaxFileContentSize = 100 * 1024 * 1024

	// CommandTimeout is the maximum time to wait for a sandbox command to complete.
	// Commands exceeding this timeout will be cancelled to prevent hanging operations.
	CommandTimeout = 30 * time.Second

	// DefaultPageSize is the default number of items per page for paginated file listings.
	// This provides a reasonable balance between response size and number of requests.
	DefaultPageSize = 100

	// MaxRetryAttempts is the number of times to retry a failed command execution.
	// Retries use exponential backoff: 1s, 2s, 4s between attempts.
	MaxRetryAttempts = 3
)
