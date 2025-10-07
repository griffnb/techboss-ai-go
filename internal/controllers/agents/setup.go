//go:generate core_generate controller Agent -modelPackage=agent -skip=authCreate,authUpdate
package agents

import (
	"github.com/CrowdShield/go-core/lib/router"
	"github.com/go-chi/chi/v5"

	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/models/agent"
)

const (
	TABLE_NAME string = agent.TABLE
	ROUTE      string = "agent"
)

// Setup sets up the router
func Setup(coreRouter *router.CoreRouter) {
	// Admin routes
	coreRouter.AddMainRoute(tools.BuildString("/admin/", ROUTE), func(r chi.Router) {
		r.Group(func(adminR chi.Router) {
			adminR.Get("/", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_READ_ADMIN: helpers.StandardRequestWrapper(adminIndex),
			}))
			adminR.Get("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_READ_ADMIN: helpers.StandardRequestWrapper(adminGet),
			}))
			adminR.Get("/count", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_READ_ADMIN: helpers.StandardRequestWrapper(adminCount),
			}))
		})
		r.Group(func(adminR chi.Router) {
			adminR.Post("/", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: helpers.StandardRequestWrapper(adminCreate),
			}))
			adminR.Put("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: helpers.StandardRequestWrapper(adminUpdate),
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
				constants.ROLE_ANY_AUTHORIZED: helpers.StandardRequestWrapper(authIndex),
			}))
			authR.Get("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: helpers.StandardRequestWrapper(authGet),
			}))

			authR.Get("/session", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: helpers.StandardRequestWrapper(authSession),
			}))
		})
	})
}
