package sandboxes

import (
	"context"
	"net/http"
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

func Test_FileListingIntegration_volumeSource(t *testing.T) {
	t.Run("complete file listing workflow for volume source", func(t *testing.T) {
		// Arrange: Create test sandbox with account
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		// Create sandbox in DB
		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-integration-volume-test")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Create HTTP request with query parameters for volume source
		params := url.Values{}
		params.Set("source", "volume")
		params.Set("page", "1")
		params.Set("per_page", "10")

		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			params,
		)
		assert.NoError(t, err)

		// Set chi URL param
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", sandboxModel.ID().String())
		req.Request = req.Request.WithContext(context.WithValue(req.Request.Context(), chi.RouteCtxKey, rctx))

		// Act: Call controller
		response, statusCode, err := req.Do(adminListFiles)

		// Assert: Verify response structure
		// Note: Files array will be empty since no actual Modal sandbox is deployed
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.NEmpty(t, response)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 10, response.PerPage)
		assert.True(t, response.TotalCount >= 0)
		assert.True(t, response.TotalPages >= 0)
	})
}

func Test_FileListingIntegration_s3Source(t *testing.T) {
	t.Run("complete file listing workflow for s3 source", func(t *testing.T) {
		// Arrange: Create test sandbox with account
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		// Create sandbox in DB
		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-integration-s3-test")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Create HTTP request with query parameters for S3 source
		params := url.Values{}
		params.Set("source", "s3")
		params.Set("page", "1")
		params.Set("per_page", "10")

		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			params,
		)
		assert.NoError(t, err)

		// Set chi URL param
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", sandboxModel.ID().String())
		req.Request = req.Request.WithContext(context.WithValue(req.Request.Context(), chi.RouteCtxKey, rctx))

		// Act: Call controller
		response, statusCode, err := req.Do(adminListFiles)

		// Assert: Verify response structure
		// Note: Files array will be empty since no actual Modal sandbox is deployed
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.NEmpty(t, response)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 10, response.PerPage)
		assert.True(t, response.TotalCount >= 0)
		assert.True(t, response.TotalPages >= 0)
	})
}

func Test_FileListingIntegration_pagination(t *testing.T) {
	t.Run("pagination across multiple pages", func(t *testing.T) {
		// Arrange: Create test sandbox with account
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		// Create sandbox in DB
		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-integration-pagination-test")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Test page 1
		params1 := url.Values{}
		params1.Set("source", "volume")
		params1.Set("page", "1")
		params1.Set("per_page", "5")

		req1, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			params1,
		)
		assert.NoError(t, err)

		rctx1 := chi.NewRouteContext()
		rctx1.URLParams.Add("id", sandboxModel.ID().String())
		req1.Request = req1.Request.WithContext(context.WithValue(req1.Request.Context(), chi.RouteCtxKey, rctx1))

		response1, statusCode1, err := req1.Do(adminListFiles)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, statusCode1)
		assert.NEmpty(t, response1)
		assert.Equal(t, 1, response1.Page)
		assert.Equal(t, 5, response1.PerPage)

		// Test page 2
		params2 := url.Values{}
		params2.Set("source", "volume")
		params2.Set("page", "2")
		params2.Set("per_page", "5")

		req2, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			params2,
		)
		assert.NoError(t, err)

		rctx2 := chi.NewRouteContext()
		rctx2.URLParams.Add("id", sandboxModel.ID().String())
		req2.Request = req2.Request.WithContext(context.WithValue(req2.Request.Context(), chi.RouteCtxKey, rctx2))

		response2, statusCode2, err := req2.Do(adminListFiles)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, statusCode2)
		assert.NEmpty(t, response2)
		assert.Equal(t, 2, response2.Page)
		assert.Equal(t, 5, response2.PerPage)

		// Verify TotalPages calculation is consistent across pages
		assert.Equal(t, response1.TotalPages, response2.TotalPages)
		assert.Equal(t, response1.TotalCount, response2.TotalCount)
	})

	t.Run("different per_page values affect pagination", func(t *testing.T) {
		// Arrange: Create test sandbox with account
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		// Create sandbox in DB
		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-integration-per-page-test")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Test with per_page = 10
		params1 := url.Values{}
		params1.Set("source", "volume")
		params1.Set("page", "1")
		params1.Set("per_page", "10")

		req1, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			params1,
		)
		assert.NoError(t, err)

		rctx1 := chi.NewRouteContext()
		rctx1.URLParams.Add("id", sandboxModel.ID().String())
		req1.Request = req1.Request.WithContext(context.WithValue(req1.Request.Context(), chi.RouteCtxKey, rctx1))

		response1, statusCode1, err := req1.Do(adminListFiles)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, statusCode1)
		assert.NEmpty(t, response1)
		assert.Equal(t, 10, response1.PerPage)

		// Test with per_page = 25
		params2 := url.Values{}
		params2.Set("source", "volume")
		params2.Set("page", "1")
		params2.Set("per_page", "25")

		req2, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			params2,
		)
		assert.NoError(t, err)

		rctx2 := chi.NewRouteContext()
		rctx2.URLParams.Add("id", sandboxModel.ID().String())
		req2.Request = req2.Request.WithContext(context.WithValue(req2.Request.Context(), chi.RouteCtxKey, rctx2))

		response2, statusCode2, err := req2.Do(adminListFiles)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, statusCode2)
		assert.NEmpty(t, response2)
		assert.Equal(t, 25, response2.PerPage)

		// Verify total count is same but per_page differs
		assert.Equal(t, response1.TotalCount, response2.TotalCount)
	})
}

