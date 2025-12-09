package billing

import (
	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/router"
	"github.com/griffnb/core/lib/router/response"
	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
)

const ROUTE string = "billing"

// Setup sets up the router
func Setup(coreRouter *router.CoreRouter) {
	coreRouter.AddMainRoute(tools.BuildString("/", ROUTE), func(r chi.Router) {
		r.Group(func(authR chi.Router) {
			authR.Post("/checkout/plan/{id}/stripe", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ORG_ADMIN: response.StandardPublicRequestWrapper(authStripeCheckout),
			}))

			authR.Post("/checkout/success", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ORG_ADMIN: response.StandardPublicRequestWrapper(authStripeCheckoutSuccess),
			}))

			authR.Post("/cancel", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ORG_ADMIN: response.StandardPublicRequestWrapper(authCancel),
			}))

			authR.Post("/resume", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ORG_ADMIN: response.StandardPublicRequestWrapper(authResume),
			}))

			//authR.Get("/portal", helpers.RoleHandler(helpers.RoleHandlerMap{
			//	constants.ROLE_ORG_ADMIN: response.StandardRequestWrapper(authPortal),
			//}))
		})

		r.Group(func(openR chi.Router) {
			openR.Post("/stripe/webhook", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_UNAUTHORIZED: openStripeHook,
			}))
		})
	})
}
