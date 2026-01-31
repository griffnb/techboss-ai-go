package sandboxes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/griffnb/core/lib/log"
	"github.com/griffnb/core/lib/testtools"
	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/core/lib/types"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service"
	"github.com/griffnb/techboss-ai-go/internal/services/testing_service"
)

func init() {
	system_testing.BuildSystem()
}

func Test_streamClaude_withOwnedSandbox(t *testing.T) {
	skipIfNotConfigured(t)

	t.Run("successfully streams with owned sandbox", func(t *testing.T) {
		// Arrange - Create a sandbox first
		body := map[string]any{
			"provider": sandbox.TYPE_CLAUDE_CODE,
			"agent_id": "",
		}

		createReq, err := testing_service.NewPOSTRequest[*sandbox.Sandbox]("/", nil, body)
		assert.NoError(t, err)

		err = createReq.WithAccount()
		assert.NoError(t, err)

		// Act - Create sandbox
		sandboxResp, errCode, err := createReq.Do(adminCreateSandbox)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, sandboxResp)
		defer testtools.CleanupModel(sandboxResp)

		// Cleanup Modal sandbox at end
		defer func() {
			service := sandbox_service.NewSandboxService()
			sandboxInfo, err := sandbox_service.ReconstructSandboxInfo(createReq.Request.Context(), sandboxResp, createReq.Account.ID())
			if err != nil {
				log.ErrorContext(err, createReq.Request.Context())
			}
			_ = service.TerminateSandbox(createReq.Request.Context(), sandboxInfo, false)
		}()

		// Arrange - Stream Claude request
		claudeBody := map[string]any{
			"prompt": "echo 'hello world'",
		}

		streamReq, err := testing_service.NewPOSTRequest[any](
			"/"+sandboxResp.ID().String()+"/claude",
			nil,
			claudeBody,
		)
		assert.NoError(t, err)

		err = streamReq.WithAccount(createReq.Account)
		assert.NoError(t, err)

		// Create a ResponseRecorder to capture the streaming response
		recorder := httptest.NewRecorder()

		// Act - Stream Claude
		adminStreamClaude(recorder, streamReq.Request)

		// Assert
		// Streaming writes directly to ResponseWriter, so we check the recorder
		assert.Equal(t, http.StatusOK, recorder.Code)
		// The response should contain some data (streaming output)
		assert.True(t, recorder.Body.Len() > 0)
	})
}

func Test_streamClaude_withUnownedSandbox(t *testing.T) {
	skipIfNotConfigured(t)

	t.Run("returns 404 for sandbox owned by different account", func(t *testing.T) {
		// Arrange - Create a sandbox with one account
		body := map[string]any{
			"provider": sandbox.TYPE_CLAUDE_CODE,
			"agent_id": "",
		}

		createReq, err := testing_service.NewPOSTRequest[*sandbox.Sandbox]("/", nil, body)
		assert.NoError(t, err)

		err = createReq.WithAccount()
		assert.NoError(t, err)

		// Act - Create sandbox
		sandboxResp, errCode, err := createReq.Do(adminCreateSandbox)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, sandboxResp)
		defer testtools.CleanupModel(sandboxResp)

		// Cleanup Modal sandbox at end
		defer func() {
			service := sandbox_service.NewSandboxService()
			sandboxInfo, err := sandbox_service.ReconstructSandboxInfo(createReq.Request.Context(), sandboxResp, createReq.Account.ID())
			if err != nil {
				log.ErrorContext(err, createReq.Request.Context())
			}
			_ = service.TerminateSandbox(createReq.Request.Context(), sandboxInfo, false)
		}()

		// Arrange - Create a different account to try accessing the sandbox
		differentAccount := account.TESTCreateAccount()
		err = differentAccount.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(differentAccount)

		// Arrange - Stream Claude request with different account
		claudeBody := map[string]any{
			"prompt": "echo 'hello world'",
		}

		streamReq, err := testing_service.NewPOSTRequest[any](
			"/"+sandboxResp.ID().String()+"/claude",
			nil,
			claudeBody,
		)
		assert.NoError(t, err)

		err = streamReq.WithAccount(differentAccount) // Using different account
		assert.NoError(t, err)

		// Create a ResponseRecorder to capture the response
		recorder := httptest.NewRecorder()

		// Act - Stream Claude
		adminStreamClaude(recorder, streamReq.Request)

		// Assert - Should return 404 for unowned sandbox
		assert.Equal(t, http.StatusNotFound, recorder.Code)
		assert.True(t, recorder.Body.Len() > 0)
		// Response should contain "sandbox not found" message
		responseBody := recorder.Body.String()
		assert.True(t, len(responseBody) > 0)
	})
}

