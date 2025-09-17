package controllers

import (
	"github.com/griffnb/techboss-ai-go/internal/controllers/admins"
	"github.com/griffnb/techboss-ai-go/internal/controllers/change_logs"

	"github.com/CrowdShield/go-core/lib/router"
	"github.com/griffnb/techboss-ai-go/internal/controllers/global_configs"
	"github.com/griffnb/techboss-ai-go/internal/controllers/login"
	"github.com/griffnb/techboss-ai-go/internal/controllers/logs"
	"github.com/griffnb/techboss-ai-go/internal/controllers/utilities"
)

// Setup Adds the controllers to the router
func Setup(coreRouter *router.CoreRouter) {
	admins.Setup(coreRouter)
	change_logs.Setup(coreRouter)
	global_configs.Setup(coreRouter)
	login.Setup(coreRouter)
	logs.Setup(coreRouter)

	utilities.Setup(coreRouter)

	// Print all routes
	// printRoutes(coreRouter.Router)
}
