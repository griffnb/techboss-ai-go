//go:generate core_gen controller Account -modelPackage=account
package accounts

import (
	"github.com/go-chi/chi/v5"
	"github.com/griffnb/core/lib/router"
	"github.com/griffnb/core/lib/router/response"

	"github.com/griffnb/core/lib/tools"
	"github.com/griffnb/techboss-ai-go/internal/constants"
	"github.com/griffnb/techboss-ai-go/internal/controllers/helpers"
	"github.com/griffnb/techboss-ai-go/internal/models/account"
)

const (
	TABLE_NAME string = account.TABLE
	ROUTE      string = "account"
)

// Setup sets up the router
func Setup(coreRouter *router.CoreRouter) {
	coreRouter.AddMainRoute(tools.BuildString("/api/", ROUTE), func(r chi.Router) {
		r.Group(func(apiR chi.Router) {
			apiR.Get("/{id}", helpers.ApiAuthRequestWrapper(response.StandardRequestWrapper(internalAPIAccount)))
		})
	})

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
			adminR.Post("/testUser", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ADMIN: response.StandardRequestWrapper(adminTestCreate),
			}))
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
				constants.ROLE_ANY_AUTHORIZED: response.StandardPublicRequestWrapper(authIndex),
			}))
			authR.Get("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: response.StandardPublicRequestWrapper(authGet),
			}))

			authR.Get("/me", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: response.StandardPublicRequestWrapper(authMe),
			}))
		})
		r.Group(func(authR chi.Router) {
			authR.Post("/", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ORG_ADMIN: response.StandardPublicRequestWrapper(authCreate),
			}))
			authR.Put("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ORG_ADMIN: response.StandardPublicRequestWrapper(authUpdate),
			}))

			authR.Post("/updatePassword", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: response.StandardPublicRequestWrapper(updatePassword),
			}))

			authR.Post("/setPassword", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: response.StandardPublicRequestWrapper(setPassword),
			}))

			authR.Delete("/{id}", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ORG_ADMIN: response.StandardPublicRequestWrapper(authDelete),
			}))

			authR.Post("/{id}/invite", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ORG_ADMIN: response.StandardPublicRequestWrapper(authResendInvite),
			}))
			authR.Post("/{id}/invite/cancel", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ORG_ADMIN: response.StandardPublicRequestWrapper(authCancelInvite),
			}))
			authR.Put("/resendVerifyEmail", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: response.StandardPublicRequestWrapper(authResendVerifyEmail),
			}))
			authR.Put("/updatePrimaryEmail", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_ANY_AUTHORIZED: response.StandardPublicRequestWrapper(updatePrimaryEmailAddress),
			}))
		})

		r.Group(func(openR chi.Router) {
			openR.Post("/signup", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_UNAUTHORIZED: response.StandardPublicRequestWrapper(openSignup),
			}))
			openR.Post("/check", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_UNAUTHORIZED: response.StandardPublicRequestWrapper(openCheckExisting),
			}))
			openR.Post("/signup/oauth", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_UNAUTHORIZED: response.StandardPublicRequestWrapper(oauthSignup),
			}))

			openR.Post("/verify/invite", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_UNAUTHORIZED: response.StandardPublicRequestWrapper(openVerifyInvite),
			}))
			openR.Post("/verify/email", helpers.RoleHandler(helpers.RoleHandlerMap{
				constants.ROLE_UNAUTHORIZED: response.StandardPublicRequestWrapper(openVerifyEmail),
			}))

			openR.Post(
				"/sendResetPassword",
				response.StandardPublicRequestWrapper(openSendResetPasswordEmail),
			)
			openR.Post("/resetPassword", response.StandardPublicRequestWrapper(openResetPassword))
		})
	})
}
