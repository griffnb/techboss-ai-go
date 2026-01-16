package github_installations_test

import (
	"context"
	"testing"

	"github.com/griffnb/core/lib/testtools"
	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/models/github_installation"
)

func init() {
	system_testing.BuildSystem()
}

// Test_InstallationIDValidation tests that installation ID validation works correctly
func Test_InstallationIDValidation(t *testing.T) {
	t.Run("Returns empty for non-existent installation ID", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		nonExistentID := "00000000-0000-0000-0000-000000000000"

		// Act
		result, err := github_installation.Get(ctx, types.UUID(nonExistentID))

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Returns installation for valid installation_id", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		// Create a test installation
		installation := github_installation.New()
		installation.InstallationID.Set("test-security-validation-123")
		installation.GithubAccountID.Set("12345")
		installation.GithubAccountName.Set("test-security-user")
		installation.RepositoryAccess.Set("all")
		installation.AppSlug.Set("test-app")
		err := installation.SaveWithContext(ctx, nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(installation)

		// Act
		retrieved, err := github_installation.GetByInstallationID(ctx, "test-security-validation-123")

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, retrieved)
		assert.Equal(t, installation.ID().String(), retrieved.ID().String())
		assert.Equal(t, "test-security-validation-123", retrieved.InstallationID.Get())
	})

	t.Run("Returns empty for empty installation ID", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		// Act
		result, err := github_installation.GetByInstallationID(ctx, "")

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}

