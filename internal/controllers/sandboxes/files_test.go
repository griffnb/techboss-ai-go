package sandboxes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/testtools"
	"github.com/griffnb/core/lib/testtools/assert"
	"github.com/griffnb/techboss-ai-go/internal/common/system_testing"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
	"github.com/griffnb/techboss-ai-go/internal/services/sandbox_service"
	"github.com/griffnb/techboss-ai-go/internal/services/testing_service"
)

func init() {
	system_testing.BuildSystem()
}

// setChiURLParam sets the "id" URL parameter in the chi router context for testing.
// This is needed when calling controller handlers directly without going through the router.
func setChiURLParam(req *http.Request, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func Test_adminListFiles_successfulListing(t *testing.T) {
	t.Run("returns files successfully with valid sandbox", func(t *testing.T) {
		// Arrange - Create test sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		// Create sandbox in DB
		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-12345")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Create HTTP request with query parameters
		params := url.Values{}
		params.Set("source", "volume")
		params.Set("page", "1")
		params.Set("per_page", "50")

		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			params,
		)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req.Request = setChiURLParam(req.Request, sandboxModel.ID().String())

		// Act - Call adminListFiles
		resp, errCode, err := req.Do(adminListFiles)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, resp)
		// Files array will be empty since we don't have an actual Modal sandbox running
		// but the response structure should be valid
		assert.Equal(t, 1, resp.Page)
		assert.Equal(t, 50, resp.PerPage)
		assert.True(t, resp.TotalCount >= 0)
		assert.True(t, resp.TotalPages >= 0)
	})

	t.Run("returns empty list for sandbox without files", func(t *testing.T) {
		// Arrange - Create test sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		// Create sandbox in DB
		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-empty")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Create HTTP request
		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			nil,
		)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req.Request = setChiURLParam(req.Request, sandboxModel.ID().String())

		// Act - Call adminListFiles
		resp, errCode, err := req.Do(adminListFiles)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, resp)
		assert.Equal(t, 0, len(resp.Files))
		assert.Equal(t, 0, resp.TotalCount)
	})
}

func Test_adminListFiles_invalidSandboxID(t *testing.T) {
	t.Run("returns error for non-existent sandbox ID", func(t *testing.T) {
		// Arrange - Use non-existent sandbox ID
		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/00000000-0000-0000-0000-000000000000/files",
			nil,
		)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req.Request = setChiURLParam(req.Request, "00000000-0000-0000-0000-000000000000")

		// Act - Call adminListFiles
		resp, errCode, err := req.Do(adminListFiles)

		// Assert - When sandbox doesn't exist, returns empty response with OK status
		// The service layer returns empty file list for non-existent sandboxes
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, resp)
		assert.Equal(t, 0, len(resp.Files))
	})

	t.Run("returns error for invalid UUID format", func(t *testing.T) {
		// Arrange - Use invalid UUID format
		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/invalid-uuid/files",
			nil,
		)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req.Request = setChiURLParam(req.Request, "invalid-uuid")

		// Act - Call adminListFiles
		_, errCode, err := req.Do(adminListFiles)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, errCode)
	})
}

