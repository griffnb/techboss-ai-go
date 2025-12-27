package sandbox

import (
	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/router"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
)

const ROUTE string = "sandbox"

// Setup configures sandbox routes with role-based access control.
// All endpoints require any authorized user (ROLE_ANY_AUTHORIZED).
// The Claude streaming endpoint uses NoTimeoutStreamingMiddleware for long-running operations.
func Setup(coreRouter *router.CoreRouter) {
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

// SetupTestRoutes configures routes for testing without CoreRouter.
// This helper function allows tests to create a chi.Router directly.
func SetupTestRoutes(r chi.Router) {
	r.Post("/sandbox", helpers.RoleHandler(helpers.RoleHandlerMap{
		constants.ROLE_ANY_AUTHORIZED: response.StandardRequestWrapper(createSandbox),
	}))

	r.Get("/sandbox/{sandboxID}", helpers.RoleHandler(helpers.RoleHandlerMap{
		constants.ROLE_ANY_AUTHORIZED: response.StandardRequestWrapper(getSandbox),
	}))

	r.Delete("/sandbox/{sandboxID}", helpers.RoleHandler(helpers.RoleHandlerMap{
		constants.ROLE_ANY_AUTHORIZED: response.StandardRequestWrapper(deleteSandbox),
	}))

	r.Post("/sandbox/{sandboxID}/claude", helpers.RoleHandler(helpers.RoleHandlerMap{
		constants.ROLE_ANY_AUTHORIZED: router.NoTimeoutStreamingMiddleware(streamClaude),
	}))
}
