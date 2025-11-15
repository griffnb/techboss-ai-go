package login

import (
	"github.com/CrowdShield/go-core/lib/router"
	"github.com/CrowdShield/go-core/lib/router/response"
	"github.com/go-chi/chi/v5"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
)

// Setup attaches the routes

// Setup attaches the routes
func Setup(coreRouter *router.CoreRouter) {
	coreRouter.Router.Post("/logout", logout)
	coreRouter.Router.Post("/login", login)
	coreRouter.Router.Get("/tokenLogin", tokenLogin)
	// Gets oauth profile
	coreRouter.Router.Post("/login/getProfile", response.StandardPublicRequestWrapper(getProfile))
	coreRouter.Router.Post("/login/link/send", response.StandardPublicRequestWrapper(sendMagicLink))
	coreRouter.Router.Post("/login/link", response.StandardPublicRequestWrapper(loginMagicLink))

	coreRouter.AddMainRoute("/admin", func(r chi.Router) {
		r.Group(func(authR chi.Router) {
			authR.Get("/login/super/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_READ_ADMIN: adminLogInAs,
			}))
		})
	})

	coreRouter.Router.Get("/admin/tokenLogin", adminTokenLogin)
}
