package sandbox_service

import (
	"testing"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
)

// Test_GenerateAppName_validUUID tests that GenerateAppName returns correct format with "app-" prefix
func Test_GenerateAppName_validUUID(t *testing.T) {
	tests := []struct {
		name      string
		accountID types.UUID
		expected  string
	}{
		{
			name:      "valid UUID generates app name with app- prefix",
			accountID: types.UUID("550e8400-e29b-41d4-a716-446655440000"),
			expected:  "app-550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:      "another valid UUID generates correct app name",
			accountID: types.UUID("123e4567-e89b-12d3-a456-426614174000"),
			expected:  "app-123e4567-e89b-12d3-a456-426614174000",
		},
		{
			name:      "zero UUID generates app name with zeros",
			accountID: types.UUID("00000000-0000-0000-0000-000000000000"),
			expected:  "app-00000000-0000-0000-0000-000000000000",
		},
		{
			name:      "empty UUID generates app name with empty string",
			accountID: types.UUID(""),
			expected:  "app-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := GenerateAppName(tt.accountID)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test_GenerateVolumeName_validUUID tests that GenerateVolumeName returns correct format with "volume-" prefix
func Test_GenerateVolumeName_validUUID(t *testing.T) {
	tests := []struct {
		name      string
		accountID types.UUID
		expected  string
	}{
		{
			name:      "valid UUID generates volume name with volume- prefix",
			accountID: types.UUID("550e8400-e29b-41d4-a716-446655440000"),
			expected:  "volume-550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:      "another valid UUID generates correct volume name",
			accountID: types.UUID("123e4567-e89b-12d3-a456-426614174000"),
			expected:  "volume-123e4567-e89b-12d3-a456-426614174000",
		},
		{
			name:      "zero UUID generates volume name with zeros",
			accountID: types.UUID("00000000-0000-0000-0000-000000000000"),
			expected:  "volume-00000000-0000-0000-0000-000000000000",
		},
		{
			name:      "empty UUID generates volume name with empty string",
			accountID: types.UUID(""),
			expected:  "volume-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := GenerateVolumeName(tt.accountID)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test_GenerateAppName_consistency tests that the same UUID always generates the same app name
func Test_GenerateAppName_consistency(t *testing.T) {
	t.Run("same UUID generates consistent app name", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("550e8400-e29b-41d4-a716-446655440000")

		// Act
		result1 := GenerateAppName(accountID)
		result2 := GenerateAppName(accountID)

		// Assert
		assert.Equal(t, result1, result2)
		assert.Equal(t, "app-550e8400-e29b-41d4-a716-446655440000", result1)
	})
}

// Test_GenerateVolumeName_consistency tests that the same UUID always generates the same volume name
func Test_GenerateVolumeName_consistency(t *testing.T) {
	t.Run("same UUID generates consistent volume name", func(t *testing.T) {
		// Arrange
		accountID := types.UUID("550e8400-e29b-41d4-a716-446655440000")

		// Act
		result1 := GenerateVolumeName(accountID)
		result2 := GenerateVolumeName(accountID)

		// Assert
		assert.Equal(t, result1, result2)
		assert.Equal(t, "volume-550e8400-e29b-41d4-a716-446655440000", result1)
	})
}
