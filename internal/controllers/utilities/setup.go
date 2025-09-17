package utilities

import (
	"github.com/CrowdShield/go-core/lib/router"
	"github.com/go-chi/chi/v5"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"

	"github.com/CrowdShield/go-core/lib/tools"
)

const ROUTE string = "utilities"

// Setup sets up the router with admin permissions
func Setup(coreRouter *router.CoreRouter) {
	coreRouter.AddMainRoute(tools.BuildString("/admin/", ROUTE), func(r chi.Router) {
		r.Group(func(authR chi.Router) {
			authR.Get("/test_error", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: helpers.StandardRequestWrapper(testError),
			}))

			authR.Get("/uploadURL", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: helpers.StandardRequestWrapper(uploadURL),
			}))
		})
		r.Group(func(open chi.Router) {
			open.Get("/hook_log", hookLog)
			open.Post("/hook_log", hookLog)
		})
	})
}