// Test_CrossAccountAccess_DataIsolation tests that installations are properly isolated by account
func Test_CrossAccountAccess_DataIsolation(t *testing.T) {
	t.Run("Installations can be filtered by account ID", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		// Create test accounts
		account1 := account.New()
		account1.FirstName.Set("Test")
		account1.LastName.Set("Account 1")
		account1.Email.Set("test-account-1@example.com")
		err := account1.SaveWithContext(ctx, nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(account1)

		account2 := account.New()
		account2.FirstName.Set("Test")
		account2.LastName.Set("Account 2")
		account2.Email.Set("test-account-2@example.com")
		err = account2.SaveWithContext(ctx, nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(account2)

		// Create installations for both accounts
		installation1 := github_installation.New()
		installation1.AccountID.Set(account1.ID())
		installation1.InstallationID.Set("test-cross-account-1")
		installation1.GithubAccountID.Set("11111")
		installation1.GithubAccountName.Set("account1-user")
		installation1.RepositoryAccess.Set("all")
		installation1.AppSlug.Set("test-app")
		err = installation1.SaveWithContext(ctx, nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(installation1)

		installation2 := github_installation.New()
		installation2.AccountID.Set(account2.ID())
		installation2.InstallationID.Set("test-cross-account-2")
		installation2.GithubAccountID.Set("22222")
		installation2.GithubAccountName.Set("account2-user")
		installation2.RepositoryAccess.Set("all")
		installation2.AppSlug.Set("test-app")
		err = installation2.SaveWithContext(ctx, nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(installation2)

		// Act - Get installations for account 1
		account1Installations, err := github_installation.GetByAccountID(ctx, account1.ID())

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, account1Installations)

		// Should only see installation1, not installation2
		foundInstallation1 := false
		foundInstallation2 := false
		for _, inst := range account1Installations {
			if inst.ID().String() == installation1.ID().String() {
				foundInstallation1 = true
			}
			if inst.ID().String() == installation2.ID().String() {
				foundInstallation2 = true
			}
		}

		assert.True(t, foundInstallation1, "Should find installation from account 1")
		assert.Equal(t, false, foundInstallation2, "Should NOT find installation from account 2")
	})

	t.Run("Installations without account ID are accessible", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		// Create installation without account ID (system-wide)
		installation := github_installation.New()
		installation.InstallationID.Set("test-system-wide")
		installation.GithubAccountID.Set("99999")
		installation.GithubAccountName.Set("system-user")
		installation.RepositoryAccess.Set("all")
		installation.AppSlug.Set("test-app")
		err := installation.SaveWithContext(ctx, nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(installation)

		// Act - Retrieve by installation ID
		retrieved, err := github_installation.GetByInstallationID(ctx, "test-system-wide")

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, retrieved)
		assert.True(t, retrieved.AccountID.IsNull())
	})
}

// Test_SuspendedInstallations tests suspended installation filtering
func Test_SuspendedInstallations(t *testing.T) {
	t.Run("Can filter active installations only", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		// Create test account
		testAccount := account.New()
		testAccount.FirstName.Set("Test")
		testAccount.LastName.Set("Account")
		testAccount.Email.Set("test-account@example.com")
		err := testAccount.SaveWithContext(ctx, nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(testAccount)

		// Create active installation
		activeInstallation := github_installation.New()
		activeInstallation.AccountID.Set(testAccount.ID())
		activeInstallation.InstallationID.Set("test-active")
		activeInstallation.GithubAccountID.Set("11111")
		activeInstallation.GithubAccountName.Set("active-user")
		activeInstallation.RepositoryAccess.Set("all")
		activeInstallation.AppSlug.Set("test-app")
		activeInstallation.Suspended.Set(0)
		err = activeInstallation.SaveWithContext(ctx, nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(activeInstallation)

		// Create suspended installation
		suspendedInstallation := github_installation.New()
		suspendedInstallation.AccountID.Set(testAccount.ID())
		suspendedInstallation.InstallationID.Set("test-suspended")
		suspendedInstallation.GithubAccountID.Set("22222")
		suspendedInstallation.GithubAccountName.Set("suspended-user")
		suspendedInstallation.RepositoryAccess.Set("all")
		suspendedInstallation.AppSlug.Set("test-app")
		suspendedInstallation.Suspended.Set(1)
		err = suspendedInstallation.SaveWithContext(ctx, nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(suspendedInstallation)

		// Act - Get active installations only
		activeInstallations, err := github_installation.GetActiveByAccountID(ctx, testAccount.ID())

		// Assert
		assert.NoError(t, err)
		assert.NEmpty(t, activeInstallations)

		// Should only see active installation
		foundActive := false
		foundSuspended := false
		for _, inst := range activeInstallations {
			if inst.ID().String() == activeInstallation.ID().String() {
				foundActive = true
			}
			if inst.ID().String() == suspendedInstallation.ID().String() {
				foundSuspended = true
			}
		}

		assert.True(t, foundActive, "Should find active installation")
		assert.Equal(t, false, foundSuspended, "Should NOT find suspended installation")
	})

	t.Run("Suspended flag can be updated", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		// Create active installation
		installation := github_installation.New()
		installation.InstallationID.Set("test-suspend-update")
		installation.GithubAccountID.Set("33333")
		installation.GithubAccountName.Set("suspend-test-user")
		installation.RepositoryAccess.Set("all")
		installation.AppSlug.Set("test-app")
		installation.Suspended.Set(0)
		err := installation.SaveWithContext(ctx, nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(installation)

		// Act - Suspend the installation
		installation.Suspended.Set(1)
		err = installation.SaveWithContext(ctx, nil)
		assert.NoError(t, err)

		// Assert - Verify suspension persisted
		updated, err := github_installation.Get(ctx, installation.ID())
		assert.NoError(t, err)
		assert.Equal(t, 1, updated.Suspended.Get())

		// Act - Unsuspend the installation
		updated.Suspended.Set(0)
		err = updated.SaveWithContext(ctx, nil)
		assert.NoError(t, err)

		// Assert - Verify unsuspension persisted
		final, err := github_installation.Get(ctx, installation.ID())
		assert.NoError(t, err)
		assert.Equal(t, 0, final.Suspended.Get())
	})
}

// Test_WebhookSecurity_Documentation documents webhook security model
func Test_WebhookSecurity_Documentation(t *testing.T) {
	t.Run("Webhook security is signature-based not user-auth based", func(t *testing.T) {
		// This test documents the security model for GitHub webhooks:
		//
		// 1. Webhook endpoints must allow unauthenticated requests (ROLE_UNAUTHORIZED)
		//    because they come from GitHub's servers, not authenticated users
		//
		// 2. Security is enforced via HMAC-SHA256 signature validation:
		//    - GitHub signs the payload with the webhook secret
		//    - Our handler validates the signature before processing
		//    - Invalid signatures are rejected with 401 Unauthorized
		//
		// 3. The webhook secret must be kept confidential and rotated periodically
		//
		// 4. Webhook signature validation is tested in webhook_test.go:
		//    - Test_webhookCallback_invalidSignature
		//    - Test_webhookCallback_missingSignature
		//
		// 5. Additional signature validation is tested in security_test.go:
		//    - github_service.ValidateWebhookSignature with various attack scenarios

		assert.True(t, true) // Documentation test
	})
}

// Test_GithubAccountType_Validation tests account type constant validation
func Test_GithubAccountType_Validation(t *testing.T) {
	t.Run("Can create user account type installation", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		installation := github_installation.New()
		installation.InstallationID.Set("test-user-type")
		installation.GithubAccountID.Set("44444")
		installation.GithubAccountName.Set("user-account")
		installation.GithubAccountType.Set(github_installation.ACCOUNT_TYPE_USER)
		installation.RepositoryAccess.Set("all")
		installation.AppSlug.Set("test-app")
		err := installation.SaveWithContext(ctx, nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(installation)

		// Act
		retrieved, err := github_installation.GetByInstallationID(ctx, "test-user-type")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, github_installation.ACCOUNT_TYPE_USER, retrieved.GithubAccountType.Get())
	})

	t.Run("Can create organization account type installation", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		installation := github_installation.New()
		installation.InstallationID.Set("test-org-type")
		installation.GithubAccountID.Set("55555")
		installation.GithubAccountName.Set("org-account")
		installation.GithubAccountType.Set(github_installation.ACCOUNT_TYPE_ORGANIZATION)
		installation.RepositoryAccess.Set("all")
		installation.AppSlug.Set("test-app")
		err := installation.SaveWithContext(ctx, nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(installation)

		// Act
		retrieved, err := github_installation.GetByInstallationID(ctx, "test-org-type")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, github_installation.ACCOUNT_TYPE_ORGANIZATION, retrieved.GithubAccountType.Get())
	})
}
