//go:generate core_generate controller Document -modelPackage=document
package documents

import (
	"github.com/CrowdShield/go-core/lib/router"
	"github.com/CrowdShield/go-core/lib/router/response"
	"github.com/go-chi/chi/v5"

	"github.com/CrowdShield/go-core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/models/document"
)

const (
	TABLE_NAME string = document.TABLE
	ROUTE      string = "document"
)

// Setup sets up the router
func Setup(coreRouter *router.CoreRouter) {
	// Admin routes
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

	// Public authenticated routes
	coreRouter.AddMainRoute(tools.BuildString("/", ROUTE), func(r chi.Router) {
		r.Group(func(authR chi.Router) {
			authR.Get("/", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: response.StandardRequestWrapper(authIndex),
			}))
			authR.Get("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: response.StandardRequestWrapper(authGet),
			}))
		})
	})
}