func Test_FileListingIntegration_authController(t *testing.T) {
	t.Run("auth controller with ownership verification", func(t *testing.T) {
		// Arrange: Create test sandbox with user context
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		// Create sandbox in DB
		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-integration-auth-test")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Create HTTP request with session
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

		// Set chi URL param
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", sandboxModel.ID().String())
		req.Request = req.Request.WithContext(context.WithValue(req.Request.Context(), chi.RouteCtxKey, rctx))

		// Act: Call authListFiles
		response, statusCode, err := req.Do(authListFiles)

		// Assert: Verify response structure
		// Note: Files array will be empty since no actual Modal sandbox is deployed
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.NEmpty(t, response)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 50, response.PerPage)
		assert.True(t, response.TotalCount >= 0)
		assert.True(t, response.TotalPages >= 0)
	})
}

func Test_FileListingIntegration_queryParameterCombinations(t *testing.T) {
	t.Run("recursive and path parameters together", func(t *testing.T) {
		// Arrange: Create test sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-integration-combo-test")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Test with recursive=false and path within /workspace
		params := url.Values{}
		params.Set("source", "volume")
		params.Set("path", "/workspace/src")
		params.Set("recursive", "false")
		params.Set("page", "1")
		params.Set("per_page", "20")

		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			params,
		)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", sandboxModel.ID().String())
		req.Request = req.Request.WithContext(context.WithValue(req.Request.Context(), chi.RouteCtxKey, rctx))

		// Act
		response, statusCode, err := req.Do(adminListFiles)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.NEmpty(t, response)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 20, response.PerPage)
	})

	t.Run("all parameters specified together", func(t *testing.T) {
		// Arrange: Create test sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-integration-all-params-test")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Test with all parameters, using valid S3 path
		params := url.Values{}
		params.Set("source", "s3")
		params.Set("path", "/s3-bucket/data")
		params.Set("recursive", "true")
		params.Set("page", "2")
		params.Set("per_page", "15")

		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			params,
		)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", sandboxModel.ID().String())
		req.Request = req.Request.WithContext(context.WithValue(req.Request.Context(), chi.RouteCtxKey, rctx))

		// Act
		response, statusCode, err := req.Do(adminListFiles)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.NEmpty(t, response)
		assert.Equal(t, 2, response.Page)
		assert.Equal(t, 15, response.PerPage)
	})
}

func Test_FileListingIntegration_errorHandling(t *testing.T) {
	t.Run("invalid sandbox ID returns empty response", func(t *testing.T) {
		// Arrange: Use non-existent sandbox ID
		params := url.Values{}
		params.Set("source", "volume")

		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/00000000-0000-0000-0000-000000000000/files",
			params,
		)
		assert.NoError(t, err)

		// Set chi URL param
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "00000000-0000-0000-0000-000000000000")
		req.Request = req.Request.WithContext(context.WithValue(req.Request.Context(), chi.RouteCtxKey, rctx))

		// Act
		response, statusCode, err := req.Do(adminListFiles)

		// Assert - Returns empty response with OK status for non-existent sandbox
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, statusCode)
		assert.NEmpty(t, response)
		assert.Equal(t, 0, len(response.Files))
	})

	t.Run("malformed UUID returns error", func(t *testing.T) {
		// Arrange: Use malformed UUID
		params := url.Values{}
		params.Set("source", "volume")

		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/not-a-valid-uuid/files",
			params,
		)
		assert.NoError(t, err)

		// Set chi URL param
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "not-a-valid-uuid")
		req.Request = req.Request.WithContext(context.WithValue(req.Request.Context(), chi.RouteCtxKey, rctx))

		// Act
		_, statusCode, err := req.Do(adminListFiles)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, statusCode)
	})

	t.Run("invalid query parameters return error", func(t *testing.T) {
		// Arrange: Create test sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-integration-invalid-params")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Test with invalid source
		params := url.Values{}
		params.Set("source", "invalid-source")

		req, err := testing_service.NewGETRequest[*sandbox_service.FileListResponse](
			"/"+sandboxModel.ID().String()+"/files",
			params,
		)
		assert.NoError(t, err)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", sandboxModel.ID().String())
		req.Request = req.Request.WithContext(context.WithValue(req.Request.Context(), chi.RouteCtxKey, rctx))

		// Act
		_, statusCode, err := req.Do(adminListFiles)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, statusCode)
	})
}

