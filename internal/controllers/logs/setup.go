package logs

import (
	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/router"
	"github.com/griffnb/core/lib/router/response"

	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
)

const ROUTE string = "logs"

// Setup sets up the router with admin permissions
func Setup(coreRouter *router.CoreRouter) {
	coreRouter.AddMainRoute(tools.BuildString("/admin/", ROUTE), func(r chi.Router) {
		r.Group(func(authR chi.Router) {
			authR.Post("/search/{logGroup}", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: response.StandardRequestWrapper(search),
			}))
			authR.Post("/searchAll/{logGroup}", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: response.StandardRequestWrapper(searchRecursive),
			}))
		})
	})
}
