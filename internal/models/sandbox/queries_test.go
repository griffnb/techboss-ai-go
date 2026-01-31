package sandbox_test

import (
	"context"
	"testing"

	"github.com/griffnb/core/lib/testtools"
	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
)

func init() {
	system_testing.BuildSystem()
}

func Test_FindByExternalID(t *testing.T) {
	tests := []struct {
		name      string
		setupFn   func(t *testing.T) (externalID string, accountID types.UUID, sandboxObj *sandbox.Sandbox, testAccount *account.Account)
		expectErr bool
		expectNil bool
	}{
		{
			name: "finds sandbox with correct account and external_id",
			setupFn: func(t *testing.T) (string, types.UUID, *sandbox.Sandbox, *account.Account) {
				testAccount := account.TESTCreateAccount()
				err := testAccount.Save(nil)
				assert.NoError(t, err)

				externalID := "sb-test-" + tools.RandString(8)
				obj := sandbox.New()
				obj.AccountID.Set(testAccount.ID())
				obj.ExternalID.Set(externalID)
				obj.Type.Set(sandbox.TYPE_CLAUDE_CODE)
				obj.Status.Set(constants.STATUS_ACTIVE)
				obj.MetaData.Set(&sandbox.MetaData{})

				err = obj.Save(nil)
				assert.NoError(t, err)

				return externalID, testAccount.ID(), obj, testAccount
			},
			expectErr: false,
			expectNil: false,
		},
		{
			name: "returns nil when sandbox not found",
			setupFn: func(t *testing.T) (string, types.UUID, *sandbox.Sandbox, *account.Account) {
				testAccount := account.TESTCreateAccount()
				err := testAccount.Save(nil)
				assert.NoError(t, err)

				return "sb-nonexistent", testAccount.ID(), nil, testAccount
			},
			expectErr: false, // Not found is not an error in this codebase
			expectNil: true,
		},
		{
			name: "returns nil with wrong account_id (ownership check)",
			setupFn: func(t *testing.T) (string, types.UUID, *sandbox.Sandbox, *account.Account) {
				correctAccount := account.TESTCreateAccount()
				err := correctAccount.Save(nil)
				assert.NoError(t, err)

				wrongAccount := account.TESTCreateAccount()
				err = wrongAccount.Save(nil)
				assert.NoError(t, err)

				externalID := "sb-test-" + tools.RandString(8)
				obj := sandbox.New()
				obj.AccountID.Set(correctAccount.ID())
				obj.ExternalID.Set(externalID)
				obj.Type.Set(sandbox.TYPE_CLAUDE_CODE)
				obj.Status.Set(constants.STATUS_ACTIVE)
				obj.MetaData.Set(&sandbox.MetaData{})

				err = obj.Save(nil)
				assert.NoError(t, err)

				return externalID, wrongAccount.ID(), obj, correctAccount
			},
			expectErr: false, // Ownership check returns nil, not error
			expectNil: true,
		},
		{
			name: "excludes deleted sandboxes",
			setupFn: func(t *testing.T) (string, types.UUID, *sandbox.Sandbox, *account.Account) {
				testAccount := account.TESTCreateAccount()
				err := testAccount.Save(nil)
				assert.NoError(t, err)

				externalID := "sb-test-" + tools.RandString(8)
				obj := sandbox.New()
				obj.AccountID.Set(testAccount.ID())
				obj.ExternalID.Set(externalID)
				obj.Type.Set(sandbox.TYPE_CLAUDE_CODE)
				obj.Status.Set(constants.STATUS_DELETED) // Use STATUS_DELETED which will set deleted=1
				obj.MetaData.Set(&sandbox.MetaData{})

				err = obj.Save(nil)
				assert.NoError(t, err)

				return externalID, testAccount.ID(), obj, testAccount
			},
			expectErr: false, // Deleted sandbox not found returns nil
			expectNil: true,
		},
		{
			name: "excludes disabled sandboxes",
			setupFn: func(t *testing.T) (string, types.UUID, *sandbox.Sandbox, *account.Account) {
				testAccount := account.TESTCreateAccount()
				err := testAccount.Save(nil)
				assert.NoError(t, err)

				externalID := "sb-test-" + tools.RandString(8)
				obj := sandbox.New()
				obj.AccountID.Set(testAccount.ID())
				obj.ExternalID.Set(externalID)
				obj.Type.Set(sandbox.TYPE_CLAUDE_CODE)
				obj.Status.Set(constants.STATUS_DISABLED) // Use STATUS_DISABLED which will set disabled=1
				obj.MetaData.Set(&sandbox.MetaData{})

				err = obj.Save(nil)
				assert.NoError(t, err)

				return externalID, testAccount.ID(), obj, testAccount
			},
			expectErr: false, // Disabled sandbox not found returns nil
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			externalID, accountID, obj, testAccount := tt.setupFn(t)
			if obj != nil {
				defer testtools.CleanupModel(obj)
			}
			if testAccount != nil {
				defer testtools.CleanupModel(testAccount)
			}

			// Act
			found, err := sandbox.FindByExternalID(context.Background(), externalID, accountID)

			// Assert
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNil {
				assert.Empty(t, found)
			} else {
				assert.NEmpty(t, found)
				assert.Equal(t, externalID, found.ExternalID.Get())
				assert.Equal(t, accountID, found.AccountID.Get())
			}
		})
	}
}