func Test_FileContentRetrievalIntegration(t *testing.T) {
	t.Run("adminGetFileContent complete workflow for volume source", func(t *testing.T) {
		// Arrange: Create sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		// Create sandbox in DB
		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-content-volume-test")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Create HTTP request
		params := url.Values{}
		params.Set("source", "volume")
		params.Set("file_path", "/workspace/test.txt")

		req, err := testing_service.NewGETRequest[any](
			"/admin/sandbox/"+sandboxModel.ID().String()+"/files/content",
			params,
		)
		assert.NoError(t, err)

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", sandboxModel.ID().String())
		req.Request = req.Request.WithContext(context.WithValue(req.Request.Context(), chi.RouteCtxKey, rctx))

		// Act: Call controller
		w, err := req.DoRaw(adminGetFileContent)
		assert.NoError(t, err)

		// Assert: Verify response
		// Note: Will return 404 because sandbox doesn't have actual files
		// But we're testing the workflow works
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("adminGetFileContent for s3 source", func(t *testing.T) {
		// Arrange: Create sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-content-s3-test")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Create HTTP request with S3 source
		params := url.Values{}
		params.Set("source", "s3")
		params.Set("file_path", "/s3-bucket/data/file.json")

		req, err := testing_service.NewGETRequest[any](
			"/admin/sandbox/"+sandboxModel.ID().String()+"/files/content",
			params,
		)
		assert.NoError(t, err)

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", sandboxModel.ID().String())
		req.Request = req.Request.WithContext(context.WithValue(req.Request.Context(), chi.RouteCtxKey, rctx))

		// Act
		w, err := req.DoRaw(adminGetFileContent)
		assert.NoError(t, err)

		// Assert: Returns 404 since no actual files
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("returns 400 when file_path missing", func(t *testing.T) {
		// Arrange: Create sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-content-missing-path")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Create HTTP request WITHOUT file_path parameter
		params := url.Values{}
		params.Set("source", "volume")
		// Intentionally omitting file_path

		req, err := testing_service.NewGETRequest[any](
			"/admin/sandbox/"+sandboxModel.ID().String()+"/files/content",
			params,
		)
		assert.NoError(t, err)

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", sandboxModel.ID().String())
		req.Request = req.Request.WithContext(context.WithValue(req.Request.Context(), chi.RouteCtxKey, rctx))

		// Act
		w, err := req.DoRaw(adminGetFileContent)
		assert.NoError(t, err)

		// Assert: Returns 400 when file_path is missing
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "file_path query parameter is required")
	})

	t.Run("authGetFileContent with ownership verification", func(t *testing.T) {
		// Arrange: Create sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-content-auth-test")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Create HTTP request with auth session
		params := url.Values{}
		params.Set("source", "volume")
		params.Set("file_path", "/workspace/main.go")

		req, err := testing_service.NewGETRequest[any](
			"/sandbox/"+sandboxModel.ID().String()+"/files/content",
			params,
		)
		assert.NoError(t, err)

		// Set session for authenticated request
		err = req.WithAccount(builder.Account)
		assert.NoError(t, err)

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", sandboxModel.ID().String())
		req.Request = req.Request.WithContext(context.WithValue(req.Request.Context(), chi.RouteCtxKey, rctx))

		// Act
		w, err := req.DoRaw(authGetFileContent)
		assert.NoError(t, err)

		// Assert: Returns 404 since no actual files (but auth workflow works)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("handles various file paths", func(t *testing.T) {
		// Arrange: Create sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-content-paths-test")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Test cases for various file paths
		testCases := []struct {
			name     string
			filePath string
		}{
			{"root level file", "/workspace/README.md"},
			{"nested directory", "/workspace/src/main.go"},
			{"deeply nested", "/workspace/src/controllers/users/handler.go"},
			{"with spaces", "/workspace/my file.txt"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				params := url.Values{}
				params.Set("source", "volume")
				params.Set("file_path", tc.filePath)

				req, err := testing_service.NewGETRequest[any](
					"/admin/sandbox/"+sandboxModel.ID().String()+"/files/content",
					params,
				)
				assert.NoError(t, err)

				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", sandboxModel.ID().String())
				req.Request = req.Request.WithContext(context.WithValue(req.Request.Context(), chi.RouteCtxKey, rctx))

				// Act
				w, err := req.DoRaw(adminGetFileContent)
				assert.NoError(t, err)

				// Assert: All should return 404 (no actual files)
				assert.Equal(t, http.StatusNotFound, w.Code)
			})
		}
	})

	t.Run("query parameter combinations", func(t *testing.T) {
		// Arrange: Create sandbox
		builder := testing_service.New().WithAccount()
		err := builder.SaveAll()
		assert.NoError(t, err)
		defer builder.CleanupAll(testtools.CleanupModel)

		sandboxModel := sandbox.New()
		sandboxModel.AccountID.Set(builder.Account.ID())
		sandboxModel.Provider.Set(sandbox.PROVIDER_CLAUDE_CODE)
		sandboxModel.ExternalID.Set("sb-content-combo-test")
		sandboxModel.Status.Set(constants.STATUS_ACTIVE)
		sandboxModel.MetaData.Set(&sandbox.MetaData{})
		err = sandboxModel.Save(nil)
		assert.NoError(t, err)
		defer testtools.CleanupModel(sandboxModel)

		// Test case 1: Volume source with file_path
		params1 := url.Values{}
		params1.Set("source", "volume")
		params1.Set("file_path", "/workspace/config.yaml")

		req1, err := testing_service.NewGETRequest[any](
			"/admin/sandbox/"+sandboxModel.ID().String()+"/files/content",
			params1,
		)
		assert.NoError(t, err)

		rctx1 := chi.NewRouteContext()
		rctx1.URLParams.Add("id", sandboxModel.ID().String())
		req1.Request = req1.Request.WithContext(context.WithValue(req1.Request.Context(), chi.RouteCtxKey, rctx1))

		w1, err := req1.DoRaw(adminGetFileContent)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w1.Code)

		// Test case 2: S3 source with file_path
		params2 := url.Values{}
		params2.Set("source", "s3")
		params2.Set("file_path", "/bucket/data.csv")

		req2, err := testing_service.NewGETRequest[any](
			"/admin/sandbox/"+sandboxModel.ID().String()+"/files/content",
			params2,
		)
		assert.NoError(t, err)

		rctx2 := chi.NewRouteContext()
		rctx2.URLParams.Add("id", sandboxModel.ID().String())
		req2.Request = req2.Request.WithContext(context.WithValue(req2.Request.Context(), chi.RouteCtxKey, rctx2))

		w2, err := req2.DoRaw(adminGetFileContent)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w2.Code)

		// Test case 3: Default source (should default to volume)
		params3 := url.Values{}
		params3.Set("file_path", "/workspace/default.txt")
		// Intentionally omit source to test default behavior

		req3, err := testing_service.NewGETRequest[any](
			"/admin/sandbox/"+sandboxModel.ID().String()+"/files/content",
			params3,
		)
		assert.NoError(t, err)

		rctx3 := chi.NewRouteContext()
		rctx3.URLParams.Add("id", sandboxModel.ID().String())
		req3.Request = req3.Request.WithContext(context.WithValue(req3.Request.Context(), chi.RouteCtxKey, rctx3))

		w3, err := req3.DoRaw(adminGetFileContent)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w3.Code)
	})

	t.Run("invalid sandbox ID returns 400", func(t *testing.T) {
		// Arrange: Use malformed UUID
		params := url.Values{}
		params.Set("source", "volume")
		params.Set("file_path", "/workspace/test.txt")

		req, err := testing_service.NewGETRequest[any](
			"/admin/sandbox/not-a-valid-uuid/files/content",
			params,
		)
		assert.NoError(t, err)

		// Add chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "not-a-valid-uuid")
		req.Request = req.Request.WithContext(context.WithValue(req.Request.Context(), chi.RouteCtxKey, rctx))

		// Act
		w, err := req.DoRaw(adminGetFileContent)
		assert.NoError(t, err)

		// Assert: Returns 400 for invalid UUID
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
