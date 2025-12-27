//go:generate core_gen controller Sandbox -modelPackage=sandbox -options=admin
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
		})
		r.Group(func(adminR chi.Router) {
			adminR.Post("/", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: response.StandardRequestWrapper(adminCreate),
			}))
			adminR.Put("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: response.StandardRequestWrapper(adminUpdate),
			}))
		})
		r.Group(func(adminR chi.Router) {
			adminR.Get("/_ts", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_READ_ADMIN: helpers.TSValidation(TABLE_NAME),
			}))
		})
	})

	coreRouter.AddMainRoute(tools.BuildString("/", ROUTE), func(r chi.Router) {
		r.Group(func(authR chi.Router) {
			// POST /sandbox - Create new sandbox
			authR.Post("/", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: response.StandardRequestWrapper(createSandbox),
			}))

			// GET /sandbox/{sandboxID} - Get sandbox status (stub for Phase 2)
			authR.Get("/{sandboxID}", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: response.StandardRequestWrapper(getSandbox),
			}))

			// DELETE /sandbox/{sandboxID} - Terminate sandbox (stub for Phase 2)
			authR.Delete("/{sandboxID}", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: response.StandardRequestWrapper(deleteSandbox),
			}))

			// POST /sandbox/{sandboxID}/sync - Sync sandbox volume to S3
			authR.Post("/{sandboxID}/sync", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: response.StandardRequestWrapper(syncSandbox),
			}))

			// POST /sandbox/{sandboxID}/claude - Execute Claude with streaming (Task 11)
			authR.Post("/{sandboxID}/claude", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: router.NoTimeoutStreamingMiddleware(streamClaude),
			}))
		})
	})
}