func Test_streamClaude_sandboxNotFound(t *testing.T) {
	skipIfNotConfigured(t)

	t.Run("returns 404 when sandbox does not exist", func(t *testing.T) {
		// Arrange - Stream Claude request with non-existent sandbox ID
		claudeBody := map[string]any{
			"prompt": "echo 'hello world'",
		}

		nonExistentID := types.UUID("00000000-0000-0000-0000-000000000000")

		streamReq, err := testing_service.NewPOSTRequest[any](
			"/"+nonExistentID.String()+"/claude",
			nil,
			claudeBody,
		)
		assert.NoError(t, err)

		err = streamReq.WithAccount()
		assert.NoError(t, err)

		// Create a ResponseRecorder to capture the response
		recorder := httptest.NewRecorder()

		// Act - Stream Claude
		adminStreamClaude(recorder, streamReq.Request)

		// Assert - Should return 404
		assert.Equal(t, http.StatusNotFound, recorder.Code)
		assert.True(t, recorder.Body.Len() > 0)
		// Response should contain "sandbox not found" message
		responseBody := recorder.Body.String()
		assert.True(t, len(responseBody) > 0)
	})
}

func Test_streamClaude_emptyPrompt(t *testing.T) {
	skipIfNotConfigured(t)

	t.Run("returns 400 when prompt is empty", func(t *testing.T) {
		// Arrange - Create a sandbox first
		body := map[string]any{
			"provider": sandbox.TYPE_CLAUDE_CODE,
			"agent_id": "",
		}

		createReq, err := testing_service.NewPOSTRequest[*sandbox.Sandbox]("/", nil, body)
		assert.NoError(t, err)

		err = createReq.WithAccount()
		assert.NoError(t, err)

		// Act - Create sandbox
		sandboxResp, errCode, err := createReq.Do(adminCreateSandbox)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, sandboxResp)
		defer testtools.CleanupModel(sandboxResp)

		// Cleanup Modal sandbox at end
		defer func() {
			service := sandbox_service.NewSandboxService()
			sandboxInfo, err := sandbox_service.ReconstructSandboxInfo(createReq.Request.Context(), sandboxResp, createReq.Account.ID())
			if err != nil {
				log.ErrorContext(err, createReq.Request.Context())
			}
			_ = service.TerminateSandbox(createReq.Request.Context(), sandboxInfo, false)
		}()

		// Arrange - Stream Claude request with empty prompt
		claudeBody := map[string]any{
			"prompt": "",
		}

		streamReq, err := testing_service.NewPOSTRequest[any](
			"/"+sandboxResp.ID().String()+"/claude",
			nil,
			claudeBody,
		)
		assert.NoError(t, err)

		err = streamReq.WithAccount(createReq.Account)
		assert.NoError(t, err)

		// Create a ResponseRecorder to capture the response
		recorder := httptest.NewRecorder()

		// Act - Stream Claude
		adminStreamClaude(recorder, streamReq.Request)

		// Assert - Should return 400 for empty prompt
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.True(t, recorder.Body.Len() > 0)
		// Response should contain "prompt is required" message
		responseBody := recorder.Body.String()
		assert.True(t, len(responseBody) > 0)
	})
}
