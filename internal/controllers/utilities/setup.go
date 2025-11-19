package utilities

import (
	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/router"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"

	"github.com/griffnb/core/lib/tools"
)

const ROUTE string = "utilities"

// Setup sets up the router with admin permissions
func Setup(coreRouter *router.CoreRouter) {
	coreRouter.AddMainRoute(tools.BuildString("/admin/", ROUTE), func(r chi.Router) {
		r.Group(func(authR chi.Router) {
			authR.Get("/test_error", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: response.StandardRequestWrapper(testError),
			}))

			authR.Get("/uploadURL", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: response.StandardRequestWrapper(uploadURL),
			}))
		})
		r.Group(func(open chi.Router) {
			open.Get("/hook_log", hookLog)
			open.Post("/hook_log", hookLog)
		})
	})
}