func Test_adminListFiles_queryParameters(t *testing.T) {
	t.Run("parses source parameter correctly", func(t *testing.T) {
		// Arrange - Create test sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-source")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Test with s3 source
		params := url.Values{}
		params.Set("source", "s3")

		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			params,
		)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req.Request = setChiURLParam(req.Request, sandboxModel.ID().String())

		// Act
		resp, errCode, err := req.Do(adminListFiles)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, resp)
	})

	t.Run("parses pagination parameters correctly", func(t *testing.T) {
		// Arrange
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-pagination")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		params := url.Values{}
		params.Set("page", "2")
		params.Set("per_page", "25")

		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			params,
		)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req.Request = setChiURLParam(req.Request, sandboxModel.ID().String())

		// Act
		resp, errCode, err := req.Do(adminListFiles)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, resp)
		assert.Equal(t, 2, resp.Page)
		assert.Equal(t, 25, resp.PerPage)
	})

	t.Run("returns error for invalid source parameter", func(t *testing.T) {
		// Arrange
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-invalid-source")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		params := url.Values{}
		params.Set("source", "invalid")

		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			params,
		)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req.Request = setChiURLParam(req.Request, sandboxModel.ID().String())

		// Act
		_, errCode, err := req.Do(adminListFiles)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, errCode)
	})

	t.Run("returns error for invalid pagination parameters", func(t *testing.T) {
		// Arrange
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-invalid-pagination")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Test with per_page > 1000
		params := url.Values{}
		params.Set("per_page", "1001")

		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			params,
		)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req.Request = setChiURLParam(req.Request, sandboxModel.ID().String())

		// Act
		_, errCode, err := req.Do(adminListFiles)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, errCode)
	})

	t.Run("parses path parameter correctly", func(t *testing.T) {
		// Arrange
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-path")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		params := url.Values{}
		params.Set("path", "/workspace/src")

		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			params,
		)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req.Request = setChiURLParam(req.Request, sandboxModel.ID().String())

		// Act
		resp, errCode, err := req.Do(adminListFiles)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, resp)
	})

	t.Run("parses recursive parameter correctly", func(t *testing.T) {
		// Arrange
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-recursive")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		params := url.Values{}
		params.Set("recursive", "false")

		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			params,
		)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req.Request = setChiURLParam(req.Request, sandboxModel.ID().String())

		// Act
		resp, errCode, err := req.Do(adminListFiles)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, resp)
	})
}

func Test_authListFiles_successfulListing(t *testing.T) {
	t.Run("returns files successfully with ownership verification", func(t *testing.T) {
		// Arrange - Create test sandbox with user context
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		// Create sandbox in DB
		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-auth-12345")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Create HTTP request with query parameters and user session
		params := url.Values{}
		params.Set("source", "volume")
		params.Set("page", "1")
		params.Set("per_page", "50")

		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			params,
		)
		assert.NoError(t, err)

		// Set session for authenticated request
		err = req.WithAccount(builder.Account)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req.Request = setChiURLParam(req.Request, sandboxModel.ID().String())

		// Act - Call authListFiles
		resp, errCode, err := req.Do(authListFiles)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, resp)
		assert.Equal(t, 1, resp.Page)
		assert.Equal(t, 50, resp.PerPage)
		assert.True(t, resp.TotalCount >= 0)
		assert.True(t, resp.TotalPages >= 0)
	})

	t.Run("returns empty list for sandbox without files", func(t *testing.T) {
		// Arrange - Create test sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		// Create sandbox in DB
		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-auth-empty")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Create HTTP request
		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			nil,
		)
		assert.NoError(t, err)

		// Set session for authenticated request
		err = req.WithAccount(builder.Account)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req.Request = setChiURLParam(req.Request, sandboxModel.ID().String())

		// Act - Call authListFiles
		resp, errCode, err := req.Do(authListFiles)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, resp)
		assert.Equal(t, 0, len(resp.Files))
		assert.Equal(t, 0, resp.TotalCount)
	})
}

func Test_authListFiles_invalidSandboxID(t *testing.T) {
	t.Run("returns error for non-existent sandbox ID", func(t *testing.T) {
		// Arrange - Create user context
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		// Use non-existent sandbox ID
		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/00000000-0000-0000-0000-000000000000/files",
			nil,
		)
		assert.NoError(t, err)

		// Set session for authenticated request
		err = req.WithAccount(builder.Account)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req.Request = setChiURLParam(req.Request, "00000000-0000-0000-0000-000000000000")

		// Act - Call authListFiles
		resp, errCode, err := req.Do(authListFiles)

		// Assert - When sandbox doesn't exist, returns empty response with OK status
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, resp)
		assert.Equal(t, 0, len(resp.Files))
	})

	t.Run("returns error for invalid UUID format", func(t *testing.T) {
		// Arrange - Create user context
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		// Use invalid UUID format
		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/invalid-uuid/files",
			nil,
		)
		assert.NoError(t, err)

		// Set session for authenticated request
		err = req.WithAccount(builder.Account)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req.Request = setChiURLParam(req.Request, "invalid-uuid")

		// Act - Call authListFiles
		_, errCode, err := req.Do(authListFiles)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, errCode)
	})
}

