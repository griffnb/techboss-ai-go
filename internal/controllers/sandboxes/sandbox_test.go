package sandboxes

import (
	"net/http"
	"testing"

	"github.com/griffnb/core/lib/testtools"
	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/integrations/modal"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service"
	"github.com/griffnb/techboss-ai-go/internal/services/testing_service"
)

func init() {
	system_testing.BuildSystem()
}

// skipIfNotConfigured skips the test if Modal integration is not configured
func skipIfNotConfigured(t *testing.T) {
	if !modal.Configured() {
		t.Skip("Modal integration not configured")
	}
}

func Test_createSandbox_withTemplateConfig(t *testing.T) {
	skipIfNotConfigured(t)

	t.Run("successful creation with template", func(t *testing.T) {
		// Arrange
		body := map[string]any{
			"provider": sandbox.PROVIDER_CLAUDE_CODE,
			"agent_id": "",
		}

		req, err := testing_service.NewPOSTRequest[*sandbox.Sandbox]("/", nil, body)
		assert.NoError(t, err)

		err = req.WithAccount()
		assert.NoError(t, err)

		// Act
		resp, errCode, err := req.Do(createSandbox)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, resp)
		assert.NEmpty(t, resp.ID())
		assert.NEmpty(t, resp.ExternalID.Get())
		assert.Equal(t, sandbox.PROVIDER_CLAUDE_CODE, resp.Provider.Get())
		assert.Equal(t, req.Account.ID(), resp.AccountID.Get())

		metadata, err := resp.MetaData.Get()
		assert.NoError(t, err)
		assert.NEmpty(t, metadata)

		// Cleanup database record
		defer testtools.CleanupModel(resp)

		// Cleanup Modal sandbox
		defer func() {
			service := sandbox_service.NewSandboxService()
			sandboxInfo := reconstructSandboxInfo(resp, req.Account.ID())
			_ = service.TerminateSandbox(req.Request.Context(), sandboxInfo, false)
		}()
	})

	t.Run("returns error for unsupported provider", func(t *testing.T) {
		// Arrange
		body := map[string]any{
			"provider": 999, // Invalid provider
			"agent_id": "",
		}

		req, err := testing_service.NewPOSTRequest[*sandbox.Sandbox]("/", nil, body)
		assert.NoError(t, err)

		err = req.WithAccount()
		assert.NoError(t, err)

		// Act
		_, errCode, err := req.Do(createSandbox)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, errCode)
	})

	t.Run("returns error when missing provider", func(t *testing.T) {
		// Arrange
		body := map[string]any{
			"agent_id": "",
		}

		req, err := testing_service.NewPOSTRequest[*sandbox.Sandbox]("/", nil, body)
		assert.NoError(t, err)

		err = req.WithAccount()
		assert.NoError(t, err)

		// Act
		_, errCode, err := req.Do(createSandbox)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, errCode)
	})
}

func Test_createSandbox_databaseSaveFailure(t *testing.T) {
	skipIfNotConfigured(t)

	t.Run("handles database save failure gracefully", func(t *testing.T) {
		// Note: This test verifies that we properly handle the case where
		// Modal sandbox is created but DB save fails. In production, this
		// would leave an orphaned sandbox that should be cleaned up.
		// For now, we just verify the error is returned correctly.

		// Arrange
		body := map[string]any{
			"provider": sandbox.PROVIDER_CLAUDE_CODE,
			"agent_id": "",
		}

		req, err := testing_service.NewPOSTRequest[*sandbox.Sandbox]("/", nil, body)
		assert.NoError(t, err)

		// Create account but don't save it (will cause FK constraint violation)
		testAccount := testing_service.New().WithAccount()
		// Don't call SaveAll() - account won't exist in DB

		err = req.WithAccount(testAccount.Account)
		assert.NoError(t, err)

		// Act
		resp, errCode, err := req.Do(createSandbox)

		// Assert
		// Should fail during Save() due to missing account FK
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, errCode)

		// If resp is not nil, cleanup the Modal sandbox that was created
		if resp != nil && !tools.Empty(resp.ExternalID.Get()) {
			defer func() {
				service := sandbox_service.NewSandboxService()
				sandboxInfo := reconstructSandboxInfo(resp, testAccount.Account.ID())
				_ = service.TerminateSandbox(req.Request.Context(), sandboxInfo, false)
			}()
		}
	})
}

