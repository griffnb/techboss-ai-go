package login

import (
	"github.com/CrowdShield/go-core/lib/router"
)

// Setup attaches the routes
func Setup(coreRouter *router.CoreRouter) {
	coreRouter.Router.Get("/admin/tokenLogin", adminTokenLogin)
	coreRouter.Router.Post("/logout", logout)
}