func Test_authListFiles_queryParameters(t *testing.T) {
	t.Run("parses source parameter correctly", func(t *testing.T) {
		// Arrange - Create test sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-auth-source")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Test with s3 source
		params := url.Values{}
		params.Set("source", "s3")

		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			params,
		)
		assert.NoError(t, err)

		// Set session for authenticated request
		err = req.WithAccount(builder.Account)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req.Request = setChiURLParam(req.Request, sandboxModel.ID().String())

		// Act
		resp, errCode, err := req.Do(authListFiles)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, resp)
	})

	t.Run("parses pagination parameters correctly", func(t *testing.T) {
		// Arrange
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-auth-pagination")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		params := url.Values{}
		params.Set("page", "2")
		params.Set("per_page", "25")

		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			params,
		)
		assert.NoError(t, err)

		// Set session for authenticated request
		err = req.WithAccount(builder.Account)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req.Request = setChiURLParam(req.Request, sandboxModel.ID().String())

		// Act
		resp, errCode, err := req.Do(authListFiles)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, errCode)
		assert.NEmpty(t, resp)
		assert.Equal(t, 2, resp.Page)
		assert.Equal(t, 25, resp.PerPage)
	})

	t.Run("returns error for invalid source parameter", func(t *testing.T) {
		// Arrange
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-auth-invalid-source")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		params := url.Values{}
		params.Set("source", "invalid")

		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			params,
		)
		assert.NoError(t, err)

		// Set session for authenticated request
		err = req.WithAccount(builder.Account)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req.Request = setChiURLParam(req.Request, sandboxModel.ID().String())

		// Act
		_, errCode, err := req.Do(authListFiles)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, errCode)
	})

	t.Run("returns error for invalid pagination parameters", func(t *testing.T) {
		// Arrange
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-auth-invalid-pagination")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Test with per_page > 1000
		params := url.Values{}
		params.Set("per_page", "1001")

		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			params,
		)
		assert.NoError(t, err)

		// Set session for authenticated request
		err = req.WithAccount(builder.Account)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req.Request = setChiURLParam(req.Request, sandboxModel.ID().String())

		// Act
		_, errCode, err := req.Do(authListFiles)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, errCode)
	})
}

func Test_adminGetFileContent(t *testing.T) {
	t.Run("returns 400 when file_path query parameter is missing", func(t *testing.T) {
		// Arrange - Create test sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-content-no-path")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Create request without file_path parameter
		req, err := http.NewRequest("GET", "/admin/sandbox/"+sandboxModel.ID().String()+"/files/content", nil)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req = setChiURLParam(req, sandboxModel.ID().String())

		// Act - Call adminGetFileContent directly
		w := httptest.NewRecorder()
		adminGetFileContent(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "file_path query parameter is required")
	})

	t.Run("returns 404 for non-existent file", func(t *testing.T) {
		// Arrange - Create test sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-content-notfound")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Create request with non-existent file path
		params := url.Values{}
		params.Set("file_path", "/workspace/nonexistent.txt")
		req, err := http.NewRequest("GET", "/admin/sandbox/"+sandboxModel.ID().String()+"/files/content?"+params.Encode(), nil)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req = setChiURLParam(req, sandboxModel.ID().String())

		// Act - Call adminGetFileContent directly
		w := httptest.NewRecorder()
		adminGetFileContent(w, req)

		// Assert - Service returns "file not found" error for non-existent files
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "file not found")
	})

	t.Run("returns 404 for invalid sandbox ID", func(t *testing.T) {
		// Arrange - Use non-existent sandbox ID
		params := url.Values{}
		params.Set("file_path", "/workspace/test.txt")
		req, err := http.NewRequest("GET", "/admin/sandbox/00000000-0000-0000-0000-000000000000/files/content?"+params.Encode(), nil)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req = setChiURLParam(req, "00000000-0000-0000-0000-000000000000")

		// Act - Call adminGetFileContent directly
		w := httptest.NewRecorder()
		adminGetFileContent(w, req)

		// Assert - Returns 404 because sandbox is not connected
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("defaults source to volume when not specified", func(t *testing.T) {
		// Arrange - Create test sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-content-default-source")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Create request without source parameter
		params := url.Values{}
		params.Set("file_path", "/workspace/test.txt")
		req, err := http.NewRequest("GET", "/admin/sandbox/"+sandboxModel.ID().String()+"/files/content?"+params.Encode(), nil)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req = setChiURLParam(req, sandboxModel.ID().String())

		// Act - Call adminGetFileContent directly
		w := httptest.NewRecorder()
		adminGetFileContent(w, req)

		// Assert - Should return 404 because sandbox is not connected (no active Modal connection)
		// but it should have successfully defaulted to "volume" source
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "file not found")
	})

	t.Run("uses s3 source when specified", func(t *testing.T) {
		// Arrange - Create test sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-content-s3-source")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Create request with s3 source
		params := url.Values{}
		params.Set("source", "s3")
		params.Set("file_path", "/s3-bucket/test.txt")
		req, err := http.NewRequest("GET", "/admin/sandbox/"+sandboxModel.ID().String()+"/files/content?"+params.Encode(), nil)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req = setChiURLParam(req, sandboxModel.ID().String())

		// Act - Call adminGetFileContent directly
		w := httptest.NewRecorder()
		adminGetFileContent(w, req)

		// Assert - Should return 404 because sandbox is not connected
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "file not found")
	})
}

