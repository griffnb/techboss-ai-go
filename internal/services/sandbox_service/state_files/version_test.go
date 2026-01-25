package state_files

import (
	"context"
	"testing"

	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/pkg/errors"
)

func Test_GetLatestVersion_success(t *testing.T) {
	t.Run("returns latest version when versions exist", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		accountID := types.UUID("test-account-123")
		bucketName := "test-bucket"
		expectedVersion := int64(1706140800)

		mockClient := &modal.MockAPIClient{
			GetLatestVersionFunc: func(_ context.Context, accID types.UUID, bucket string) (int64, error) {
				assert.Equal(t, accountID, accID)
				assert.Equal(t, bucketName, bucket)
				return expectedVersion, nil
			},
		}

		// Act
		result, err := GetLatestVersion(ctx, mockClient, accountID, bucketName)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedVersion, result)
	})

	t.Run("returns zero when no versions exist", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		accountID := types.UUID("test-account-456")
		bucketName := "test-bucket"

		mockClient := &modal.MockAPIClient{
			GetLatestVersionFunc: func(_ context.Context, _ types.UUID, _ string) (int64, error) {
				return 0, nil
			},
		}

		// Act
		result, err := GetLatestVersion(ctx, mockClient, accountID, bucketName)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, int64(0), result)
	})

	t.Run("returns highest timestamp when multiple versions exist", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		accountID := types.UUID("test-account-789")
		bucketName := "test-bucket"
		highestTimestamp := int64(1706227200) // Latest

		mockClient := &modal.MockAPIClient{
			GetLatestVersionFunc: func(_ context.Context, _ types.UUID, _ string) (int64, error) {
				return highestTimestamp, nil
			},
		}

		// Act
		result, err := GetLatestVersion(ctx, mockClient, accountID, bucketName)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, highestTimestamp, result)
	})
}

func Test_GetLatestVersion_validationErrors(t *testing.T) {
	t.Run("returns error when client is nil", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		accountID := types.UUID("test-account-nil")
		bucketName := "test-bucket"

		// Act
		result, err := GetLatestVersion(ctx, nil, accountID, bucketName)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client cannot be nil")
		assert.Equal(t, int64(0), result)
	})

	t.Run("returns error when accountID is empty", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		accountID := types.UUID("")
		bucketName := "test-bucket"

		mockClient := &modal.MockAPIClient{}

		// Act
		result, err := GetLatestVersion(ctx, mockClient, accountID, bucketName)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "accountID cannot be empty")
		assert.Equal(t, int64(0), result)
	})

	t.Run("returns error when bucketName is empty", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		accountID := types.UUID("test-account-empty-bucket")
		bucketName := ""

		mockClient := &modal.MockAPIClient{}

		// Act
		result, err := GetLatestVersion(ctx, mockClient, accountID, bucketName)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bucketName cannot be empty")
		assert.Equal(t, int64(0), result)
	})
}

func Test_GetLatestVersion_modalClientErrors(t *testing.T) {
	t.Run("returns error when modal client fails", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		accountID := types.UUID("test-account-fail")
		bucketName := "test-bucket"
		expectedErr := errors.New("failed to list S3 prefixes")

		mockClient := &modal.MockAPIClient{
			GetLatestVersionFunc: func(_ context.Context, _ types.UUID, _ string) (int64, error) {
				return 0, expectedErr
			},
		}

		// Act
		result, err := GetLatestVersion(ctx, mockClient, accountID, bucketName)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get latest version")
		assert.Equal(t, int64(0), result)
	})

	t.Run("returns error when modal client returns S3 access denied", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		accountID := types.UUID("test-account-denied")
		bucketName := "test-bucket"
		expectedErr := errors.New("access denied")

		mockClient := &modal.MockAPIClient{
			GetLatestVersionFunc: func(_ context.Context, _ types.UUID, _ string) (int64, error) {
				return 0, expectedErr
			},
		}

		// Act
		result, err := GetLatestVersion(ctx, mockClient, accountID, bucketName)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get latest version")
		assert.Equal(t, int64(0), result)
	})

	t.Run("returns error when modal client times out", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		accountID := types.UUID("test-account-timeout")
		bucketName := "test-bucket"
		expectedErr := errors.New("context deadline exceeded")

		mockClient := &modal.MockAPIClient{
			GetLatestVersionFunc: func(_ context.Context, _ types.UUID, _ string) (int64, error) {
				return 0, expectedErr
			},
		}

		// Act
		result, err := GetLatestVersion(ctx, mockClient, accountID, bucketName)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get latest version")
		assert.Equal(t, int64(0), result)
	})
}