func Test_FindAllByAccount(t *testing.T) {
	tests := []struct {
		name          string
		setupFn       func(t *testing.T) (accountID types.UUID, sandboxes []*sandbox.Sandbox, testAccount *account.Account)
		expectedCount int
		expectErr     bool
	}{
		{
			name: "returns all active sandboxes for account",
			setupFn: func(t *testing.T) (types.UUID, []*sandbox.Sandbox, *account.Account) {
				testAccount := account.TESTCreateAccount()
				err := testAccount.Save(nil)
				assert.NoError(t, err)

				sandboxes := []*sandbox.Sandbox{}

				// Create 3 active sandboxes
				for i := 0; i < 3; i++ {
					obj := sandbox.New()
					obj.AccountID.Set(testAccount.ID())
					obj.ExternalID.Set("sb-test-" + tools.RandString(8))
					obj.Type.Set(sandbox.TYPE_CLAUDE_CODE)
					obj.Status.Set(constants.STATUS_ACTIVE)
					obj.MetaData.Set(&sandbox.MetaData{})

					err := obj.Save(nil)
					assert.NoError(t, err)
					sandboxes = append(sandboxes, obj)
				}

				return testAccount.ID(), sandboxes, testAccount
			},
			expectedCount: 3,
			expectErr:     false,
		},
		{
			name: "excludes deleted sandboxes",
			setupFn: func(t *testing.T) (types.UUID, []*sandbox.Sandbox, *account.Account) {
				testAccount := account.TESTCreateAccount()
				err := testAccount.Save(nil)
				assert.NoError(t, err)

				sandboxes := []*sandbox.Sandbox{}

				// Create 2 active sandboxes
				for i := 0; i < 2; i++ {
					obj := sandbox.New()
					obj.AccountID.Set(testAccount.ID())
					obj.ExternalID.Set("sb-test-" + tools.RandString(8))
					obj.Type.Set(sandbox.TYPE_CLAUDE_CODE)
					obj.Status.Set(constants.STATUS_ACTIVE)
					obj.MetaData.Set(&sandbox.MetaData{})

					err := obj.Save(nil)
					assert.NoError(t, err)
					sandboxes = append(sandboxes, obj)
				}

				// Create 1 deleted sandbox
				deletedObj := sandbox.New()
				deletedObj.AccountID.Set(testAccount.ID())
				deletedObj.ExternalID.Set("sb-test-" + tools.RandString(8))
				deletedObj.Type.Set(sandbox.TYPE_CLAUDE_CODE)
				deletedObj.Status.Set(constants.STATUS_DELETED) // Use STATUS_DELETED
				deletedObj.MetaData.Set(&sandbox.MetaData{})

				err = deletedObj.Save(nil)
				assert.NoError(t, err)
				sandboxes = append(sandboxes, deletedObj)

				return testAccount.ID(), sandboxes, testAccount
			},
			expectedCount: 2, // Only active ones
			expectErr:     false,
		},
		{
			name: "excludes disabled sandboxes",
			setupFn: func(t *testing.T) (types.UUID, []*sandbox.Sandbox, *account.Account) {
				testAccount := account.TESTCreateAccount()
				err := testAccount.Save(nil)
				assert.NoError(t, err)

				sandboxes := []*sandbox.Sandbox{}

				// Create 1 active sandbox
				obj := sandbox.New()
				obj.AccountID.Set(testAccount.ID())
				obj.ExternalID.Set("sb-test-" + tools.RandString(8))
				obj.Type.Set(sandbox.TYPE_CLAUDE_CODE)
				obj.Status.Set(constants.STATUS_ACTIVE)
				obj.MetaData.Set(&sandbox.MetaData{})

				err = obj.Save(nil)
				assert.NoError(t, err)
				sandboxes = append(sandboxes, obj)

				// Create 1 disabled sandbox
				disabledObj := sandbox.New()
				disabledObj.AccountID.Set(testAccount.ID())
				disabledObj.ExternalID.Set("sb-test-" + tools.RandString(8))
				disabledObj.Type.Set(sandbox.TYPE_CLAUDE_CODE)
				disabledObj.Status.Set(constants.STATUS_DISABLED) // Use STATUS_DISABLED
				disabledObj.MetaData.Set(&sandbox.MetaData{})

				err = disabledObj.Save(nil)
				assert.NoError(t, err)
				sandboxes = append(sandboxes, disabledObj)

				return testAccount.ID(), sandboxes, testAccount
			},
			expectedCount: 1, // Only active one
			expectErr:     false,
		},
		{
			name: "returns empty list for account with no sandboxes",
			setupFn: func(t *testing.T) (types.UUID, []*sandbox.Sandbox, *account.Account) {
				testAccount := account.TESTCreateAccount()
				err := testAccount.Save(nil)
				assert.NoError(t, err)

				return testAccount.ID(), []*sandbox.Sandbox{}, testAccount
			},
			expectedCount: 0,
			expectErr:     false,
		},
		{
			name: "does not return sandboxes from other accounts",
			setupFn: func(t *testing.T) (types.UUID, []*sandbox.Sandbox, *account.Account) {
				targetAccount := account.TESTCreateAccount()
				err := targetAccount.Save(nil)
				assert.NoError(t, err)

				otherAccount := account.TESTCreateAccount()
				err = otherAccount.Save(nil)
				assert.NoError(t, err)

				sandboxes := []*sandbox.Sandbox{}

				// Create sandbox for target account
				obj1 := sandbox.New()
				obj1.AccountID.Set(targetAccount.ID())
				obj1.ExternalID.Set("sb-test-" + tools.RandString(8))
				obj1.Type.Set(sandbox.TYPE_CLAUDE_CODE)
				obj1.Status.Set(constants.STATUS_ACTIVE)
				obj1.MetaData.Set(&sandbox.MetaData{})

				err = obj1.Save(nil)
				assert.NoError(t, err)
				sandboxes = append(sandboxes, obj1)

				// Create sandbox for other account
				obj2 := sandbox.New()
				obj2.AccountID.Set(otherAccount.ID())
				obj2.ExternalID.Set("sb-test-" + tools.RandString(8))
				obj2.Type.Set(sandbox.TYPE_CLAUDE_CODE)
				obj2.Status.Set(constants.STATUS_ACTIVE)
				obj2.MetaData.Set(&sandbox.MetaData{})

				err = obj2.Save(nil)
				assert.NoError(t, err)
				sandboxes = append(sandboxes, obj2)

				// Clean up other account's sandbox
				defer testtools.CleanupModel(obj2)
				defer testtools.CleanupModel(otherAccount)

				return targetAccount.ID(), sandboxes, targetAccount
			},
			expectedCount: 1, // Only target account's sandbox
			expectErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			accountID, sandboxes, testAccount := tt.setupFn(t)
			for _, sb := range sandboxes {
				defer testtools.CleanupModel(sb)
			}
			if testAccount != nil {
				defer testtools.CleanupModel(testAccount)
			}

			// Act
			found, err := sandbox.FindAllByAccount(context.Background(), accountID)

			// Assert
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedCount, len(found))

			// Verify all returned sandboxes belong to the correct account
			for _, sb := range found {
				assert.Equal(t, accountID, sb.AccountID.Get())
				assert.Equal(t, 0, sb.Deleted.Get())
				assert.Equal(t, 0, sb.Disabled.Get())
			}
		})
	}
}