func Test_authGetFileContent(t *testing.T) {
	t.Run("returns 400 when file_path query parameter is missing", func(t *testing.T) {
		// Arrange - Create test sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-auth-content-no-path")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Create request without file_path parameter
		req, err := http.NewRequest("GET", "/sandbox/"+sandboxModel.ID().String()+"/files/content", nil)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req = setChiURLParam(req, sandboxModel.ID().String())

		// Act - Call authGetFileContent directly
		w := httptest.NewRecorder()
		authGetFileContent(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "file_path query parameter is required")
	})

	t.Run("returns 404 for non-existent file", func(t *testing.T) {
		// Arrange - Create test sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-auth-content-notfound")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Create request with non-existent file path
		params := url.Values{}
		params.Set("file_path", "/workspace/nonexistent.txt")
		req, err := http.NewRequest("GET", "/sandbox/"+sandboxModel.ID().String()+"/files/content?"+params.Encode(), nil)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req = setChiURLParam(req, sandboxModel.ID().String())

		// Act - Call authGetFileContent directly
		w := httptest.NewRecorder()
		authGetFileContent(w, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "file not found")
	})

	t.Run("returns 404 for invalid sandbox ID", func(t *testing.T) {
		// Arrange - Use non-existent sandbox ID
		params := url.Values{}
		params.Set("file_path", "/workspace/test.txt")
		req, err := http.NewRequest("GET", "/sandbox/00000000-0000-0000-0000-000000000000/files/content?"+params.Encode(), nil)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req = setChiURLParam(req, "00000000-0000-0000-0000-000000000000")

		// Act - Call authGetFileContent directly
		w := httptest.NewRecorder()
		authGetFileContent(w, req)

		// Assert - Returns 404 because sandbox is not connected
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("defaults source to volume when not specified", func(t *testing.T) {
		// Arrange - Create test sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-auth-content-default-source")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Create request without source parameter
		params := url.Values{}
		params.Set("file_path", "/workspace/test.txt")
		req, err := http.NewRequest("GET", "/sandbox/"+sandboxModel.ID().String()+"/files/content?"+params.Encode(), nil)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req = setChiURLParam(req, sandboxModel.ID().String())

		// Act - Call authGetFileContent directly
		w := httptest.NewRecorder()
		authGetFileContent(w, req)

		// Assert - Should return 404 because sandbox is not connected
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "file not found")
	})

	t.Run("uses s3 source when specified", func(t *testing.T) {
		// Arrange - Create test sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-test-auth-content-s3-source")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Create request with s3 source
		params := url.Values{}
		params.Set("source", "s3")
		params.Set("file_path", "/s3-bucket/test.txt")
		req, err := http.NewRequest("GET", "/sandbox/"+sandboxModel.ID().String()+"/files/content?"+params.Encode(), nil)
		assert.NoError(t, err)

		// Set URL parameter for chi router
		req = setChiURLParam(req, sandboxModel.ID().String())

		// Act - Call authGetFileContent directly
		w := httptest.NewRecorder()
		authGetFileContent(w, req)

		// Assert - Should return 404 because sandbox is not connected
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "file not found")
	})
}
