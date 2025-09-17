package global_configs

import (
	"github.com/CrowdShield/go-core/lib/router"
	"github.com/go-chi/chi/v5"

	"github.com/CrowdShield/go-core/lib/tools"

	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	genericmodel "github.com/griffnb/techboss-ai-go/internal/models/global_config"
)

const (
	TABLE_NAME string = genericmodel.TABLE
	ROUTE      string = "global_config"
)

// Setup sets up the router with admin permissions
func Setup(coreRouter *router.CoreRouter) {
	coreRouter.AddMainRoute(tools.BuildString("/admin/", ROUTE), func(r chi.Router) {
		r.Group(func(authR chi.Router) {
			authR.Get("/", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: helpers.StandardRequestWrapper(adminIndex),
			}))
			authR.Get("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: helpers.StandardRequestWrapper(adminGet),
			}))

			authR.Get("/count", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: helpers.StandardRequestWrapper(adminCount),
			}))
		})
		r.Group(func(authR chi.Router) {
			authR.Post("/", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: helpers.StandardRequestWrapper(adminCreate),
			}))
			authR.Put("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: helpers.StandardRequestWrapper(adminUpdate),
			}))
		})
		r.Group(func(adminR chi.Router) {
			adminR.Get("/_ts", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: helpers.TSValidation(TABLE_NAME),
			}))
		})
	})
}