func Test_CountByAccount(t *testing.T) {
	tests := []struct {
		name          string
		setupFn       func(t *testing.T) (accountID types.UUID, sandboxes []*sandbox.Sandbox, testAccount *account.Account)
		expectedCount int64
		expectErr     bool
	}{
		{
			name: "counts all active sandboxes for account",
			setupFn: func(t *testing.T) (types.UUID, []*sandbox.Sandbox, *account.Account) {
				testAccount := account.TESTCreateAccount()
				err := testAccount.Save(nil)
				assert.NoError(t, err)

				sandboxes := []*sandbox.Sandbox{}

				// Create 5 active sandboxes
				for i := 0; i < 5; i++ {
					obj := sandbox.New()
					obj.AccountID.Set(testAccount.ID())
					obj.ExternalID.Set("sb-test-" + tools.RandString(8))
					obj.Type.Set(sandbox.TYPE_CLAUDE_CODE)
					obj.Status.Set(constants.STATUS_ACTIVE)
					obj.MetaData.Set(&sandbox.MetaData{})

					err := obj.Save(nil)
					assert.NoError(t, err)
					sandboxes = append(sandboxes, obj)
				}

				return testAccount.ID(), sandboxes, testAccount
			},
			expectedCount: 5,
			expectErr:     false,
		},
		{
			name: "excludes deleted and disabled sandboxes from count",
			setupFn: func(t *testing.T) (types.UUID, []*sandbox.Sandbox, *account.Account) {
				testAccount := account.TESTCreateAccount()
				err := testAccount.Save(nil)
				assert.NoError(t, err)

				sandboxes := []*sandbox.Sandbox{}

				// Create 3 active sandboxes
				for i := 0; i < 3; i++ {
					obj := sandbox.New()
					obj.AccountID.Set(testAccount.ID())
					obj.ExternalID.Set("sb-test-" + tools.RandString(8))
					obj.Type.Set(sandbox.TYPE_CLAUDE_CODE)
					obj.Status.Set(constants.STATUS_ACTIVE)
					obj.MetaData.Set(&sandbox.MetaData{})

					err := obj.Save(nil)
					assert.NoError(t, err)
					sandboxes = append(sandboxes, obj)
				}

				// Create 1 deleted sandbox
				deletedObj := sandbox.New()
				deletedObj.AccountID.Set(testAccount.ID())
				deletedObj.ExternalID.Set("sb-test-" + tools.RandString(8))
				deletedObj.Type.Set(sandbox.TYPE_CLAUDE_CODE)
				deletedObj.Status.Set(constants.STATUS_DELETED) // Use STATUS_DELETED
				deletedObj.MetaData.Set(&sandbox.MetaData{})

				err = deletedObj.Save(nil)
				assert.NoError(t, err)
				sandboxes = append(sandboxes, deletedObj)

				// Create 1 disabled sandbox
				disabledObj := sandbox.New()
				disabledObj.AccountID.Set(testAccount.ID())
				disabledObj.ExternalID.Set("sb-test-" + tools.RandString(8))
				disabledObj.Type.Set(sandbox.TYPE_CLAUDE_CODE)
				disabledObj.Status.Set(constants.STATUS_DISABLED) // Use STATUS_DISABLED
				disabledObj.MetaData.Set(&sandbox.MetaData{})

				err = disabledObj.Save(nil)
				assert.NoError(t, err)
				sandboxes = append(sandboxes, disabledObj)

				return testAccount.ID(), sandboxes, testAccount
			},
			expectedCount: 3, // Only active ones
			expectErr:     false,
		},
		{
			name: "returns zero for account with no sandboxes",
			setupFn: func(t *testing.T) (types.UUID, []*sandbox.Sandbox, *account.Account) {
				testAccount := account.TESTCreateAccount()
				err := testAccount.Save(nil)
				assert.NoError(t, err)

				return testAccount.ID(), []*sandbox.Sandbox{}, testAccount
			},
			expectedCount: 0,
			expectErr:     false,
		},
		{
			name: "does not count sandboxes from other accounts",
			setupFn: func(t *testing.T) (types.UUID, []*sandbox.Sandbox, *account.Account) {
				targetAccount := account.TESTCreateAccount()
				err := targetAccount.Save(nil)
				assert.NoError(t, err)

				otherAccount := account.TESTCreateAccount()
				err = otherAccount.Save(nil)
				assert.NoError(t, err)

				sandboxes := []*sandbox.Sandbox{}

				// Create 2 sandboxes for target account
				for i := 0; i < 2; i++ {
					obj := sandbox.New()
					obj.AccountID.Set(targetAccount.ID())
					obj.ExternalID.Set("sb-test-" + tools.RandString(8))
					obj.Type.Set(sandbox.TYPE_CLAUDE_CODE)
					obj.Status.Set(constants.STATUS_ACTIVE)
					obj.MetaData.Set(&sandbox.MetaData{})

					err := obj.Save(nil)
					assert.NoError(t, err)
					sandboxes = append(sandboxes, obj)
				}

				// Create 3 sandboxes for other account
				for i := 0; i < 3; i++ {
					obj := sandbox.New()
					obj.AccountID.Set(otherAccount.ID())
					obj.ExternalID.Set("sb-test-" + tools.RandString(8))
					obj.Type.Set(sandbox.TYPE_CLAUDE_CODE)
					obj.Status.Set(constants.STATUS_ACTIVE)
					obj.MetaData.Set(&sandbox.MetaData{})

					err := obj.Save(nil)
					assert.NoError(t, err)
					sandboxes = append(sandboxes, obj)

					// Clean up other account's sandbox
					defer testtools.CleanupModel(obj)
				}

				defer testtools.CleanupModel(otherAccount)

				return targetAccount.ID(), sandboxes, targetAccount
			},
			expectedCount: 2, // Only target account's count
			expectErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			accountID, sandboxes, testAccount := tt.setupFn(t)
			for _, sb := range sandboxes {
				defer testtools.CleanupModel(sb)
			}
			if testAccount != nil {
				defer testtools.CleanupModel(testAccount)
			}

			// Act
			count, err := sandbox.CountByAccount(context.Background(), accountID)

			// Assert
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedCount, count)
		})
	}
}
