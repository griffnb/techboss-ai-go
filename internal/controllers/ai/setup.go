package ai

import (
	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/router"

	"github.com/griffnb/core/lib/tools"

	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
)

const (
	ROUTE string = "ai"
)

// Setup sets up the router with admin permissions
func Setup(coreRouter *router.CoreRouter) {
	coreRouter.AddMainRoute(tools.BuildString("/", ROUTE), func(r chi.Router) {
		r.Group(func(authR chi.Router) {
			authR.Post("/openai/responses", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_UNAUTHORIZED: router.NoTimeoutMiddleware(authRun),
			}))
			authR.Post("/openai/stream/responses", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_UNAUTHORIZED: router.NoTimeoutStreamingMiddleware(authStream),
			}))
		})
	})
}
