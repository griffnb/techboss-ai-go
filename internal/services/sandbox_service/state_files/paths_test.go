package state_files_test

import (
	"testing"
	"time"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service/state_files"
)

// Test_GenerateS3Path_validinputs tests generating S3 paths with valid inputs
func Test_GenerateS3Path_validinputs(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		accountID  types.UUID
		timestamp  int64
		expected   string
	}{
		{
			name:       "Standard valid inputs",
			bucketName: "mybucket",
			accountID:  types.UUID("550e8400-e29b-41d4-a716-446655440000"),
			timestamp:  1234567890,
			expected:   "s3://mybucket/docs/550e8400-e29b-41d4-a716-446655440000/1234567890/",
		},
		{
			name:       "Different bucket and account",
			bucketName: "production-bucket",
			accountID:  types.UUID("abc12345-6789-0def-1234-567890abcdef"),
			timestamp:  9876543210,
			expected:   "s3://production-bucket/docs/abc12345-6789-0def-1234-567890abcdef/9876543210/",
		},
		{
			name:       "Zero timestamp",
			bucketName: "testbucket",
			accountID:  types.UUID("00000000-0000-0000-0000-000000000000"),
			timestamp:  0,
			expected:   "s3://testbucket/docs/00000000-0000-0000-0000-000000000000/0/",
		},
		{
			name:       "Negative timestamp",
			bucketName: "negativebucket",
			accountID:  types.UUID("11111111-2222-3333-4444-555555555555"),
			timestamp:  -1234567890,
			expected:   "s3://negativebucket/docs/11111111-2222-3333-4444-555555555555/-1234567890/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := state_files.GenerateS3Path(tt.bucketName, tt.accountID, tt.timestamp)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test_GenerateS3Path_edgecases tests edge cases with empty inputs
func Test_GenerateS3Path_edgecases(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		accountID  types.UUID
		timestamp  int64
		expected   string
	}{
		{
			name:       "Empty bucket name",
			bucketName: "",
			accountID:  types.UUID("550e8400-e29b-41d4-a716-446655440000"),
			timestamp:  1234567890,
			expected:   "s3:///docs/550e8400-e29b-41d4-a716-446655440000/1234567890/",
		},
		{
			name:       "Empty account ID",
			bucketName: "mybucket",
			accountID:  types.UUID(""),
			timestamp:  1234567890,
			expected:   "s3://mybucket/docs//1234567890/",
		},
		{
			name:       "All empty except timestamp",
			bucketName: "",
			accountID:  types.UUID(""),
			timestamp:  9999999999,
			expected:   "s3:///docs//9999999999/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := state_files.GenerateS3Path(tt.bucketName, tt.accountID, tt.timestamp)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test_GenerateS3Path_trailingslash verifies path always ends with trailing slash
func Test_GenerateS3Path_trailingslash(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		accountID  types.UUID
		timestamp  int64
	}{
		{
			name:       "Standard case",
			bucketName: "mybucket",
			accountID:  types.UUID("550e8400-e29b-41d4-a716-446655440000"),
			timestamp:  1234567890,
		},
		{
			name:       "Zero timestamp",
			bucketName: "testbucket",
			accountID:  types.UUID("00000000-0000-0000-0000-000000000000"),
			timestamp:  0,
		},
		{
			name:       "Large timestamp",
			bucketName: "largestamp",
			accountID:  types.UUID("ffffffff-ffff-ffff-ffff-ffffffffffff"),
			timestamp:  9999999999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := state_files.GenerateS3Path(tt.bucketName, tt.accountID, tt.timestamp)

			// Assert - Must end with trailing slash
			lastChar := result[len(result)-1]
			assert.Equal(t, byte('/'), lastChar)
		})
	}
}

// Test_GenerateS3Path_format verifies correct S3 path structure
func Test_GenerateS3Path_format(t *testing.T) {
	t.Run("Path contains all required components", func(t *testing.T) {
		// Arrange
		bucketName := "testbucket"
		accountID := types.UUID("550e8400-e29b-41d4-a716-446655440000")
		timestamp := int64(1234567890)

		// Act
		result := state_files.GenerateS3Path(bucketName, accountID, timestamp)

		// Assert - Check path structure
		expected := "s3://testbucket/docs/550e8400-e29b-41d4-a716-446655440000/1234567890/"
		assert.Equal(t, expected, result)

		// Verify it starts with s3://
		assert.Equal(t, "s3://", result[:5])

		// Verify it contains /docs/
		assert.Contains(t, result, "/docs/")

		// Verify it ends with /
		assert.Equal(t, byte('/'), result[len(result)-1])
	})
}

// Test_GetCurrentTimestamp_nonzero tests that GetCurrentTimestamp returns non-zero value
func Test_GetCurrentTimestamp_nonzero(t *testing.T) {
	t.Run("Returns non-zero Unix timestamp", func(t *testing.T) {
		// Act
		timestamp := state_files.GetCurrentTimestamp()

		// Assert
		assert.NotEqual(t, int64(0), timestamp)
	})
}

// Test_GetCurrentTimestamp_recent tests that GetCurrentTimestamp returns recent time
func Test_GetCurrentTimestamp_recent(t *testing.T) {
	t.Run("Returns timestamp within last minute", func(t *testing.T) {
		// Arrange
		beforeCall := time.Now().Unix()

		// Act
		timestamp := state_files.GetCurrentTimestamp()

		// Arrange (after)
		afterCall := time.Now().Unix()

		// Assert - Timestamp should be between beforeCall and afterCall
		if timestamp < beforeCall || timestamp > afterCall {
			t.Errorf("Timestamp %d is not between %d and %d", timestamp, beforeCall, afterCall)
		}
	})

	t.Run("Returns reasonable timestamp not in distant past or future", func(t *testing.T) {
		// Arrange
		now := time.Now().Unix()
		oneMinuteAgo := now - 60
		oneMinuteFromNow := now + 60

		// Act
		timestamp := state_files.GetCurrentTimestamp()

		// Assert - Should be within 1 minute of now
		if timestamp < oneMinuteAgo || timestamp > oneMinuteFromNow {
			t.Errorf("Timestamp %d is not within 1 minute of now (%d)", timestamp, now)
		}
	})
}

// Test_GetCurrentTimestamp_positive tests that GetCurrentTimestamp is always positive
func Test_GetCurrentTimestamp_positive(t *testing.T) {
	t.Run("Returns positive Unix timestamp", func(t *testing.T) {
		// Act
		timestamp := state_files.GetCurrentTimestamp()

		// Assert
		if timestamp <= 0 {
			t.Errorf("Expected positive timestamp, got %d", timestamp)
		}
	})
}

// Test_GetCurrentTimestamp_consistent tests that consecutive calls return increasing or equal values
func Test_GetCurrentTimestamp_consistent(t *testing.T) {
	t.Run("Consecutive calls return increasing or equal timestamps", func(t *testing.T) {
		// Act
		timestamp1 := state_files.GetCurrentTimestamp()
		timestamp2 := state_files.GetCurrentTimestamp()
		timestamp3 := state_files.GetCurrentTimestamp()

		// Assert - Later calls should be >= earlier calls
		if timestamp2 < timestamp1 {
			t.Errorf("Second timestamp %d is less than first timestamp %d", timestamp2, timestamp1)
		}
		if timestamp3 < timestamp2 {
			t.Errorf("Third timestamp %d is less than second timestamp %d", timestamp3, timestamp2)
		}
	})
}
