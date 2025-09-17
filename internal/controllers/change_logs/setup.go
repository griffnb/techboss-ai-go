package change_logs

import (
	"github.com/CrowdShield/go-core/lib/router"
	"github.com/go-chi/chi/v5"

	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/models/change_log"
)

const (
	TABLE_NAME string = change_log.TABLE
	ROUTE      string = "change_log"
)

// Setup sets up the router
func Setup(coreRouter *router.CoreRouter) {
	coreRouter.AddMainRoute(tools.BuildString("/admin/", ROUTE), func(r chi.Router) {
		r.Group(func(authR chi.Router) {
			authR.Get("/", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: helpers.StandardRequestWrapper(adminIndex),
			}))

			authR.Get("/count", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: helpers.StandardRequestWrapper(adminCount),
			}))
		})

		r.Group(func(adminR chi.Router) {
			adminR.Get("/_ts", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: helpers.TSValidation(TABLE_NAME),
			}))
		})
	})
}
