//go:generate core_gen controller Sandbox -modelPackage=sandbox -skip=authCreate,authUpdate
package sandboxes

import (
	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/router"

	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/models/sandbox"
)

const (
	TABLE_NAME string = sandbox.TABLE
	ROUTE      string = "sandbox"
)

// Setup sets up the router
func Setup(coreRouter *router.CoreRouter) {
	coreRouter.AddMainRoute(tools.BuildString("/admin/", ROUTE), func(r chi.Router) {
		r.Group(func(adminR chi.Router) {
			adminR.Get("/", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_READ_ADMIN: response.StandardRequestWrapper(adminIndex),
			}))
			adminR.Get("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_READ_ADMIN: response.StandardRequestWrapper(adminGet),
			}))
			adminR.Get("/count", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_READ_ADMIN: response.StandardRequestWrapper(adminCount),
			}))
			// GET /admin/sandbox/{id}/files - List files in sandbox volume or S3
			adminR.Get("/{id}/files", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_READ_ADMIN: response.StandardRequestWrapper(adminListFiles),
			}))
			// GET /admin/sandbox/{id}/files/content - Get file content from sandbox
			adminR.Get("/{id}/files/content", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_READ_ADMIN: adminGetFileContent,
			}))
			// GET /admin/sandbox/{id}/files/tree - Get hierarchical tree structure of files
			adminR.Get("/{id}/files/tree", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_READ_ADMIN: response.StandardRequestWrapper(adminGetFileTree),
			}))
		})
		r.Group(func(adminR chi.Router) {
			adminR.Post("/", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: response.StandardPublicRequestWrapper(adminCreateSandbox),
			}))
			adminR.Put("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: response.StandardRequestWrapper(adminUpdate),
			}))

			// POST /sandbox/{sandboxID}/claude - Execute Claude with streaming (Task 11)
			adminR.Post("/{id}/claude", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_READ_ADMIN: router.NoTimeoutStreamingMiddleware(adminStreamClaude),
			}))
		})
		r.Group(func(adminR chi.Router) {
			adminR.Get("/_ts", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_READ_ADMIN: helpers.TSValidation(TABLE_NAME),
			}))
		})
	})

	// Public authenticated routes
	coreRouter.AddMainRoute(tools.BuildString("/", ROUTE), func(r chi.Router) {
		r.Group(func(authR chi.Router) {
			authR.Get("/", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: response.StandardPublicRequestWrapper(authIndex),
			}))
			authR.Get("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: response.StandardPublicRequestWrapper(authGet),
			}))
			// GET /sandbox/{id}/files - List files in sandbox volume or S3
			authR.Get("/{id}/files", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: response.StandardPublicRequestWrapper(authListFiles),
			}))
			// GET /sandbox/{id}/files/content - Get file content from sandbox
			authR.Get("/{id}/files/content", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: authGetFileContent,
			}))
			// GET /sandbox/{id}/files/tree - Get hierarchical tree structure of files
			authR.Get("/{id}/files/tree", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: response.StandardPublicRequestWrapper(authGetFileTree),
			}))
		})

		r.Group(func(authR chi.Router) {
			//authR.Post("/", helpers.RoleHandler(helpers.RoleHandlerMap{
			//	constants.ROLE_ANY_AUTHORIZED: response.StandardPublicRequestWrapper(createSandbox),
			//}))
			authR.Put("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: response.StandardPublicRequestWrapper(authUpdate),
			}))

			authR.Delete("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: response.StandardPublicRequestWrapper(authDelete),
			}))
			// POST /sandbox/{id}/sync - Sync sandbox volume to S3
			authR.Post("/{id}/sync", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: response.StandardRequestWrapper(syncSandbox),
			}))
			// POST /sandbox/{sandboxID}/claude - Execute Claude with streaming (Task 11)
			authR.Post("/{id}/claude", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_READ_ADMIN: router.NoTimeoutStreamingMiddleware(adminStreamClaude),
			}))
		})
	})
}
