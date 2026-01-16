//go:generate core_gen controller GithubInstallation -modelPackage=github_installation -options=admin
package github_installations

import (
	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/router"

	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/models/github_installation"
)

const (
	TABLE_NAME string = github_installation.TABLE
	ROUTE      string = "github_installation"
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
		r.Group(func(openR chi.Router) {
			openR.Post("/callback", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_UNAUTHORIZED: WebhookCallback,
			}))
		})
	})
}