func Test_authDelete_successfulDeletion(t *testing.T) {
	skipIfNotConfigured(t)

	t.Run("successfully deletes sandbox and updates status", func(t *testing.T) {
		// Arrange - Create a sandbox first
		body := map[string]any{
			"provider": sandbox.PROVIDER_CLAUDE_CODE,
			"agent_id": "",
		}

		createReq, err := testing_service.NewPOSTRequest[*sandbox.Sandbox]("/", nil, body)
		assert.NoError(t, err)

		err = createReq.WithAccount()
		assert.NoError(t, err)

		// Act - Create sandbox
		sandboxResp, errCode, err := createReq.Do(createSandbox)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, sandboxResp)
		defer testtools.CleanupModel(sandboxResp)

		// Arrange - Delete request
		deleteReq, err := testing_service.NewDELETERequest[*sandbox.Sandbox]("/"+sandboxResp.ID().String(), nil)
		assert.NoError(t, err)

		err = deleteReq.WithAccount(createReq.Account)
		assert.NoError(t, err)

		// Act - Delete sandbox
		deleteResp, errCode, err := deleteReq.Do(authDelete)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, deleteResp)
		assert.Equal(t, 1, deleteResp.Deleted.Get())
		assert.Equal(t, constants.STATUS_DELETED, deleteResp.Status.Get())
	})

	t.Run("continues with soft delete when Modal termination fails", func(t *testing.T) {
		// Arrange - Create a sandbox
		body := map[string]any{
			"provider": sandbox.PROVIDER_CLAUDE_CODE,
			"agent_id": "",
		}

		createReq, err := testing_service.NewPOSTRequest[*sandbox.Sandbox]("/", nil, body)
		assert.NoError(t, err)

		err = createReq.WithAccount()
		assert.NoError(t, err)

		sandboxResp, errCode, err := createReq.Do(createSandbox)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		defer testtools.CleanupModel(sandboxResp)

		// Terminate the Modal sandbox directly to simulate failure
		service := sandbox_service.NewSandboxService()
		sandboxInfo := reconstructSandboxInfo(sandboxResp, createReq.Account.ID())
		_ = service.TerminateSandbox(createReq.Request.Context(), sandboxInfo, false)

		// Arrange - Delete request (Modal termination will fail since already terminated)
		deleteReq, err := testing_service.NewDELETERequest[*sandbox.Sandbox]("/"+sandboxResp.ID().String(), nil)
		assert.NoError(t, err)

		err = deleteReq.WithAccount(createReq.Account)
		assert.NoError(t, err)

		// Act - Delete sandbox (Modal termination will fail but should continue)
		deleteResp, errCode, err := deleteReq.Do(authDelete)

		// Assert - Should succeed with soft delete even though Modal termination failed
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, deleteResp)
		assert.Equal(t, 1, deleteResp.Deleted.Get())
		assert.Equal(t, constants.STATUS_DELETED, deleteResp.Status.Get())
	})
}

func Test_syncSandbox_updatesMetadata(t *testing.T) {
	skipIfNotConfigured(t)

	t.Run("successfully syncs sandbox and updates metadata", func(t *testing.T) {
		// Arrange - Create a sandbox first
		body := map[string]any{
			"provider": sandbox.PROVIDER_CLAUDE_CODE,
			"agent_id": "",
		}

		createReq, err := testing_service.NewPOSTRequest[*sandbox.Sandbox]("/", nil, body)
		assert.NoError(t, err)

		err = createReq.WithAccount()
		assert.NoError(t, err)

		// Act - Create sandbox
		sandboxResp, errCode, err := createReq.Do(createSandbox)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, sandboxResp)
		defer testtools.CleanupModel(sandboxResp)

		// Cleanup Modal sandbox at end
		defer func() {
			service := sandbox_service.NewSandboxService()
			sandboxInfo := reconstructSandboxInfo(sandboxResp, createReq.Account.ID())
			_ = service.TerminateSandbox(createReq.Request.Context(), sandboxInfo, false)
		}()

		// Arrange - Sync request with proper URL including ID
		syncReq, err := testing_service.NewPOSTRequest[*SyncSandboxResponse](
			"/"+sandboxResp.ID().String()+"/sync",
			nil,
			nil,
		)
		assert.NoError(t, err)

		err = syncReq.WithAccount(createReq.Account)
		assert.NoError(t, err)

		// Act - Sync sandbox
		syncResp, errCode, err := syncReq.Do(syncSandbox)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, syncResp)
		assert.Equal(t, sandboxResp.ID().String(), syncResp.SandboxID)

		// Verify sync stats in response
		assert.True(t, syncResp.FilesProcessed >= 0)
		assert.True(t, syncResp.BytesTransferred >= 0)
		assert.True(t, syncResp.DurationMs >= 0)

		// Verify metadata was updated in database
		updatedSandbox, err := sandbox.Get(createReq.Request.Context(), sandboxResp.ID())
		assert.NoError(t, err)
		assert.NEmpty(t, updatedSandbox)

		metadata, err := updatedSandbox.MetaData.Get()
		assert.NoError(t, err)
		assert.NEmpty(t, metadata)
		assert.NEmpty(t, metadata.LastS3Sync)
		assert.NEmpty(t, metadata.SyncStats)

		// Verify sync stats structure in metadata
		assert.True(t, metadata.SyncStats.FilesProcessed >= 0)
		assert.True(t, metadata.SyncStats.BytesTransferred >= 0)
		assert.True(t, metadata.SyncStats.DurationMs >= 0)
	})

	t.Run("returns error when sandbox not found", func(t *testing.T) {
		// Arrange - Sync request with non-existent ID
		syncReq, err := testing_service.NewPOSTRequest[*SyncSandboxResponse](
			"/00000000-0000-0000-0000-000000000000/sync",
			nil,
			nil,
		)
		assert.NoError(t, err)

		err = syncReq.WithAccount()
		assert.NoError(t, err)

		// Act - Sync sandbox
		_, errCode, err := syncReq.Do(syncSandbox)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, errCode)
	})
}
